package data

import "io"

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
	v, err := w.w.Write([]byte{byte(n)})
	if err == nil && v != 1 {
		return io.ErrShortWrite
	}
	return err
}
func (w *writer) WriteBytes(b []byte) error {
	switch l := uint64(len(b)); {
	case l == 0:
		v, err := w.w.Write([]byte{0})
		if err == nil && v != 1 {
			return io.ErrShortWrite
		}
		return err
	case l < LimitSmall:
		if v, err := w.w.Write([]byte{1, byte(l)}); err != nil {
			return err
		} else if v != 2 {
			return io.ErrShortWrite
		}
	case l < LimitMedium:
		if v, err := w.w.Write([]byte{3, byte(l >> 8), byte(l)}); err != nil {
			return err
		} else if v != 3 {
			return io.ErrShortWrite
		}
	case l < LimitLarge:
		if v, err := w.w.Write([]byte{5, byte(l >> 24), byte(l >> 16), byte(l >> 8), byte(l)}); err != nil {
			return err
		} else if v != 5 {
			return io.ErrShortWrite
		}
	default:
		if v, err := w.w.Write([]byte{
			7, byte(l >> 56), byte(l >> 48), byte(l >> 40), byte(l >> 32),
			byte(l >> 24), byte(l >> 16), byte(l >> 8), byte(l),
		}); err != nil {
			return nil
		} else if v != 9 {
			return io.ErrShortWrite
		}
	}
	_, err := w.w.Write(b)
	return err
}
func (w *writer) WriteUint16(n uint16) error {
	v, err := w.w.Write([]byte{byte(n >> 8), byte(n)})
	if err == nil && v != 2 {
		return io.ErrShortWrite
	}
	return err
}
func (w *writer) WriteUint32(n uint32) error {
	v, err := w.w.Write([]byte{byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)})
	if err == nil && v != 4 {
		return io.ErrShortWrite
	}
	return err
}
func (w *writer) WriteUint64(n uint64) error {
	v, err := w.w.Write([]byte{
		byte(n >> 56), byte(n >> 48), byte(n >> 40), byte(n >> 32),
		byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n),
	})
	if err == nil && v != 8 {
		return io.ErrShortWrite
	}
	return err
}
func (w *writer) WriteString(s string) error {
	return w.WriteBytes([]byte(s))
}
func (w *writer) Write(b []byte) (int, error) {
	return w.w.Write(b)
}
func (w *writer) WriteFloat32(f float32) error {
	return w.WriteUint32(float32ToInt(f))
}
func (w *writer) WriteFloat64(f float64) error {
	return w.WriteUint64(float64ToInt(f))
}
