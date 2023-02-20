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

package util

import "unsafe"

// A Builder is used to efficiently build a string using Write methods. It
// minimizes memory copying. The zero value is ready to use. Do not copy a
// non-zero Builder.
//
// Re-implemented to remove UTF8 dependency and added some useful functions.
// Copy-check was also removed.
type Builder struct {
	b []byte
}

// Reset resets the Builder to be empty.
func (b *Builder) Reset() {
	b.b = nil
}

// Len returns the number of accumulated bytes; b.Len() == len(b.String()).
func (b *Builder) Len() int {
	return len(b.b)
}

// Cap returns the capacity of the builder's underlying byte slice. It is the
// total space allocated for the string being built and includes any bytes
// already written.
func (b *Builder) Cap() int {
	return cap(b.b)
}

// Grow grows b's capacity, if necessary, to guarantee space for another n bytes.
//
// After Grow(n), at least n bytes can be written to b without another allocation.
// If n is negative, Grow is a NOP.
func (b *Builder) Grow(n int) {
	if n < 0 || cap(b.b)-len(b.b) >= n {
		return
	}
	v := make([]byte, len(b.b), 2*cap(b.b)+n)
	copy(v, b.b)
	b.b, v = v, nil
}

// String returns the accumulated string.
func (b *Builder) String() string {
	return *(*string)(unsafe.Pointer(&b.b))
}

// Output returns the accumulated string, then resets the value of this Builder.
func (b *Builder) Output() string {
	s := *(*string)(unsafe.Pointer(&b.b))
	b.Reset()
	return s
}

// WriteByte appends the byte c to b's buffer.
//
// The returned error is always nil.
func (b *Builder) WriteByte(c byte) error {
	b.b = append(b.b, c)
	return nil
}

// InsertByte appends the byte c to b's buffer at the zero position.
//
// The returned error is always nil.
func (b *Builder) InsertByte(c byte) error {
	b.b = append(b.b, 0)
	copy(b.b[1:], b.b)
	b.b[0] = c
	return nil
}

// Write appends the contents of p to b's buffer.
//
// Write always returns len(p), nil.
func (b *Builder) Write(p []byte) (int, error) {
	b.b = append(b.b, p...)
	return len(p), nil
}

// WriteString appends the contents of s to b's buffer.
//
// It returns the length of s and a nil error.
func (b *Builder) WriteString(s string) (int, error) {
	b.b = append(b.b, s...)
	return len(s), nil
}
