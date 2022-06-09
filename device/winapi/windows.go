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
	"sync"
	"syscall"
	"unsafe"

	"github.com/iDigitalFlame/xmt/data"
)

const (
	// SwHide hides the window and activates another window.
	SwHide uint8 = iota
	// SwNormal activates and displays a window. If the window is minimized or
	// maximized, the system restores it to its original size and position. An
	// application should specify this flag when displaying the window for the
	// first time.
	SwNormal
	// SwMinimized activates the window and displays it as a minimized window.
	SwMinimized
	// SwMaximize activates the window and displays it as a maximized window.
	SwMaximize
	// SwNoActive displays a window in its most recent size and position. This
	// value is similar to SwNormal, except that the window is not activated.
	SwNoActive
	// SwShow activates the window and displays it in its current size and
	// position.
	SwShow
	// SwMinimize minimizes the specified window and activates the next top-level
	// window in the Z order.
	SwMinimize
	// SwMinimizeNoActive displays the window as a minimized window. This value
	// is similar to SwMinimizeNoActive, except the window is not activated.
	SwMinimizeNoActive
	// SwShowNoActive displays the window in its current size and position.
	// This value is similar to SwShow, except that the window is not activated.
	SwShowNoActive
	// SwRestore activates and displays the window. If the window is minimized
	// or maximized, the system restores it to its original size and position.
	// An application should specify this flag when restoring a minimized window.
	SwRestore
	// SwDefault sets the show state based on the SW_ value specified in the
	// STARTUPINFO structure passed to the CreateProcess function by the program
	// that started the application.
	SwDefault
	// SwMinimizeForce minimizes a window, even if the thread that owns the
	// window is not responding. This flag should only be used when minimizing
	// windows from a different thread.
	SwMinimizeForce
)

var winCb windowSearcher
var enumWindowsOnce struct {
	sync.Once
	f uintptr
}

type key struct {
	Key   uint16
	_     uint16
	Flags uint32
	_     uint32
	_     uint64
}
type input struct {
	Type uint32
	Key  key
	_    uint64
}

