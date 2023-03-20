//go:build windows && heap
// +build windows,heap

// Copyright (C) 2020 - 2023 iDigitalFlame
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

package data

import (
	"io"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/iDigitalFlame/xmt/util/bugtrack"
)

var heapBase uintptr
var heapBaseInit sync.Once

// Chunk is a low level data container. Chunks allow for simple read/write
// operations on static containers.
//
// Chunk fulfils the Reader, Seeker, Writer, Flusher and Closer interfaces.
// Seeking on Chunks is only supported in a read-only fashion.
//
// If the underlying device is running Windows and the "heap" build tag is used,
// Chunks will be created on the Process Heap not managed by Go. This prevents
// over-allocation and relieves pressure on the GC. By default, this is off.
type Chunk struct {
	h          uintptr
	buf        []byte
	rpos, wpos int

	Limit int
	id    uint32
}
type header struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

// Reset resets the Chunk buffer to be empty but retains the underlying storage
// for use by future writes.
func (c *Chunk) Reset() {
	if atomic.LoadUintptr(&c.h) == 0 {
		return
	}
	// NOTE(dij): Not sure if this is needed, this would allow us to zero out
	//            the Chunk data.
	/*if c.wpos == 0 {
		for i := 0; i < c.rpos; i++ {
			c.buf[i] = 0
		}
	} else {
		for i := 0; i < c.wpos; i++ {
			c.buf[i] = 0
		}
	}*/
	c.rpos, c.wpos = 0, 0
}
func heapBaseInitFunc() {
	var err error
	if heapBase, err = heapCreate(0); err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("data.heapBaseInitFunc(): Failed with error: %s!", err)
		}
		panic(err)
	}
	if bugtrack.Enabled {
		bugtrack.Track("data.heapBaseInitFunc(): Created heapBase at 0x%X!", heapBase)
	}
}

// Clear is similar to Reset, but discards the buffer, which must be allocated
// again. If using the buffer the 'Reset' function is preferable.
func (c *Chunk) Clear() {
	if atomic.LoadUintptr(&c.h) == 0 {
		return
	}
	if err := heapFree(heapBase, c.h); err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("data.(*Chunk).Close(): Failed to free Chunk 0x%X: %s!", c.h, err)
		}
		panic(err)
	}
	if atomic.StoreUintptr(&c.h, 0); bugtrack.Enabled {
		bugtrack.Track("data.(*Chunk).Close(): Freed Chunk 0x%X.", c.h)
	}
	c.rpos, c.wpos, c.buf = 0, 0, nil
}

// Size returns the internal size of the backing buffer, similar to len(b).
func (c *Chunk) Size() int {
	return c.wpos
}

// Empty returns true if this Chunk's buffer is empty or has been drained by
// reads.
func (c *Chunk) Empty() bool {
	return c == nil || atomic.LoadUintptr(&c.h) == 0 || c.wpos <= c.rpos
}

// NewChunk creates a new Chunk struct and will use the provided byte array as
// the underlying backing buffer.
func NewChunk(b []byte) *Chunk {
	var c Chunk
	if _, err := c.Write(b); err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("data.NewChunk(): Creating a new []byte-based Chunk failed: %s!", err)
		}
		panic(err)
	}
	return &c
}

//go:linkname heapFree github.com/iDigitalFlame/xmt/device/winapi.heapFree
func heapFree(h, m uintptr) error
func (c *Chunk) grow(n int) error {
	if heapBaseInit.Do(heapBaseInitFunc); atomic.LoadUintptr(&c.h) == 0 {
		v := uint64(n) * 2
		if v < 4096 {
			// NOTE(dij): 4096 is chosen as it's a good middleground to prevent
			//            a lot of reallocations.
			v = 4096
		}
		var r bool
		if c.Limit > 0 && int(v) > c.Limit {
			v, r = uint64(c.Limit), true
		}
		h, err := heapAlloc(heapBase, v, true)
		if err != nil {
			if bugtrack.Enabled {
				bugtrack.Track("data.(*Chunk).grow(): Creating a new Chunk of size %d failed: %s!", v, err)
			}
			return err
		}
		if atomic.StoreUintptr(&c.h, h); bugtrack.Enabled {
			bugtrack.Track("data.(*Chunk).grow(): Created a new Chunk 0x%X with size %d.", c.h, v)
		}
		c.buf, c.wpos, c.rpos = *(*[]byte)(unsafe.Pointer(&header{Data: unsafe.Pointer(h), Len: int(v), Cap: int(v)})), 0, 0
		if r {
			return ErrLimit
		}
		return nil
	}
	if c.wpos+n < len(c.buf) {
		return nil
	}
	var (
		v = uint64(len(c.buf)+n) * 2
		r bool
	)
	if c.Limit > 0 && int(v) > c.Limit {
		v, r = uint64(c.Limit), true
	}
	h, err := heapReAlloc(heapBase, c.h, v, true)
	if err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("data.(*Chunk).grow(): Reallocating a Chunk 0x%X with size %d failed: %s!", c.h, v, err)
		}
		return err
	}
	if bugtrack.Enabled {
		bugtrack.Track("data.(*Chunk).grow(): Reallocated Chunk 0x%X with a new size of %d.", c.h, v)
	}
	c.buf = nil
	c.buf = *(*[]byte)(unsafe.Pointer(&header{Data: unsafe.Pointer(h), Len: int(v), Cap: int(v)}))
	if atomic.StoreUintptr(&c.h, h); r {
		return ErrLimit
	}
	return nil
}

