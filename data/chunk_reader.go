package data

import "io"

// Int reads the value from the Chunk payload buffer.
func (c *Chunk) Int() (int, error) {
	v, err := c.Uint64()
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

// Uint reads the value from the Chunk payload buffer.
func (c *Chunk) Uint() (uint, error) {
	v, err := c.Uint64()
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

// Bool reads the value from the Chunk payload buffer.
func (c *Chunk) Bool() (bool, error) {
	v, err := c.Uint8()
	if err != nil {
		return false, err
	}
	return v == 1, nil
}

// Int8 reads the value from the Chunk payload buffer.
func (c *Chunk) Int8() (int8, error) {
	v, err := c.Uint8()
	if err != nil {
		return 0, err
	}
	return int8(v), nil
}

// ReadInt reads the value from the Chunk payload buffer into the provided pointer.
func (c *Chunk) ReadInt(p *int) error {
	v, err := c.Int()
	if err != nil {
		return err
	}
	*p = v
	return nil
}

// Int16 reads the value from the Chunk payload buffer.
func (c *Chunk) Int16() (int16, error) {
	v, err := c.Uint16()
	if err != nil {
		return 0, err
	}
	return int16(v), nil
}

// Int32 reads the value from the Chunk payload buffer.
func (c *Chunk) Int32() (int32, error) {
	v, err := c.Uint32()
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}

// Int64 reads the value from the Chunk payload buffer.
func (c *Chunk) Int64() (int64, error) {
	v, err := c.Uint64()
	if err != nil {
		return 0, err
	}
	return int64(v), nil
}

// Uint8 reads the value from the Chunk payload buffer.
func (c *Chunk) Uint8() (uint8, error) {
	if c.rpos+1 > len(c.buf) {
		return 0, io.EOF
	}
	v := uint8(c.buf[c.rpos])
	c.rpos++
	return v, nil
}

// Bytes reads the value from the Chunk payload buffer.
func (c *Chunk) Bytes() ([]byte, error) {
	t, err := c.Uint8()
	if err != nil {
		return nil, err
	}
	var l int
	switch t {
	case 0:
		return nil, nil
	case 1, 2:
		n, err := c.Uint8()
		if err != nil {
			return nil, err
		}
		l = int(n)
	case 3, 4:
		n, err := c.Uint16()
		if err != nil {
			return nil, err
		}
		l = int(n)
	case 5, 6:
		n, err := c.Uint32()
		if err != nil {
			return nil, err
		}
		l = int(n)
	case 7, 8:
		n, err := c.Uint64()
		if err != nil {
			return nil, err
		}
		l = int(n)
	default:
		return nil, ErrInvalidBytes
	}
	b := make([]byte, l)
	n, err := ReadFully(c, b)
	if err != nil && ((err != io.EOF && err != ErrLimit) || n != l) {
		return nil, err
	}
	if n != l {
		return nil, io.EOF
	}
	return b, nil
}

// ReadUint reads the value from the Chunk payload buffer into the provided pointer.
func (c *Chunk) ReadUint(p *uint) error {
	v, err := c.Uint()
	if err != nil {
		return err
	}
	*p = v
	return nil
}

// ReadInt8 reads the value from the Chunk payload buffer into the provided pointer.
func (c *Chunk) ReadInt8(p *int8) error {
	v, err := c.Int8()
	if err != nil {
		return err
	}
	*p = v
	return nil
}

// ReadBool reads the value from the Chunk payload buffer into the provided pointer.
func (c *Chunk) ReadBool(p *bool) error {
	v, err := c.Bool()
	if err != nil {
		return err
	}
	*p = v
	return nil
}

// Uint16 reads the value from the Chunk payload buffer.
func (c *Chunk) Uint16() (uint16, error) {
	if c.rpos+2 > len(c.buf) {
		return 0, io.EOF
	}
	_ = c.buf[c.rpos+1]
	v := uint16(c.buf[c.rpos+1]) | uint16(c.buf[c.rpos])<<8
	c.rpos += 2
	return v, nil
}

// Uint32 reads the value from the Chunk payload buffer.
func (c *Chunk) Uint32() (uint32, error) {
	if c.rpos+4 > len(c.buf) {
		return 0, io.EOF
	}
	_ = c.buf[c.rpos+3]
	v := uint32(c.buf[c.rpos+3]) | uint32(c.buf[c.rpos+2])<<8 | uint32(c.buf[c.rpos+1])<<16 | uint32(c.buf[c.rpos])<<24
	c.rpos += 4
	return v, nil
}

// Uint64 reads the value from the Chunk payload buffer.
func (c *Chunk) Uint64() (uint64, error) {
	if c.rpos+8 > len(c.buf) {
		return 0, io.EOF
	}
	_ = c.buf[c.rpos+7]
	v := uint64(c.buf[c.rpos+7]) | uint64(c.buf[c.rpos+6])<<8 | uint64(c.buf[c.rpos+5])<<16 | uint64(c.buf[c.rpos+4])<<24 |
		uint64(c.buf[c.rpos+3])<<32 | uint64(c.buf[c.rpos+2])<<40 | uint64(c.buf[c.rpos+1])<<48 | uint64(c.buf[c.rpos])<<56
	c.rpos += 8
	return v, nil
}

// ReadInt16 reads the value from the Chunk payload buffer into the provided pointer.
func (c *Chunk) ReadInt16(p *int16) error {
	v, err := c.Int16()
	if err != nil {
		return err
	}
	*p = v
	return nil
}

// ReadInt32 reads the value from the Chunk payload buffer into the provided pointer.
func (c *Chunk) ReadInt32(p *int32) error {
	v, err := c.Int32()
	if err != nil {
		return err
	}
	*p = v
	return nil
}

// ReadInt64 reads the value from the Chunk payload buffer into the provided pointer.
func (c *Chunk) ReadInt64(p *int64) error {
	v, err := c.Int64()
	if err != nil {
		return err
	}
	*p = v
	return nil
}

// ReadUint8 reads the value from the Chunk payload buffer into the provided pointer.
func (c *Chunk) ReadUint8(p *uint8) error {
	v, err := c.Uint8()
	if err != nil {
		return err
	}
	*p = v
	return nil
}

// Float32 reads the value from the Chunk payload buffer.
func (c *Chunk) Float32() (float32, error) {
	v, err := c.Uint32()
	if err != nil {
		return 0, nil
	}
	return float32FromInt(v), nil
}

// Float64 reads the value from the Chunk payload buffer.
func (c *Chunk) Float64() (float64, error) {
	v, err := c.Uint64()
	if err != nil {
		return 0, nil
	}
	return float64FromInt(v), nil
}

// StringVal reads the value from the Chunk payload buffer.
func (c *Chunk) StringVal() (string, error) {
	b, err := c.Bytes()
	if err != nil {
		if err == ErrInvalidBytes {
			return "", ErrInvalidString
		}
		return "", err
	}
	return string(b), nil
}

// ReadUint16 reads the value from the Chunk payload buffer into the provided pointer.
func (c *Chunk) ReadUint16(p *uint16) error {
	v, err := c.Uint16()
	if err != nil {
		return err
	}
	*p = v
	return nil
}

// ReadUint32 reads the value from the Chunk payload buffer into the provided pointer.
func (c *Chunk) ReadUint32(p *uint32) error {
	v, err := c.Uint32()
	if err != nil {
		return err
	}
	*p = v
	return nil
}

// ReadUint64 reads the value from the Chunk payload buffer into the provided pointer.
func (c *Chunk) ReadUint64(p *uint64) error {
	v, err := c.Uint64()
	if err != nil {
		return err
	}
	*p = v
	return nil
}

// ReadString reads the value from the Chunk payload buffer into the provided pointer.
func (c *Chunk) ReadString(p *string) error {
	v, err := c.StringVal()
	if err != nil {
		return err
	}
	*p = v
	return nil
}

// ReadFloat32 reads the value from the Chunk payload buffer into the provided pointer.
func (c *Chunk) ReadFloat32(p *float32) error {
	v, err := c.Float32()
	if err != nil {
		return err
	}
	*p = v
	return nil
}

// ReadFloat64 reads the value from the Chunk payload buffer into the provided pointer.
func (c *Chunk) ReadFloat64(p *float64) error {
	v, err := c.Float64()
	if err != nil {
		return err
	}
	*p = v
	return nil
}
