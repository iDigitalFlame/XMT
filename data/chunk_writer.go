package data

import "io"

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
	if c.Limit > 0 && !c.Avaliable(1) {
		return ErrLimit
	}
	v, err := c.Write([]byte{byte(n)})
	if err == nil && v != 1 {
		return io.ErrShortWrite
	}
	return err
}

// WriteBytes writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteBytes(b []byte) error {
	if c.Limit > 0 && !c.Avaliable(1) {
		return ErrLimit
	}
	switch l := uint64(len(b)); {
	case l == 0:
		v, err := c.Write([]byte{0})
		if err == nil && v != 1 {
			return io.ErrShortWrite
		}
		return err
	case l < LimitSmall:
		if c.Limit > 0 && !c.Avaliable(2+int(l)) {
			return ErrLimit
		}
		if v, err := c.Write([]byte{1, byte(l)}); err != nil {
			return err
		} else if v != 2 {
			return io.ErrShortWrite
		}
	case l < LimitMedium:
		if c.Limit > 0 && !c.Avaliable(3+int(l)) {
			return ErrLimit
		}
		if v, err := c.Write([]byte{3, byte(l >> 8), byte(l)}); err != nil {
			return err
		} else if v != 3 {
			return io.ErrShortWrite
		}
	case l < LimitLarge:
		if c.Limit > 0 && !c.Avaliable(5+int(l)) {
			return ErrLimit
		}
		if v, err := c.Write([]byte{5, byte(l >> 24), byte(l >> 16), byte(l >> 8), byte(l)}); err != nil {
			return err
		} else if v != 5 {
			return io.ErrShortWrite
		}
	default:
		if c.Limit > 0 && !c.Avaliable(9+int(l)) {
			return ErrLimit
		}
		if v, err := c.Write([]byte{
			7, byte(l >> 56), byte(l >> 48), byte(l >> 40), byte(l >> 32),
			byte(l >> 24), byte(l >> 16), byte(l >> 8), byte(l),
		}); err != nil {
			return nil
		} else if v != 9 {
			return io.ErrShortWrite
		}
	}
	_, err := c.Write(b)
	return err
}

// WriteUint16 writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteUint16(n uint16) error {
	if c.Limit > 0 && !c.Avaliable(2) {
		return ErrLimit
	}
	v, err := c.Write([]byte{byte(n >> 8), byte(n)})
	if err == nil && v != 2 {
		return io.ErrShortWrite
	}
	return err
}

// WriteUint32 writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteUint32(n uint32) error {
	if c.Limit > 0 && !c.Avaliable(4) {
		return ErrLimit
	}
	v, err := c.Write([]byte{byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)})
	if err == nil && v != 4 {
		return io.ErrShortWrite
	}
	return err
}

// WriteUint64 writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteUint64(n uint64) error {
	if c.Limit > 0 && !c.Avaliable(8) {
		return ErrLimit
	}
	v, err := c.Write([]byte{
		byte(n >> 56), byte(n >> 48), byte(n >> 40), byte(n >> 32),
		byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n),
	})
	if err == nil && v != 8 {
		return io.ErrShortWrite
	}
	return err
}

// WriteString writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteString(s string) error {
	return c.WriteBytes([]byte(s))
}

// WriteFloat32 writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteFloat32(f float32) error {
	if c.Limit > 0 && !c.Avaliable(4) {
		return ErrLimit
	}
	return c.WriteUint32(float32ToInt(f))
}

// WriteFloat64 writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteFloat64(f float64) error {
	if c.Limit > 0 && !c.Avaliable(8) {
		return ErrLimit
	}
	return c.WriteUint64(float64ToInt(f))
}
