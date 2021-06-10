// +build windows

package cmd

import (
	"os"
	"strconv"
	"strings"
	"unsafe"

	"github.com/iDigitalFlame/xmt/device/devtools"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/xerr"
	"golang.org/x/sys/windows"
)

// Select will attempt to find a process with the specified Filter options. If a suitable process
// is found, the Process ID will be returned. An 'ErrNoProcessFound' error will be returned if no
// processes that match the Filter match. This function returns 'ErrNoWindows' on non-Windows devices.
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

// SelectFunc will attempt to find a process with the specified Filter options. If a suitable process
// is found, the Process ID will be returned. An 'ErrNoProcessFound' error will be returned if no
// processes that match the Filter match. This function returns 'ErrNoWindows' on non-Windows devices.
//
// This function allows for a filtering function to be passed along that will be supplied with
// the ProcessID, if the process is elevated, the process handle and process name. The function supplied
// should return true if the process passes the filter. The function argument may be nil.
func (f Filter) SelectFunc(x filter) (uint32, error) {
	// Process ID values in Windows technically begin at 5, [0-4] are System, we can't really use them or
	// get handles to them.
	if f.PID > 4 && x == nil {
		// No need to check info (we have no extra filter)
		return f.PID, nil
	}
	h, err := f.open(windows.PROCESS_QUERY_INFORMATION, false, x)
	if err != nil {
		return 0, err
	}
	p, err := windows.GetProcessId(h)
	if windows.CloseHandle(h); err != nil {
		return 0, err
	}
	return p, nil
}
func (f Filter) handle(a uint32) (windows.Handle, error) {
	if f.PID > 4 {
		h, err := windows.OpenProcess(a, true, uint32(f.PID))
		if h == 0 || err != nil {
			return 0, xerr.Wrap("winapi OpenProcess PID "+strconv.Itoa(int(f.PID))+" error", err)
		}
		return h, nil
	}
	return f.open(a, false, nil)
}
func (f Filter) open(a uint32, r bool, x filter) (windows.Handle, error) {
	h, err := windows.CreateToolhelp32Snapshot(0x0002, 0)
	if err != nil {
		return 0, xerr.Wrap("winapi CreateToolhelp32Snapshot error", err)
	}
	devtools.AdjustPrivileges("SeDebugPrivilege")
	var (
		e     windows.ProcessEntry32
		l     []windows.Handle
		z     windows.Token
		s     string
		o     windows.Handle
		y, yr uint32
		j     bool
		p     = uint32(os.Getpid())
	)
	e.Size = uint32(unsafe.Sizeof(e))
	for err = windows.Process32First(h, &e); err == nil; err = windows.Process32Next(h, &e) {
		if e.ProcessID == p || e.ProcessID <= 4 {
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
			l = append(l, o)
			/*
				// Left this commented to be un-commented if you want a fast-path to select.
				// However, this prevents selecting a random process and grabs the first one instead.
				if len(f.Include) == 1 {
					break
				}
			*/
			continue
		}
		if err = windows.OpenProcessToken(o, windows.TOKEN_QUERY, &z); err != nil {
			windows.CloseHandle(o)
			continue
		}
		if j, y = z.IsElevated(), 0; f.Session != Empty {
			if err = windows.GetTokenInformation(z, windows.TokenSessionId, (*byte)(unsafe.Pointer(&y)), 4, &yr); err != nil || yr != 4 {
				y = 0
			}
		}
		if z.Close(); (f.Elevated == True && !j) || (f.Elevated == False && j) || (f.Session == True && y == 0) || (f.Session == False && y > 0) {
			windows.CloseHandle(o)
			continue
		}
		if (x != nil && !x(e.ProcessID, j, s, uintptr(o))) || (f.PID > 4 && e.ProcessID == f.PID) {
			windows.CloseHandle(o)
			continue
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
		return f.open(a, true, x)
	}
	return 0, ErrNoProcessFound
}
