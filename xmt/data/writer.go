package data

// Writer is a basic interface that supports writing of all core Golang builtin types.
type Writer interface {
	Close() error
	Flush() error

	Write([]byte) error

	WriteBool(bool) error

	WriteInt(uint) error
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
}
