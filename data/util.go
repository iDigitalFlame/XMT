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
	"context"
	"io"
	"os"
	"unsafe"
)

type ctxReader struct {
	ctx    context.Context
	cancel context.CancelFunc
	io.ReadCloser
}
type nopReadCloser struct {
	_ [0]func()
	io.Reader
}
type nopWriteCloser struct {
	_ [0]func()
	io.Writer
}

func (r *ctxReader) Close() error {
	r.cancel()
	return r.ReadCloser.Close()
}
func (nopReadCloser) Close() error {
	return nil
}
func float32ToInt(f float32) uint32 {
	return *(*uint32)(unsafe.Pointer(&f))
}
func float64ToInt(f float64) uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
}
func (nopWriteCloser) Close() error {
	return nil
}
func float32FromInt(i uint32) float32 {
	return *(*float32)(unsafe.Pointer(&i))
}
func float64FromInt(i uint64) float64 {
	return *(*float64)(unsafe.Pointer(&i))
}

// ReadFile reads the named file and returns the contents. A successful call
// returns err == nil, not err == EOF.
//
// Because ReadFile reads the whole file, it does not treat an EOF from Read as
// an error to be reported.
//
// This is a pre go1.16 compatibility helper.
func ReadFile(n string) ([]byte, error) {
	f, err := os.OpenFile(n, 0, 0)
	if err != nil {
		return nil, err
	}
	var s int
	if i, err := f.Stat(); err == nil {
		if x := i.Size(); int64(int(x)) == x {
			s = int(x)
		}
	}
	if s++; s < 512 {
		s = 512
	}
	var (
		d = make([]byte, 0, s)
		c int
	)
	for {
		if len(d) >= cap(d) {
			q := append(d[:cap(d)], 0)
			d = q[:len(d)]
		}
		c, err = f.Read(d[len(d):cap(d)])
		if d = d[:len(d)+c]; err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
	}
	f.Close()
	return d, err
}

// ReadAll reads from r until an error or EOF and returns the data it read. A
// successful call returns err == nil, not err == EOF.
//
// Because ReadAll is defined to read from src until EOF, it does not treat an
// EOF from Read as an error to be reported.
//
// This is a pre go1.16 compatibility helper.
func ReadAll(r io.Reader) ([]byte, error) {
	var (
		b   = make([]byte, 0, 512)
		n   int
		err error
	)
	for {
		if len(b) == cap(b) {
			b = append(b, 0)[:len(b)]
		}
		n, err = r.Read(b[len(b):cap(b)])
		if b = b[:len(b)+n]; err != nil {
			if err == io.EOF {
				return b, nil
			}
			return b, err
		}
	}
}

// ReadCloser is a function that will wrap the supplied Reader in a NopReadCloser.
func ReadCloser(r io.Reader) io.ReadCloser {
	if v, ok := r.(io.ReadCloser); ok {
		return v
	}
	return &nopReadCloser{Reader: r}
}

// WriteCloser is a function that will wrap the supplied Writer in a NopWriteCloser.
func WriteCloser(w io.Writer) io.WriteCloser {
	if v, ok := w.(io.WriteCloser); ok {
		return v
	}
	return &nopWriteCloser{Writer: w}
}
func (r *ctxReader) Read(b []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		if err := r.ReadCloser.Close(); err != nil {
			return 0, err
		}
		return 0, r.ctx.Err()
	default:
		return r.ReadCloser.Read(b)
	}
}

// ReadStringList attempts to read a string list written using the 'WriteStringList'
// function from the supplied string into the string list pointer. If the provided
// array is nil or not large enough, it will be resized.
func ReadStringList(r Reader, s *[]string) error {
	t, err := r.Uint8()
	if err != nil {
		return err
	}
	var l int
	switch t {
	case 0:
		return nil
	case 1, 2:
		n, err := r.Uint8()
		if err != nil {
			return err
		}
		l = int(n)
	case 3, 4:
		n, err := r.Uint16()
		if err != nil {
			return err
		}
		l = int(n)
	case 5, 6:
		n, err := r.Uint32()
		if err != nil {
			return err
		}
		l = int(n)
	case 7, 8:
		n, err := r.Uint64()
		if err != nil {
			return err
		}
		l = int(n)
	default:
		return ErrInvalidType
	}
	if s == nil || len(*s) < l {
		*s = make([]string, l)
	}
	for x := 0; x < l; x++ {
		if err := r.ReadString(&(*s)[x]); err != nil {
			return err
		}
	}
	return nil
}

// WriteStringList will attempt to write the supplied string list to the writer.
// If the string list is nil or empty, it will write a zero byte to the Writer.
// The resulting data can be read using the 'ReadStringList' function.
func WriteStringList(w Writer, s []string) error {
	switch l := uint64(len(s)); {
	case l == 0:
		v, err := w.Write([]byte{0})
		if err == nil && v != 1 {
			return io.ErrShortWrite
		}
		return err
	case l < LimitSmall:
		if v, err := w.Write([]byte{1, byte(l)}); err != nil {
			return err
		} else if v != 2 {
			return io.ErrShortWrite
		}
	case l < LimitMedium:
		if v, err := w.Write([]byte{3, byte(l >> 8), byte(l)}); err != nil {
			return err
		} else if v != 3 {
			return io.ErrShortWrite
		}
	case l < LimitLarge:
		if v, err := w.Write([]byte{5, byte(l >> 24), byte(l >> 16), byte(l >> 8), byte(l)}); err != nil {
			return err
		} else if v != 5 {
			return io.ErrShortWrite
		}
	default:
		if v, err := w.Write([]byte{
			7, byte(l >> 56), byte(l >> 48), byte(l >> 40), byte(l >> 32),
			byte(l >> 24), byte(l >> 16), byte(l >> 8), byte(l),
		}); err != nil {
			return nil
		} else if v != 9 {
			return io.ErrShortWrite
		}
	}
	for i := range s {
		if err := w.WriteString(s[i]); err != nil {
			return err
		}
	}
	return nil
}

// NewCtxReader creates a reader backed by the supplied Reader and Context. This
// reader will automatically close when the parent context is canceled. This is
// useful in situations when direct copies using 'io.Copy' on threads or timed
// operations are required.
func NewCtxReader(x context.Context, r io.Reader) io.ReadCloser {
	var i ctxReader
	if c, ok := r.(io.ReadCloser); ok {
		i.ReadCloser = c
	} else {
		i.ReadCloser = &nopReadCloser{Reader: r}
	}
	i.ctx, i.cancel = context.WithCancel(x)
	return &i
}
