//go:build !windows

package screen

import (
	"image"
	"io"

	"github.com/iDigitalFlame/xmt/device"
)

// TODO(dij): Add non-windows (Linux, MacOS) support.
//            The list for this would be heavy-ish. Screenshot libraries
//            https://github.com/kbinani/screenshot are nice, but are dependency
//            heavy. Also MacOS requires CGO enabled, which makes it harder to
//            cross-compile. Also Linux needs XOrg and doesn't seem to support
//            Wayland.

// Capture attempts to take a PNG-encoded screenshot of all the active
// displays on the device into the supplied io.Writer.
//
// This function will return an error getting the screen color info fails or
// encoding the image fails.
//
// This function calculates the dimensions of all the active displays together
// and calls 'CaptureRange' underneath.
//
// TODO(dij): Currently works only on Windows devices.
func Capture(_ io.Writer) error {
	return device.ErrNoWindows
}

// ActiveDisplays returns the count of current active displays enabled on the
// device.
//
// This function returns an error if any error occurs when retriving the display
// count.
//
// TODO(dij): Currently works only on Windows devices.
func ActiveDisplays() (uint32, error) {
	return 0, nil
}

// DisplayBounds returns the bounds of the supplied display index.
//
// This function will return the bounds of the first monitor if the index is out
// of bounds of the current display count.
//
// TODO(dij): Currently works only on Windows devices.
func DisplayBounds(_ uint32) (image.Rectangle, error) {
	return image.Rectangle{}, nil
}

// CaptureRange attempts to take a PNG-encoded screenshot of the current
// deminsions specified into the supplied io.Writer.
//
// This function will return an error getting the screen color info fails or
// encoding the image fails.
//
// TODO(dij): Currently works only on Windows devices.
func CaptureRange(_, _, _, _ uint32, _ io.Writer) error {
	return device.ErrNoWindows
}
