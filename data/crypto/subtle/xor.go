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

// Package subtle is similar to the 'cipher/subtle', only needed for very specific
// crypto operations.
//
package subtle

import (
	// Importing unsafe to link "xorBytes"
	_ "unsafe"
	// Importing crypto/cipher" to link "xorBytes"
	_ "crypto/cipher"
)

// XorOp will call the 'XorBytes' function but write the value back to the value
// instead. This one assumes that the key value is less than or equal to the
// value.
//
// If you need finer control, use the 'XorBytes' function.
func XorOp(value, key []byte) {
	if len(key) == 0 || len(value) == 0 {
		return
	}
	if len(key) == len(value) {
		xorBytes(value, key, value)
		return
	}
	for n := 0; n < len(value); {
		n += xorBytes(value[n:], key, value[n:])
	}
}

//go:linkname xorBytes crypto/cipher.xorBytes
func xorBytes(dst, a, b []byte) int

// XorBytes is the runtime import from "crypto/cipher.xorBytes" that can use
// hardware instructions for a 200% faster XOR operation.
//
// This variant will overlap the xor for the entire backing array, depending
// on which one is larger.
func XorBytes(dst, a, b []byte) int {
	var n int
	if len(a) < len(b) {
		for n < len(b) {
			n += xorBytes(dst[n:], a, b[n:])
		}
		return n
	}
	for n < len(a) {
		n += xorBytes(dst[n:], b, a[n:])
	}
	return n
}
