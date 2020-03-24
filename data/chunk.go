package data

import (
	"errors"
	"fmt"
	"io"
	"sync"
)

const (
	max   = int(^uint(0) >> 1)
	small = 64
	empty = "0xNULL"
)

var (
	// ErrLimit is an error that is returned when a Limit is set on a Chunk and the size limit was hit when
	// attempting to write to the Chunk. This error wraps the io.EOF error, which allows this error to match
	// io.EOF for sanity checking.
	ErrLimit = fmt.Errorf("buffer size limit reached: %w", io.EOF)
	// ErrTooLarge is raised if memory cannot be allocated to store data in a Chunk.
	ErrTooLarge = errors.New("buffer size is too large")
	// ErrInvalidIndex is raised if a specified Grow or index function is supplied with an
	// negative or out of bounds number or when a Seek index is not valid.
	ErrInvalidIndex = errors.New("buffer index provided is not valid")
	// ErrInvalidWhence is returned when the provided seek whence is not a valid whence value.
	ErrInvalidWhence = errors.New("buffer seek whence is invalid")

	bufs = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 512)
		},
	}
)

// Chunk is a low level data container. Chunks allow for simple read/write
// operations on static containers. Chunk fulfils the Reader, Seeker, Writer, Flusher
// and Closer interfaces.
type Chunk struct {
	Limit int

	buf        []byte
	rpos, wpos int
}

// Reset resets the Chunk buffer to be empty but retains the underlying storage for use
// by future writes.
func (c *Chunk) Reset() {
	c.wpos = 0
	c.rpos = 0
	c.buf = c.buf[:0]
}

// Clear is similar to Reset, but discards the buffer, which must be allocated again. If using
// the buffer the Reset function is preferable.
func (c *Chunk) Clear() {
	c.Reset()
	c.buf = nil
}

// Rewind will seek the writing and reading positions back to zero. This function can be used
// to 'reset' the Chunk without deleting any data.
func (c *Chunk) Rewind() {
	c.rpos, c.wpos = 0, 0
}

// Len returns the same result as Size. This function returns the amount of bytes written or
// contained in this Chunk.
func (c Chunk) Len() int {
	return c.Size()
}

// Left returns the amount of bytes avaliable in this Chunk when a Limit is set. This function will
// return -1 if there us no limit set.
func (c Chunk) Left() int {
	if c.Limit <= 0 {
		return -1
	}
	return c.Limit - c.Size()
}

// Size returns the amount of bytes written or contained in this Chunk.
func (c Chunk) Size() int {
	if c.buf == nil {
		return 0
	}
	return len(c.buf) - c.wpos
}

// Flush does nothing for the Chunk struct. Just here for compatibility.
func (Chunk) Flush() error {
	return nil
}

// Empty returns true if this Chunk's buffer is empty.
func (c Chunk) Empty() bool {
	return len(c.buf) == 0
}

// Close will truncate the writing buffer if this Chunk has been written to.
func (c *Chunk) Close() error {
	if c.wpos > 0 {
		c.buf = c.buf[:c.wpos]
	}
	return nil
}

// String returns a string representation of this Chunk's buffer.
func (c Chunk) String() string {
	if len(c.buf) == 0 {
		return empty
	}
	return string(c.buf[c.wpos:])
}

// NewChunk creates a new Chunk struct and will use the provided byte array as the underlying
// structure if is is not null.
func NewChunk(b []byte) *Chunk {
	c := &Chunk{buf: b}
	if b != nil {
		c.wpos = len(b)
	}
	return c
}

// Payload returns a copy of the underlying buffer contained in this Chunk.
func (c Chunk) Payload() []byte {
	if c.buf == nil {
		return nil
	}
	return c.buf[c.wpos:]
}

// Grow grows the Chunk's buffer capacity, if necessary, to guarantee space for another n bytes.
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

