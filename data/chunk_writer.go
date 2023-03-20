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

import "io"

// WriteInt writes the supplied value to the Chunk payload buffer.
//
// If this function returns 'ErrLimit' this indicates that the write was NOT
// preformed.
func (c *Chunk) WriteInt(n int) error {
	return c.WriteUint64(uint64(n))
}

// WriteUint writes the supplied value to the Chunk payload buffer.
//
// If this function returns 'ErrLimit' this indicates that the write was NOT
// preformed.
func (c *Chunk) WriteUint(n uint) error {
	return c.WriteUint64(uint64(n))
}

// WriteInt8 writes the supplied value to the Chunk payload buffer.
//
// If this function returns 'ErrLimit' this indicates that the write was NOT
// preformed.
func (c *Chunk) WriteInt8(n int8) error {
	return c.WriteUint8(uint8(n))
}

// WriteBool writes the supplied value to the Chunk payload buffer.
//
// If this function returns 'ErrLimit' this indicates that the write was NOT
// preformed.
func (c *Chunk) WriteBool(b bool) error {
	if b {
		return c.WriteUint8(1)
	}
	return c.WriteUint8(0)
}

// WriteInt16 writes the supplied value to the Chunk payload buffer.
//
// If this function returns 'ErrLimit' this indicates that the write was NOT
// preformed.
func (c *Chunk) WriteInt16(n int16) error {
	return c.WriteUint16(uint16(n))
}

// WriteInt32 writes the supplied value to the Chunk payload buffer.
//
// If this function returns 'ErrLimit' this indicates that the write was NOT
// preformed.
func (c *Chunk) WriteInt32(n int32) error {
	return c.WriteUint32(uint32(n))
}

// WriteInt64 writes the supplied value to the Chunk payload buffer.
//
// If this function returns 'ErrLimit' this indicates that the write was NOT
// preformed.
func (c *Chunk) WriteInt64(n int64) error {
	return c.WriteUint64(uint64(n))
}

// WriteUint8 writes the supplied value to the Chunk payload buffer.
//
// If this function returns 'ErrLimit' this indicates that the write was NOT
// preformed.
func (c *Chunk) WriteUint8(n uint8) error {
	i, err := c.checkWriteSize(1)
	if i == -1 {
		return err
	}
	_ = c.buf[i]
	c.buf[i] = n
	return err
}

// WriteBytes writes the supplied value to the Chunk payload buffer.
func (c *Chunk) WriteBytes(b []byte) error {
	i, err := c.checkWriteSize(1)
	if i == -1 {
		return err
	}
	var x int
	switch l := uint64(len(b)); {
	case l == 0:
		_ = c.buf[i]
		c.buf[i] = 0
		return err
	case l < LimitSmall:
		if x, err = c.checkWriteSize(1 + int(l)); x == -1 {
			return err
		}
		_, x = c.buf[i+1+int(l)], x+1
		c.buf[i], c.buf[i+1] = 1, byte(l)
	case l < LimitMedium:
		if x, err = c.checkWriteSize(2 + int(l)); x == -1 {
			return err
		}
		_, x = c.buf[i+2+int(l)], x+2
		c.buf[i], c.buf[i+1], c.buf[i+2] = 3, byte(l>>8), byte(l)
	case l < LimitLarge:
		if x, err = c.checkWriteSize(4 + int(l)); x == -1 {
			return err
		}
		_, x = c.buf[i+4+int(l)], x+4
		c.buf[i], c.buf[i+1], c.buf[i+2] = 5, byte(l>>24), byte(l>>16)
		c.buf[i+3], c.buf[i+4] = byte(l>>8), byte(l)
	default:
		if x, err = c.checkWriteSize(8 + int(l)); x == -1 {
			return err
		}
		_, x = c.buf[i+8+int(l)], x+8
		c.buf[i], c.buf[i+1], c.buf[i+2] = 7, byte(l>>56), byte(l>>48)
		c.buf[i+3], c.buf[i+4] = byte(l>>40), byte(l>>32)
		c.buf[i+5], c.buf[i+6] = byte(l>>24), byte(l>>16)
		c.buf[i+7], c.buf[i+8] = byte(l>>8), byte(l)
	}
	if n := copy(c.buf[x:], b); n != len(b) {
		return io.ErrShortWrite
	}
	return err
}

// WriteUint16 writes the supplied value to the Chunk payload buffer.
//
// If this function returns 'ErrLimit' this indicates that the write was NOT
// preformed.
func (c *Chunk) WriteUint16(n uint16) error {
	i, err := c.checkWriteSize(2)
	if i == -1 {
		return err
	}
	_ = c.buf[i+1]
	c.buf[i], c.buf[i+1] = byte(n>>8), byte(n)
	return err
}

// WriteUint32 writes the supplied value to the Chunk payload buffer.
//
// If this function returns 'ErrLimit' this indicates that the write was NOT
// preformed.
func (c *Chunk) WriteUint32(n uint32) error {
	i, err := c.checkWriteSize(4)
	if i == -1 {
		return err
	}
	_ = c.buf[i+3]
	c.buf[i], c.buf[i+1] = byte(n>>24), byte(n>>16)
	c.buf[i+2], c.buf[i+3] = byte(n>>8), byte(n)
	return err
}

