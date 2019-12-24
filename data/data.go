package data

import (
	"io"
)

// Reader is a basic interface that supports all types of read functions of the core Golang
// builtin types. Functions pointer functions are avaliable to allow for easier usage and
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

// Writeable is an interface that supports writing the target
// struct data to a Writer.
type Writeable interface {
	MarshalStream(Writer) error
}

// Readable is an interface that supports reading the target
// struct data from a Reader.
type Readable interface {
	UnmarshalStream(Reader) error
}
