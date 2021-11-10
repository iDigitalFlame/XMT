//go:build windows
// +build windows

package cmd

import (
	"os"
	"strconv"
	"strings"
	"unsafe"

	"github.com/iDigitalFlame/xmt/device/devtools"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
	"golang.org/x/sys/windows"
)

// Select will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process ID will be returned.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the Filter
// are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices  if a PID is not set.
func (f Filter) Select() (uint32, error) {
	return f.SelectFunc(nil)
}
func inStrList(s string, l []string) bool {
	for i := range l {
		if strings.EqualFold(s, l[i]) {
			return true
		}
	}
	return false
}
func isTokenElevated(t windows.Token) bool {
	if !t.IsElevated() {
		return false
	}
	// TODO(dij): Should add a component of using token strealing here.
	switch u, err := t.GetTokenUser(); {
	case err != nil:
		fallthrough
	case u.User.Sid.IsWellKnown(windows.WinLocalServiceSid):
		fallthrough
	case u.User.Sid.IsWellKnown(windows.WinNetworkServiceSid):
		return false
	}
	return true
}

// Handle will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process Handle will be returned.
// The first argument is the access rights requested, expressed as a uint32.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the Filter
// are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices.
func (f Filter) Handle(a uint32) (uintptr, error) {
	return f.HandleFunc(a, nil)
}

// SelectFunc will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process ID will be returned.
//
// This function allows for a filtering function to be passed along that will be
// supplied with the ProcessID, if the process is elevated, the process name
// and process handle. The function supplied should return true if the process
// passes the filter. The function argument may be nil.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the Filter
// are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices  if a PID is not set.
func (f Filter) SelectFunc(x filter) (uint32, error) {
	// Process ID values in Windows technically begin at 5, [0-4] are System, we can't really use them or
	// get handles to them.
	if f.PID > 4 && x == nil {
		// No need to check info (we have no extra filter)
		return f.PID, nil
	}
	h, err := f.get(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ, x)
	if err != nil {
		return 0, err
	}
	p, err := windows.GetProcessId(h)
	if windows.CloseHandle(h); err != nil {
		return 0, err
	}
	return p, nil
}

// SelectFunc will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process Handle will be returned.
//
// This function allows for a filtering function to be passed along that will be
// supplied with the ProcessID, if the process is elevated, the process name
// and process handle. The function supplied should return true if the process
// passes the filter. The function argument may be nil.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the Filter
// are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices.
func (f Filter) HandleFunc(a uint32, x filter) (uintptr, error) {
	h, err := f.get(a, x)
	if err != nil {
		return 0, err
	}
	return uintptr(h), nil
}
func (f Filter) get(a uint32, x filter) (windows.Handle, error) {
	// If we have a specific PID in mind that's valid.
	if f.PID > 4 {
		h, err := windows.OpenProcess(a, true, f.PID)
		if h == 0 || err != nil {
			return 0, xerr.Wrap("OpenProcess "+strconv.Itoa(int(f.PID)), err)
		}
		if x == nil {
			// Quick path since we don't have a filter func.
			return h, nil
		}
		var n [256]uint16
		if err = windows.GetModuleBaseName(h, 0, &n[0], 256); err != nil {
			windows.CloseHandle(h)
			return 0, xerr.Wrap("GetModuleBaseName", err)
		}
		var t windows.Token
		if err = windows.OpenProcessToken(h, windows.TOKEN_QUERY, &t); err != nil {
			windows.CloseHandle(h)
			return 0, xerr.Wrap("OpenProcessToken", err)
		}
		r := x(f.PID, isTokenElevated(t), windows.UTF16ToString(n[:]), uintptr(h))
		if t.Close(); r {
			return h, nil
		}
		windows.CloseHandle(h)
		return 0, ErrNoProcessFound
	}
	return f.open(a, false, x)
}
func (f Filter) open(a uint32, r bool, x filter) (windows.Handle, error) {
	h, err := windows.CreateToolhelp32Snapshot(0x0002, 0)
	if err != nil {
		return 0, xerr.Wrap("winapi CreateToolhelp32Snapshot error", err)
	}
	devtools.AdjustPrivileges("SeDebugPrivilege")
	var (
		e    windows.ProcessEntry32
		l    []windows.Handle
		t    windows.Token
		s    string
		o    windows.Handle
		d, z uint32
		j    bool
		p    = uint32(os.Getpid())
	)
	e.Size = uint32(unsafe.Sizeof(e))
	for err = windows.Process32First(h, &e); err == nil; err = windows.Process32Next(h, &e) {
		if e.ProcessID == p || e.ProcessID < 5 {
			continue
		}
		if s = windows.UTF16ToString(e.ExeFile[:]); len(s) == 0 {
			continue
		}
		if len(f.Exclude) > 0 && inStrList(s, f.Exclude) {
			continue
		}
		if len(f.Include) > 0 && !inStrList(s, f.Include) {
			continue
		}
		if o, err = windows.OpenProcess(a, true, e.ProcessID); err != nil || o == 0 {
			continue
		}
		if x == nil && ((f.Elevated == Empty && f.Session == Empty) || r) {
			if bugtrack.Enabled {
				bugtrack.Track("cmd.Filter.open(): Added process s=%q, e.ProcessID=%d for eval.", s, e.ProcessID)
			}
			l = append(l, o)
			// NOTE(dij): Left this commented to be un-commented if you want a fast-path to select.
			//            However, this prevents selecting a random process and grabs the first one instead.
			//            Also produces less handles opened.
			// if len(f.Include) == 1 {
			// 	break
			// }
			continue
		}
		if err = windows.OpenProcessToken(o, windows.TOKEN_QUERY, &t); err != nil {
			windows.CloseHandle(o)
			continue
		}
		if j, d = isTokenElevated(t), 0; f.Session != Empty {
			if err = windows.GetTokenInformation(t, windows.TokenSessionId, (*byte)(unsafe.Pointer(&d)), 4, &z); err != nil || z != 4 {
				d = 0
			}
		}
		if t.Close(); (f.Elevated == True && !j) || (f.Elevated == False && j) || (f.Session == True && d == 0) || (f.Session == False && d > 0) {
			windows.CloseHandle(o)
			continue
		}
		if x != nil && !x(e.ProcessID, j, s, uintptr(o)) {
			windows.CloseHandle(o)
			continue
		}
		if bugtrack.Enabled {
			bugtrack.Track("cmd.Filter.open(): Added process s=%q, e.ProcessID=%d, j=%t, d=%d for eval.", s, e.ProcessID, j, d)
		}
		l = append(l, o)
	}
	if windows.CloseHandle(h); len(l) == 1 {
		return l[0], nil
	}
	if len(l) > 1 {
		o = l[int(util.FastRandN(len(l)))]
		for i := range l {
			if l[i] == o {
				continue
			}
			windows.CloseHandle(l[i])
		}
		return o, nil
	}
	if !r && x == nil && f.Fallback {
		if bugtrack.Enabled {
			bugtrack.Track("cmd.Filter.open(): First run failed, starting fallback!")
		}
		return f.open(a, true, x)
	}
	return 0, ErrNoProcessFound
}
