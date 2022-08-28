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

package data

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const (
	max     = int(^uint(0) >> 1)
	bufSize = 2 << 13
)

var bufs = sync.Pool{
	New: func() any {
		var b [bufSize]byte
		return &b
	},
}

// KeySize is the size of the Key array.
const KeySize = 32

// Chunk is a low level data container. Chunks allow for simple read/write
// operations on static containers.
//
// Chunk fulfils the Reader, Seeker, Writer, Flusher and Closer interfaces.
// Seeking on Chunks is only supported in a read-only fashion.
type Chunk struct {
	buf []byte
	pos int

	Limit int
}

// Key is an alias for an encryption key that can be used to protect Chunks
// when non-empty.
type Key [KeySize]byte

// Reset resets the Chunk buffer to be empty but retains the underlying storage
// for use by future writes.
func (c *Chunk) Reset() {
	c.pos, c.buf = 0, c.buf[:0]
}

// Clear is similar to Reset, but discards the buffer, which must be allocated
// again. If using the buffer the 'Reset' function is preferable.
func (c *Chunk) Clear() {
	c.Reset()
	c.buf = nil
}

// Size returns the internal size of the backing buffer, similar to len(b).
func (c *Chunk) Size() int {
	if c.buf == nil {
		return 0
	}
	return len(c.buf)
}

// Flush allows Chunk to support the io.Flusher interface.
func (Chunk) Flush() error {
	return nil
}

// Close allows Chunk to support the io.Closer interface.
func (Chunk) Close() error {
	return nil
}

// Space returns the amount of bytes available in this Chunk when a Limit is
// set.
//
// This function will return -1 if there is no limit set and returns 0 (zero)
// when a limit is set, but no byte space is available.
func (c *Chunk) Space() int {
	if c.Limit <= 0 {
		return -1
	}
	if r := c.Limit - len(c.buf); r > 0 {
		return r
	}
	return 0
}

// Empty returns true if this Chunk's buffer is empty or has been drained by
// reads.
func (c *Chunk) Empty() bool {
	return len(c.buf) == 0 || len(c.buf) <= c.pos
}

// NewChunk creates a new Chunk struct and will use the provided byte array as
// the underlying backing buffer.
func NewChunk(b []byte) *Chunk {
	return &Chunk{buf: b}
}

// String returns a string representation of this Chunk's buffer.
func (c *Chunk) String() string {
	if c == nil || len(c.buf) == 0 || len(c.buf) <= c.pos {
		return "<nil>"
	}
	_ = c.buf[c.pos]
	return string(c.buf[c.pos:])
}

// Remaining returns the number of bytes left to be read in this Chunk. This is
// the length 'Size' minus the read cursor.
func (c *Chunk) Remaining() int {
	if c.buf == nil {
		return 0
	}
	return len(c.buf) - c.pos
}

