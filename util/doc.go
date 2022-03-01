// Package util is a very generic package that is used to contain simple
// functions that may be used in multiple packages, such as the simple random
// number generator.
//
// Generic re-implementations of built-in Golang structs or functions also
// sometimes land in here.
//
// This package is affected by the "stdrand" build tag, which will replace the
// "fastrand" implementation with the "math/rand" random struct.
//
package util
