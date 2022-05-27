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
// instead.
//
// If you need finer control, use the 'XorBytes' function.
func XorOp(value, key []byte) {
	XorBytes(value, key, value)
}

// XorBytes is the runtime import from "crypto/cipher.xorBytes" that can use
// hardware instructions for a 200% faster XOR operation.
//go:linkname XorBytes crypto/cipher.xorBytes
func XorBytes(dst, a, b []byte) int
