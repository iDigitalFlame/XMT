package com

import (
	"math"

	data "github.com/iDigitalFlame/xmt/xmt-data"
)

const (
	bufMaxSize   = int(^uint(0) >> 1)
	bufSizeSmall = 64
)

// Flush does nothing for the Packet struct. Just
// here for compatibility.
func (Packet) Flush() error {
	return nil
}

// Flush will push out any Packets to the underlying Packet
// Writer (if not nil).
func (s *Stream) Flush() error {
	s.flushPackets(false)
	return nil
}

// Close will truncate the writing stream if this Packet
// has been written to. This will allow the packet payload to be
// uniform to the data written to it.
func (p *Packet) Close() error {
	if p.stream != nil {
		p.stream.Clear()
	}
	if p.wpos > 0 {
		p.buf = p.buf[:p.wpos]
	}
	return nil
}

// Grow grows the payload buffer's capacity, if necessary, to guarantee space for
// another n bytes.
func (p *Packet) Grow(n int) error {
	if n < 0 {
		return ErrInvalidIndex
	}
	m, err := p.grow(n)
	if err != nil {
		return err
	}
	p.buf = p.buf[:m]
	return nil
}

// WriteInt writes the supplied value to the Packet payload buffer.
func (p *Packet) WriteInt(n int) error {
	return p.WriteUint64(uint64(n))
}

// WriteInt writes the supplied value to the Stream payload buffer.
func (s *Stream) WriteInt(n int) error {
	return s.WriteUint64(uint64(n))
}
func (p *Packet) small(b ...byte) error {
	_, err := p.Write(b)
	return err
}
func (s *Stream) small(b ...byte) error {
	_, err := s.Write(b)
	return err
}

// WriteUint writes the supplied value to the Packet payload buffer.
func (p *Packet) WriteUint(n uint) error {
	return p.WriteUint64(uint64(n))
}

// WriteUint writes the supplied value to the Stream payload buffer.
func (s *Stream) WriteUint(n uint) error {
	return s.WriteUint64(uint64(n))
}

// WriteInt8 writes the supplied value to the Packet payload buffer.
func (p *Packet) WriteInt8(n int8) error {
	return p.WriteUint8(uint8(n))
}

// WriteInt8 writes the supplied value to the Stream payload buffer.
func (s *Stream) WriteInt8(n int8) error {
	return s.WriteUint8(uint8(n))
}

// WriteBool writes the supplied value to the Packet payload buffer.
func (p *Packet) WriteBool(n bool) error {
	if n {
		return p.WriteUint8(1)
	}
	return p.WriteUint8(0)
}

// WriteBool writes the supplied value to the Stream payload buffer.
func (s *Stream) WriteBool(n bool) error {
	if n {
		return s.WriteUint8(1)
	}
	return s.WriteUint8(0)
}
func (p *Packet) grow(n int) (int, error) {
	m := len(p.buf) - p.wpos
	if m == 0 && p.wpos != 0 {
		p.Reset()
	}
	if i, ok := p.reslice(n); ok {
		return i, nil
	}
	if p.buf == nil && n <= bufSizeSmall {
		p.buf = make([]byte, n, bufSizeSmall)
		return 0, nil
	}
	c := cap(p.buf)
	if n <= c/2-m {
		copy(p.buf, p.buf[p.wpos:])
	} else if c > bufMaxSize-c-n {
		return 0, ErrTooLarge
	} else {
		b, err := trySlice(2*c + n)
		if err != nil {
			return 0, err
		}
		copy(b, p.buf[p.wpos:])
		p.buf = b
	}
	p.wpos = 0
	p.buf = p.buf[:m+n]
	return m, nil
}
func trySlice(n int) (b []byte, err error) {
	defer func() {
		if recover() != nil {
			err = ErrTooLarge
		}
	}()
	return make([]byte, n), nil
}

// WriteInt16 writes the supplied value to the Packet payload buffer.
func (p *Packet) WriteInt16(n int16) error {
	return p.WriteUint16(uint16(n))
}

// WriteInt16 writes the supplied value to the Stream payload buffer.
func (s *Stream) WriteInt16(n int16) error {
	return s.WriteUint16(uint16(n))
}

// WriteInt32 writes the supplied value to the Packet payload buffer.
func (p *Packet) WriteInt32(n int32) error {
	return p.WriteUint32(uint32(n))
}

// WriteInt32 writes the supplied value to the Stream payload buffer.
func (s *Stream) WriteInt32(n int32) error {
	return s.WriteUint32(uint32(n))
}

// WriteInt64 writes the supplied value to the Packet payload buffer.
func (p *Packet) WriteInt64(n int64) error {
	return p.WriteUint64(uint64(n))
}

// WriteInt64 writes the supplied value to the Stream payload buffer.
func (s *Stream) WriteInt64(n int64) error {
	return s.WriteUint64(uint64(n))
}

// WriteUint8 writes the supplied value to the Packet payload buffer.
func (p *Packet) WriteUint8(n uint8) error {
	return p.small(byte(n))
}

// WriteUint8 writes the supplied value to the Stream payload buffer.
func (s *Stream) WriteUint8(n uint8) error {
	return s.small(byte(n))
}

