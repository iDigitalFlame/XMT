package data

import (
	"io"
)

// Reader is a basic interface that supports all types of read functions of the core Golang
// builtin types. Functions pointer functions are avaliable to allow for easier usage and more
// fluid operation.
type Reader interface {
	Bool() (bool, error)

	Bytes() ([]byte, error)

	Int() (int, error)
	Int8() (int8, error)
	Int16() (int16, error)
	Int32() (int32, error)
	Int64() (int64, error)

	Uint() (uint, error)
	Uint8() (uint8, error)
	Uint16() (uint16, error)
	Uint32() (uint32, error)
	Uint64() (uint64, error)

	Float32() (float32, error)
	Float64() (float64, error)

	// StringVal is used instead of 'String' for
	// compatibility with fmt.Stringer.
	StringVal() (string, error)

	ReadBool(*bool) error

	ReadInt(*int) error
	ReadInt8(*int8) error
	ReadInt16(*int16) error
	ReadInt32(*int32) error
	ReadInt64(*int64) error

	ReadUint(*uint) error
	ReadUint8(*uint8) error
	ReadUint16(*uint16) error
	ReadUint32(*uint32) error
	ReadUint64(*uint64) error

	ReadString(*string) error

	ReadFloat32(*float32) error
	ReadFloat64(*float64) error

	io.ReadCloser
}

// Writer is a basic interface that supports writing
// of all core Golang builtin types.
type Writer interface {
	WriteBool(bool) error

	WriteBytes([]byte) error

	WriteInt(int) error
	WriteInt8(int8) error
	WriteInt16(int16) error
	WriteInt32(int32) error
	WriteInt64(int64) error

	WriteUint(uint) error
	WriteUint8(uint8) error
	WriteUint16(uint16) error
	WriteUint32(uint32) error
	WriteUint64(uint64) error

	WriteString(string) error

	WriteFloat32(float32) error
	WriteFloat64(float64) error

	io.WriteCloser
	Flusher
}

// Flusher is an interface that supports Flushing the
// stream output to the underlying Writer.
type Flusher interface {
	Flush() error
}

// Writable is an interface that supports writing the target
// struct data to a Writer.
type Writable interface {
	MarshalStream(Writer) error
}

// Readable is an interface that supports reading the target
// struct data from a Reader.
type Readable interface {
	UnmarshalStream(Reader) error
}

// ReadStringList attempts to read a string list written using
// the 'WriteStringList' function from the supplied string into
// the string list pointer. If the provided array is nil or not large
// enough, it will be resized.
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

// WriteStringList will attempt to write the supplied string list to
// the writer. If the string list is nil or empty, it will write a zero
// byte to the Writer. The resulting data can be read using the 'ReadStringList'
// function.
func WriteStringList(w Writer, s []string) error {
	if s == nil {
		return w.WriteUint8(0)
	}
	switch l := len(s); {
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

// ReadFully attempts to Read all the bytes from the
// specified reader until the length of the array or EOF.
func ReadFully(r io.Reader, b []byte) (int, error) {
	var n int
	for n < len(b) {
		i, err := r.Read(b[n:])
		if err != nil && (err != io.EOF || n != len(b)) {
			return n, err
		}
		n += i
	}
	return n, nil
}
