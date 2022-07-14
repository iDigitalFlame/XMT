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
	"image"
	"image/png"
	"io"
	"runtime"
	"runtime/debug"
	"sync"
	"syscall"
	"unsafe"

	"github.com/iDigitalFlame/xmt/data"
)

var screenFunctions struct {
	sync.Once
	c, b uintptr
}

type rect struct {
	Left, Top, Right, Bottom int32
}
type point struct {
	X, Y int32
}
type rgbQuad struct {
	Blue  byte
	Green byte
	Red   byte
	_     byte
}
type devMode struct {
	_        [68]byte
	Size     uint16
	_        [6]byte
	Position point
	_        [86]byte
	Width    uint32
	Height   uint32
	_        [40]byte
}
type boundsInfo struct {
	Index uint32
	Rect  rect
	Count uint32
}
type bitmapInfo struct {
	Header bitmapInfoHeader
	Colors *rgbQuad
}
type monitorInfo struct {
	Size    uint32
	Monitor rect
	Work    rect
	Flags   uint32
}
type monitorInfoEx struct {
	monitorInfo
	Name [32]uint16
}
type bitmapInfoHeader struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

func initCallbacks() {
	screenFunctions.c = syscall.NewCallback(monitorCountCallback)
	screenFunctions.b = syscall.NewCallback(monitorBoundsCallback)
}
func releaseDC(w, h uintptr) error {
	r, _, err := syscall.SyscallN(funcReleaseDC.address(), w, h)
	if r == 0 {
		return unboxError(err)
	}
	return nil
}

// ActiveDisplays returns the count of current active displays enabled on the
// device.
//
// This function returns an error if any error occurs when retriving the display
// count.
func ActiveDisplays() (uint32, error) {
	screenFunctions.Do(initCallbacks)
	var (
		c   uint32
		err = enumDisplayMonitors(0, nil, screenFunctions.c, uintptr(unsafe.Pointer(&c)))
	)
	return c, err
}
func getDC(w uintptr) (uintptr, error) {
	r, _, err := syscall.SyscallN(funcGetDC.address(), w)
	if r == 0 {
		return 0, unboxError(err)
	}
	return r, nil
}
func getDesktopWindow() (uintptr, error) {
	r, _, err := syscall.SyscallN(funcGetDesktopWindow.address())
	if r == 0 {
		return 0, unboxError(err)
	}
	return r, nil
}
func getMonitorRealSize(h uintptr) *rect {
	i := monitorInfoEx{}
	i.Size = uint32(unsafe.Sizeof(i))
	if err := getMonitorInfo(h, &i); err != nil {
		return nil
	}
	d := devMode{}
	d.Size = uint16(unsafe.Sizeof(d))
	if err := enumDisplaySettings(i.Name, true, &d); err != nil {
		return nil
	}
	return &rect{
		Left:   d.Position.X,
		Right:  d.Position.X + int32(d.Width),
		Top:    d.Position.Y,
		Bottom: d.Position.Y + int32(d.Height),
	}
}
func deleteDC(h uintptr) (uintptr, error) {
	r, _, err := syscall.SyscallN(funcDeleteDC.address(), h)
	if r == 0 {
		return 0, unboxError(err)
	}
	return r, nil
}
func deleteObject(h uintptr) (uintptr, error) {
	r, _, err := syscall.SyscallN(funcDeleteObject.address(), h)
	if r == 0 {
		return 0, unboxError(err)
	}
	return r, nil
}
func selectObject(h, sel uintptr) (uintptr, error) {
	r, _, err := syscall.SyscallN(funcSelectObject.address(), h, sel)
	if r == 0 {
		return 0, unboxError(err)
	}
	return r, nil
}
func createCompatibleDC(m uintptr) (uintptr, error) {
	r, _, err := syscall.SyscallN(funcCreateCompatibleDC.address(), m)
	if r == 0 {
		return 0, unboxError(err)
	}
	return r, nil
}

// DisplayBounds returns the bounds of the supplied display index.
//
// This function will return the bounds of the first monitor if the index is out
// of bounds of the current display count.
func DisplayBounds(i uint32) (image.Rectangle, error) {
	screenFunctions.Do(initCallbacks)
	v := boundsInfo{Index: i}
	enumDisplayMonitors(0, nil, screenFunctions.b, uintptr(unsafe.Pointer(&v)))
	return image.Rect(int(v.Rect.Left), int(v.Rect.Top), int(v.Rect.Right), int(v.Rect.Bottom)), nil
}
func getMonitorInfo(h uintptr, m *monitorInfoEx) error {
	r, _, err := syscall.SyscallN(funcGetMonitorInfo.address(), h, uintptr(unsafe.Pointer(m)))
	if r == 0 {
		return unboxError(err)
	}
	return nil
}

