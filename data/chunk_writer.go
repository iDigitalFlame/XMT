package data

import (
	"math"
)

// WriteInt writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteInt(n int) error {
	return c.WriteUint64(uint64(n))
}

// WriteUint writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteUint(n uint) error {
	return c.WriteUint64(uint64(n))
}

// WriteInt8 writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteInt8(n int8) error {
	return c.WriteUint8(uint8(n))
}

// WriteBool writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteBool(b bool) error {
	if b {
		return c.WriteUint8(1)
	}
	return c.WriteUint8(0)
}

// WriteInt16 writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteInt16(n int16) error {
	return c.WriteUint16(uint16(n))
}

// WriteInt32 writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteInt32(n int32) error {
	return c.WriteUint32(uint32(n))
}

// WriteInt64 writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteInt64(n int64) error {
	return c.WriteUint64(uint64(n))
}

// WriteUint8 writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteUint8(n uint8) error {
	if !c.Avaliable(1) {
		return ErrLimit
	}
	return c.small(byte(n))
}

// WriteBytes writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteBytes(b []byte) error {
	if !c.Avaliable(1) {
		return ErrLimit
	}
	switch l := len(b); {
	case l == 0:
		return c.small(0)
	case l < DataLimitSmall:
		if !c.Avaliable(2 + l) {
			return ErrLimit
		}
		if err := c.WriteUint8(1); err != nil {
			return err
		}
		if err := c.WriteUint8(uint8(l)); err != nil {
			return err
		}
	case l < DataLimitMedium:
		if !c.Avaliable(3 + l) {
			return ErrLimit
		}
		if err := c.WriteUint8(3); err != nil {
			return err
		}
		if err := c.WriteUint16(uint16(l)); err != nil {
			return err
		}
	case l < DataLimitLarge:
		if !c.Avaliable(5 + l) {
			return ErrLimit
		}
		if err := c.WriteUint8(5); err != nil {
			return err
		}
		if err := c.WriteUint32(uint32(l)); err != nil {
			return err
		}
	default:
		if !c.Avaliable(9 + l) {
			return ErrLimit
		}
		if err := c.WriteUint8(7); err != nil {
			return err
		}
		if err := c.WriteUint64(uint64(l)); err != nil {
			return err
		}
	}
	_, err := c.Write(b)
	return err
}

// WriteUint16 writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteUint16(n uint16) error {
	if !c.Avaliable(2) {
		return ErrLimit
	}
	return c.small(byte(n>>8), byte(n))
}

// WriteUint32 writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteUint32(n uint32) error {
	if !c.Avaliable(4) {
		return ErrLimit
	}
	return c.small(byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}

// WriteUint64 writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteUint64(n uint64) error {
	if !c.Avaliable(8) {
		return ErrLimit
	}
	return c.small(
		byte(n>>56), byte(n>>48), byte(n>>40), byte(n>>32),
		byte(n>>24), byte(n>>16), byte(n>>8), byte(n),
	)
}

// WriteString writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteString(s string) error {
	return c.WriteBytes([]byte(s))
}

// WriteFloat32 writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteFloat32(f float32) error {
	if !c.Avaliable(4) {
		return ErrLimit
	}
	return c.WriteUint32(math.Float32bits(f))
}

// WriteFloat64 writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteFloat64(f float64) error {
	if !c.Avaliable(8) {
		return ErrLimit
	}
	return c.WriteUint64(math.Float64bits(f))
}