// Avaliable returns if a limit will block the writing of n bytes. This function can
// be used to check if there is space to write before commiting a write.
func (c Chunk) Avaliable(n int) bool {
	return c.Limit <= 0 || c.Size()+n <= c.Limit
}
func (c *Chunk) small(b ...byte) error {
	_, err := c.Write(b)
	return err
}
func (c *Chunk) grow(n int) (int, error) {
	x := len(c.buf) - c.wpos
	if x == 0 && c.wpos != 0 {
		c.Reset()
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
	if c.buf == nil && n <= small {
		c.buf = make([]byte, n, small)
		return 0, nil
	}
	m := cap(c.buf)
	switch {
	case n <= m/2-x:
		copy(c.buf, c.buf[c.wpos:])
	case c.Limit > 0 && m > c.Limit-m-n:
		return 0, ErrLimit
	case m > max-m-n:
		return 0, ErrTooLarge
	default:
		b, err := trySlice(2*m + n)
		if err != nil {
			return 0, err
		}
		copy(b, c.buf[c.wpos:])
		c.buf = b
	}
	c.wpos = 0
	c.buf = c.buf[:x+n]
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
	defer func() {
		if recover() != nil {
			err = ErrTooLarge
		}
	}()
	return make([]byte, n), nil
}

// Read reads the next len(p) bytes from the Chunk or until the Chunk
// is drained. The return value n is the number of bytes read.
func (c *Chunk) Read(b []byte) (int, error) {
	if len(c.buf) <= c.rpos {
		c.Reset()
		return 0, io.EOF
	}
	n := copy(b, c.buf[c.rpos:])
	c.rpos += n
	return n, nil
}

// Write appends the contents of p to the buffer, growing the buffer as needed. If the buffer becomes
// too large, Write will return ErrTooLarge. If there is a limit set, this function will return ErrLimit
// if the Limit is being hit. If an ErrLimit is returned, check the returned bytes as ErrLimit is returned
// as a warning that not all bytes have been written before refusing writes.
func (c *Chunk) Write(b []byte) (int, error) {
	m, ok := c.reslice(len(b))
	if !ok {
		var err error
		if m, err = c.grow(len(b)); err != nil {
			return 0, err
		}
	}
	n := copy(c.buf[m:], b)
	if n < len(b) && c.Limit > 0 && c.Size() >= c.Limit {
		return n, ErrLimit
	}
	return n, nil
}

// MarshalStream writes the Chunk data into a binary data representation. This function will return an error
// if any part of the writes fail.
func (c Chunk) MarshalStream(w Writer) error {
	return w.WriteBytes(c.buf[c.wpos:])
}

// UnmarshalStream reads the Chunk data from a binary data representation. This function will return an error
// if any part of the writes fail.
func (c *Chunk) UnmarshalStream(r Reader) error {
	var err error
	c.buf, err = r.Bytes()
	c.rpos, c.wpos = 0, 0
	return err
}

// WriteTo writes data to the supplied Writer until there's no more data to write or when an error occurs. The return
// value is the number of bytes written. Any error encountered during the write is also returned.
func (c *Chunk) WriteTo(w io.Writer) (int64, error) {
	if c.Empty() {
		return 0, nil
	}
	n, err := w.Write(c.buf[c.rpos:])
	c.rpos += n
	return int64(n), err
}

// Seek will attempt to seek to the provided offset index and whence. This function will return
// the new offset if successful and will return an error if the offset and/or whence are invalid.
func (c *Chunk) Seek(o int64, w int) (int64, error) {
	switch w {
	case io.SeekStart:
		if o < 0 {
			return 0, ErrInvalidWhence
		}
	case io.SeekCurrent:
		o += int64(c.rpos)
	case io.SeekEnd:
		o += int64(c.Size())
	default:
		return 0, ErrInvalidWhence
	}
	if o < 0 || int(o) > c.Size() {
		return 0, ErrInvalidIndex
	}
	c.rpos = int(o)
	return o, nil

}

// ReadFrom reads data from the supplied Reader until EOF or error. The return value is the number of
// bytes read. Any error except io.EOF encountered during the read is also returned.
func (c *Chunk) ReadFrom(r io.Reader) (int64, error) {
	b := bufs.Get().([]byte)
	var (
		n   int
		w   int
		t   int64
		err error
	)
	for {
		if c.Limit > 0 {
			x := c.Limit - c.Size()
			if x <= 0 {
				break
			}
			n, err = r.Read(b[:x])
		} else {
			n, err = r.Read(b)
		}
		if n > 0 {
			w, err = c.Write(b[:n])
			if w < n {
				t += int64(w)
			} else {
				t += int64(n)
			}
			if err != nil {
				break
			}
		}
		if n < len(b) || err != nil {
			break
		}
	}
	bufs.Put(b)
	// TODO: Should delete this?
	//if err != nil && errors.Is(err, io.EOF) {
	//	err = nil
	//}
	return t, err
}
