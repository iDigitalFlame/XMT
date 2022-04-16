//go:build windows

package filter

import (
	"strconv"
	"strings"
	"unsafe"

	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

func isTokenElevated(h uintptr) bool {
	if !winapi.IsTokenElevated(h) {
		return false
	}
	switch u, err := winapi.GetTokenUser(h); {
	case err != nil:
		fallthrough
	case u.User.Sid.IsWellKnown(0x17):
		fallthrough
	case u.User.Sid.IsWellKnown(0x18):
		return false
	}
	return true
}

// Select will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process ID will be returned.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the
// Filter are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices if a PID is not set.
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

// Token will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process Token Handle will be returned.
// The first argument is the access rights requested, expressed as a uint32.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the
// Filter are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices.
func (f Filter) Token(a uint32) (uintptr, error) {
	return f.TokenFunc(a, nil)
}

// Handle will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process Handle will be returned.
// The first argument is the access rights requested, expressed as a uint32.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the
// Filter are found.
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
// An'ErrNoProcessFound' error will be returned if no processes that match the
// Filter are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices if a PID is not set.
func (f Filter) SelectFunc(x filter) (uint32, error) {
	// NOTE(dij): Process ID values in Windows technically begin at 5, [0-4] are
	//            System, we can't really use them or get handles to them.
	if f.PID > 4 && x == nil {
		// NOTE(dij): No need to check info (we have no extra filter)
		return f.PID, nil
	}
	h, err := f.HandleFunc(0x410, x)
	if err != nil {
		return 0, err
	}
	p, err := winapi.GetProcessID(h)
	if winapi.CloseHandle(h); err != nil {
		return 0, err
	}
	return p, nil
}

// TokenFunc will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process Token Handle will be returned.
// The first argument is the access rights requested, expressed as a uint32.
//
// This function allows for a filtering function to be passed along that will be
// supplied with the ProcessID, if the process is elevated, the process name
// and process handle. The function supplied should return true if the process
// passes the filter. The function argument may be nil.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the
// Filter are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices.
func (f Filter) TokenFunc(a uint32, x filter) (uintptr, error) {
	h, err := f.HandleFunc(0x440, x)
	if err != nil {
		return 0, err
	}
	var t uintptr
	err = winapi.OpenProcessToken(h, a, &t)
	winapi.CloseHandle(h)
	return t, err
}

// HandleFunc will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process Handle will be returned.
// The first argument is the access rights requested, expressed as a uint32.
//
// This function allows for a filtering function to be passed along that will be
// supplied with the ProcessID, if the process is elevated, the process name
// and process handle. The function supplied should return true if the process
// passes the filter. The function argument may be nil.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the
// Filter are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices.
func (f Filter) HandleFunc(a uint32, x filter) (uintptr, error) {
	// If we have a specific PID in mind that's valid.
	if f.PID > 4 {
		h, err := winapi.OpenProcess(a, true, f.PID)
		if h == 0 || err != nil {
			if xerr.Concat {
				return 0, xerr.Wrap("OpenProcess "+strconv.Itoa(int(f.PID)), err)
			}
			return 0, xerr.Wrap("OpenProcess", err)
		}
		if x == nil {
			// NOTE(dij): Quick path since we don't have a filter func.
			return h, nil
		}
		var n string
		if n, err = winapi.GetProcessFileName(h); err != nil {
			winapi.CloseHandle(h)
			return 0, xerr.Wrap("GetProcessFileName", err)
		}
		var t uintptr
		if err = winapi.OpenProcessToken(h, 0x8, &t); err != nil {
			winapi.CloseHandle(h)
			return 0, xerr.Wrap("OpenProcessToken", err)
		}
		r := x(f.PID, isTokenElevated(t), n, uintptr(h))
		if winapi.CloseHandle(t); r {
			return h, nil
		}
		winapi.CloseHandle(h)
		return 0, ErrNoProcessFound
	}
	return f.open(a, false, x)
}
func (f Filter) open(a uint32, r bool, x filter) (uintptr, error) {
	h, err := winapi.CreateToolhelp32Snapshot(2, 0)
	if err != nil {
		return 0, xerr.Wrap("CreateToolhelp32Snapshot", err)
	}
	var (
		e    winapi.ProcessEntry32
		l    []uintptr
		s    string
		o, t uintptr
		d, z uint32
		j    bool
		p    = winapi.GetCurrentProcessID()
	)
	e.Size = uint32(unsafe.Sizeof(e))
	if err = winapi.GetDebugPrivilege(); err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("filter.Filter.open(): GetDebugPrivilege failed with err=%s", err)
		}
	}
	for err = winapi.Process32First(h, &e); err == nil; err = winapi.Process32Next(h, &e) {
		if e.ProcessID == p || e.ProcessID < 5 {
			continue
		}
		if s = winapi.UTF16ToString(e.ExeFile[:]); len(s) == 0 {
			continue
		}
		if len(f.Exclude) > 0 && inStrList(s, f.Exclude) {
			continue
		}
		if len(f.Include) > 0 && !inStrList(s, f.Include) {
			continue
		}
		if o, err = winapi.OpenProcess(a, true, e.ProcessID); err != nil || o == 0 {
			if bugtrack.Enabled {
				bugtrack.Track("filter.Filter.open(): OpenProcess e.ProcessID=%d, failed err=%s.", e.ProcessID, err)
			}
			continue
		}
		if x == nil && ((f.Elevated == Empty && f.Session == Empty) || r) {
			if bugtrack.Enabled {
				bugtrack.Track("filter.Filter.open(): Added process s=%q, e.ProcessID=%d for eval.", s, e.ProcessID)
			}
			l = append(l, o)
			// NOTE(dij): Left this commented to be un-commented if you want a fast-path to select.
			//            However, this prevents selecting a random process and grabs the first one instead.
			//            Also produces less handles opened.
			// if len(f.Include) == 1 {
			//     break
			// }
			continue
		}
		if err = winapi.OpenProcessToken(o, 0x8, &t); err != nil {
			winapi.CloseHandle(o)
			continue
		}
		if j, d = isTokenElevated(t), 0; f.Session != Empty {
			if err = winapi.GetTokenInformation(t, 0xC, (*byte)(unsafe.Pointer(&d)), 4, &z); err != nil || z != 4 {
				d = 0
			}
		}
		if winapi.CloseHandle(t); (f.Elevated == True && !j) || (f.Elevated == False && j) || (f.Session == True && d == 0) || (f.Session == False && d > 0) {
			winapi.CloseHandle(o)
			continue
		}
		if x != nil && !x(e.ProcessID, j, s, uintptr(o)) {
			winapi.CloseHandle(o)
			continue
		}
		if bugtrack.Enabled {
			bugtrack.Track("filter.Filter.open(): Added process s=%q, e.ProcessID=%d, j=%t, d=%d for eval.", s, e.ProcessID, j, d)
		}
		l = append(l, o)
	}
	if winapi.CloseHandle(h); len(l) == 1 {
		return l[0], nil
	}
	if len(l) > 1 {
		o = l[int(util.FastRandN(len(l)))]
		for i := range l {
			if l[i] == o {
				continue
			}
			winapi.CloseHandle(l[i])
		}
		return o, nil
	}
	if !r && x == nil && f.Fallback {
		if bugtrack.Enabled {
			bugtrack.Track("filter.Filter.open(): First run failed, starting fallback!")
		}
		return f.open(a, true, x)
	}
	return 0, ErrNoProcessFound
}
