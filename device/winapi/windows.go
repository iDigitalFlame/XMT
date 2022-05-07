//go:build windows

package winapi

import (
	"sync"
	"syscall"
	"unsafe"

	"github.com/iDigitalFlame/xmt/data"
)

var winCb windowSearcher
var enumWindowsOnce struct {
	sync.Once
	f uintptr
}

// Window is a struct that represents a Windows Window. The handles are the same
// for the duration of the Window's existence.
type Window struct {
	Name   string
	Handle uintptr
	X, Y   uint32
	Width  uint32
	Height uint32
}
type windowInfo struct {
	Size    uint32
	Window  rect
	Client  rect
	Style   uint32
	ExStyle uint32
	Status  uint32
	_, _    uint32
	_, _    uint16
}
type windowSearcher struct {
	sync.Mutex
	e []Window
}

func initWindowsEnumFunc() {
	enumWindowsOnce.f = syscall.NewCallback(enumWindowsCallback)
}

// TopLevelWindows returns a list of the current (non-dialog) Windows as an
// slice with their Name, Handle, Size and Position.
//
// The handles may be used for multiple functions and are valid until the window
// is closed.
func TopLevelWindows() ([]Window, error) {
	winCb.Lock()
	enumWindowsOnce.Do(initWindowsEnumFunc)
	var (
		e         []Window
		r, _, err = syscall.SyscallN(funcEnumWindows.address(), enumWindowsOnce.f, 0)
	)
	e, winCb.e = winCb.e, nil
	if winCb.Unlock(); r == 0 {
		return nil, unboxError(err)
	}
	return e, nil
}
func enumWindowsCallback(h, _ uintptr) uintptr {
	n, _, _ := syscall.SyscallN(funcGetWindowTextLength.address(), h)
	if n == 0 {
		return 1
	}
	var i windowInfo
	i.Size = uint32(unsafe.Sizeof(i))
	if r, _, _ := syscall.SyscallN(funcGetWindowInfo.address(), h, uintptr(unsafe.Pointer(&i))); r == 0 {
		return 1
	}
	if i.Style&0x80000000 != 0 || i.Style == 0x4C00000 {
		return 1
	}
	v := make([]uint16, n+1)
	if n, _, _ = syscall.SyscallN(funcGetWindowText.address(), h, uintptr(unsafe.Pointer(&v[0])), n+1); n == 0 {
		return 1
	}
	winCb.e = append(winCb.e, Window{
		X:      uint32(i.Window.Left),
		Y:      uint32(i.Window.Top),
		Name:   UTF16ToString(v[:n]),
		Width:  uint32(i.Window.Right - i.Window.Left),
		Height: uint32(i.Window.Bottom - i.Window.Top),
		Handle: h,
	})
	return 1
}

// MarshalStream transforms this struct into a binary format and writes to the
// supplied data.Writer.
func (i Window) MarshalStream(w data.Writer) error {
	if err := w.WriteUint64(uint64(i.Handle)); err != nil {
		return err
	}
	if err := w.WriteString(i.Name); err != nil {
		return err
	}
	if err := w.WriteUint32(i.X); err != nil {
		return err
	}
	if err := w.WriteUint32(i.Y); err != nil {
		return err
	}
	if err := w.WriteUint32(i.Width); err != nil {
		return err
	}
	if err := w.WriteUint32(i.Height); err != nil {
		return err
	}
	return nil
}

// EnableWindow Windows API Call
//   Enables or disables mouse and keyboard input to the specified window or
//   control. When input is disabled, the window does not receive input such as
//   mouse clicks and key presses. When input is enabled, the window receives
//   all input.
//
// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-enablewindow
func EnableWindow(h uintptr, e bool) (bool, error) {
	var v uint32
	if e {
		v = 1
	}
	r, _, err := syscall.SyscallN(funcEnableWindow.address(), h, uintptr(v))
	if err > 0 {
		return r > 0, unboxError(err)
	}
	return r > 0, nil
}

// SetWindowTransparency will attempt to set the transparency of the window handle
// to 0-255, 0 being completely transparent and 255 being opaque.
func SetWindowTransparency(h uintptr, t byte) error {
	r, _, err := syscall.SyscallN(funcGetWindowLongPtr.address(), h, layeredPtr)
	if r == 0 && err > 0 {
		return unboxError(err)
	}
	syscall.SyscallN(funcSetWindowLongPtr.address(), h, r|0x80000)
	if r, _, err = syscall.SyscallN(funcSetLayeredWindowAttributes.address(), h, 0, uintptr(t), 3); r == 0 {
		return unboxError(err)
	}
	return nil
}