// ScreenShot attempts to take a PNG-encoded screenshot of the current deminsions
// specified into the supplied io.Writer.
//
// This function will return an error if any of the API calls or encoding the
// image fails.
func ScreenShot(x, y, width, height uint32, w io.Writer) error {
	p, err := heapCreate(uint64(((int64(width)*32 + 31) / 32) * 4 * int64(height)))
	if err != nil {
		return err
	}
	v, err := getDesktopWindow()
	if err != nil {
		return err
	}
	m, err := getDC(v)
	if err != nil {
		return err
	}
	d, err := createCompatibleDC(m)
	if err != nil {
		releaseDC(v, m)
		return err
	}
	var b uintptr
	if b, err = createCompatibleBitmap(m, width, height); err == nil {
		h := bitmapInfoHeader{
			Width:       int32(width),
			Planes:      1,
			Height:      -int32(height),
			BitCount:    32,
			SizeImage:   0,
			Compression: 0,
		}
		h.Size = uint32(unsafe.Sizeof(h))
		var l, o uintptr
		if l, err = heapAlloc(p, uint64(((int64(width)*32+31)/32)*4*int64(height)), false); err == nil {
			if o, err = selectObject(d, b); err == nil {
				if err = bitBlt(d, 0, 0, width, height, m, x, y, 0xCC0020); err == nil {
					if _, err = getDIBits(m, b, 0, height, (*uint8)(unsafe.Pointer(l)), (*bitmapInfo)(unsafe.Pointer(&h)), 0); err == nil {
						var r *image.RGBA
						if r, err = tryCreateImage(image.Rect(0, 0, int(width), int(height))); err == nil {
							_ = r.Pix[len(r.Pix)-1]
							for z, i, k := l, 0, uint32(0); k < height && i < len(r.Pix); k++ {
								for j := uint32(0); j < width; j++ {
									r.Pix[i], r.Pix[i+1], r.Pix[i+2], r.Pix[i+3] = *(*uint8)(unsafe.Pointer(z + 2)), *(*uint8)(unsafe.Pointer(z + 1)), *(*uint8)(unsafe.Pointer(z)), 0xFF
									i += 4
									z += 4
								}
							}
							err = png.Encode(w, r)
							for i := range r.Pix {
								r.Pix[i] = 0 // Should be a memclr.
							}
							r.Pix, r = nil, nil
						}
					}
				}
				selectObject(d, o)
			}
			heapFree(p, l)
		}
		deleteObject(b)
	}
	deleteDC(d)
	releaseDC(v, m)
	heapDestroy(p)
	CloseHandle(p)
	runtime.GC()
	debug.FreeOSMemory()
	return err
}
func tryCreateImage(r image.Rectangle) (i *image.RGBA, err error) {
	defer func() {
		if recover() != nil {
			err = data.ErrTooLarge
		}
	}()
	return image.NewRGBA(r), nil
}
func monitorCountCallback(_, _ uintptr, _ *rect, d uintptr) uintptr {
	n := (*uint32)(unsafe.Pointer(d))
	*n = *n + 1
	return 1
}
func monitorBoundsCallback(h, _ uintptr, p *rect, d uintptr) uintptr {
	v := (*boundsInfo)(unsafe.Pointer(d))
	if v.Count != v.Index {
		v.Count = v.Count + 1
		return 1
	}
	if r := getMonitorRealSize(h); r != nil {
		v.Rect = *r
	} else {
		v.Rect = *p
	}
	return 0
}
func createCompatibleBitmap(m uintptr, x, y uint32) (uintptr, error) {
	r, _, err := syscall.SyscallN(funcCreateCompatibleBitmap.address(), m, uintptr(x), uintptr(y))
	if r == 0 {
		return 0, unboxError(err)
	}
	return r, nil
}
func enumDisplaySettings(n [32]uint16, current bool, d *devMode) error {
	var m uint32
	if current {
		m = 0xFFFFFFFF
	}
	r, _, err := syscall.SyscallN(
		funcEnumDisplaySettings.address(), uintptr(unsafe.Pointer(&n[0])), uintptr(m), uintptr(unsafe.Pointer(d)),
	)
	if r == 0 {
		return unboxError(err)
	}
	return nil
}
func enumDisplayMonitors(h uintptr, p *rect, f uintptr, d uintptr) error {
	r, _, err := syscall.SyscallN(funcEnumDisplayMonitors.address(), h, uintptr(unsafe.Pointer(p)), f, d)
	if r == 0 {
		return unboxError(err)
	}
	return nil
}
func bitBlt(h uintptr, x, y, w, g uint32, s uintptr, x1, y1, f uint32) error {
	r, _, err := syscall.SyscallN(
		funcBitBlt.address(), h, uintptr(x), uintptr(y), uintptr(w), uintptr(g), s, uintptr(x1), uintptr(y1), uintptr(f),
	)
	if r == 0 {
		return unboxError(err)
	}
	return nil
}
func getDIBits(h, b uintptr, s, l uint32, m *uint8, i *bitmapInfo, f uint32) (uint32, error) {
	r, _, err := syscall.SyscallN(
		funcGetDIBits.address(), h, b, uintptr(s), uintptr(l), uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(i)),
		uintptr(f),
	)
	if r == 0 {
		return 0, unboxError(err)
	}
	return uint32(r), nil
}
