//go:build !windows || !heap
// +build !windows !heap

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

import "github.com/iDigitalFlame/xmt/util/bugtrack"

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
	buf  []byte
	rpos int

	Limit int
}

// Reset resets the Chunk buffer to be empty but retains the underlying storage
// for use by future writes.
func (c *Chunk) Reset() {
	c.rpos, c.buf = 0, c.buf[:0]
}

// Clear is similar to Reset, but discards the buffer, which must be allocated
// again. If using the buffer the 'Reset' function is preferable.
func (c *Chunk) Clear() {
	c.rpos, c.buf = 0, nil
}

// Size returns the internal size of the backing buffer, similar to len(b).
func (c *Chunk) Size() int {
	return len(c.buf)
}

// Empty returns true if this Chunk's buffer is empty or has been drained by
// reads.
func (c *Chunk) Empty() bool {
	return c == nil || len(c.buf) <= c.rpos
}

// NewChunk creates a new Chunk struct and will use the provided byte array as
// the underlying backing buffer.
func NewChunk(b []byte) *Chunk {
	return &Chunk{buf: b}
}

// Grow grows the Chunk's buffer capacity, if necessary, to guarantee space for
// another n bytes.
func (c *Chunk) Grow(n int) error {
	if n <= 0 {
		return ErrInvalidIndex
	}
	m, err := c.grow(n)
	if err != nil {
		return err
	}
	c.buf = c.buf[:m]
	return nil
}

// Truncate discards all but the first n unread bytes from the Chunk but
// continues to use the same allocated storage.
//
// This will return an error if n is negative or greater than the length of the
// buffer.
func (c *Chunk) Truncate(n int) error {
	if n == 0 {
		c.Reset()
		return nil
	}
	if n < 0 || n > len(c.buf)-c.rpos {
		return ErrInvalidIndex
	}
	c.buf = c.buf[:c.rpos+n]
	return nil
}
func (c *Chunk) checkBounds(n int) bool {
	return c.rpos+n > len(c.buf)
}
func (c *Chunk) grow(n int) (int, error) {
	x := len(c.buf) - c.rpos
	if x == 0 && c.rpos != 0 {
		c.rpos, c.buf = 0, c.buf[:0]
	}
	if c.Limit > 0 {
		if x >= c.Limit {
			return 0, ErrLimit
		}
		if n > c.Limit {
			n = c.Limit
		}
	}
	if i, ok := c.reslice(n); ok {
		return i, nil
	}
	if c.buf == nil && n <= 64 {
		c.buf = make([]byte, n, 64)
		return 0, nil
	}
	switch m := cap(c.buf); {
	case n <= m/2-x:
		// From the Golang source:
		//	We can slide things down instead of allocating a new
		//	slice. We only need m+n <= c to slide, but
		//	we instead let capacity get twice as large so we
		//	don't spend all our time copying.
		copy(c.buf, c.buf[c.rpos:])
	case c.Limit > 0 && (m > c.Limit+n || x+n > c.Limit):
		return 0, ErrLimit
	case m > max-m-n:
		return 0, ErrTooLarge
	default:
		b, err := trySlice(c.buf[c.rpos:], c.rpos+n)
		if err != nil {
			return 0, err
		}
		c.buf = nil // Reset and set.
		c.buf = b
	}
	c.rpos, c.buf = 0, c.buf[:x+n]
	return x, nil
}
func (c *Chunk) reslice(n int) (int, bool) {
	if l := len(c.buf); n <= cap(c.buf)-l {
		if c.Limit > 0 {
			if l >= c.Limit {
				return 0, false
			}
			if l+n >= c.Limit {
				n = c.Limit - l
			}
		}
		c.buf = c.buf[:l+n]
		return l, true
	}
	return 0, false
}
func (c *Chunk) quickSlice(n int) (int, error) {
	m, ok := c.reslice(n)
	if ok {
		return m, nil
	}
	return c.grow(n)
}

// UnmarshalStream reads the Chunk data from a binary data representation. This
// function will return an error if any part of the read fails.
func (c *Chunk) UnmarshalStream(r Reader) error {
	if bugtrack.Enabled {
		if _, ok := r.(*Chunk); ok {
			bugtrack.Track("data.(*Chunk).UnmarshalStream(): UnmarshalStream was called from a Chunk to a Chunk!")
		}
	}
	c.buf = nil
	err := r.ReadBytes(&c.buf)
	c.rpos = 0
	return err
}
func trySlice(b []byte, n int) (x []byte, err error) {
	if n > MaxSlice {
		return nil, ErrTooLarge
	}
	defer func() {
		if recover() != nil {
			err = ErrTooLarge
		}
	}()
	c := len(b) + n
	if c < 2*cap(b) {
		c = 2 * cap(b)
	}
	x = append([]byte(nil), make([]byte, c)...)
	copy(x, b)
	return x[:len(b)], nil
}
