package c2

import (
	"io"
	"math"

	"github.com/iDigitalFlame/xmt/xmt/data"
)

func (w *wrapBuffer) Int() (int, error) {
	v, err := w.Uint64()
	if err != nil {
		return 0, err
	}
	return int(v), nil
}
func (w *wrapBuffer) Uint() (uint, error) {
	v, err := w.Uint64()
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}
func (w *wrapBuffer) Bool() (bool, error) {
	v, err := w.Uint8()
	if err != nil {
		return false, err
	}
	return v == 1, nil
}
func (w *wrapBuffer) Int8() (int8, error) {
	v, err := w.Uint8()
	if err != nil {
		return 0, err
	}
	return int8(v), nil
}
func (w *wrapBuffer) ReadInt(i *int) error {
	v, err := w.Int()
	if err != nil {
		return err
	}
	*i = v
	return nil
}
func (w *wrapBuffer) Int16() (int16, error) {
	v, err := w.Uint16()
	if err != nil {
		return 0, err
	}
	return int16(v), nil
}
func (w *wrapBuffer) Int32() (int32, error) {
	v, err := w.Uint32()
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}
func (w *wrapBuffer) Int64() (int64, error) {
	v, err := w.Uint64()
	if err != nil {
		return 0, err
	}
	return int64(v), nil
}
func (w *wrapBuffer) Uint8() (uint8, error) {
	n, err := w.Read(w.rbuf[0:1])
	if err != nil {
		return 0, err
	}
	if n < 1 {
		return 0, io.EOF
	}
	return uint8(w.rbuf[0]), nil
}
func (w *wrapBuffer) Bytes() ([]byte, error) {
	t, err := w.Uint8()
	if err != nil {
		return nil, err
	}
	var l int
	switch t {
	case 0:
		return nil, nil
	case 1, 2:
		n, err := w.Uint8()
		if err != nil {
			return nil, err
		}
		l = int(n)
	case 3, 4:
		n, err := w.Uint16()
		if err != nil {
			return nil, err
		}
		l = int(n)
	case 5, 6:
		n, err := w.Uint32()
		if err != nil {
			return nil, err
		}
		l = int(n)
	case 7, 8:
		n, err := w.Uint64()
		if err != nil {
			return nil, err
		}
		l = int(n)
	default:
		return nil, data.ErrInvalidBytes
	}
	b := make([]byte, l)
	n, err := w.Read(b)
	if err != nil {

		return nil, err
	}
	if n != l {
		return nil, io.EOF
	}
	return b, nil
}
func (w *wrapBuffer) ReadUint(i *uint) error {
	v, err := w.Uint()
	if err != nil {
		return err
	}
	*i = v
	return nil
}
func (w *wrapBuffer) ReadInt8(i *int8) error {
	v, err := w.Int8()
	if err != nil {
		return err
	}
	*i = v
	return nil
}
func (w *wrapBuffer) ReadBool(i *bool) error {
	v, err := w.Bool()
	if err != nil {
		return err
	}
	*i = v
	return nil
}
func (b *buffer) Read(p []byte) (int, error) {
	if len(b.buf) <= b.rpos {
		b.Reset()
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n := copy(p, b.buf[b.rpos:])
	b.rpos += n
	return n, nil
}
func (w *wrapBuffer) Uint16() (uint16, error) {
	n, err := w.Read(w.rbuf[0:2])
	if err != nil {
		return 0, err
	}
	if n < 2 {
		return 0, io.EOF
	}
	return uint16(w.rbuf[1]) | uint16(w.rbuf[0])<<8, nil
}
func (w *wrapBuffer) Uint32() (uint32, error) {
	n, err := w.Read(w.rbuf[0:4])
	if err != nil {
		return 0, err
	}
	if n < 4 {
		return 0, io.EOF
	}
	return uint32(w.rbuf[3]) | uint32(w.rbuf[2])<<8 | uint32(w.rbuf[1])<<16 | uint32(w.rbuf[0])<<24, nil
}
func (w *wrapBuffer) Uint64() (uint64, error) {
	n, err := w.Read(w.rbuf)
	if err != nil {
		return 0, err
	}
	if n < 8 {
		return 0, io.EOF
	}
	return uint64(w.rbuf[7]) | uint64(w.rbuf[6])<<8 | uint64(w.rbuf[5])<<16 | uint64(w.rbuf[4])<<24 |
		uint64(w.rbuf[3])<<32 | uint64(w.rbuf[2])<<40 | uint64(w.rbuf[1])<<48 | uint64(w.rbuf[0])<<56, nil
}
func (w *wrapBuffer) ReadInt16(i *int16) error {
	v, err := w.Int16()
	if err != nil {
		return err
	}
	*i = v
	return nil
}
func (w *wrapBuffer) ReadInt32(i *int32) error {
	v, err := w.Int32()
	if err != nil {
		return err
	}
	*i = v
	return nil
}
func (w *wrapBuffer) ReadInt64(i *int64) error {
	v, err := w.Int64()
	if err != nil {
		return err
	}
	*i = v
	return nil
}
func (w *wrapBuffer) ReadUint8(i *uint8) error {
	v, err := w.Uint8()
	if err != nil {
		return err
	}
	*i = v
	return nil
}
func (w *wrapBuffer) Float32() (float32, error) {
	v, err := w.Uint32()
	if err != nil {
		return 0, nil
	}
	return math.Float32frombits(v), nil
}
func (w *wrapBuffer) Float64() (float64, error) {
	v, err := w.Uint64()
	if err != nil {
		return 0, nil
	}
	return math.Float64frombits(v), nil
}
func (w *wrapBuffer) UTFString() (string, error) {
	t, err := w.Uint8()
	if err != nil {
		return "", err
	}
	var l int
	switch t {
	case 0:
		return "", nil
	case 1, 2:
		n, err := w.Uint8()
		if err != nil {
			return "", err
		}
		l = int(n)
	case 3, 4:
		n, err := w.Uint16()
		if err != nil {
			return "", err
		}
		l = int(n)
	case 5, 6:
		n, err := w.Uint32()
		if err != nil {
			return "", err
		}
		l = int(n)
	case 7, 8:
		n, err := w.Uint64()
		if err != nil {
			return "", err
		}
		l = int(n)
	default:
		return "", data.ErrInvalidString
	}
	if t%2 == 0 {
		b := make([]rune, l)
		for i := range b {
			v, err := w.Uint16()
			if err != nil {
				return "", err
			}
			b[i] = rune(v)
		}
		return string(b), nil
	}
	b := make([]byte, l)
	n, err := w.Read(b)
	if err != nil {
		return "", err
	}
	if n != l {
		return "", io.EOF
	}
	return string(b), nil
}
func (w *wrapBuffer) ReadUint16(i *uint16) error {
	v, err := w.Uint16()
	if err != nil {
		return err
	}
	*i = v
	return nil
}
func (w *wrapBuffer) ReadUint32(i *uint32) error {
	v, err := w.Uint32()
	if err != nil {
		return err
	}
	*i = v
	return nil
}
func (w *wrapBuffer) ReadUint64(i *uint64) error {
	v, err := w.Uint64()
	if err != nil {
		return err
	}
	*i = v
	return nil
}
func (w *wrapBuffer) ReadString(i *string) error {
	v, err := w.UTFString()
	if err != nil {
		return err
	}
	*i = v
	return nil
}
func (w *wrapBuffer) Read(b []byte) (int, error) {
	n, err := w.r.Read(b)
	if err == io.EOF && n == len(b) {
		return n, nil
	}
	return n, err
}
func (w *wrapBuffer) ReadFloat32(i *float32) error {
	v, err := w.Float32()
	if err != nil {
		return err
	}
	*i = v
	return nil
}
func (w *wrapBuffer) ReadFloat64(i *float64) error {
	v, err := w.Float64()
	if err != nil {
		return err
	}
	*i = v
	return nil
}
