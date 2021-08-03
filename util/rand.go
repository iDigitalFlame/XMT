package util

import (
	// Import unsafe to use faster "cputicks" function instead of "time.Now().UnixNano()"
	_ "unsafe"
)

// Rand is the custom Random number generator, based on the current time as a seed.
// This struct is overridden by the tag "stdrand". By default, it will use the "unsafe" fastrand() implentation
// which is faster, but contains less entropy than the built-in, 'rand.Rand', which requires more memory and
// binary storage space.
var Rand = getRandom()

//go:linkname cputicks runtime.cputicks
func cputicks() int64

// FastRand is a fast thread local random function. This should be used in place instead of 'Rand.Uint32()'.
//
// Taken from https://github.com/dgraph-io/ristretto/blob/master/z/rtutil.go Thanks!
//go:linkname FastRand runtime.fastrand
func FastRand() uint32

// FastRandN is a fast thread local random function. This should be used in place instead of 'Rand.Uint32n()'.
// This function will take a max value to specify.
func FastRandN(n int) uint32 {
	return FastRand() % uint32(n)
}
