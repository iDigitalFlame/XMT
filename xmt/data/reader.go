package data

import (
	"errors"
	"io"
	"math"
)

var (
	// ErrInvalidBytes is an error that occurs when the Bytes function
	// could not propertly determine the type of byte array from the Reader.
	ErrInvalidBytes = errors.New("could not understand string type")
	// ErrInvalidString is an error that occurs when the ReadString or String functions
	// could not propertly determine the type of string from the Reader.
	ErrInvalidString = errors.New("could not understand string type")
)

type readerBase struct {
	r   io.Reader
	buf []byte
}

func (r *readerBase) Close() error {
	if c, ok := r.r.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// NewReader creates a simple Reader struct from the base Reader
// provided.
func NewReader(r io.Reader) Reader {
	return &readerBase{r: r, buf: make([]byte, 8)}
}
func (r *readerBase) Int() (int, error) {
	v, err := r.Uint64()
	if err != nil {
		return 0, err
	}
	return int(v), nil
}
func (r *readerBase) Uint() (uint, error) {
	v, err := r.Uint64()
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}
func (r *readerBase) Bool() (bool, error) {
	v, err := r.Uint8()
	if err != nil {
		return false, err
	}
	return v == 1, nil
}
func (r *readerBase) Int8() (int8, error) {
	v, err := r.Uint8()
	if err != nil {
		return 0, err
	}
	return int8(v), nil
}
func (r *readerBase) ReadInt(p *int) error {
	v, err := r.Int()
	if err != nil {
		return err
	}
	*p = v
	return nil
}
func (r *readerBase) Int16() (int16, error) {
	v, err := r.Uint16()
	if err != nil {
		return 0, err
	}
	return int16(v), nil
}
func (r *readerBase) Int32() (int32, error) {
	v, err := r.Uint32()
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}
func (r *readerBase) Int64() (int64, error) {
	v, err := r.Uint64()
	if err != nil {
		return 0, err
	}
	return int64(v), nil
}
func (r *readerBase) Uint8() (uint8, error) {
	n, err := r.r.Read(r.buf[0:1])
	if err != nil {
		return 0, err
	}
	if n < 1 {
		return 0, io.EOF
	}
	return uint8(r.buf[0]), nil
}
func (r *readerBase) Bytes() ([]byte, error) {
	t, err := r.Uint8()
	if err != nil {
		return nil, err
	}
	var l int
	switch t {
	case 0:
		return nil, nil
	case 1, 2:
		n, err := r.Uint8()
		if err != nil {
			return nil, err
		}
		l = int(n)
	case 3, 4:
		n, err := r.Uint16()
		if err != nil {
			return nil, err
		}
		l = int(n)
	case 5, 6:
		n, err := r.Uint32()
		if err != nil {
			return nil, err
		}
		l = int(n)
	case 7, 8:
		n, err := r.Uint64()
		if err != nil {
			return nil, err
		}
		l = int(n)
	default:
		return nil, ErrInvalidBytes
	}
	b := make([]byte, l)
	n, err := ReadFully(r.r, b)
	if err != nil && (err != io.EOF || n != l) {
		return nil, err
	}
	if n != l {
		return nil, io.EOF
	}
	return b, nil
}
func (r *readerBase) ReadUint(p *uint) error {
	v, err := r.Uint()
	if err != nil {
		return err
	}
	*p = v
	return nil
}
func (r *readerBase) ReadInt8(p *int8) error {
	v, err := r.Int8()
	if err != nil {
		return err
	}
	*p = v
	return nil
}
func (r *readerBase) ReadBool(p *bool) error {
	v, err := r.Bool()
	if err != nil {
		return err
	}
	*p = v
	return nil
}
func (r *readerBase) Uint16() (uint16, error) {
	n, err := ReadFully(r.r, r.buf[0:2])
	if err != nil {
		return 0, err
	}
	if n < 2 {
		return 0, io.EOF
	}
	return uint16(r.buf[1]) | uint16(r.buf[0])<<8, nil
}
func (r *readerBase) Uint32() (uint32, error) {
	n, err := ReadFully(r.r, r.buf[0:4])
	if err != nil {
		return 0, err
	}
	if n < 4 {
		return 0, io.EOF
	}
	return uint32(r.buf[3]) | uint32(r.buf[2])<<8 | uint32(r.buf[1])<<16 | uint32(r.buf[0])<<24, nil
}
func (r *readerBase) Uint64() (uint64, error) {
	n, err := ReadFully(r.r, r.buf)
	if err != nil {
		return 0, err
	}
	if n < 8 {
		return 0, io.EOF
	}
	return uint64(r.buf[7]) | uint64(r.buf[6])<<8 | uint64(r.buf[5])<<16 | uint64(r.buf[4])<<24 |
		uint64(r.buf[3])<<32 | uint64(r.buf[2])<<40 | uint64(r.buf[1])<<48 | uint64(r.buf[0])<<56, nil
}
func (r *readerBase) ReadInt16(p *int16) error {
	v, err := r.Int16()
	if err != nil {
		return err
	}
	*p = v
	return nil
}
func (r *readerBase) ReadInt32(p *int32) error {
	v, err := r.Int32()
	if err != nil {
		return err
	}
	*p = v
	return nil
}
func (r *readerBase) ReadInt64(p *int64) error {
	v, err := r.Int64()
	if err != nil {
		return err
	}
	*p = v
	return nil
}
func (r *readerBase) ReadUint8(p *uint8) error {
	v, err := r.Uint8()
	if err != nil {
		return err
	}
	*p = v
	return nil
}
func (r *readerBase) Float32() (float32, error) {
	v, err := r.Uint32()
	if err != nil {
		return 0, nil
	}
	return math.Float32frombits(v), nil
}
func (r *readerBase) Float64() (float64, error) {
	v, err := r.Uint64()
	if err != nil {
		return 0, nil
	}
	return math.Float64frombits(v), nil
}
func (r *readerBase) Read(b []byte) (int, error) {
	return r.r.Read(b)
}
func (r *readerBase) ReadUint16(p *uint16) error {
	v, err := r.Uint16()
	if err != nil {
		return err
	}
	*p = v
	return nil
}
func (r *readerBase) ReadUint32(p *uint32) error {
	v, err := r.Uint32()
	if err != nil {
		return err
	}
	*p = v
	return nil
}
func (r *readerBase) ReadUint64(p *uint64) error {
	v, err := r.Uint64()
	if err != nil {
		return err
	}
	*p = v
	return nil
}
func (r *readerBase) ReadString(p *string) error {
	v, err := r.StringVal()
	if err != nil {
		return err
	}
	*p = v
	return nil
}
func (r *readerBase) StringVal() (string, error) {
	b, err := r.Bytes()
	if err != nil {
		if err == ErrInvalidBytes {
			return "", ErrInvalidString
		}
		return "", err
	}
	return string(b), nil
}
func (r *readerBase) ReadFloat32(p *float32) error {
	v, err := r.Float32()
	if err != nil {
		return err
	}
	*p = v
	return nil
}
func (r *readerBase) ReadFloat64(p *float64) error {
	v, err := r.Float64()
	if err != nil {
		return err
	}
	*p = v
	return nil
}
