//go:build windows

// Copyright (C) 2020 - 2023 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package filter

import (
	"strings"

	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

var emptyProc winapi.ProcessEntry

func isTokenElevated(h uintptr) bool {
	if !winapi.IsTokenElevated(h) {
		return false
	}
	switch u, err := winapi.GetTokenUser(h); {
	case err != nil:
		fallthrough
	case u.User.Sid.IsWellKnown(0x17): // 0x17 - WinLocalServiceSid
		fallthrough
	case u.User.Sid.IsWellKnown(0x18): // 0x18 - WinNetworkServiceSid
		return false
	}
	return true
}

// Select will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process ID will be returned.
//
// An 'ErrNoProcessFound' error will be returned if no processes that match the
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
// The first argument is the access rights requested, expressed as an uint32.
//
// An 'ErrNoProcessFound' error will be returned if no processes that match the
// Filter are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices.
func (f Filter) Token(a uint32) (uintptr, error) {
	return f.TokenFunc(a, nil)
}

// Thread will attempt to find a process with the specified Filter options.
// If a suitable process is found, a handle to the first Thread in the Process
// will be returned. The first argument is the access rights requested, expressed
// as an uint32.
//
// An 'ErrNoProcessFound' error will be returned if no processes that match the
// Filter are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices.
func (f Filter) Thread(a uint32) (uintptr, error) {
	return f.ThreadFunc(a, nil)
}

// Handle will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process Handle will be returned.
// The first argument is the access rights requested, expressed as an uint32.
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
// An 'ErrNoProcessFound' error will be returned if no processes that match the
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
	// 0x400 - PROCESS_QUERY_LIMITED_INFORMATION
	h, err := f.open(0x1000, false, x)
	if err != nil {
		return 0, err
	}
	return h.PID, nil
}

