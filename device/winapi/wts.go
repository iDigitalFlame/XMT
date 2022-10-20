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
	"time"
	"unsafe"

	"github.com/iDigitalFlame/xmt/data"
)

const (
	epoch          = 0x19DB1DED53E8000
	wtsProcSize    = unsafe.Sizeof(wtsProcess{})
	wtsSessionSize = unsafe.Sizeof(wtsSession{})
)

// Session is a struct that is used to indicate Windows Terminal Services (WTS)
// Login/Session data.
//
// This struct is similar to 'device.Login' but contains more non-generic data.
type Session struct {
	_         [0]func()
	User      string
	Host      string
	Domain    string
	Login     int64
	LastInput int64
	ID        uint32
	From      [16]byte
	Remote    bool
	Status    uint8
}
type wtsAddr struct {
	// DO NOT REORDER
	Family  uint32
	Address [16]byte
	_       uint32
}
type wtsInfo struct {
	// DO NOT REORDER
	_         uint32
	_, _, _   uint64
	_         [32]uint16
	Domain    [17]uint16
	User      [21]uint16
	_, _      int64
	LastInput int64
	Logon     int64
	Now       int64
}
type wtsSession struct {
	// DO NOT REORDER
	SessionID uint32
	Station   *uint16
	State     uint32
}
type wtsProcess struct {
	// DO NOT REORDER
	SessionID uint32
	PID       uint32
	Name      *uint16
	SID       *SID
}

// SessionProcess is a struct that contains information about a Process reterived
// via a 'WTSEnumerateProcesses' call.
type SessionProcess struct {
	_         [0]func()
	Name      string
	User      string
	SessionID uint32
	PID       uint32
}

// WTSCloseServer Windows API Call
//   Closes an open handle to a Remote Desktop Session Host (RD Session Host)
//   server.
//
// https://learn.microsoft.com/en-us/windows/win32/api/wtsapi32/nf-wtsapi32-wtscloseserver
func WTSCloseServer(h uintptr) {
	syscall.SyscallN(funcWTSCloseServer.address(), h)
}
func (s *Session) getSessionInfo(h uintptr) error {
	var (
		x uint32
		i *wtsInfo
	)
	r, _, err := syscall.SyscallN(
		funcWTSQuerySessionInformation.address(), h, uintptr(s.ID), 0x18, uintptr(unsafe.Pointer(&i)),
		uintptr(unsafe.Pointer(&x)),
	)
	if r == 0 {
		return unboxError(err)
	}
	if s.User, s.Domain = UTF16ToString(i.User[:]), UTF16ToString(i.Domain[:]); i.Logon > 0 {
		s.Login = time.Unix(0, (i.Logon-epoch)*100).Unix()
	}
	if i.LastInput > 0 {
		s.LastInput = time.Unix(0, (i.LastInput-epoch)*100).Unix()
	} else if i.Logon > 0 {
		s.LastInput = time.Unix(0, (i.Now-epoch)*100).Unix()
	}
	syscall.SyscallN(funcWTSFreeMemory.address(), uintptr(unsafe.Pointer(i)))
	var a *wtsAddr
	r, _, err = syscall.SyscallN(
		funcWTSQuerySessionInformation.address(), h, uintptr(s.ID), 0xE, uintptr(unsafe.Pointer(&a)),
		uintptr(unsafe.Pointer(&x)),
	)
	if s.Remote = false; r == 0 {
		return unboxError(err)
	}
	switch a.Family {
	case 0x2:
		copy(s.From[0:], a.Address[2:6])
		s.Remote = true
	case 0x17:
		copy(s.From[0:], a.Address[0:])
		s.Remote = true
	default:
		s.Remote = false
	}
	syscall.SyscallN(funcWTSFreeMemory.address(), uintptr(unsafe.Pointer(a)))
	return nil
}

// WTSOpenServer Windows API Call
//   Opens a handle to the specified Remote Desktop Session Host (RD Session Host)
//   server.
//
// https://learn.microsoft.com/en-us/windows/win32/api/wtsapi32/nf-wtsapi32-wtsopenserverw
func WTSOpenServer(server string) (uintptr, error) {
	if len(server) == 0 {
		return 0, nil
	}
	n, err := UTF16PtrFromString(server)
	if err != nil {
		return invalid, err
	}
	r, _, err1 := syscall.SyscallN(funcWTSOpenServer.address(), uintptr(unsafe.Pointer(n)))
	if r == invalid {
		return invalid, unboxError(err1)
	}
	return r, nil
}

// WTSGetSessions will attempt to reterive a detailed list of all Sessions
// on the target server handle (use 0 for the current host or use 'WTSOpenServer')
//
// This function will return a 'Session' struct for each Session found or any
// errors that may occur during enumeration.
func WTSGetSessions(server uintptr) ([]Session, error) {
	var (
		b          uintptr
		c          uint32
		r, _, err1 = syscall.SyscallN(funcWTSEnumerateSessions.address(), server, 0, 1, uintptr(unsafe.Pointer(&b)), uintptr(unsafe.Pointer(&c)))
	)
	if r == 0 {
		return nil, unboxError(err1)
	}
	if c == 0 {
		syscall.SyscallN(funcWTSFreeMemory.address(), b)
		return nil, nil
	}
	var (
		o   = make([]Session, 0, c)
		err error
	)
	for i := uint32(0); i < c; i++ {
		var (
			s = *(*wtsSession)(unsafe.Pointer(b + (wtsSessionSize * uintptr(i))))
			v = Session{ID: s.SessionID, Status: uint8(s.State), Host: UTF16PtrToString(s.Station)}
		)
		if err = v.getSessionInfo(server); err != nil {
			break
		}
		o = append(o, v)
	}
	syscall.SyscallN(funcWTSFreeMemory.address(), b)
	return o, err
}

