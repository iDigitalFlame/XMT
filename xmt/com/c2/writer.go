package c2

import (
	"math"

	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/data"
)

const (
	bufMaxSize   = int(^uint(0) >> 1)
	bufSizeSmall = 64
)

func (b *buffer) Reset() {
	b.r = nil
	b.w = nil
	b.wpos = 0
	b.rpos = 0
	b.buf = b.buf[:0]
}
func (b *buffer) Flush() error {
	return nil
}
func (b *buffer) Close() error {
	if b.wpos > 0 {
		b.buf = b.buf[:b.wpos]
	}
	return nil
}
func (w *wrapBuffer) Flush() error {
	if f, ok := w.w.(data.Flusher); ok {
		return f.Flush()
	}
	return nil
}
func (w *wrapBuffer) Close() error {
	if w.w != nil {
		return w.w.Close()
	}
	if w.r != nil {
		return w.r.Close()
	}
	return nil
}
func (b *buffer) Grow(n int) error {
	if n < 0 {
		return com.ErrInvalidIndex
	}
	m, err := b.grow(n)
	if err != nil {
		return err
	}
	b.buf = b.buf[:m]
	return nil
}
func (b *buffer) grow(n int) (int, error) {
	m := len(b.buf) - b.wpos
	if m == 0 && b.wpos != 0 {
		b.Reset()
	}
	if i, ok := b.reslice(n); ok {
		return i, nil
	}
	if b.buf == nil && n <= bufSizeSmall {
		b.buf = make([]byte, n, bufSizeSmall)
		return 0, nil
	}
	c := cap(b.buf)
	if n <= c/2-m {
		copy(b.buf, b.buf[b.wpos:])
	} else if c > bufMaxSize-c-n {
		return 0, com.ErrTooLarge
	} else {
		t, err := trySlice(2*c + n)
		if err != nil {
			return 0, err
		}
		copy(t, b.buf[b.wpos:])
		b.buf = t
	}
	b.wpos = 0
	b.buf = b.buf[:m+n]
	return m, nil
}
func trySlice(n int) (b []byte, err error) {
	defer func() {
		if recover() != nil {
			err = com.ErrTooLarge
		}
	}()
	return make([]byte, n), nil
}
func (w *wrapBuffer) WriteInt(n int) error {
	return w.WriteUint64(uint64(n))
}
func (w *wrapBuffer) small(b ...byte) error {
	_, err := w.Write(b)
	return err
}
func (b *buffer) reslice(n int) (int, bool) {
	if l := len(b.buf); n <= cap(b.buf)-l {
		b.buf = b.buf[:l+n]
		return l, true
	}
	return 0, false
}
func (w *wrapBuffer) WriteUint(n uint) error {
	return w.WriteUint64(uint64(n))
}
func (w *wrapBuffer) WriteInt8(n int8) error {
	return w.WriteUint8(uint8(n))
}
func (w *wrapBuffer) WriteBool(n bool) error {
	if n {
		return w.WriteUint8(1)
	}
	return w.WriteUint8(0)
}
func (b *buffer) Write(p []byte) (int, error) {
	m, ok := b.reslice(len(p))
	if !ok {
		var err error
		if m, err = b.grow(len(p)); err != nil {
			return 0, err
		}
	}
	return copy(b.buf[m:], p), nil
}
func (w *wrapBuffer) WriteInt16(n int16) error {
	return w.WriteUint16(uint16(n))
}
func (w *wrapBuffer) WriteInt32(n int32) error {
	return w.WriteUint32(uint32(n))
}
func (w *wrapBuffer) WriteInt64(n int64) error {
	return w.WriteUint64(uint64(n))
}
func (w *wrapBuffer) WriteUint8(n uint8) error {
	return w.small(byte(n))
}
func (w *wrapBuffer) WriteBytes(b []byte) error {
	switch l := len(b); {
	case l == 0:
		return w.small(0)
	case l < data.WriteStringSmall:
		if err := w.WriteUint8(1); err != nil {
			return err
		}
		if err := w.WriteUint8(uint8(l)); err != nil {
			return err
		}
	case l < data.WriteStringMedium:
		if err := w.WriteUint8(3); err != nil {
			return err
		}
		if err := w.WriteUint16(uint16(l)); err != nil {
			return err
		}
	case l < data.WriteStringLarge:
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
	if _, err := w.Write(b); err != nil {
		return err
	}
	return nil
}
func (w *wrapBuffer) WriteUint16(n uint16) error {
	return w.small(byte(n>>8), byte(n))
}
func (w *wrapBuffer) WriteUint32(n uint32) error {
	return w.small(
		byte(n>>24), byte(n>>16), byte(n>>8), byte(n),
	)
}
func (w *wrapBuffer) WriteUint64(n uint64) error {
	return w.small(
		byte(n>>56), byte(n>>48), byte(n>>40), byte(n>>32),
		byte(n>>24), byte(n>>16), byte(n>>8), byte(n),
	)
}
func (w *wrapBuffer) WriteString(n string) error {
	return w.WriteBytes([]byte(n))
}
func (w *wrapBuffer) Write(b []byte) (int, error) {
	return w.w.Write(b)
}
func (w *wrapBuffer) WriteFloat32(n float32) error {
	return w.WriteUint32(math.Float32bits(n))
}
func (w *wrapBuffer) WriteFloat64(n float64) error {
	return w.WriteUint64(math.Float64bits(n))
}
func (w *wrapBuffer) WriteUTF8String(n string) error {
	return w.WriteBytes([]byte(n))
}
func (w *wrapBuffer) WriteUTF16String(n string) error {
	switch l := len(n); {
	case l == 0:
		return w.small(0, 0)
	case l < data.WriteStringSmall:
		if err := w.WriteUint8(2); err != nil {
			return err
		}
		if err := w.WriteUint8(uint8(l)); err != nil {
			return err
		}
	case l < data.WriteStringMedium:
		if err := w.WriteUint8(4); err != nil {
			return err
		}
		if err := w.WriteUint16(uint16(l)); err != nil {
			return err
		}
	case l < data.WriteStringLarge:
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
