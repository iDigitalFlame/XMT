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

// Package util is a very generic package that is used to contain simple
// functions that may be used in multiple packages, such as the simple random
// number generator.
//
// Generic re-implementations of built-in Golang structs or functions also
// sometimes land in here.
//
// This package is affected by the "stdrand" build tag, which will replace the
// "fastrand" implementation with the "math/rand" random struct.
package util

import (
	// Import unsafe to use faster "fastrand" function
	_ "unsafe"
)

// Rand is the custom Random number generator, based on the current time as a
// seed.
//
// This struct is overridden by the tag "stdrand". By default, it will use the
// "unsafe" fastrand() implementation which is faster, but contains less entropy
// than the built-in, 'rand.Rand', which requires more memory and binary storage
// space.
var Rand = getRandom()

// FastRand is a fast thread local random function. This should be used in place
// instead of 'Rand.Uint32()'.
//
// Taken from https://github.com/dgraph-io/ristretto/blob/master/z/rtutil.go
// Thanks!
//
//go:linkname FastRand runtime.fastrand
func FastRand() uint32

// FastRandN is a fast thread local random function. This should be used in
// place instead of 'Rand.Uint32n()'.
//
// This function will take a max value to specify.
func FastRandN(n int) uint32 {
	// return FastRand() % uint32(n)
	// NOTE(dij): The below code is supposed to be faster.
	return uint32(uint64(FastRand()) * uint64(n) >> 32)
}
