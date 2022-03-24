//go:build windows

package screen

import (
	"image"
	"io"

	"github.com/iDigitalFlame/xmt/device/winapi"
)

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
func Capture(w io.Writer) error {
	c, err := winapi.ActiveDisplays()
	if err != nil {
		return err
	}
	var h, k uint32
	for i := uint32(0); i < c; i++ {
		r, err := winapi.DisplayBounds(i)
		if err != nil {
			return err
		}
		if v := uint32(r.Dy()); v > h {
			h = v
		}
		k += uint32(r.Dx())
	}
	return winapi.ScreenShot(0, 0, k, h, w)
}

// ActiveDisplays returns the count of current active displays enabled on the
// device.
//
// This function returns an error if any error occurs when retriving the display
// count.
//
// TODO(dij): Currently works only on Windows devices.
func ActiveDisplays() (uint32, error) {
	return winapi.ActiveDisplays()
}

// DisplayBounds returns the bounds of the supplied display index.
//
// This function will return the bounds of the first monitor if the index is out
// of bounds of the current display count.
//
// TODO(dij): Currently works only on Windows devices.
func DisplayBounds(i uint32) (image.Rectangle, error) {
	return winapi.DisplayBounds(i)
}

// CaptureRange attempts to take a PNG-encoded screenshot of the current
// deminsions specified into the supplied io.Writer.
//
// This function will return an error getting the screen color info fails or
// encoding the image fails.
//
// TODO(dij): Currently works only on Windows devices.
func CaptureRange(x, y, width, height uint32, w io.Writer) error {
	return winapi.ScreenShot(x, y, width, height, w)
}