// WriteBytes writes the supplied value to the Packet payload buffer.
func (p *Packet) WriteBytes(b []byte) error {
	if b == nil {
		return p.small(0)
	}
	switch l := len(b); {
	case l == 0:
		return p.small(0)
	case l < data.DataLimitSmall:
		if err := p.WriteUint8(1); err != nil {
			return err
		}
		if err := p.WriteUint8(uint8(l)); err != nil {
			return err
		}
	case l < data.DataLimitMedium:
		if err := p.WriteUint8(3); err != nil {
			return err
		}
		if err := p.WriteUint16(uint16(l)); err != nil {
			return err
		}
	case l < data.DataLimitLarge:
		if err := p.WriteUint8(5); err != nil {
			return err
		}
		if err := p.WriteUint32(uint32(l)); err != nil {
			return err
		}
	default:
		if err := p.WriteUint8(7); err != nil {
			return err
		}
		if err := p.WriteUint64(uint64(l)); err != nil {
			return err
		}
	}
	_, err := p.Write(b)
	return err
}

// WriteBytes writes the supplied value to the Stream payload buffer.
func (s *Stream) WriteBytes(b []byte) error {
	if b == nil {
		return s.small(0)
	}
	switch l := len(b); {
	case l == 0:
		return s.small(0)
	case l < data.DataLimitSmall:
		if err := s.WriteUint8(1); err != nil {
			return err
		}
		if err := s.WriteUint8(uint8(l)); err != nil {
			return err
		}
	case l < data.DataLimitMedium:
		if err := s.WriteUint8(3); err != nil {
			return err
		}
		if err := s.WriteUint16(uint16(l)); err != nil {
			return err
		}
	case l < data.DataLimitLarge:
		if err := s.WriteUint8(5); err != nil {
			return err
		}
		if err := s.WriteUint32(uint32(l)); err != nil {
			return err
		}
	default:
		if err := s.WriteUint8(7); err != nil {
			return err
		}
		if err := s.WriteUint64(uint64(l)); err != nil {
			return err
		}
	}
	_, err := s.Write(b)
	return err
}
func (p *Packet) reslice(n int) (int, bool) {
	if l := len(p.buf); n <= cap(p.buf)-l {
		p.buf = p.buf[:l+n]
		return l, true
	}
	return 0, false
}

// WriteUint16 writes the supplied value to the Packet payload buffer.
func (p *Packet) WriteUint16(n uint16) error {
	return p.small(byte(n>>8), byte(n))
}

// WriteUint16 writes the supplied value to the Stream payload buffer.
func (s *Stream) WriteUint16(n uint16) error {
	return s.small(byte(n>>8), byte(n))
}

// WriteUint32 writes the supplied value to the Packet payload buffer.
func (p *Packet) WriteUint32(n uint32) error {
	return p.small(
		byte(n>>24), byte(n>>16), byte(n>>8), byte(n),
	)
}

// WriteUint32 writes the supplied value to the Stream payload buffer.
func (s *Stream) WriteUint32(n uint32) error {
	return s.small(
		byte(n>>24), byte(n>>16), byte(n>>8), byte(n),
	)
}

// WriteUint64 writes the supplied value to the Packet payload buffer.
func (p *Packet) WriteUint64(n uint64) error {
	return p.small(
		byte(n>>56), byte(n>>48), byte(n>>40), byte(n>>32),
		byte(n>>24), byte(n>>16), byte(n>>8), byte(n),
	)
}

// WriteUint64 writes the supplied value to the Stream payload buffer.
func (s *Stream) WriteUint64(n uint64) error {
	return s.small(
		byte(n>>56), byte(n>>48), byte(n>>40), byte(n>>32),
		byte(n>>24), byte(n>>16), byte(n>>8), byte(n),
	)
}

// WriteString writes the supplied value to the Packet payload buffer.
func (p *Packet) WriteString(n string) error {
	return p.WriteBytes([]byte(n))
}

// WriteString writes the supplied value to the Stream payload buffer.
func (s *Stream) WriteString(n string) error {
	return s.WriteBytes([]byte(n))
}

// Write appends the contents of p to the buffer, growing the buffer as
// needed. If the buffer becomes too large, Write will return ErrTooLarge.
func (p *Packet) Write(b []byte) (int, error) {
	m, ok := p.reslice(len(b))
	if !ok {
		var err error
		if m, err = p.grow(len(b)); err != nil {
			return 0, err
		}
	}
	return copy(p.buf[m:], b), nil
}

// WriteFloat32 writes the supplied value to the Packet payload buffer.
func (p *Packet) WriteFloat32(n float32) error {
	return p.WriteUint32(math.Float32bits(n))
}

// WriteFloat32 writes the supplied value to the Stream payload buffer.
func (s *Stream) WriteFloat32(n float32) error {
	return s.WriteUint32(math.Float32bits(n))
}

// WriteFloat64 writes the supplied value to the Packet payload buffer.
func (p *Packet) WriteFloat64(n float64) error {
	return p.WriteUint64(math.Float64bits(n))
}

// WriteFloat64 writes the supplied value to the Stream payload buffer.
func (s *Stream) WriteFloat64(n float64) error {
	return s.WriteUint64(math.Float64bits(n))
}
