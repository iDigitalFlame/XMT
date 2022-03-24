//go:build windows

package man

import (
	"unsafe"

	"github.com/iDigitalFlame/xmt/com/pipe"
	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

type objListener uintptr

func (objListener) Listen() {}
func (l objListener) Close() error {
	return winapi.CloseHandle(uintptr(l))
}
func mutexCheck(s string) (bool, error) {
	if len(s) == 0 || len(s) > 248 {
		return false, xerr.Sub("invalid name", 0xA)
	}
	if s[0] != '\\' {
		s = prefix + s
	}
	m, err := winapi.OpenMutex(0x120000, false, s)
	if err != nil {
		return false, err
	}
	winapi.CloseHandle(m)
	return true, nil
}
func eventCheck(s string) (bool, error) {
	if len(s) == 0 || len(s) > 248 {
		return false, xerr.Sub("invalid name", 0xA)
	}
	if s[0] != '\\' {
		s = prefix + s
	}
	e, err := winapi.OpenEvent(0x120000, false, s)
	if err != nil {
		return false, err
	}
	winapi.CloseHandle(e)
	return true, nil
}
func mailslotCheck(s string) (bool, error) {
	if len(s) == 0 || len(s) > 243 {
		return false, xerr.Sub("invalid name", 0xA)
	}
	if len(s) < 4 || (s[0] != '\\' && s[1] != '\\' && s[2] != '.' && s[3] != '\\') {
		s = slot + s
	}
	m, err := winapi.CreateFile(s, 0xc0000000, 0x3, nil, 0x3, 0, 0)
	if err != nil {
		return false, err
	}
	winapi.CloseHandle(m)
	return true, nil
}
func semaphoreCheck(s string) (bool, error) {
	if len(s) == 0 || len(s) > 248 {
		return false, xerr.Sub("invalid name", 0xA)
	}
	if s[0] != '\\' {
		s = prefix + s
	}
	r, err := winapi.OpenSemaphore(0x120000, false, s)
	if err != nil {
		return false, err
	}
	winapi.CloseHandle(r)
	return true, nil
}
func mutexCreate(s string) (listener, error) {
	if len(s) == 0 || len(s) > 248 {
		return nil, xerr.Sub("invalid name", 0xA)
	}
	if s[0] != '\\' {
		s = prefix + s
	}
	var (
		v   = winapi.SecurityAttributes{InheritHandle: 1}
		err error
	)
	if v.SecurityDescriptor, err = winapi.SecurityDescriptorFromString(pipe.PermEveryone); err != nil {
		return nil, err
	}
	v.Length = uint32(unsafe.Sizeof(s))
	m, err := winapi.CreateMutex(&v, true, s)
	if err != nil {
		return nil, err
	}
	return objListener(m), nil
}
func eventCreate(s string) (listener, error) {
	if len(s) == 0 || len(s) > 248 {
		return nil, xerr.Sub("invalid name", 0xA)
	}
	if s[0] != '\\' {
		s = prefix + s
	}
	var (
		v   = winapi.SecurityAttributes{InheritHandle: 1}
		err error
	)
	if v.SecurityDescriptor, err = winapi.SecurityDescriptorFromString(pipe.PermEveryone); err != nil {
		return nil, err
	}
	v.Length = uint32(unsafe.Sizeof(s))
	e, err := winapi.CreateEvent(&v, true, true, s)
	if err != nil {
		return nil, err
	}
	return objListener(e), nil
}
func (o objSync) check(s string) (bool, error) {
	switch o {
	case Mutex:
		return mutexCheck(s)
	case Event:
		return eventCheck(s)
	case Mailslot:
		return mailslotCheck(s)
	case Semaphore:
		return semaphoreCheck(s)
	}
	return false, xerr.Sub("invalid value", 0xD)
}
func mailslotCreate(s string) (listener, error) {
	if len(s) == 0 || len(s) > 243 {
		return nil, xerr.Sub("invalid name", 0xA)
	}
	if len(s) < 4 || (s[0] != '\\' && s[1] != '\\' && s[2] != '.' && s[3] != '\\') {
		s = slot + s
	}
	var (
		v   = winapi.SecurityAttributes{InheritHandle: 1}
		err error
	)
	if v.SecurityDescriptor, err = winapi.SecurityDescriptorFromString(pipe.PermEveryone); err != nil {
		return nil, err
	}
	v.Length = uint32(unsafe.Sizeof(s))
	r, err := winapi.CreateMailslot(s, 0, -1, &v)
	if err != nil {
		return nil, err
	}
	return objListener(r), nil
}
func semaphoreCreate(s string) (listener, error) {
	if len(s) == 0 || len(s) > 248 {
		return nil, xerr.Sub("invalid name", 0xA)
	}
	if s[0] != '\\' {
		s = prefix + s
	}
	var (
		v   = winapi.SecurityAttributes{InheritHandle: 1}
		err error
	)
	if v.SecurityDescriptor, err = winapi.SecurityDescriptorFromString(pipe.PermEveryone); err != nil {
		return nil, err
	}
	v.Length = uint32(unsafe.Sizeof(s))
	r, err := winapi.CreateSemaphore(&v, 0, 1, s)
	if err != nil {
		return nil, err
	}
	return objListener(r), nil
}
func (o objSync) create(s string) (listener, error) {
	switch o {
	case Mutex:
		return mutexCreate(s)
	case Event:
		return eventCreate(s)
	case Mailslot:
		return mailslotCreate(s)
	case Semaphore:
		return semaphoreCreate(s)
	}
	return nil, xerr.Sub("invalid type", 0xC)
}
