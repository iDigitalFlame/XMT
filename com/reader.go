package com

import (
	"io"
	"math"

	"github.com/iDigitalFlame/xmt/data"
)

// Int reads the value from the Packet payload buffer.
func (p *Packet) Int() (int, error) {
	v, err := p.Uint64()
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

// Uint reads the value from the Packet payload buffer.
func (p *Packet) Uint() (uint, error) {
	v, err := p.Uint64()
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

// Bool reads the value from the Packet payload buffer.
func (p *Packet) Bool() (bool, error) {
	v, err := p.Uint8()
	if err != nil {
		return false, err
	}
	return v == 1, nil
}

// Int8 reads the value from the Packet payload buffer.
func (p *Packet) Int8() (int8, error) {
	v, err := p.Uint8()
	if err != nil {
		return 0, err
	}
	return int8(v), nil
}

// ReadInt reads the value from the Packet payload buffer into
// the provided pointer.
func (p *Packet) ReadInt(i *int) error {
	v, err := p.Int()
	if err != nil {
		return err
	}
	*i = v
	return nil
}

// Int16 reads the value from the Packet payload buffer.
func (p *Packet) Int16() (int16, error) {
	v, err := p.Uint16()
	if err != nil {
		return 0, err
	}
	return int16(v), nil
}

// Int32 reads the value from the Packet payload buffer.
func (p *Packet) Int32() (int32, error) {
	v, err := p.Uint32()
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}

// Int64 reads the value from the Packet payload buffer.
func (p *Packet) Int64() (int64, error) {
	v, err := p.Uint64()
	if err != nil {
		return 0, err
	}
	return int64(v), nil
}

// Uint8 reads the value from the Packet payload buffer.
func (p *Packet) Uint8() (uint8, error) {
	if p.stream != nil {
		n, err := p.stream.Read(p.buf[0:1])
		if err != nil {
			return 0, err
		}
		if n < 1 {
			return 0, io.EOF
		}
	} else {
		if p.rpos+1 > len(p.buf) {
			return 0, io.EOF
		}
	}
	v := uint8(p.buf[p.rpos])
	if p.stream == nil {
		p.rpos++
	}
	return v, nil
}

// Bytes reads the value from the Packet payload buffer.
func (p *Packet) Bytes() ([]byte, error) {
	t, err := p.Uint8()
	if err != nil {
		return nil, err
	}
	var l int
	switch t {
	case 0:
		return nil, nil
	case 1, 2:
		n, err := p.Uint8()
		if err != nil {
			return nil, err
		}
		l = int(n)
	case 3, 4:
		n, err := p.Uint16()
		if err != nil {
			return nil, err
		}
		l = int(n)
	case 5, 6:
		n, err := p.Uint32()
		if err != nil {
			return nil, err
		}
		l = int(n)
	case 7, 8:
		n, err := p.Uint64()
		if err != nil {
			return nil, err
		}
		l = int(n)
	default:
		return nil, data.ErrInvalidBytes
	}
	b := make([]byte, l)
	n, err := data.ReadFully(p, b)
	if err != nil && (err != io.EOF || n != l) {
		return nil, err
	}
	if n != l {
		return nil, io.EOF
	}
	return b, nil
}

// ReadUint reads the value from the Packet payload buffer into
// the provided pointer.
func (p *Packet) ReadUint(i *uint) error {
	v, err := p.Uint()
	if err != nil {
		return err
	}
	*i = v
	return nil
}

// ReadInt8 reads the value from the Packet payload buffer into
// the provided pointer.
func (p *Packet) ReadInt8(i *int8) error {
	v, err := p.Int8()
	if err != nil {
		return err
	}
	*i = v
	return nil
}

// ReadBool reads the value from the Packet payload buffer into
// the provided pointer.
func (p *Packet) ReadBool(i *bool) error {
	v, err := p.Bool()
	if err != nil {
		return err
	}
	*i = v
	return nil
}

// Uint16 reads the value from the Packet payload buffer.
func (p *Packet) Uint16() (uint16, error) {
	if p.stream != nil {
		n, err := p.stream.Read(p.buf[0:2])
		if err != nil {
			return 0, err
		}
		if n < 2 {
			return 0, io.EOF
		}
	} else {
		if p.rpos+2 > len(p.buf) {
			return 0, io.EOF
		}
	}
	_ = p.buf[p.rpos+1]
	v := uint16(p.buf[p.rpos+1]) | uint16(p.buf[p.rpos])<<8
	if p.stream == nil {
		p.rpos += 2
	}
	return v, nil
}

// Uint32 reads the value from the Packet payload buffer.
func (p *Packet) Uint32() (uint32, error) {
	if p.stream != nil {
		n, err := p.stream.Read(p.buf[0:4])
		if err != nil {
			return 0, err
		}
		if n < 4 {
			return 0, io.EOF
		}
	} else {
		if p.rpos+4 > len(p.buf) {
			return 0, io.EOF
		}
	}
	_ = p.buf[p.rpos+3]
	v := uint32(p.buf[p.rpos+3]) | uint32(p.buf[p.rpos+2])<<8 | uint32(p.buf[p.rpos+1])<<16 | uint32(p.buf[p.rpos])<<24
	if p.stream == nil {
		p.rpos += 4
	}
	return v, nil
}

// Uint64 reads the value from the Packet payload buffer.
func (p *Packet) Uint64() (uint64, error) {
	if p.stream != nil {
		n, err := p.stream.Read(p.buf)
		if err != nil {
			return 0, err
		}
		if n < 8 {
			return 0, io.EOF
		}
	} else {
		if p.rpos+8 > len(p.buf) {
			return 0, io.EOF
		}
	}
	_ = p.buf[p.rpos+7]
	v := uint64(p.buf[p.rpos+7]) | uint64(p.buf[p.rpos+6])<<8 | uint64(p.buf[p.rpos+5])<<16 | uint64(p.buf[p.rpos+4])<<24 |
		uint64(p.buf[p.rpos+3])<<32 | uint64(p.buf[p.rpos+2])<<40 | uint64(p.buf[p.rpos+1])<<48 | uint64(p.buf[p.rpos])<<56
	if p.stream == nil {
		p.rpos += 8
	}
	return v, nil
}

// ReadInt16 reads the value from the Packet payload buffer into
// the provided pointer.
func (p *Packet) ReadInt16(i *int16) error {
	v, err := p.Int16()
	if err != nil {
		return err
	}
	*i = v
	return nil
}

// ReadInt32 reads the value from the Packet payload buffer into
// the provided pointer.
func (p *Packet) ReadInt32(i *int32) error {
	v, err := p.Int32()
	if err != nil {
		return err
	}
	*i = v
	return nil
}

// ReadInt64 reads the value from the Packet payload buffer into
// the provided pointer.
func (p *Packet) ReadInt64(i *int64) error {
	v, err := p.Int64()
	if err != nil {
		return err
	}
	*i = v
	return nil
}

// ReadUint8 reads the value from the Packet payload buffer into
// the provided pointer.
func (p *Packet) ReadUint8(i *uint8) error {
	v, err := p.Uint8()
	if err != nil {
		return err
	}
	*i = v
	return nil
}

// Float32 reads the value from the Packet payload buffer.
func (p *Packet) Float32() (float32, error) {
	v, err := p.Uint32()
	if err != nil {
		return 0, nil
	}
	return math.Float32frombits(v), nil
}

// Float64 reads the value from the Packet payload buffer.
func (p *Packet) Float64() (float64, error) {
	v, err := p.Uint64()
	if err != nil {
		return 0, nil
	}
	return math.Float64frombits(v), nil
}

// StringVal reads the value from the Packet payload buffer.
func (p *Packet) StringVal() (string, error) {
	b, err := p.Bytes()
	if err != nil {
		if err == data.ErrInvalidBytes {
			return "", data.ErrInvalidString
		}
		return "", err
	}
	return string(b), nil
}

// ReadUint16 reads the value from the Packet payload buffer into
// the provided pointer.
func (p *Packet) ReadUint16(i *uint16) error {
	v, err := p.Uint16()
	if err != nil {
		return err
	}
	*i = v
	return nil
}

// ReadUint32 reads the value from the Packet payload buffer into
// the provided pointer.
func (p *Packet) ReadUint32(i *uint32) error {
	v, err := p.Uint32()
	if err != nil {
		return err
	}
	*i = v
	return nil
}

// ReadUint64 reads the value from the Packet payload buffer into
// the provided pointer.
func (p *Packet) ReadUint64(i *uint64) error {
	v, err := p.Uint64()
	if err != nil {
		return err
	}
	*i = v
	return nil
}

// ReadString reads the value from the Packet payload buffer into
// the provided pointer.
func (p *Packet) ReadString(i *string) error {
	v, err := p.StringVal()
	if err != nil {
		return err
	}
	*i = v
	return nil
}

// Read reads the next len(p) bytes from the buffer or until the buffer
// is drained. The return value n is the number of bytes read.
func (p *Packet) Read(b []byte) (int, error) {
	if p.stream != nil {
		n, err := p.stream.Read(b)
		return n, err
	}
	if len(p.buf) <= p.rpos {
		p.Reset()
		if len(b) == 0 {
			return 0, io.EOF
		}
		return 0, io.EOF
	}
	n := copy(b, p.buf[p.rpos:])
	p.rpos += n
	return n, nil
}

// ReadFloat32 reads the value from the Packet payload buffer into
// the provided pointer.
func (p *Packet) ReadFloat32(i *float32) error {
	v, err := p.Float32()
	if err != nil {
		return err
	}
	*i = v
	return nil
}

// ReadFloat64 reads the value from the Packet payload buffer into
// the provided pointer.
func (p *Packet) ReadFloat64(i *float64) error {
	v, err := p.Float64()
	if err != nil {
		return err
	}
	*i = v
	return nil
}
