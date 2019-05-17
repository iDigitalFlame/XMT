package dio

import (
	"errors"
	"io"
)

const (
	// StringMaxLength is the max size a string can be to use the WriteString function. This size is the
	// limit of a uint16, which is used to track the size.
	StringMaxLength = 65535
)

var (
	// ErrStringTooLarge is returned when an operation of WriteString encounters a string larger than
	// 'StingMaxLength'. The string may be still written, but must use other methods, such as WriteByte.
	// Note that not using WriteString will not track string length.
	ErrStringTooLarge = errors.New("string too large")
)

// Writer is a basic interface that supports writing of all core Golang builtin types.
type Writer interface {
	Close() error
	Flush() error
	WriteBool(bool) error
	WriteInt8(int8) error
	WriteUint8(uint8) error
	WriteInt16(int16) error
	WriteInt32(int32) error
	WriteInt64(int64) error
	WriteUint16(uint16) error
	WriteUint32(uint32) error
	WriteUint64(uint64) error
	WriteByte([]byte) (int, error)
	WriteString(string) (int, error)
}
type writeFlusher interface {
	Flush() error
}

// streamWriter is an struct that wraps the Golang io.Writer interface. This allows for us to assign capabilities
// and functions to this interface, such as writing more complex types.
type streamWriter struct {
	buf [8]byte
	w   io.Writer
}

// NewWriter wraps a Writer around the specified io.Writer 'w'.
func NewWriter(w io.Writer) Writer {
	return &streamWriter{w: w}
}
func (w *streamWriter) Close() error {
	if c, ok := w.w.(io.WriteCloser); ok {
		return c.Close()
	}
	return nil
}
func (w *streamWriter) Flush() error {
	if c, ok := w.w.(writeFlusher); ok {
		return c.Flush()
	}
	return nil
}
func (w *streamWriter) WriteBool(b bool) error {
	if b {
		return w.WriteUint8(1)
	}
	return w.WriteUint8(0)
}
func (w *streamWriter) WriteInt8(i int8) error {
	return w.WriteUint8(uint8(i))
}
func (w *streamWriter) WriteUint8(i uint8) error {
	w.buf[0] = byte(i)
	n, err := w.w.Write(w.buf[:1])
	if err != nil {
		return err
	}
	if n < 1 {
		return io.ErrShortWrite
	}
	return nil
}
func (w *streamWriter) WriteInt16(i int16) error {
	return w.WriteUint16(uint16(i))
}
func (w *streamWriter) WriteInt32(i int32) error {
	return w.WriteUint32(uint32(i))
}
func (w *streamWriter) WriteInt64(i int64) error {
	return w.WriteUint64(uint64(i))
}
func (w *streamWriter) WriteUint16(i uint16) error {
	w.buf[0] = byte(i >> 8)
	w.buf[1] = byte(i)
	n, err := w.w.Write(w.buf[:2])
	if err != nil {
		return err
	}
	if n < 2 {
		return io.ErrShortWrite
	}
	return nil
}
func (w *streamWriter) WriteUint32(i uint32) error {
	w.buf[0] = byte(i >> 24)
	w.buf[1] = byte(i >> 16)
	w.buf[2] = byte(i >> 8)
	w.buf[3] = byte(i)
	n, err := w.w.Write(w.buf[:4])
	if err != nil {
		return err
	}
	if n < 4 {
		return io.ErrShortWrite
	}
	return nil
}
func (w *streamWriter) WriteUint64(i uint64) error {
	w.buf[0] = byte(i >> 56)
	w.buf[1] = byte(i >> 48)
	w.buf[2] = byte(i >> 40)
	w.buf[3] = byte(i >> 32)
	w.buf[4] = byte(i >> 24)
	w.buf[5] = byte(i >> 16)
	w.buf[6] = byte(i >> 8)
	w.buf[7] = byte(i)
	n, err := w.w.Write(w.buf[:8])
	if err != nil {
		return err
	}
	if n < 8 {
		return io.ErrShortWrite
	}
	return nil
}
func (w *streamWriter) WriteByte(b []byte) (int, error) {
	return w.w.Write(b)
}
func (w *streamWriter) WriteString(s string) (int, error) {
	l := len(s)
	if l > StringMaxLength {
		return 0, ErrStringTooLarge
	}
	if err := w.WriteUint16(uint16(l)); err != nil {
		return 0, err
	}
	b := make([]byte, l)
	n := copy(b, s)
	o, err := w.w.Write(b)
	if err != nil {
		return 0, err
	}
	if n != o {
		return 0, io.ErrShortWrite
	}
	return n, nil
}