// WriteUint64 writes the supplied value to the Chunk payload buffer.
//
// If this function returns 'ErrLimit' this indicates that the write was NOT
// preformed.
func (c *Chunk) WriteUint64(n uint64) error {
	i, err := c.checkWriteSize(8)
	if i == -1 {
		return err
	}
	_ = c.buf[i+7]
	c.buf[i], c.buf[i+1] = byte(n>>56), byte(n>>48)
	c.buf[i+2], c.buf[i+3] = byte(n>>40), byte(n>>32)
	c.buf[i+4], c.buf[i+5] = byte(n>>24), byte(n>>16)
	c.buf[i+6], c.buf[i+7] = byte(n>>8), byte(n)
	return err
}

// WriteString writes the supplied value to the Chunk payload buffer.
//
// If this function returns 'ErrLimit' this indicates that the write was NOT
// preformed.
func (c *Chunk) WriteString(s string) error {
	return c.WriteBytes([]byte(s))
}

// WriteFloat32 writes the supplied value to the Chunk payload buffer.
//
// If this function returns 'ErrLimit' this indicates that the write was NOT
// preformed.
func (c *Chunk) WriteFloat32(f float32) error {
	return c.WriteUint32(float32ToInt(f))
}

// WriteFloat64 writes the supplied value to the Chunk payload buffer.
//
// If this function returns 'ErrLimit' this indicates that the write was NOT
// preformed.
func (c *Chunk) WriteFloat64(f float64) error {
	return c.WriteUint64(float64ToInt(f))
}

// WriteBoolPos writes the supplied boolean value to the Chunk payload buffer at
// the supplied index 'p'.
//
// The error 'io.EOF' is returned if the position specified is greater than the
// Chunk buffer size, or 'ErrLimit' if this position is greater than the set
// Chunk limit.
func (c *Chunk) WriteBoolPos(p int, b bool) error {
	if p >= c.Size() {
		return io.EOF
	}
	if c.Limit > 0 && p >= c.Limit {
		return ErrLimit
	}
	if _ = c.buf[p]; b {
		c.buf[p] = 1
	} else {
		c.buf[p] = 0
	}
	return nil
}
func (c *Chunk) checkWriteSize(n int) (int, error) {
	if c.Limit > 0 && !c.Available(n) {
		return -1, ErrLimit
	}
	i, err := c.quickSlice(n)
	switch {
	case i == 0 && err != nil:
		return -1, err
	case c.Limit <= 0 && c.Size() < i+n:
		return -1, io.ErrShortWrite
	}
	return i, err
}

// WriteUint8Pos writes the supplied uint8 value to the Chunk payload buffer at
// the supplied index 'p'.
//
// The error 'io.EOF' is returned if the position specified is greater than the
// Chunk buffer size, or 'ErrLimit' if this position is greater than the set
// Chunk limit.
func (c *Chunk) WriteUint8Pos(p int, n uint8) error {
	if p >= c.Size() {
		return io.EOF
	}
	if c.Limit > 0 && p >= c.Limit {
		return ErrLimit
	}
	_ = c.buf[p]
	c.buf[p] = n
	return nil
}

// WriteUint16Pos writes the supplied uint16 value to the Chunk payload buffer
// at the supplied index 'p'.
//
// The error 'io.EOF' is returned if the position specified is greater than the
// Chunk buffer size, or 'ErrLimit' if this position is greater than the set
// Chunk limit.
func (c *Chunk) WriteUint16Pos(p int, n uint16) error {
	if p >= c.Size() || p+1 >= c.Size() {
		return io.EOF
	}
	if c.Limit > 0 && (p >= c.Limit || p+1 >= c.Limit) {
		return ErrLimit
	}
	_ = c.buf[p+1]
	c.buf[p], c.buf[p+1] = byte(n>>8), byte(n)
	return nil
}

// WriteUint32Pos writes the supplied uint16 value to the Chunk payload buffer
// at the supplied index 'p'.
//
// The error 'io.EOF' is returned if the position specified is greater than the
// Chunk buffer size, or 'ErrLimit' if this position is greater than the set
// Chunk limit.
func (c *Chunk) WriteUint32Pos(p int, n uint32) error {
	if p >= c.Size() || p+3 >= c.Size() {
		return io.EOF
	}
	if c.Limit > 0 && (p >= c.Limit || p+3 >= c.Limit) {
		return ErrLimit
	}
	_ = c.buf[p+3]
	c.buf[p], c.buf[p+1], c.buf[p+2], c.buf[p+3] = byte(n>>24), byte(n>>16), byte(n>>8), byte(n)
	return nil
}

// WriteUint64Pos writes the supplied uint16 value to the Chunk payload buffer
// at the supplied index 'p'.
//
// The error 'io.EOF' is returned if the position specified is greater than the
// Chunk buffer size, or 'ErrLimit' if this position is greater than the set
// Chunk limit.
func (c *Chunk) WriteUint64Pos(p int, n uint64) error {
	if p >= c.Size() || p+7 >= c.Size() {
		return io.EOF
	}
	if c.Limit > 0 && (p >= c.Limit || p+7 >= c.Limit) {
		return ErrLimit
	}
	_ = c.buf[p+7]
	c.buf[p], c.buf[p+1], c.buf[p+2], c.buf[p+3] = byte(n>>56), byte(n>>48), byte(n>>40), byte(n>>32)
	c.buf[p+4], c.buf[p+5], c.buf[p+6], c.buf[p+7] = byte(n>>24), byte(n>>16), byte(n>>8), byte(n)
	return nil
}
