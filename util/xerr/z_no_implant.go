//go:build !implant

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

// Package xerr is a simplistic (and more efficient) re-write of the "errors"
// built-in package.
//
// This is used to create comparable and (sometimes) un-wrapable error structs.
//
// This package acts differently when the "implant" build tag is used. If enabled,
// Most error string values are stripped to prevent identification and debugging.
//
// It is recommended if errors are needed to be compared even when in an implant
// build, to use the "Sub" function, which will ignore error strings and use
// error codes instead.
//
package xerr

// ExtendedInfo is a compile time constant to help signal if complex string
// values should be concatenated inline.
//
// This helps prevent debugging when the "-tags implant" option is enabled.
const ExtendedInfo = true

type strErr string
type wrapErr struct {
	e error
	s string
}

// New creates a new string backed error interface and returns it.
// This error struct does not support Unwrapping.
//
// The resulting errors created will be comparable.
func New(s string) error {
	return strErr(s)
}
func (e strErr) Error() string {
	return string(e)
}
func (e wrapErr) Error() string {
	return e.s
}
func (e wrapErr) Unwrap() error {
	return e.e
}

// Sub creates a new string backed error interface and returns it.
// This error struct does not support Unwrapping.
//
// If the "-tags implant" option is selected, the second value, the error code,
// will be used instead, otherwise it's ignored.
//
// The resulting errors created will be comparable.
func Sub(s string, _ uint8) error {
	return strErr(s)
}

// Wrap creates a new error that wraps the specified error.
//
// If not nil, this function will append ": " + 'Error()' to the resulting
// string message and will keep the original error for unwrapping.
//
// If "-tags implant" is specified, this will instead return the wrapped error
// directly.
func Wrap(s string, e error) error {
	if e != nil {
		return &wrapErr{s: s + ": " + e.Error(), e: e}
	}
	return &wrapErr{s: s}
}
