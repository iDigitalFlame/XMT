//go:build windows

// Copyright (C) 2020 - 2022 iDigitalFlame
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

package winapi

import (
	"syscall"
	"unsafe"
)

// ThreadEntry is a basic struct passed to the user supplied function during
// a call to 'EnumThreads'. This struct supplies basic Thread information and
// can be used to gain more information about a Thread.
type ThreadEntry struct {
	_   [0]func()
	TID uint32
	PID uint32
	sus uint8
}

// ProcessEntry is a basic struct passed to the user supplied function during
// a call to 'EnumProcesses'. This struct supplies basic Process information
// and can be used to gain more information about a Process.
type ProcessEntry struct {
	_       [0]func()
	Name    string
	PID     uint32
	PPID    uint32
	Threads uint32
	session int32
}

// User attempts to reterive a string version of the username that this Process
// is running under.
//
// A string username and any errors during reterival will be returned.
func (p ProcessEntry) User() (string, error) {
	// 0x400 - PROCESS_QUERY_INFORMATION
	h, err := OpenProcess(0x400, false, p.PID)
	if err != nil {
		return "", err
	}
	var t uintptr
	// 0x8 - TOKEN_QUERY
	if err = OpenProcessToken(h, 0x8, &t); err != nil {
		CloseHandle(h)
		return "", err
	}
	u, err := UserFromToken(t)
	CloseHandle(t)
	CloseHandle(h)
	return u, err
}

// IsSuspended will attempt to determine if the current Thread is suspended. If
// the state information was supplied initially during discovery, it will be
// immediately returned, otherwise a Suspend/Resume cycle will be done to get the
// Thread suspension count.
//
// The return result will be true if the Thread is currently suspended and any
// errors that may have occurred.
func (t ThreadEntry) IsSuspended() (bool, error) {
	if t.sus > 0 {
		return t.sus == 2, nil
	}
	// 0x42 - THREAD_QUERY_INFORMATION | THREAD_SUSPEND_RESUME
	h, err := OpenThread(0x42, false, t.TID)
	if err != nil {
		return false, err
	}
	s, err := t.suspended(h)
	if CloseHandle(h); err != nil {
		return false, err
	}
	return s, nil
}

// Handle is a convenience function that calls 'OpenThread' on the Thread with
// the supplied access mask and returns a Thread handle that must be closed
// when you are done using it.
//
// This function does NOT make handles inheritable.
//
// Any errors that occur during the operation will be returned.
func (t ThreadEntry) Handle(a uint32) (uintptr, error) {
	return OpenThread(a, false, t.TID)
}

// Handle is a convenience function that calls 'OpenProcess' on the Process with
// the supplied access mask and returns a Process handle that must be closed
// when you are done using it.
//
// This function does NOT make handles inheritable.
//
// Any errors that occur during the operation will be returned.
func (p ProcessEntry) Handle(a uint32) (uintptr, error) {
	return OpenProcess(a, false, p.PID)
}
func (t ThreadEntry) suspended(h uintptr) (bool, error) {
	if t.sus > 0 {
		return t.sus == 2, nil
	}
	if getCurrentThreadID() == t.TID {
		// Can't do a suspend/resume cycle on ourselves.
		return false, syscall.EINVAL
	}
	if _, err := SuspendThread(h); err != nil {
		return false, err
	}
	c, err := ResumeThread(h)
	if err != nil {
		return false, err
	}
	return c > 1, nil
}

// Info will attempt to retrieve the Process session and Token elevation status
// and return it as a boolean (true if elevated) and a Session ID.
//
// The access mask can be used to determine the open permissions for the Process
// and this function will automatically add the PROCESS_QUERY_INFORMATION mask.
// If no access testing is desired, a value of zero is accepted.
//
// Boolean values for the elevation and session checks are passed as parameters
// to disable/enable checking of the value. If the value check is disabled (false)
// the return result will be the default value.
//
// Any errors during checking will be returned.
//
// To gain access to the underlying handle instead of opening a new one, use the
// 'InfoEx' function.
func (p ProcessEntry) Info(a uint32, elevated, session bool) (bool, uint32, error) {
	_, e, s, err := p.InfoEx(a, elevated, session, false)
	return e, s, err
}

// InfoEx will attempt to retrieve the Process handle (optional) session and Token
// elevation status and return it as a boolean (true if elevated) and a Session ID.
//
// The access mask can be used to determine the open permissions for the Process
// and this function will automatically add the PROCESS_QUERY_INFORMATION mask.
// If no access testing is desired, a value of zero is accepted. Unlike the non-Ex
// function 'Info', this function will return the un-closed Process handle if
// the last Boolean value for handle is true.
//
// Boolean values for the elevation and session checks are passed as parameters
// to disable/enable checking of the value. If the value check is disabled (false)
// the return result will be the default value.
//
// Any errors during checking will be returned.
func (p ProcessEntry) InfoEx(a uint32, elevated, session, handle bool) (uintptr, bool, uint32, error) {
	if !handle && !elevated && !session {
		return 0, false, 0, nil
	}
	if !handle && !elevated && session && p.session >= 0 {
		return 0, false, uint32(p.session), nil
	}
	// 0x400 - PROCESS_QUERY_INFORMATION
	//
	// NOTE(dij): The reason we have an access param is so we can only open
	//            the handle once while we're doing this to "check" if we can
	//            access it with the requested access we want.
	//            When Filters call this function, we do a quick 'Handle' check
	//            to make sure we can open it before adding to the eval list.
	h, err := OpenProcess(a|0x400, false, p.PID)
	if err != nil {
		return 0, false, 0, err
	}
	if !elevated && !session {
		if !handle {
			CloseHandle(h)
			return 0, false, 0, nil
		}
		return h, false, 0, nil
	}
	var t uintptr
	// 0x20008 - TOKEN_READ
	if err = OpenProcessToken(h, 0x20008, &t); err != nil {
		CloseHandle(h)
		return 0, false, 0, err
	}
	var v uint32
	if p.session >= 0 {
		v = uint32(p.session)
	} else {
		var s uint32
		// 0xC - TokenSessionId
		if err = GetTokenInformation(t, 0xC, (*byte)(unsafe.Pointer(&v)), 4, &s); err != nil || s != 4 {
			CloseHandle(t)
			CloseHandle(h)
			return 0, false, 0, err
		}
	}
	e := IsTokenElevated(t)
	if e {
		switch u, err := GetTokenUser(t); {
		case err != nil:
			fallthrough
		case u.User.Sid.IsWellKnown(0x17): // 0x17 - WinLocalServiceSid
			fallthrough
		case u.User.Sid.IsWellKnown(0x18): // 0x18 - WinNetworkServiceSid
			e = false
		}
	}
	if CloseHandle(t); !handle {
		CloseHandle(h)
		return 0, e, v, nil
	}
	return h, e, v, nil
}
