package data

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
)

const (
	// DataLimitSmall is the size value allowed for small strings using the WriteString and WriteBytes functions.
	DataLimitSmall uint64 = 2 << 7
	// DataLimitLarge is the size value allowed for large strings using the WriteString and WriteBytes functions.
	DataLimitLarge uint64 = 2 << 31
	// DataLimitMedium is the size value allowed for medium strings using the WriteString and WriteBytes functions.
	DataLimitMedium uint64 = 2 << 15
)

var (
	// ErrInvalidBytes is an error that occurs when the Bytes function could not propertly determine
	// the type of byte array from the Reader.
	ErrInvalidBytes = errors.New("could not understand bytes type")
	// ErrInvalidString is an error that occurs when the ReadString or String functions could not propertly
	// determine the type of string from the Reader.
	ErrInvalidString = errors.New("could not understand string type")
)

// Reader is a basic interface that supports all types of read functions of the core Golang builtin types.
// Pointer functions are avaliable to allow for easier usage and fluid operation.
type Reader interface {
	Int() (int, error)
	Bool() (bool, error)
	Int8() (int8, error)
	Uint() (uint, error)
	Int16() (int16, error)
	Int32() (int32, error)
	Int64() (int64, error)
	Uint8() (uint8, error)
	Bytes() ([]byte, error)
	Uint16() (uint16, error)
	Uint32() (uint32, error)
	Uint64() (uint64, error)
	Float32() (float32, error)
	Float64() (float64, error)
	StringVal() (string, error)

	ReadInt(*int) error
	ReadBool(*bool) error
	ReadInt8(*int8) error
	ReadUint(*uint) error
	ReadInt16(*int16) error
	ReadInt32(*int32) error
	ReadInt64(*int64) error
	ReadUint8(*uint8) error
	ReadUint16(*uint16) error
	ReadUint32(*uint32) error
	ReadUint64(*uint64) error
	ReadString(*string) error
	ReadFloat32(*float32) error
	ReadFloat64(*float64) error
	io.ReadCloser
}

// Writer is a basic interface that supports writing of all core Golang builtin types.
type Writer interface {
	WriteInt(int) error
	WriteBool(bool) error
	WriteInt8(int8) error
	WriteUint(uint) error
	WriteInt16(int16) error
	WriteInt32(int32) error
	WriteInt64(int64) error
	WriteUint8(uint8) error
	WriteBytes([]byte) error
	WriteUint16(uint16) error
	WriteUint32(uint32) error
	WriteUint64(uint64) error
	WriteString(string) error
	WriteFloat32(float32) error
	WriteFloat64(float64) error
	io.WriteCloser
	Flusher
}
type ctxReader struct {
	ctx    context.Context
	cancel context.CancelFunc
	io.ReadCloser
}

// Flusher is an interface that supports Flushing the stream output to the underlying Writer.
type Flusher interface {
	Flush() error
}

// Writeable is an interface that supports writing the target struct data to a Writer.
type Writeable interface {
	MarshalStream(Writer) error
}

// Readable is an interface that supports reading the target struct data from a Reader.
type Readable interface {
	UnmarshalStream(Reader) error
}

func (r *ctxReader) Close() error {
	r.cancel()
	return r.ReadCloser.Close()
}
func (r *ctxReader) Read(b []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		if err := r.ReadCloser.Close(); err != nil {
			return 0, err
		}
		return 0, r.ctx.Err()
	default:
		return r.ReadCloser.Read(b)
	}
}

// ReadStringList attempts to read a string list written using the 'WriteStringList' function from the supplied
// string into the string list pointer. If the provided array is nil or not large enough, it will be resized.
func ReadStringList(r Reader, s *[]string) error {
	t, err := r.Uint8()
	if err != nil {
		return err
	}
	var l int
	switch t {
	case 0:
		return nil
	case 1, 2:
		n, err := r.Uint8()
		if err != nil {
			return err
		}
		l = int(n)
	case 3, 4:
		n, err := r.Uint16()
		if err != nil {
			return err
		}
		l = int(n)
	case 5, 6:
		n, err := r.Uint32()
		if err != nil {
			return err
		}
		l = int(n)
	case 7, 8:
		n, err := r.Uint64()
		if err != nil {
			return err
		}
		l = int(n)
	default:
		return ErrInvalidString
	}
	if s == nil || len(*s) < l {
		*s = make([]string, l)
	}
	for x := 0; x < l; x++ {
		if err := r.ReadString(&(*s)[x]); err != nil {
			return err
		}
	}
	return nil
}

// WriteStringList will attempt to write the supplied string list to the writer. If the string list is nil or
// empty, it will write a zero byte to the Writer. The resulting data can be read using the 'ReadStringList' function.
func WriteStringList(w Writer, s []string) error {
	switch l := uint64(len(s)); {
	case l == 0:
		return w.WriteUint8(0)
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
	for i := range s {
		if err := w.WriteString(s[i]); err != nil {
			return err
		}
	}
	return nil
}

// ReadFully attempts to Read all the bytes from the specified reader until the length of the array or EOF.
func ReadFully(r io.Reader, b []byte) (int, error) {
	var n int
	for n < len(b) {
		i, err := r.Read(b[n:])
		if n += i; err != nil && ((err != io.EOF && err != ErrLimit) || n != len(b)) {
			return n, err
		}
	}
	return n, nil
}

// NewCtxReader creates a reader backed by the supplied Reader and Context. This reader will automatically close
// when the parent context is canceled. This is useful in situations when direct copies using 'io.Copy' on threads
// or timed operations are required.
func NewCtxReader(x context.Context, r io.Reader) io.ReadCloser {
	i := new(ctxReader)
	if c, ok := r.(io.ReadCloser); ok {
		i.ReadCloser = c
	} else {
		i.ReadCloser = ioutil.NopCloser(r)
	}
	i.ctx, i.cancel = context.WithCancel(x)
	return i
}
