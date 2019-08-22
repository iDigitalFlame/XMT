package data

// Reader is a basic interface that supports all types of read functions of the core Golang
// builtin types. Functions pointer functions are avaliable to allow for easier usage and more
// fulent operation.
type Reader interface {
	Close() error

	Bool() (bool, error)

	ReadBool(*bool) error

	Int() (int, error)
	Int8() (int8, error)
	Int16() (int16, error)
	Int32() (int32, error)
	Int64() (int64, error)

	ReadInt(*int) error
	ReadInt8(*int8) error
	ReadInt16(*int16) error
	ReadInt32(*int32) error
	ReadInt64(*int64) error

	Uint() (uint, error)
	Uint8() (uint8, error)
	Uint16() (uint16, error)
	Uint32() (uint32, error)
	Uint64() (uint64, error)

	ReadUint(*uint) error
	ReadUint8(*uint8) error
	ReadUint16(*uint16) error
	ReadUint32(*uint32) error
	ReadUint64(*uint64) error

	String() (string, error)

	ReadString(*string) error

	Read([]byte) (int, error)

	Float32() (float32, error)
	Float64() (float64, error)

	ReadFloat32(*float32) error
	ReadFloat64(*float64) error
}
