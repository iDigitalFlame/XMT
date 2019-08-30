package data

import (
	"io"
	"math"
)

const (
	// WriteStringSmall is the size value allowed for small strings
	// using the WriteString and WriteBytes functions.
	WriteStringSmall = 2 << 7
	// WriteStringLarge is the size value allowed for large strings
	// using the WriteString and WriteBytes functions.
	WriteStringLarge = 2 << 31
	// WriteStringMedium is the size value allowed for medium strings
	// using the WriteString and WriteBytes functions.
	WriteStringMedium = 2 << 15
)

type writerBase struct {
	w io.Writer
}

func (w *writerBase) Flush() error {
	if f, ok := w.w.(Flusher); ok {
		return f.Flush()
	}
	return nil
}
func (w *writerBase) Close() error {
	if c, ok := w.w.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// NewWriter creates a simple Writer struct from the base Writer
// provided.
func NewWriter(w io.Writer) Writer {
	return &writerBase{w: w}
}
func (w *writerBase) WriteInt(n int) error {
	return w.WriteUint64(uint64(n))
}
func (w *writerBase) small(b ...byte) error {
	_, err := w.w.Write(b)
	return err
}
func (w *writerBase) WriteUint(n uint) error {
	return w.WriteUint64(uint64(n))
}
func (w *writerBase) WriteInt8(n int8) error {
	return w.WriteUint8(uint8(n))
}
func (w *writerBase) WriteBool(n bool) error {
	if n {
		return w.WriteUint8(1)
	}
	return w.WriteUint8(0)
}
func (w *writerBase) WriteInt16(n int16) error {
	return w.WriteUint16(uint16(n))
}
func (w *writerBase) WriteInt32(n int32) error {
	return w.WriteUint32(uint32(n))
}
func (w *writerBase) WriteInt64(n int64) error {
	return w.WriteUint64(uint64(n))
}
func (w *writerBase) WriteUint8(n uint8) error {
	return w.small(byte(n))
}
func (w *writerBase) WriteBytes(b []byte) error {
	switch l := len(b); {
	case l == 0:
		return w.small(0)
	case l < WriteStringSmall:
		if err := w.WriteUint8(1); err != nil {
			return err
		}
		if err := w.WriteUint8(uint8(l)); err != nil {
			return err
		}
	case l < WriteStringMedium:
		if err := w.WriteUint8(3); err != nil {
			return err
		}
		if err := w.WriteUint16(uint16(l)); err != nil {
			return err
		}
	case l < WriteStringLarge:
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
	if _, err := w.w.Write(b); err != nil {
		return err
	}
	return nil
}
func (w *writerBase) WriteUint16(n uint16) error {
	return w.small(byte(n>>8), byte(n))
}
func (w *writerBase) WriteUint32(n uint32) error {
	return w.small(
		byte(n>>24), byte(n>>16), byte(n>>8), byte(n),
	)
}
func (w *writerBase) WriteUint64(n uint64) error {
	return w.small(
		byte(n>>56), byte(n>>48), byte(n>>40), byte(n>>32),
		byte(n>>24), byte(n>>16), byte(n>>8), byte(n),
	)
}
func (w *writerBase) WriteString(n string) error {
	return w.WriteBytes([]byte(n))
}
func (w *writerBase) Write(b []byte) (int, error) {
	return w.w.Write(b)
}
func (w *writerBase) WriteFloat32(n float32) error {
	return w.WriteUint32(math.Float32bits(n))
}
func (w *writerBase) WriteFloat64(n float64) error {
	return w.WriteUint64(math.Float64bits(n))
}
func (w *writerBase) WriteUTF8String(n string) error {
	return w.WriteBytes([]byte(n))
}
func (w *writerBase) WriteUTF16String(n string) error {
	switch l := len(n); {
	case l == 0:
		return w.small(0, 0)
	case l < WriteStringSmall:
		if err := w.WriteUint8(2); err != nil {
			return err
		}
		if err := w.WriteUint8(uint8(l)); err != nil {
			return err
		}
	case l < WriteStringMedium:
		if err := w.WriteUint8(4); err != nil {
			return err
		}
		if err := w.WriteUint16(uint16(l)); err != nil {
			return err
		}
	case l < WriteStringLarge:
		if err := w.WriteUint8(6); err != nil {
			return err
		}
		if err := w.WriteUint32(uint32(l)); err != nil {
			return err
		}
	default:
		if err := w.WriteUint8(8); err != nil {
			return err
		}
		if err := w.WriteUint64(uint64(l)); err != nil {
			return err
		}
	}
	for i := range n {
		w.small(byte(uint16(n[i])>>8), byte(n[i]))
	}
	return nil
}