// Grow grows the Chunk's buffer capacity, if necessary, to guarantee space for
// another n bytes.
func (c *Chunk) Grow(n int) error {
	if n <= 0 {
		return ErrInvalidIndex
	}
	if err := c.grow(n); err != nil {
		return err
	}
	return nil
}

// Truncate discards all but the first n unread bytes from the Chunk but
// continues to use the same allocated storage.
//
// This will return an error if n is negative or greater than the length of the
// buffer.
func (c *Chunk) Truncate(n int) error {
	if n == 0 {
		if c.Empty() {
			return nil
		}
		c.Reset()
		return nil
	}
	if c.Empty() || n < 0 || n > len(c.buf)-c.rpos {
		return ErrInvalidIndex
	}
	for i := c.rpos + n; i < len(c.buf); i++ {
		c.buf[i] = 0
	}
	return nil
}
func (c *Chunk) checkBounds(n int) bool {
	return c.rpos+n > len(c.buf) || c.rpos+n > c.wpos
}

//go:linkname heapCreate github.com/iDigitalFlame/xmt/device/winapi.heapCreate
func heapCreate(n uint64) (uintptr, error)
func (c *Chunk) reslice(n int) (int, bool) {
	if c.Empty() {
		return 0, false
	}
	if c.wpos+n < len(c.buf) {
		return c.wpos, true
	}
	return 0, false
}
func (c *Chunk) quickSlice(n int) (int, error) {
	if x, ok := c.reslice(n); ok {
		c.wpos += n
		return x, nil
	}
	err, v := c.grow(n), c.wpos
	c.wpos += n
	return v, err
}

// UnmarshalStream reads the Chunk data from a binary data representation. This
// function will return an error if any part of the read fails.
func (c *Chunk) UnmarshalStream(r Reader) error {
	// NOTE(dij): We have to re-write this here as we don't want the runtime
	// to directly allocate our new buffer and instead we want to read it directly
	// into the heap.
	if bugtrack.Enabled {
		if _, ok := r.(*Chunk); ok {
			bugtrack.Track("data.(*Chunk).UnmarshalStream(): UnmarshalStream was called from a Chunk to a Chunk!")
		}
	}
	if !c.Empty() {
		c.Reset()
	}
	t, err := r.Uint8()
	if err != nil {
		return err
	}
	var l uint64
	switch t {
	case 0:
		return nil
	case 1, 2:
		n, err2 := r.Uint8()
		if err2 != nil {
			return err2
		}
		l = uint64(n)
	case 3, 4:
		n, err2 := r.Uint16()
		if err2 != nil {
			return err2
		}
		l = uint64(n)
	case 5, 6:
		n, err2 := r.Uint32()
		if err2 != nil {
			return err2
		}
		l = uint64(n)
	case 7, 8:
		n, err2 := r.Uint64()
		if err2 != nil {
			return err2
		}
		l = n
	default:
		return ErrInvalidType
	}
	if l == 0 {
		return io.ErrUnexpectedEOF
	}
	if l > MaxSlice {
		return ErrTooLarge
	}
	if err = c.grow(int(l)); err != nil {
		return err
	}
	x, err := r.Read(c.buf[0:l])
	if c.rpos, c.wpos, c.wpos = 0, 0, x; x != int(l) {
		return io.ErrUnexpectedEOF
	}
	return err
}

//go:linkname heapAlloc github.com/iDigitalFlame/xmt/device/winapi.heapAlloc
func heapAlloc(h uintptr, s uint64, z bool) (uintptr, error)

//go:linkname heapReAlloc github.com/iDigitalFlame/xmt/device/winapi.heapReAlloc
func heapReAlloc(h, m uintptr, s uint64, z bool) (uintptr, error)