// Window is a struct that represents a Windows Window. The handles are the same
// for the duration of the Window's existence.
type Window struct {
	Name          string
	Flags         uint8
	Handle        uintptr
	X, Y          int32
	Width, Height int32
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
func sendText(s string) error {
	var b [256]input
	if len(s) < 64 {
		return sendKeys(&b, s)
	}
	for i, e := 0, 0; i < len(s); {
		if e = i + 64; e > len(s) {
			e = len(s)
		}
		if err := sendKeys(&b, s[i:e]); err != nil {
			return err
		}
		i = e
	}
	return nil
}

// CloseWindow is a helper function that sends the WM_DESTROY to the supplied
// Window handle.
//
// If the value of h is 0, this will target ALL FOUND WINDOWS.
func CloseWindow(h uintptr) error {
	if h > 0 {
		return closeWindow(h)
	}
	w, err := TopLevelWindows()
	if err != nil {
		return err
	}
	for i := range w {
		closeWindow(w[i].Handle)
	}
	w = nil
	return err
}
func closeWindow(h uintptr) error {
	r, _, err := syscall.SyscallN(funcSendNotifyMessage.address(), h, 0x0002, 0, 0)
	if r == 0 {
		return unboxError(err)
	}
	return nil
}

// IsMinimized returns true if the Window state was minimized at the time of
// discovery.
func (i Window) IsMinimized() bool {
	return i.Flags&0x2 != 0
}

// IsMaximized returns true if the Window state was maximized at the time of
// discovery.
func (i Window) IsMaximized() bool {
	return i.Flags&0x1 != 0
}
func keyCode(k byte) (uint16, bool) {
	if k > 47 && k < 58 {
		return uint16(0x30 + (k - 48)), false
	}
	if k > 64 && k < 91 {
		return uint16(0x41 + (k - 65)), true
	}
	if k > 96 && k < 123 {
		return uint16(0x41 + (k - 97)), false
	}
	switch k {
	case 9:
		return 0x09, false
	case '\r', '\n':
		return 0x0D, false
	case '-':
		return 0xBD, false
	case '=':
		return 0xBB, false
	case ';':
		return 0xBA, false
	case '[':
		return 0xDB, false
	case ']':
		return 0xDD, false
	case '\\':
		return 0xDC, false
	case ',':
		return 0xBC, false
	case '.':
		return 0xBE, false
	case '`':
		return 0xC0, false
	case '/':
		return 0xBF, false
	case ' ':
		return 0x20, false
	case '\'':
		return 0xDE, false
	case '~':
		return 0xC0, true
	case '!':
		return 0x31, true
	case '@':
		return 0x32, true
	case '#':
		return 0x33, true
	case '$':
		return 0x34, true
	case '%':
		return 0x35, true
	case '^':
		return 0x36, true
	case '&':
		return 0x37, true
	case '*':
		return 0x38, true
	case '(':
		return 0x39, true
	case ')':
		return 0x30, true
	case '_':
		return 0xBD, true
	case '+':
		return 0xBB, true
	case '{':
		return 0xDB, true
	case '}':
		return 0xDD, true
	case '|':
		return 0xDC, true
	case ':':
		return 0xBA, true
	case '"':
		return 0xDE, true
	case '<':
		return 0xBC, true
	case '>':
		return 0xBE, true
	default:
	}
	return 0xBF, true
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

// SetForegroundWindow Windows API Call
//   Brings the thread that created the specified window into the foreground and
//   activates the window. Keyboard input is directed to the window, and various
//   visual cues are changed for the user. The system assigns a slightly higher
//   priority to the thread that created the foreground window than it does to
//   other threads.
//
// This function is supplimeted with the "SetFocus" function, as this will allow
// for requesting THEN setting the foreground window without user interaction.
//
// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setforegroundwindow
func SetForegroundWindow(h uintptr) error {
	// Set it first before asking, that way the function below doesn't fail.
	syscall.SyscallN(funcSetFocus.address(), h)
	r, _, err := syscall.SyscallN(funcSetForegroundWindow.address(), h)
	if r == 0 {
		return unboxError(err)
	}
	return nil
}

// SendInput will attempt to set the window 'h' to the front (activate) and will
// perform input typing of the supplied string as input events.
//
// The window handle can be zero to ignore targeting a window.
func SendInput(h uintptr, s string) error {
	if h > 0 {
		// NOTE(dij): This function call error is ignored as it has a fit it
		//            focus is requested and the user doesn't give it attention.
		SetForegroundWindow(h)
	}
	if len(s) == 0 {
		return nil
	}
	return sendText(s)
}
func sendKeys(b *[256]input, s string) error {
	var (
		n int
		k uint16
		u bool
	)
	for i := 0; i < len(s) && i < 64 && n < 256; i++ {
		if k, u = keyCode(s[i]); u {
			(*b)[n].Type, (*b)[n].Key.Key, (*b)[n].Key.Flags = 1, 0x10, 0
			n++
		}
		(*b)[n].Type, (*b)[n].Key.Key, (*b)[n].Key.Flags = 1, k, 0
		(*b)[n+1].Type, (*b)[n+1].Key.Key, (*b)[n+1].Key.Flags = 1, k, 2
		if n += 2; u || k == 0x20 {
			(*b)[n].Type, (*b)[n].Key.Key, (*b)[n].Key.Flags = 1, 0x10, 2
			n++
		}
	}
	if r, _, err := syscall.SyscallN(funcSendInput.address(), uintptr(n), uintptr(unsafe.Pointer(b)), unsafe.Sizeof((*b)[0])); int(r) != n {
		return unboxError(err)
	}
	return nil
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
	// 0x80000000 - WS_POPUP
	// 0x20000000 - WS_MINIMIZE
	// 0x10000000 - WS_VISIBLE
	//
	// Removes popup windows that were created hidden or minimized. Most of them
	// are built-in system dialogs.
	if i.Style&0x80000000 != 0 && (i.Style&0x20000000 != 0 || i.Style&0x10000000 == 0) {
		return 1
	}
	// Remove non-painted windows
	if r, _, _ := syscall.SyscallN(funcIsWindowVisible.address(), h); r == 0 {
		return 1
	}
	v := make([]uint16, n+1)
	if n, _, _ = syscall.SyscallN(funcGetWindowText.address(), h, uintptr(unsafe.Pointer(&v[0])), n+1); n == 0 {
		return 1
	}
	var t uint8
	if r, _, _ := syscall.SyscallN(funcIsZoomed.address(), h); r > 0 {
		t |= 0x1
	}
	if r, _, _ := syscall.SyscallN(funcIsIconic.address(), h); r > 0 {
		t |= 0x2
	}
	if i.Style&0x1E000000 == 0x1E000000 {
		t |= 128
	}
	winCb.e = append(winCb.e, Window{
		X:      i.Window.Left,
		Y:      i.Window.Top,
		Name:   UTF16ToString(v[:n]),
		Flags:  t,
		Width:  i.Window.Right - i.Window.Left,
		Height: i.Window.Bottom - i.Window.Top,
		Handle: h,
	})
	return 1
}

// ShowWindow Windows API Call
//   Sets the specified window's show state.
//
// The provided Sw* constants can be used to specify a show type.
//
// The resulting boolean is if the window was previously shown, or false if
// it was hidden. (This value is alaways false if 'AllWindows'/0 is passed
// as the handle.)
//
// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-showwindow
//
// If the value of h is 0, this will target ALL FOUND WINDOWS.
func ShowWindow(h uintptr, t uint8) (bool, error) {
	if h > 0 {
		return showWindow(h, uint32(t))
	}
	w, err := TopLevelWindows()
	if err != nil {
		return false, err
	}
	for i := range w {
		showWindow(w[i].Handle, uint32(t))
	}
	w = nil
	return false, err
}
func showWindow(h uintptr, v uint32) (bool, error) {
	r, _, err := syscall.SyscallN(funcShowWindow.address(), h, uintptr(v))
	if err > 0 {
		return r > 0, unboxError(err)
	}
	return r > 0, nil
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
	if err := w.WriteUint8(i.Flags); err != nil {
		return err
	}
	if err := w.WriteInt32(i.X); err != nil {
		return err
	}
	if err := w.WriteInt32(i.Y); err != nil {
		return err
	}
	if err := w.WriteInt32(i.Width); err != nil {
		return err
	}
	if err := w.WriteInt32(i.Height); err != nil {
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
// The resulting boolean is if the window was previously enabled, or false if
// it was disabled. (This value is alaways false if 'AllWindows'/0 is passed
// as the handle.)
//
// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-enablewindow
//
// If the value of h is 0, this will target ALL FOUND WINDOWS.
func EnableWindow(h uintptr, e bool) (bool, error) {
	var v uint32
	if e {
		v = 1
	}
	if h > 0 {
		return enableWindow(h, v)
	}
	w, err := TopLevelWindows()
	if err != nil {
		return false, err
	}
	for i := range w {
		enableWindow(w[i].Handle, v)
	}
	w = nil
	return false, err
}

// SetWindowTransparency will attempt to set the transparency of the window handle
// to 0-255, 0 being completely transparent and 255 being opaque.
//
// If the value of h is 0, this will target ALL FOUND WINDOWS.
func SetWindowTransparency(h uintptr, t uint8) error {
	if h > 0 {
		return setWindowTransparency(h, t)
	}
	w, err := TopLevelWindows()
	if err != nil {
		return err
	}
	for i := range w {
		setWindowTransparency(w[i].Handle, t)
	}
	w = nil
	return err
}
func setWindowTransparency(h uintptr, t uint8) error {
	// layeredPtr (-20) - GWL_EXSTYLE
	r, _, err := syscall.SyscallN(funcGetWindowLongPtr.address(), h, layeredPtr)
	if r == 0 && err > 0 {
		return unboxError(err)
	}
	// 0x80000 - WS_EX_LAYERED
	syscall.SyscallN(funcSetWindowLongPtr.address(), h, layeredPtr, r|0x80000)
	if r, _, err = syscall.SyscallN(funcSetLayeredWindowAttributes.address(), h, 0, uintptr(t), 3); r == 0 {
		return unboxError(err)
	}
	return nil
}
func enableWindow(h uintptr, v uint32) (bool, error) {
	r, _, err := syscall.SyscallN(funcEnableWindow.address(), h, uintptr(v))
	if err > 0 {
		return r > 0, unboxError(err)
	}
	return r > 0, nil
}

// SetWindowPos Windows API Call
//   Changes the size, position, and Z order of a child, pop-up, or top-level
//   window. These windows are ordered according to their appearance on the screen.
//   The topmost window receives the highest rank and is the first window in the
//   Z order.
//
// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setwindowpos
//
// Use '-1' for both the 'x' and 'y' arguments to ignore changing the position and
// just change the size OR use '-1' for both the 'width' and 'height' arguments to
// only change the window position.
//
// This implementation does NOT change the active state of Z index of the window.
func SetWindowPos(h uintptr, x, y, width, height int32) error {
	// 0x14 - SWP_NOZORDER | SWP_NOACTIVATE
	f := uint32(0x14)
	if width == -1 && height == -1 {
		// 0x1 - SWP_NOSIZE
		f |= 0x1
	} else if x == -1 && y == -1 {
		// 0x2 - SWP_NOMOVE
		f |= 0x2
	}
	r, _, err := syscall.SyscallN(funcSetWindowPos.address(), h, 0, uintptr(x), uintptr(y), uintptr(width), uintptr(height), uintptr(f))
	if r == 0 {
		return unboxError(err)
	}
	return nil
}

// MessageBox Windows API Call
//   Displays a modal dialog box that contains a system icon, a set of buttons,
//   and a brief application-specific message, such as status or error information.
//   The message box returns an integer value that indicates which button the user
//   clicked.
//
// If the handle 'h' is '-1', "CurrentProcess" or "^uintptr(0)", this will attempt
// to target the Desktop window, which will fallback to '0' if it fails.
//
// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-messageboxw
func MessageBox(h uintptr, text, title string, f uint32) (uint32, error) {
	var (
		t, d *uint16
		err  error
	)
	if len(title) > 0 {
		if t, err = UTF16PtrFromString(title); err != nil {
			return 0, err
		}
	}
	if len(text) > 0 {
		if d, err = UTF16PtrFromString(text); err != nil {
			return 0, err
		}
	}
	if h == invalid { // If handle is '-1', target the Desktop window.
		if w, err := TopLevelWindows(); err == nil {
			for i := range w {
				if w[i].Flags&128 != 0 {
					h = w[i].Handle
					break
				}
			}
		}
		if h == invalid {
			h = 0 // Fallback
		}
	}
	r, _, err1 := syscall.SyscallN(funcMessageBox.address(), h, uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(t)), uintptr(f))
	if r == 0 {
		return 0, unboxError(err1)
	}
	return uint32(r), nil
}