// WTSGetSessionsHost will attempt to reterive a detailed list of all Sessions
// on the target server name (use an empty string for the local host).
//
// This function will return a 'Session' struct for each Session found or any
// errors that may occur during enumeration.
//
// This function calls 'WTSOpenServer(server)' then enumerates the Sessions and
// closes the handle after. If you would like more control, use the 'WTSGetSessions'
// function which takes a server handle instead.
func WTSGetSessionsHost(server string) ([]Session, error) {
	if len(server) == 0 {
		return WTSGetSessions(0)
	}
	h, err := WTSOpenServer(server)
	if err != nil {
		return nil, err
	}
	r, err := WTSGetSessions(h)
	if h > 0 {
		WTSCloseServer(h)
	}
	return r, err
}

// MarshalStream transforms this struct into a binary format and writes to the
// supplied data.Writer.
func (p SessionProcess) MarshalStream(w data.Writer) error {
	if err := w.WriteUint32(p.PID); err != nil {
		return err
	}
	if err := w.WriteUint32(0); err != nil {
		return err
	}
	if err := w.WriteString(p.Name); err != nil {
		return err
	}
	if err := w.WriteString(p.User); err != nil {
		return err
	}
	return nil
}

// WTSLogoffSession Windows API Call
//   Logs off a specified Remote Desktop Services session.
//
// https://learn.microsoft.com/en-us/windows/win32/api/wtsapi32/nf-wtsapi32-wtslogoffsession
func WTSLogoffSession(server uintptr, sid int32, wait bool) error {
	var w uint32
	if wait {
		w = 1
	}
	if r, _, err := syscall.SyscallN(funcWTSLogoffSession.address(), server, uintptr(sid), uintptr(w)); r == 0 {
		return unboxError(err)
	}
	return nil
}

// WTSDisconnectSession Windows API Call
//   Disconnects the logged-on user from the specified Remote Desktop Services
//   session without closing the session. If the user subsequently logs on to
//   the same Remote Desktop Session Host (RD Session Host) server, the user is
//   reconnected to the same session.
//
// https://learn.microsoft.com/en-us/windows/win32/api/wtsapi32/nf-wtsapi32-wtsdisconnectsession
func WTSDisconnectSession(server uintptr, sid int32, wait bool) error {
	var w uint32
	if wait {
		w = 1
	}
	if r, _, err := syscall.SyscallN(funcWTSDisconnectSession.address(), server, uintptr(sid), uintptr(w)); r == 0 {
		return unboxError(err)
	}
	return nil
}

// WTSEnumerateProcesses Windows API Call
//   Retrieves information about the active processes on a specified Remote
//   Desktop Session Host (RD Session Host) server.
//
// https://learn.microsoft.com/en-us/windows/win32/api/wtsapi32/nf-wtsapi32-wtsenumerateprocessesw
func WTSEnumerateProcesses(server uintptr, sid int32) ([]SessionProcess, error) {
	var (
		c         uint32
		b         uintptr
		r, _, err = syscall.SyscallN(funcWTSEnumerateProcesses.address(), server, 0, 1, uintptr(unsafe.Pointer(&b)), uintptr(unsafe.Pointer(&c)))
	)
	if r == 0 {
		return nil, unboxError(err)
	}
	if c == 0 {
		return nil, nil
	}
	o := make([]SessionProcess, 0, c)
	for i := uint32(0); i < c; i++ {
		if s := *(*wtsProcess)(unsafe.Pointer(b + (wtsProcSize * uintptr(i)))); sid < 0 || uint32(sid) == s.SessionID {
			u, _ := s.SID.UserName()
			o = append(o, SessionProcess{Name: UTF16PtrToString(s.Name), User: u, SessionID: s.SessionID, PID: s.PID})
		}
	}
	syscall.SyscallN(funcWTSFreeMemory.address(), b)
	return o, nil
}

// WTSSendMessage Windows API Call
//   Displays a message box on the client desktop of a specified Remote Desktop
//   Services session.
//
// https://learn.microsoft.com/en-us/windows/win32/api/wtsapi32/nf-wtsapi32-wtssendmessagew
func WTSSendMessage(server uintptr, sid int32, title, text string, f, secs uint32, wait bool) (uint32, error) {
	t, err := UTF16PtrFromString(title)
	if err != nil {
		return 0, err
	}
	d, err := UTF16PtrFromString(text)
	if err != nil {
		return 0, err
	}
	var o, w uint32
	if wait {
		w = 1
	}
	r, _, err1 := syscall.SyscallN(
		funcWTSSendMessage.address(), server, uintptr(sid), uintptr(unsafe.Pointer(t)), uintptr(len(title)*2), uintptr(unsafe.Pointer(d)),
		uintptr(len(text)*2), uintptr(f), uintptr(secs), uintptr(unsafe.Pointer(&o)), uintptr(w),
	)
	if r == 0 {
		return 0, unboxError(err1)
	}
	return o, nil
}
