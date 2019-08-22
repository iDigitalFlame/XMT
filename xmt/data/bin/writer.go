package bin

import (
	"bytes"

	"golang.org/x/xerrors"
)

const (
	// Size is the default byte buffer size created by NewByteWriter if a length is not specified.
	Size = 256
	// SizeMax is the maximum length of a string converted into bytes via the Write and Read string function.
	SizeMax = 1073741824
	// StrSizeStep is the maximum length of a string before having to write a secondary byte to the buffer
	// for describing string length.
	StrSizeStep = 128
)

var (
	// DefaultUTF16 determines if all string and byte converting funtions use UTF8 or UTF16.
	// By default, most functions will use UTF8 (one byte per char), versus UTF16 (two bytes per char). Set this
	// to 'true' if support for non-western or English character sets are needed. This can be set on a per-reader/writer
	// basis also.
	DefaultUTF16 = false
	// ErrStringInvalid will be returned by the ReadString function if the reader returns an invalid string value
	// of less than zero.
	ErrStringInvalid = xerrors.New("bytes: received invalid string size")
	// ErrStringTooLarge will be returned by the WriteString function if the supplied string is too
	// large to be supported using a 2^30 size integer.
	ErrStringTooLarge = xerrors.New("bytes: string size is too large")
)

// Writer allows for writing to a byte array seamlessly.
// This struct also gives support for writing Strings.
type Writer struct {
	UTF16      bool

	buf *bytes.Buffer
}

// NewByteWriter creates a Writer struct with a default size of the 'BytesBufSize' constant.
func NewByteWriter() *Writer {
	return NewByteWriterLen(BytesBufSize)
}

func (w *Writer) Close() error {
	return nil
}

func (w *Writer) Bytes() []byte {
	return w.buf.Bytes()
}

func NewByteWriterLen(n int) *Writer {
	return &Writer{
		buf: bytes.NewBuffer(make([]byte, 0, n)),
		UTF16: DefaultUTF16,
	}
}

func (w *Writer) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

func (w *Writer) WriteUint8(n uint8) (int, error) {
	return 1, w.buf.WriteByte(n)
}

func (w *Writer) WriteUint16(n uint16) (int, error) {
	if _, err := w.buf.Write([]byte{byte(n >> 8), byte(n)}); err != nil {
		return 0, err
	}
	return 2, nil
}

func (w *Writer) WriteString(s string) (int, error) {
	n := len(s)
	if ByteStrUTF16 {
		n *= 2
	}
	if n > ByteStrSizeMax {
		return 0, ErrStringTooLarge
	}
	if n > ByteStrSizeStep {
		if _, err := w.buf.Write([]byte{byte(n>>8) | (1 << 7), byte(n)}); err != nil {
			return 0, err
		}
		n += 2
	} else {
		if err := w.buf.WriteByte(byte(n)); err != nil {
			return 0, err
		}
		n++
	}
	for v := 0; v < len(s); v++ {
		if w.UTF16 {
			if _, err := w.buf.Write([]byte{byte(uint16(s[v]) >> 8), byte(s[v])}); err != nil {
				return 0, err
			}
		} else {
			if err := w.buf.WriteByte(byte(s[v])); err != nil {
				return 0, err
			}
		}
	}
	return n, nil
}
