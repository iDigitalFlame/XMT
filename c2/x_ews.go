//go:build ews && implant

package c2

import (
	"unsafe"

	"github.com/iDigitalFlame/xmt/util"
)

type header struct {
	Data uintptr
	Len  int
	Cap  int
}
type container struct {
	// Alloc free in-memory crypt.
	v []byte
	k [16]byte
}

func (c *container) Wrap() {
	for i := 0; i < 16; i++ {
		c.k[i] = byte(util.FastRand())
	}
	if c.k[0] == 0 {
		c.k[0] = 1
	}
	for i := 0; i < len(c.v); i++ {
		c.v[i] = c.v[i] ^ c.k[i%16]
	}
}
func (c *container) Unwrap() {
	if c.k[0] == 0 {
		return
	}
	for i := 0; i < len(c.v); i++ {
		c.v[i] = c.v[i] ^ c.k[i%16]
	}
}
func (c *container) Set(s string) {
	if c.k[0] = 0; len(c.v) == 0 {
		c.v = []byte(s)
		return
	}
	if len(c.v) < len(s) {
		c.v = append(c.v, make([]byte, len(s)-len(c.v))...)
	}
	n := copy(c.v, s)
	c.v = c.v[:n]
}
func (c container) String() string {
	// NOTE(dij): No allocs baby!
	return *(*string)(unsafe.Pointer((*header)(unsafe.Pointer(&c.v))))
}
