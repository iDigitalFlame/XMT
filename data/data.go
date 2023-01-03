// Copyright (C) 2020 - 2023 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

// Package data provides interfaces and helper methods that can be used for data
// writing and reading. This package exports the data Reader and Writer interfaces
// which are an extension of the standard io.Reader and io.Writer interfaces.
//
package data

const (
	// LimitSmall is the size value allowed for small strings using the WriteString
	// and WriteBytes functions.
	LimitSmall uint64 = 2 << 7
	// LimitLarge is the size value allowed for large strings using the WriteString
	// and WriteBytes functions.
	LimitLarge uint64 = 2 << 31
	// LimitMedium is the size value allowed for medium strings using the WriteString
	// and WriteBytes functions.
	LimitMedium uint64 = 2 << 15
)

// Reader is a basic interface that supports all types of read functions of the
// core Golang builtin types.
//
// Pointer functions are available to allow for easier usage and fluid operation.
type Reader interface {
	Close() error
	Read([]byte) (int, error)

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
	ReadBytes(*[]byte) error
	ReadUint16(*uint16) error
	ReadUint32(*uint32) error
	ReadUint64(*uint64) error
	ReadString(*string) error
	ReadFloat32(*float32) error
	ReadFloat64(*float64) error
}

// Writer is a basic interface that supports writing of all core Golang builtin
// types.
type Writer interface {
	Close() error
	Flush() error
	Write([]byte) (int, error)

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
}