// Payload returns a copy of the underlying UNREAD buffer contained in this
// Chunk.
//
// This may be empty depending on the read status of this chunk. To retrieve the
// full buffer, use the 'Seek' function to set the read cursor to zero.
func (c *Chunk) Payload() []byte {
	if len(c.buf) == 0 || len(c.buf) <= c.pos {
		return nil
	}
	_ = c.buf[c.pos]
	return c.buf[c.pos:]
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

// Available returns if a limit will block the writing of n bytes. This function
// can be used to check if there is space to write before committing a write.
func (c *Chunk) Available(n int) bool {
	return c.Limit <= 0 || c.Limit-len(c.buf) > n
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
	if n < 0 || n > len(c.buf) || n+c.pos > len(c.buf) {
		return ErrInvalidIndex
	}
	c.buf = c.buf[c.pos : c.pos+n]
	return nil
}
func (c *Chunk) grow(n int) (int, error) {
	x := len(c.buf) - c.pos
	if x == 0 && c.pos != 0 {
		c.pos, c.buf = 0, c.buf[:0]
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
		if bugtrack.Enabled {
			bugtrack.Track("Well dam, the weird copy-overlap code was triggered!??!! Trace this PTR: %p", c)
		}
		copy(c.buf, c.buf[c.pos:])
	case c.Limit > 0 && m > c.Limit-m-n:
		return 0, ErrLimit
	case m > max-m-n:
		return 0, ErrTooLarge
	default:
		b, err := trySlice(2*m + n)
		if err != nil {
			return 0, err
		}
		copy(b, c.buf[c.pos:])
		c.buf = b
	}
	c.pos, c.buf = 0, c.buf[:x+n]
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
func trySlice(n int) (b []byte, err error) {
	if n > MaxSlice {
		return nil, ErrTooLarge
	}
	defer func() {
		if recover() != nil {
			err = ErrTooLarge
		}
	}()
	return make([]byte, n), nil
}

// Read reads the next len(p) bytes from the Chunk or until the Chunk is
// drained. The return value n is the number of bytes read and any errors that
// may have occurred.
func (c *Chunk) Read(b []byte) (int, error) {
	if len(c.buf) <= c.pos {
		c.Reset()
		return 0, io.EOF
	}
	n := copy(b, c.buf[c.pos:])
	c.pos += n
	return n, nil
}

// Write appends the contents of b to the buffer, growing the buffer as needed.
//
// If the buffer becomes too large, Write will return 'ErrTooLarge.' If there is
// a limit set, this function will return 'ErrLimit' if the Limit is being hit.
//
// If an 'ErrLimit' is returned, check the returned bytes as 'ErrLimit' is
// returned as a warning that not all bytes have been written before refusing
// writes.
func (c *Chunk) Write(b []byte) (int, error) {
	m, ok := c.reslice(len(b))
	if !ok {
		var err error
		if m, err = c.grow(len(b)); err != nil {
			return 0, err
		}
	}
	n := copy(c.buf[m:], b)
	if n < len(b) && c.Limit > 0 && len(c.buf) >= c.Limit {
		return n, ErrLimit
	}
	return n, nil
}

// MarshalStream writes the unread Chunk data into a binary data representation.
// This function will return an error if any part of the write fails.
func (c *Chunk) MarshalStream(w Writer) error {
	if _, ok := w.(*Chunk); ok {
		panic("MarshalStream: Chunk -> Chunk")
	}
	return w.WriteBytes(c.buf[c.pos:])
}

// UnmarshalStream reads the Chunk data from a binary data representation. This
// function will return an error if any part of the read fails.
func (c *Chunk) UnmarshalStream(r Reader) error {
	if _, ok := r.(*Chunk); ok {
		panic("UnmarshalStream: Chunk -> Chunk")
	}
	c.buf = nil
	err := r.ReadBytes(&c.buf)
	c.pos = 0
	return err
}

// WriteTo writes data to the supplied Writer until there's no more data to
// write or when an error occurs.
//
// The return value is the number of bytes written. Any error encountered
// during the write is also returned.
func (c *Chunk) WriteTo(w io.Writer) (int64, error) {
	if c.Empty() {
		return 0, nil
	}
	var (
		n   int
		err error
	)
	for v, s, e := 0, c.pos, c.pos+bufSize; n < len(c.buf) && err == nil; {
		if e > len(c.buf) {
			e = len(c.buf)
		}
		if s == e {
			break
		}
		v, err = w.Write(c.buf[s:e])
		if n += v; err != nil {
			break
		}
		s = e
		e += v
	}
	c.pos += n
	return int64(n), err
}

// Seek will attempt to seek to the provided read offset index and whence. This
// function will return the new offset if successful and will return an error
// if the offset and/or whence are invalid.
//
// NOTE: This only affects read operations.
func (c *Chunk) Seek(o int64, w int) (int64, error) {
	switch w {
	case io.SeekStart:
		if o < 0 {
			return 0, ErrInvalidIndex
		}
	case io.SeekCurrent:
		o += int64(c.pos)
	case io.SeekEnd:
		o += int64(len(c.buf))
	default:
		return 0, xerr.Sub("invalid whence", 0x27)
	}
	if o < 0 || int(o) > len(c.buf) {
		return 0, ErrInvalidIndex
	}
	c.pos = int(o)
	return o, nil
}

// ReadFrom reads data from the supplied Reader until EOF or error.
//
// The return value is the number of bytes read.
// Any error except 'io.EOF' encountered during the read is also returned.
func (c *Chunk) ReadFrom(r io.Reader) (int64, error) {
	var (
		b         = bufs.Get().(*[bufSize]byte)
		t         int64
		n, w      int
		err, err2 error
	)
	for {
		if c.Limit > 0 {
			x := c.Space()
			if x <= 0 {
				break
			}
			if x > bufSize {
				x = bufSize
			}
			n, err = r.Read((*b)[:x])
		} else {
			n, err = r.Read((*b)[:])
		}
		if n > 0 {
			w, err2 = c.Write((*b)[:n])
			if w < n {
				t += int64(w)
			} else {
				t += int64(n)
			}
			if err2 != nil {
				break
			}
		}
		if bugtrack.Enabled {
			bugtrack.Track("data.Chunk.ReadFrom(): n=%d, t=%d, len(b)=%d, err=%s, err2=%s", n, t, len(*b), err, err2)
		}
		if n == 0 || err != nil || err2 != nil || (c.Limit > 0 && n >= c.Limit) {
			if err == io.EOF || err == ErrLimit {
				err = nil
			}
			break
		}
	}
	if bufs.Put(b); bugtrack.Enabled {
		bugtrack.Track("data.Chunk.ReadFrom(): return t=%d, err=%s", t, err)
	}
	return t, err
}

// ReadDeadline reads data from the supplied net.Conn until EOF or error.
//
// The return value is the number of bytes read.
// Any error except 'io.EOF' encountered during the read is also returned.
//
// If the specific duration is greater than zero, the read deadline will be
// applied. Timeout errors will NOT be returned and will instead break a read.
func (c *Chunk) ReadDeadline(r net.Conn, d time.Duration) (int64, error) {
	var (
		b         = bufs.Get().(*[bufSize]byte)
		t         int64
		n, w      int
		err, err2 error
	)
	if bugtrack.Enabled {
		bugtrack.Track("data.Chunk.ReadDeadline(): start, d=%s", d.String())
	}
	for {
		if c.Limit > 0 {
			x := c.Space()
			if x <= 0 {
				break
			}
			if x > bufSize {
				x = bufSize
			}
			n, err = r.Read((*b)[:x])
		} else {
			n, err = r.Read((*b)[:])
		}
		if n > 0 {
			w, err2 = c.Write((*b)[:n])
			if w < n {
				t += int64(w)
			} else {
				t += int64(n)
			}
			if err2 != nil {
				break
			}
		}
		if bugtrack.Enabled {
			bugtrack.Track("data.Chunk.ReadDeadline(): n=%d, t=%d, len(b)=%d, err=%s, err2=%s", n, t, len(*b), err, err2)
		}
		if n == 0 || err != nil || err2 != nil || (c.Limit > 0 && n >= c.Limit) {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				err = nil
			} else if err == io.EOF || err == ErrLimit {
				err = nil
			}
			break
		}
		if d > 0 {
			r.SetReadDeadline(time.Now().Add(d))
		}
	}
	if bufs.Put(b); bugtrack.Enabled {
		bugtrack.Track("data.Chunk.ReadDeadline(): return t=%d, err=%s", t, err)
	}
	return t, err
}
