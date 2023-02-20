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
	New: func() interface{} {
		var b [bufSize]byte
		return &b
	},
}

// Flush allows Chunk to support the io.Flusher interface.
func (*Chunk) Flush() error {
	return nil
}

// Close allows Chunk to support the io.Closer interface.
func (*Chunk) Close() error {
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
	//if c.Empty() {
	//	return 0
	//}
	if r := c.Limit - c.Size(); r > 0 {
		return r
	}
	return 0
}

// String returns a string representation of this Chunk's buffer.
func (c *Chunk) String() string {
	if c.Empty() {
		return "<nil>"
	}
	_ = c.buf[c.rpos]
	return string(c.buf[c.rpos:])
}

// Remaining returns the number of bytes left to be read in this Chunk. This is
// the length 'Size' minus the read cursor.
func (c *Chunk) Remaining() int {
	if c.Empty() {
		return 0
	}
	return c.Size() - c.rpos
}

// Payload returns a copy of the underlying UNREAD buffer contained in this
// Chunk.
//
// This may be empty depending on the read status of this chunk. To retrieve the
// full buffer, use the 'Seek' function to set the read cursor to zero.
func (c *Chunk) Payload() []byte {
	if c.Empty() {
		return nil
	}
	_ = c.buf[c.rpos]
	return c.buf[c.rpos:]
}

// Available returns if a limit will block the writing of n bytes. This function
// can be used to check if there is space to write before committing a write.
func (c *Chunk) Available(n int) bool {
	return c.Limit <= 0 || c.Limit-c.Size() > n
}

// Read reads the next len(p) bytes from the Chunk or until the Chunk is
// drained. The return value n is the number of bytes read and any errors that
// may have occurred.
func (c *Chunk) Read(b []byte) (int, error) {
	if c.Empty() && c.buf != nil {
		if c.Reset(); len(b) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n := copy(b, c.buf[c.rpos:])
	c.rpos += n
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
	m, err := c.quickSlice(len(b))
	if err != nil {
		return 0, err
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
	if bugtrack.Enabled {
		if _, ok := w.(*Chunk); ok {
			bugtrack.Track("data.(*Chunk).MarshalStream(): MarshalStream was called from a Chunk to a Chunk!")
		}
	}
	return w.WriteBytes(c.buf[c.rpos:])
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
	for v, s, e := 0, c.rpos, c.rpos+bufSize; n < c.Size() && err == nil; {
		if e > c.Size() {
			e = c.Size()
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
	c.rpos += n
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
		o += int64(c.rpos)
	case io.SeekEnd:
		o += int64(c.Size())
	default:
		return 0, xerr.Sub("invalid whence", 0x27)
	}
	if o < 0 || int(o) > c.Size() {
		return 0, ErrInvalidIndex
	}
	c.rpos = int(o)
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
			bugtrack.Track("data.(*Chunk).ReadFrom(): n=%d, t=%d, len(b)=%d, err=%s, err2=%s", n, t, len(*b), err, err2)
		}
		if n == 0 || err != nil || err2 != nil || (c.Limit > 0 && n >= c.Limit) {
			if err == io.EOF || err == ErrLimit {
				err = nil
			}
			break
		}
	}
	if bufs.Put(b); bugtrack.Enabled {
		bugtrack.Track("data.(*Chunk).ReadFrom(): return t=%d, err=%s, err2=%s", t, err, err2)
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
		bugtrack.Track("data.(*Chunk).ReadDeadline(): start, d=%s", d.String())
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
			bugtrack.Track("data.(*Chunk).ReadDeadline(): n=%d, t=%d, len(b)=%d, err=%s, err2=%s", n, t, len(*b), err, err2)
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
		bugtrack.Track("data.(*Chunk).ReadDeadline(): return t=%d, err=%s", t, err)
	}
	return t, err
}
