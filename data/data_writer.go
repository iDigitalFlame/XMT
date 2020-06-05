package data

import (
	"io"
	"math"
)

type writer struct {
	_ [0]func()
	w io.Writer
}

func (w *writer) Flush() error {
	if f, ok := w.w.(Flusher); ok {
		return f.Flush()
	}
	return nil
}
func (w *writer) Close() error {
	if c, ok := w.w.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// NewWriter creates a simple Writer struct from the base Writer provided.
func NewWriter(w io.Writer) Writer {
	return &writer{w: w}
}
func (w *writer) WriteInt(n int) error {
	return w.WriteUint64(uint64(n))
}
func (w *writer) small(b ...byte) error {
	_, err := w.w.Write(b)
	return err
}
func (w *writer) WriteUint(n uint) error {
	return w.WriteUint64(uint64(n))
}
func (w *writer) WriteInt8(n int8) error {
	return w.WriteUint8(uint8(n))
}
func (w *writer) WriteBool(b bool) error {
	if b {
		return w.WriteUint8(1)
	}
	return w.WriteUint8(0)
}
func (w *writer) WriteInt16(n int16) error {
	return w.WriteUint16(uint16(n))
}
func (w *writer) WriteInt32(n int32) error {
	return w.WriteUint32(uint32(n))
}
func (w *writer) WriteInt64(n int64) error {
	return w.WriteUint64(uint64(n))
}
func (w *writer) WriteUint8(n uint8) error {
	return w.small(byte(n))
}
func (w *writer) WriteBytes(b []byte) error {
	switch l := uint64(len(b)); {
	case l == 0:
		return w.small(0)
	case l < DataLimitSmall:
		if err := w.WriteUint8(1); err != nil {
			return err
		}
		if err := w.WriteUint8(uint8(l)); err != nil {
			return err
		}
	case l < DataLimitMedium:
		if err := w.WriteUint8(3); err != nil {
			return err
		}
		if err := w.WriteUint16(uint16(l)); err != nil {
			return err
		}
	case l < DataLimitLarge:
		if err := w.WriteUint8(5); err != nil {
			return err
		}
		if err := w.WriteUint32(uint32(l)); err != nil {
			return err
		}
	default:
		if err := w.WriteUint8(7); err != nil {
			return err
		}
		if err := w.WriteUint64(uint64(l)); err != nil {
			return err
		}
	}
	_, err := w.w.Write(b)
	return err
}
func (w *writer) WriteUint16(n uint16) error {
	return w.small(byte(n>>8), byte(n))
}
func (w *writer) WriteUint32(n uint32) error {
	return w.small(byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}
func (w *writer) WriteUint64(n uint64) error {
	return w.small(
		byte(n>>56), byte(n>>48), byte(n>>40), byte(n>>32),
		byte(n>>24), byte(n>>16), byte(n>>8), byte(n),
	)
}
func (w *writer) WriteString(s string) error {
	return w.WriteBytes([]byte(s))
}
func (w *writer) Write(b []byte) (int, error) {
	return w.w.Write(b)
}
func (w *writer) WriteFloat32(f float32) error {
	return w.WriteUint32(math.Float32bits(f))
}
func (w *writer) WriteFloat64(f float64) error {
	return w.WriteUint64(math.Float64bits(f))
}