// TokenFunc will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process Token Handle will be returned.
// The first argument is the access rights requested, expressed as an uint32.
//
// This function allows for a filtering function to be passed along that will be
// supplied with the ProcessID, if the process is elevated, the process name
// and process handle. The function supplied should return true if the process
// passes the filter. The function argument may be nil.
//
// An 'ErrNoProcessFound' error will be returned if no processes that match the
// Filter are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices.
func (f Filter) TokenFunc(a uint32, x filter) (uintptr, error) {
	// 0x400 - PROCESS_QUERY_INFORMATION
	h, err := f.HandleFunc(0x400, x)
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
// The first argument is the access rights requested, expressed as an uint32.
//
// This function allows for a filtering function to be passed along that will be
// supplied with the ProcessID, if the process is elevated, the process name
// and process handle. The function supplied should return true if the process
// passes the filter. The function argument may be nil.
//
// An 'ErrNoProcessFound' error will be returned if no processes that match the
// Filter are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices.
func (f Filter) HandleFunc(a uint32, x filter) (uintptr, error) {
	// If we have a specific PID in mind that's valid.
	if f.PID > 4 {
		if err := winapi.GetDebugPrivilege(); err != nil {
			if bugtrack.Enabled {
				bugtrack.Track("filter.(Filter).open(): GetDebugPrivilege failed with err=%s", err.Error())
			}
		}
		h, err := winapi.OpenProcess(a, false, f.PID)
		if h == 0 || err != nil {
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
		// (old 0x8 - TOKEN_QUERY)
		// 0x20008 - TOKEN_READ | TOKEN_QUERY
		if err = winapi.OpenProcessToken(h, 0x20008, &t); err != nil {
			winapi.CloseHandle(h)
			return 0, xerr.Wrap("OpenProcessToken", err)
		}
		r := x(f.PID, isTokenElevated(t), n, h)
		if winapi.CloseHandle(t); r {
			return h, nil
		}
		winapi.CloseHandle(h)
		return 0, ErrNoProcessFound
	}
	h, err := f.open(a, false, x)
	if err != nil {
		return 0, err
	}
	return h.Handle(a)
}

// ThreadFunc will attempt to find a process with the specified Filter options.
// If a suitable process is found, a handle to the first Thread in the Process
// will be returned. The first argument is the access rights requested, expressed
// as an uint32.
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
func (f Filter) ThreadFunc(a uint32, x filter) (uintptr, error) {
	i, err := f.SelectFunc(x)
	if err != nil {
		return 0, err
	}
	if err = winapi.GetDebugPrivilege(); err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("filter.(Filter).open(): GetDebugPrivilege failed with err=%s", err.Error())
		}
	}
	var v uintptr
	err = winapi.EnumThreads(i, func(e winapi.ThreadEntry) error {
		if v, err = e.Handle(a); err == nil {
			return winapi.ErrNoMoreFiles
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	if v == 0 {
		return 0, ErrNoProcessFound
	}
	return v, nil
}
func (f Filter) open(a uint32, r bool, x filter) (winapi.ProcessEntry, error) {
	var (
		z = make([]winapi.ProcessEntry, 0, 64)
		p = winapi.GetCurrentProcessID()
		s = f.Session > Empty
		v = f.Elevated > Empty || x != nil
	)
	if err := winapi.GetDebugPrivilege(); err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("filter.(Filter).open(): GetDebugPrivilege failed with err=%s", err.Error())
		}
	}
	err := winapi.EnumProcesses(func(e winapi.ProcessEntry) error {
		if e.PID == p || e.PID < 5 || len(e.Name) == 0 || (f.PID > 0 && f.PID != e.PID) {
			return nil
		}
		if (len(f.Exclude) > 0 && inStrList(e.Name, f.Exclude)) || (len(f.Include) > 0 && !inStrList(e.Name, f.Include)) {
			return nil
		}
		if (x == nil && !s && !v) || r {
			h, err := e.Handle(a)
			if err != nil {
				return nil
			}
			if winapi.CloseHandle(h); bugtrack.Enabled {
				bugtrack.Track("filter.(Filter).open(): Added process e.Name=%s, e.PID=%d for eval.", e.Name, e.PID)
			}
			z = append(z, e)
			// NOTE(dij): Left this commented to be un-commented if you want a
			//            fast-path to select. However, this prevents selecting
			//            a random process and grabs the first one instead, but
			//            also produces fewer handles opened. YMMV
			//
			//	if len(f.Include) == 1 {
			//	    return false
			//	}
			//
			return nil
		}
		h, k, i, err := e.InfoEx(a, v, s, x != nil)
		if err != nil {
			return nil
		}
		if v && ((k && f.Elevated == False) || (!k && f.Elevated == True)) {
			return nil
		}
		if s && ((i > 0 && f.Session == False) || (i == 0 && f.Session == True)) {
			return nil
		}
		if x != nil {
			q := x(e.PID, k, e.Name, h)
			if winapi.CloseHandle(h); !q {
				return nil
			}
		}
		if z = append(z, e); bugtrack.Enabled {
			bugtrack.Track("filter.(Filter).open(): Added process e.Name=%s, e.PID=%d for eval.", e.Name, e.PID)
		}
		return nil
	})
	if err != nil {
		return emptyProc, err
	}
	switch len(z) {
	case 0:
		if !r && x == nil && f.Fallback {
			if bugtrack.Enabled {
				bugtrack.Track("filter.(Filter).open(): First run failed, starting fallback!")
			}
			return f.open(a, true, x)
		}
		return emptyProc, ErrNoProcessFound
	case 1:
		if bugtrack.Enabled {
			bugtrack.Track("filter.(Filter).open(): Choosing process e.Name=%s, e.PID=%d.", z[0].Name, z[0].PID)
		}
		return z[0], nil
	}
	n := z[int(util.FastRandN(len(z)))]
	if bugtrack.Enabled {
		bugtrack.Track("filter.(Filter).open(): Choosing process e.Name=%s, e.PID=%d.", n.Name, n.PID)
	}
	return n, nil
}
