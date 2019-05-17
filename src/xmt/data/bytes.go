package data

import (
	"bytes"
	"errors"
	"io"
	"math"

	"../../xmt/util"
)

const (
	// BytesBufSize is the default byte buffer size created by NewByteWriter if a length is not specified.
	BytesBufSize = 256
	// ByteStrSizeMax is the maximum length of a string converted into bytes via the Write and Read string function.
	ByteStrSizeMax = 1073741824
	// ByteStrSizeStep is the maximum length of a string before having to write a secondary byte to the buffer
	// for describing string length.
	ByteStrSizeStep = 128
)

var (
	// ByteStrUTF16 determines if all string and byte converting funtions use UTF8 or UTF16.
	// By default, most functions will use UTF8 (one byte per char), versus UTF16 (two bytes per char). Set this
	// to 'true' if support for non-western or English character sets are needed. This can be set on a per-reader/writer
	// basis also.
	ByteStrUTF16 = false
	// ErrStringInvalid will be returned by the ReadString function if the reader returns an invalid string value
	// of less than zero.
	ErrStringInvalid = errors.New("bytes: received invalid string size")
	// ErrStringTooLarge will be returned by the WriteString function if the supplied string is too
	// large to be supported using a 2^30 size integer.
	ErrStringTooLarge = errors.New("bytes: string size is too large")
)

// ByteReader is a struct that allows for reading a byte array seamlessly.
// This struct also gives support for reading strings.
type ByteReader struct {
	UTF16      bool
	byteIndex  int
	byteBuffer []byte
}

// ByteWriter allows for writing to a byte array seamlessly.
// This struct also gives support for writing Strings.
type ByteWriter struct {
	UTF16      bool
	byteBuffer *bytes.Buffer
}

// StreamMarshaller is an interface that defines support for Marshalling data to and from a data stream
// into a struct or other type of data.
type StreamMarshaller interface {
	MarshalStream(io.Writer) error
	UnmarshalStream(io.Reader) error
}

// BinaryMarshaller is an interface that defines support for Marshalling data to and from a binary format
// into a struct or other type of data.
type BinaryMarshaller interface {
	UnmarshalBinary([]byte) error
	MarshalBinary() ([]byte, error)
}

// NewByteWriter creates a ByteWriter struct with a default size of the 'BytesBufSize' constant.
func NewByteWriter() *ByteWriter {
	return NewByteWriterLen(BytesBufSize)
}

func (w *ByteWriter) Close() error {
	return nil
}

func (r *ByteReader) Close() error {
	return nil
}

func (w *ByteWriter) Bytes() []byte {
	return w.byteBuffer.Bytes()
}

func (r *ByteReader) Bytes() []byte {
	if r.byteIndex >= len(r.byteBuffer) {
		return nil
	}
	return r.byteBuffer[r.byteIndex:]
}

func NewByteReader(b []byte) *ByteReader {
	return &ByteReader{UTF16: ByteStrUTF16, byteBuffer: b}
}

func NewByteWriterLen(n int) *ByteWriter {
	return &ByteWriter{UTF16: ByteStrUTF16, byteBuffer: bytes.NewBuffer(make([]byte, 0, n))}
}

func (r *ByteReader) ReadUint8() (uint8, error) {
	if r.byteIndex >= len(r.byteBuffer) {
		return 0, io.EOF
	}
	b := r.byteBuffer[r.byteIndex]
	r.byteIndex++
	return b, nil
}

func (r *ByteReader) Read(b []byte) (int, error) {
	if r.byteIndex >= len(r.byteBuffer) {
		return 0, io.EOF
	}
	n := int(math.Min(float64(len(b)), float64(len(r.byteBuffer)-r.byteIndex)))
	copy(b, r.byteBuffer[r.byteIndex:r.byteIndex+n])
	return n, nil
}

func (w *ByteWriter) Write(b []byte) (int, error) {
	return w.byteBuffer.Write(b)
}

func (r *ByteReader) ReadUint16() (uint16, error) {
	if r.byteIndex+1 >= len(r.byteBuffer) {
		return 0, io.EOF
	}
	b := uint16(uint16(r.byteBuffer[r.byteIndex+1]) | uint16(r.byteBuffer[r.byteIndex])<<8)
	r.byteIndex += 2
	return b, nil
}

func (r *ByteReader) ReadString() (string, error) {
	if r.byteIndex >= len(r.byteBuffer) {
		return util.Empty, io.EOF
	}
	n := 0
	if (r.byteBuffer[r.byteIndex] & (1 << 7)) > 0 {
		if r.byteIndex+1 >= len(r.byteBuffer) {
			return util.Empty, io.EOF
		}
		n = int(uint16(r.byteBuffer[r.byteIndex+1]) | uint16(byte(int(r.byteBuffer[r.byteIndex])&^(1<<7)))<<8)
		r.byteIndex += 2
	} else {
		n = int(r.byteBuffer[r.byteIndex])
		r.byteIndex++
	}
	if n <= 0 {
		return util.Empty, ErrStringInvalid
	}
	if r.byteIndex+n >= len(r.byteBuffer) {
		return util.Empty, io.EOF
	}
	s := make([]byte, n)
	for v := 0; v < n; r.byteIndex++ {
		s[v] = r.byteBuffer[r.byteIndex]
		if r.UTF16 {
			s[v+1] = r.byteBuffer[r.byteIndex+1]
			r.byteIndex++
			v++
		}
		v++
	}
	return string(s), nil
}

func (w *ByteWriter) WriteUint8(n uint8) (int, error) {
	return 1, w.byteBuffer.WriteByte(n)
}

func (w *ByteWriter) WriteUint16(n uint16) (int, error) {
	if _, err := w.byteBuffer.Write([]byte{byte(n >> 8), byte(n)}); err != nil {
		return 0, err
	}
	return 2, nil
}

func (w *ByteWriter) WriteString(s string) (int, error) {
	n := len(s)
	if ByteStrUTF16 {
		n *= 2
	}
	if n > ByteStrSizeMax {
		return 0, ErrStringTooLarge
	}
	if n > ByteStrSizeStep {
		if _, err := w.byteBuffer.Write([]byte{byte(n>>8) | (1 << 7), byte(n)}); err != nil {
			return 0, err
		}
		n += 2
	} else {
		if err := w.byteBuffer.WriteByte(byte(n)); err != nil {
			return 0, err
		}
		n++
	}
	for v := 0; v < len(s); v++ {
		if w.UTF16 {
			if _, err := w.byteBuffer.Write([]byte{byte(uint16(s[v]) >> 8), byte(s[v])}); err != nil {
				return 0, err
			}
		} else {
			if err := w.byteBuffer.WriteByte(byte(s[v])); err != nil {
				return 0, err
			}
		}
	}
	return n, nil
}
