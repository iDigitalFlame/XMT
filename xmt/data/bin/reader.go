package bin

import (
	"io"
	"math"
)

// Reader is a struct that allows for reading a byte array seamlessly.
// This struct also gives support for reading strings.
type Reader struct {
	UTF16      bool

	pos  int
	buf []byte
}

func (r *Reader) Close() error {
	return nil
}

func (r *Reader) Bytes() []byte {
	if r.pos >= len(r.buf) {
		return nil
	}
	return r.buf[r.pos:]
}

func NewReader(b []byte) *Reader {
	return &Reader{UTF16: ByteStrUTF16, buf: b}
}

func (r *Reader) ReadUint8() (uint8, error) {
	if r.pos >= len(r.buf) {
		return 0, io.EOF
	}
	b := r.buf[r.pos]
	r.pos++
	return b, nil
}

func (r *Reader) Read(b []byte) (int, error) {
	if r.pos >= len(r.buf) {
		return 0, io.EOF
	}
	n := int(math.Min(float64(len(b)), float64(len(r.buf)-r.pos)))
	copy(b, r.buf[r.pos:r.pos+n])
	return n, nil
}

func (r *Reader) ReadUint16() (uint16, error) {
	if r.pos+1 >= len(r.buf) {
		return 0, io.EOF
	}
	b := uint16(uint16(r.buf[r.pos+1]) | uint16(r.buf[r.pos])<<8)
	r.pos += 2
	return b, nil
}

func (r *Reader) ReadString() (string, error) {
	if r.pos >= len(r.buf) {
		return "", io.EOF
	}
	n := 0
	if (r.buf[r.pos] & (1 << 7)) > 0 {
		if r.pos+1 >= len(r.buf) {
			return "", io.EOF
		}
		n = int(uint16(r.buf[r.pos+1]) | uint16(byte(int(r.buf[r.pos])&^(1<<7)))<<8)
		r.pos += 2
	} else {
		n = int(r.buf[r.pos])
		r.pos++
	}
	if n <= 0 {
		return "", ErrStringInvalid
	}
	if r.pos+n >= len(r.buf) {
		return "", io.EOF
	}
	s := make([]byte, n)
	for v := 0; v < n; r.pos++ {
		s[v] = r.buf[r.pos]
		if r.UTF16 {
			s[v+1] = r.buf[r.pos+1]
			r.pos++
			v++
		}
		v++
	}
	return string(s), nil
}
