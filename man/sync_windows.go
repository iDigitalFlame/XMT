//go:build windows
// +build windows

package man

import (
	"unsafe"

	"github.com/iDigitalFlame/xmt/com/pipe"
	"github.com/iDigitalFlame/xmt/util/xerr"
	"golang.org/x/sys/windows"
)

var (
	dllKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	funcOpenSemaphore   = dllKernel32.NewProc("OpenSemaphoreW")
	funcCreateMailslot  = dllKernel32.NewProc("CreateMailslotW")
	funcCreateSemaphore = dllKernel32.NewProc("CreateSemaphoreW")
)

type objListener uintptr

func (objListener) Listen() {}

func (l objListener) Close() error {
	return windows.CloseHandle(windows.Handle(l))
}
func mutexCheck(s string) (bool, error) {
	if len(s) == 0 || len(s) > 248 {
		return false, xerr.New("invalid Mutex name")
	}
	if s[0] != '\\' {
		s = "Global\\" + s
	}
	n, err := windows.UTF16PtrFromString(s)
	if err != nil {
		return false, xerr.Wrap("could not convert Mutex name", err)
	}
	m, err := windows.OpenMutex(windows.READ_CONTROL|windows.SYNCHRONIZE, false, n)
	if err != nil {
		return false, xerr.Wrap("could not open Mutex", err)
	}
	if m == 0 || m == windows.InvalidHandle {
		return false, xerr.New("could not open Mutex: (unknown error)")
	}
	windows.CloseHandle(m)
	return true, nil
}
func eventCheck(s string) (bool, error) {
	if len(s) == 0 || len(s) > 248 {
		return false, xerr.New("invalid Event name")
	}
	if s[0] != '\\' {
		s = "Global\\" + s
	}
	n, err := windows.UTF16PtrFromString(s)
	if err != nil {
		return false, xerr.Wrap("could not convert Event name", err)
	}
	e, err := windows.OpenEvent(windows.READ_CONTROL|windows.SYNCHRONIZE, false, n)
	if err != nil {
		return false, xerr.Wrap("could not open Event", err)
	}
	if e == 0 || e == windows.InvalidHandle {
		return false, xerr.New("could not open Event: (unknown error)")
	}
	windows.CloseHandle(e)
	return true, nil
}
func mailslotCheck(s string) (bool, error) {
	if len(s) == 0 || len(s) > 243 {
		return false, xerr.New("invalid Mailslot name")
	}
	if len(s) < 4 || (s[0] != '\\' && s[1] != '\\' && s[2] != '.' && s[3] != '\\') {
		s = `\\.\mailslot` + "\\" + s
	}
	n, err := windows.UTF16PtrFromString(s)
	if err != nil {
		return false, xerr.Wrap("could not convert Mailslot name", err)
	}
	m, err := windows.CreateFile(
		n, windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil, windows.OPEN_EXISTING, 0, 0,
	)
	if m == 0 || err != nil {
		return false, xerr.Wrap("could not open Mailslot", err)
	}
	windows.CloseHandle(m)
	return true, nil
}
func semaphoreCheck(s string) (bool, error) {
	if len(s) == 0 || len(s) > 248 {
		return false, xerr.New("invalid Semaphore name")
	}
	if s[0] != '\\' {
		s = "Global\\" + s
	}
	n, err := windows.UTF16PtrFromString(s)
	if err != nil {
		return false, xerr.Wrap("could not convert Semaphore name", err)
	}
	r, _, err := funcOpenSemaphore.Call(windows.READ_CONTROL|windows.SYNCHRONIZE, 0, uintptr(unsafe.Pointer(n)))
	if r == 0 || r == uintptr(windows.InvalidHandle) {
		return false, xerr.Wrap("could not open Semaphore", err)
	}
	windows.CloseHandle(windows.Handle(r))
	return true, nil
}
func mutexCreate(s string) (listener, error) {
	if len(s) == 0 || len(s) > 248 {
		return nil, xerr.New("invalid Mutex name")
	}
	if s[0] != '\\' {
		s = "Global\\" + s
	}
	n, err := windows.UTF16PtrFromString(s)
	if err != nil {
		return nil, xerr.Wrap("could not convert Mutex name", err)
	}
	v := windows.SecurityAttributes{InheritHandle: 1}
	if v.SecurityDescriptor, err = windows.SecurityDescriptorFromString(pipe.PermEveryone); err != nil {
		return nil, err
	}
	v.Length = uint32(unsafe.Sizeof(s))
	m, err := windows.CreateMutex(&v, true, n)
	if err != nil {
		return nil, xerr.Wrap("could not create Mutex", err)
	}
	return objListener(m), nil
}
func eventCreate(s string) (listener, error) {
	if len(s) == 0 || len(s) > 248 {
		return nil, xerr.New("invalid Event name")
	}
	if s[0] != '\\' {
		s = "Global\\" + s
	}
	n, err := windows.UTF16PtrFromString(s)
	if err != nil {
		return nil, xerr.Wrap("could not convert Event name", err)
	}
	v := windows.SecurityAttributes{InheritHandle: 1}
	if v.SecurityDescriptor, err = windows.SecurityDescriptorFromString(pipe.PermEveryone); err != nil {
		return nil, err
	}
	v.Length = uint32(unsafe.Sizeof(s))
	e, err := windows.CreateEvent(&v, 1, 1, n)
	if err != nil {
		return nil, xerr.Wrap("could not create Event", err)
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
	return false, xerr.New("invalid object type")
}
func mailslotCreate(s string) (listener, error) {
	if len(s) == 0 || len(s) > 243 {
		return nil, xerr.New("invalid Mailslot name")
	}
	if len(s) < 4 || (s[0] != '\\' && s[1] != '\\' && s[2] != '.' && s[3] != '\\') {
		s = `\\.\mailslot` + "\\" + s
	}
	n, err := windows.UTF16PtrFromString(s)
	if err != nil {
		return nil, xerr.Wrap("could not convert Mailslot name", err)
	}
	v := windows.SecurityAttributes{InheritHandle: 1}
	if v.SecurityDescriptor, err = windows.SecurityDescriptorFromString(pipe.PermEveryone); err != nil {
		return nil, err
	}
	v.Length = uint32(unsafe.Sizeof(s))
	r, _, err := funcCreateMailslot.Call(uintptr(unsafe.Pointer(n)), 0, ^uintptr(0), uintptr(unsafe.Pointer(&v)))
	if r == 0 || r == uintptr(windows.InvalidHandle) || err == windows.ERROR_ALREADY_EXISTS {
		return nil, xerr.Wrap("could not create Mailslot", err)
	}
	return objListener(r), nil
}
func semaphoreCreate(s string) (listener, error) {
	if len(s) == 0 || len(s) > 248 {
		return nil, xerr.New("invalid Semaphore name")
	}
	if s[0] != '\\' {
		s = "Global\\" + s
	}
	n, err := windows.UTF16PtrFromString(s)
	if err != nil {
		return nil, xerr.Wrap("could not convert Semaphore name", err)
	}
	v := windows.SecurityAttributes{InheritHandle: 1}
	if v.SecurityDescriptor, err = windows.SecurityDescriptorFromString(pipe.PermEveryone); err != nil {
		return nil, err
	}
	v.Length = uint32(unsafe.Sizeof(s))
	r, _, err := funcCreateSemaphore.Call(uintptr(unsafe.Pointer(&v)), 0, 1, uintptr(unsafe.Pointer(n)))
	if r == 0 || err == windows.ERROR_ALREADY_EXISTS {
		return nil, xerr.Wrap("could not create Semaphore", err)
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
	return nil, xerr.New("invalid object type")
}
