// Package subtle is similar to the 'cipher/subtle', only needed for very specific
// crypto operations.
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
			//println("n", n, "b", len(b), "a", len(a))
			n += xorBytes(dst[n:], a, b[n:])
		}
		return n
	}
	for n < len(a) {
		//println("n", n, "b", len(b), "a", len(a))
		n += xorBytes(dst[n:], b, a[n:])
	}
	return n
}
