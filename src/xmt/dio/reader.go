package dio

import (
	"fmt"
	"io"
)

// Reader is a basic interface that supports all types of read functions of the core Golang
// builtin types. These functions accept pointers to allow for easier usage and more
// fulent operation.
type Reader interface {
	ReadBool(*bool) error
	ReadInt8(*int8) error
	ReadUint8(*uint8) error
	ReadInt16(*int16) error
	ReadInt32(*int32) error
	ReadInt64(*int64) error
	ReadUint16(*uint16) error
	ReadUint32(*uint32) error
	ReadUint64(*uint64) error
	ReadString(*string) error
	ReadBytes([]byte) (int, error)
}

// byteReader is a struct that implements the Reader interface. This struct reads all bytes from
// a supplied byte array.
type byteReader struct {
	byteBuffer   []byte
	bytePosition int
}

// streamReader is an struct that wraps the Golang io.Reader interface. This allows for us to assign capabilities
// and functions to this interface, such as reading more complex types.
type streamReader struct {
	buf [8]byte
	r   io.Reader
}

// NewReader wraps a Reader around the specified io.Reader 'r'.
func NewReader(r io.Reader) Reader {
	return &streamReader{r: r}
}

// NewByteReader returns a Reader backed by the supplied byte array 'b'.
func NewByteReader(b []byte) Reader {
	return &byteReader{byteBuffer: b}
}
func (r *byteReader) ReadBool(b *bool) error {
	if r.bytePosition > len(r.byteBuffer) {
		return io.EOF
	}
	*b = r.byteBuffer[r.bytePosition] == 1
	r.bytePosition++
	return nil
}
func (r *byteReader) ReadInt8(i *int8) error {
	if r.bytePosition > len(r.byteBuffer) {
		return io.EOF
	}
	*i = int8(r.byteBuffer[r.bytePosition])
	r.bytePosition++
	return nil
}
func (r *streamReader) ReadBool(b *bool) error {
	n, err := r.r.Read(r.buf[:1])
	if err != nil {
		return err
	}
	if n < 1 {
		return io.EOF
	}
	*b = r.buf[0] == 1
	return nil
}
func (r *streamReader) ReadInt8(i *int8) error {
	n, err := r.r.Read(r.buf[:1])
	if err != nil {
		return err
	}
	if n < 1 {
		return io.EOF
	}
	*i = int8(r.buf[0])
	return nil
}
func (r *byteReader) ReadUint8(i *uint8) error {
	if r.bytePosition > len(r.byteBuffer) {
		return io.EOF
	}
	*i = uint8(r.byteBuffer[r.bytePosition])
	r.bytePosition++
	return nil
}
func (r *byteReader) ReadInt16(i *int16) error {
	if (r.bytePosition + 2) > len(r.byteBuffer) {
		return io.EOF
	}
	*i = int16(int16(r.byteBuffer[r.bytePosition+1]) | int16(r.byteBuffer[r.bytePosition])<<8)
	r.bytePosition += 2
	return nil
}
func (r *byteReader) ReadInt32(i *int32) error {
	if (r.bytePosition + 4) > len(r.byteBuffer) {
		return io.EOF
	}
	*i = int32(
		int32(r.byteBuffer[r.bytePosition+3]) | int32(r.byteBuffer[r.bytePosition+2])<<8 |
			int32(r.byteBuffer[r.bytePosition+1])<<16 | int32(r.byteBuffer[r.bytePosition])<<24,
	)
	r.bytePosition += 4
	return nil
}
func (r *byteReader) ReadInt64(i *int64) error {
	if (r.bytePosition + 8) > len(r.byteBuffer) {
		return io.EOF
	}
	*i = int64(
		int64(r.byteBuffer[r.bytePosition+7]) | int64(r.byteBuffer[r.bytePosition+6])<<8 |
			int64(r.byteBuffer[r.bytePosition+5])<<16 | int64(r.byteBuffer[r.bytePosition+4])<<24 |
			int64(r.byteBuffer[r.bytePosition+3])<<32 | int64(r.byteBuffer[r.bytePosition+2])<<40 |
			int64(r.byteBuffer[r.bytePosition+1])<<48 | int64(r.byteBuffer[r.bytePosition])<<56,
	)
	r.bytePosition += 8
	return nil
}
func (r *streamReader) ReadUint8(i *uint8) error {
	n, err := r.r.Read(r.buf[:1])
	if err != nil {
		return err
	}
	if n < 1 {
		return io.EOF
	}
	*i = uint8(r.buf[0])
	return nil
}
func (r *streamReader) ReadInt16(i *int16) error {
	n, err := r.r.Read(r.buf[:2])
	if err != nil {
		return err
	}
	if n < 2 {
		return io.EOF
	}
	*i = int16(int16(r.buf[1]) | int16(r.buf[0])<<8)
	return nil
}
func (r *streamReader) ReadInt32(i *int32) error {
	n, err := r.r.Read(r.buf[:4])
	if err != nil {
		return err
	}
	if n < 4 {
		return io.EOF
	}
	*i = int32(int32(r.buf[3]) | int32(r.buf[2])<<8 | int32(r.buf[1])<<16 | int32(r.buf[0])<<24)
	return nil
}
func (r *streamReader) ReadInt64(i *int64) error {
	n, err := r.r.Read(r.buf[:8])
	if err != nil {
		return err
	}
	if n < 8 {
		return io.EOF
	}
	*i = int64(
		int64(r.buf[7]) | int64(r.buf[6])<<8 | int64(r.buf[5])<<16 | int64(r.buf[4])<<24 |
			int64(r.buf[3])<<32 | int64(r.buf[2])<<40 | int64(r.buf[1])<<48 | int64(r.buf[0])<<56,
	)
	return nil
}
func (r *byteReader) ReadUint16(i *uint16) error {
	if (r.bytePosition + 2) > len(r.byteBuffer) {
		return io.EOF
	}
	*i = uint16(uint16(r.byteBuffer[r.bytePosition+1]) | uint16(r.byteBuffer[r.bytePosition])<<8)
	r.bytePosition += 2
	return nil
}
func (r *byteReader) ReadUint32(i *uint32) error {
	if (r.bytePosition + 4) > len(r.byteBuffer) {
		return io.EOF
	}
	*i = uint32(
		uint32(r.byteBuffer[r.bytePosition+3]) | uint32(r.byteBuffer[r.bytePosition+2])<<8 |
			uint32(r.byteBuffer[r.bytePosition+1])<<16 | uint32(r.byteBuffer[r.bytePosition])<<24,
	)
	r.bytePosition += 4
	return nil
}
func (r *byteReader) ReadUint64(i *uint64) error {
	if (r.bytePosition + 8) > len(r.byteBuffer) {
		return io.EOF
	}
	*i = uint64(
		uint64(r.byteBuffer[r.bytePosition+7]) | uint64(r.byteBuffer[r.bytePosition+6])<<8 |
			uint64(r.byteBuffer[r.bytePosition+5])<<16 | uint64(r.byteBuffer[r.bytePosition+4])<<24 |
			uint64(r.byteBuffer[r.bytePosition+3])<<32 | uint64(r.byteBuffer[r.bytePosition+2])<<40 |
			uint64(r.byteBuffer[r.bytePosition+1])<<48 | uint64(r.byteBuffer[r.bytePosition])<<56,
	)
	r.bytePosition += 8
	return nil
}
func (r *byteReader) ReadString(s *string) error {
	if (r.bytePosition + 2) > len(r.byteBuffer) {
		return io.EOF
	}
	l := uint16(uint16(r.byteBuffer[r.bytePosition+1]) | uint16(r.byteBuffer[r.bytePosition])<<8)
	r.bytePosition += 2
	if l < 0 {
		return io.ErrShortBuffer
	}
	if l == 0 {
		*s = ""
	} else {
		if r.bytePosition > len(r.byteBuffer) {
			return io.EOF
		}
		b := make([]byte, l)
		n := copy(b, r.byteBuffer[r.bytePosition:])
		if uint16(n) != l {
			return io.EOF
		}
		*s = string(b)
		r.bytePosition += n
	}
	fmt.Printf("left: %d\n", r.bytePosition)
	return nil
}
func (r *streamReader) ReadUint16(i *uint16) error {
	n, err := r.r.Read(r.buf[:2])
	if err != nil {
		return err
	}
	if n < 2 {
		return io.EOF
	}
	*i = uint16(uint16(r.buf[1]) | uint16(r.buf[0])<<8)
	return nil
}
func (r *streamReader) ReadUint32(i *uint32) error {
	n, err := r.r.Read(r.buf[:4])
	if err != nil {
		return err
	}
	if n < 4 {
		return io.EOF
	}
	*i = uint32(uint32(r.buf[3]) | uint32(r.buf[2])<<8 | uint32(r.buf[1])<<16 | uint32(r.buf[0])<<24)
	return nil
}
func (r *streamReader) ReadUint64(i *uint64) error {
	n, err := r.r.Read(r.buf[:8])
	if err != nil {
		return err
	}
	if n < 8 {
		return io.EOF
	}
	*i = uint64(
		uint64(r.buf[7]) | uint64(r.buf[6])<<8 | uint64(r.buf[5])<<16 | uint64(r.buf[4])<<24 |
			uint64(r.buf[3])<<32 | uint64(r.buf[2])<<40 | uint64(r.buf[1])<<48 | uint64(r.buf[0])<<56,
	)
	return nil
}
func (r *streamReader) ReadString(s *string) error {
	n, err := r.r.Read(r.buf[:2])
	if err != nil {
		return err
	}
	if n < 2 {
		return io.EOF
	}
	l := uint16(uint16(r.buf[1]) | uint16(r.buf[0])<<8)
	if l < 0 {
		return io.ErrShortBuffer
	}
	if l == 0 {
		*s = ""
	} else {
		b := make([]byte, l)
		n, err := r.r.Read(b)
		if err != nil {
			return err
		}
		if uint16(n) != l {
			return io.EOF
		}
		*s = string(b)
	}
	return nil
}
func (r *byteReader) ReadBytes(b []byte) (int, error) {
	if r.bytePosition > len(r.byteBuffer) {
		return 0, io.EOF
	}
	n := copy(b, r.byteBuffer[r.bytePosition:])
	r.bytePosition += n
	return n, nil
}
func (r *streamReader) ReadBytes(b []byte) (int, error) {
	return r.r.Read(b)
}
