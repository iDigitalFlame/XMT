// Copyright (C) 2020 - 2022 iDigitalFlame
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
package data

import (
	"io"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

const (
	// ErrTooLarge is raised if memory cannot be allocated to store data in a Chunk.
	ErrTooLarge = dataError(3)
	// ErrInvalidType is an error that occurs when the Bytes, ReadBytes, StringVal
	// or ReadString functions could not
	// propertly determine the underlying type of array from the Reader.
	ErrInvalidType = dataError(1)
	// ErrInvalidIndex is raised if a specified Grow or index function is supplied
	// with an negative or out of
	// bounds number or when a Seek index is not valid.
	ErrInvalidIndex = dataError(2)
)

// ErrLimit is an error that is returned when a Limit is set on a Chunk and the
// size limit was hit when attempting to write to the Chunk. This error wraps the
// io.EOF error, which allows this error to match io.EOF for sanity checking.
var ErrLimit = new(limitError)

type dataError uint8
type limitError struct{}

func (limitError) Error() string {
	if xerr.ExtendedInfo {
		return "buffer limit reached"
	}
	return "0x23"
}
func (limitError) Unwrap() error {
	return io.EOF
}
func (e dataError) Error() string {
	if xerr.ExtendedInfo {
		switch e {
		case ErrTooLarge:
			return "buffer is too large"
		case ErrInvalidType:
			return "invalid buffer type"
		case ErrInvalidIndex:
			return "invalid index"
		}
		return "unknown error"
	}
	switch e {
	case ErrTooLarge:
		return "0x26"
	case ErrInvalidType:
		return "0x24"
	case ErrInvalidIndex:
		return "0x25"
	}
	return "0x1"
}
