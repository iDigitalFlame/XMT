//go:build ews && implant
// +build ews,implant

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

package c2

import (
	"unsafe"

	"github.com/iDigitalFlame/xmt/data/crypto/subtle"
	"github.com/iDigitalFlame/xmt/util"
)

type header struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}
type container struct {
	// Alloc free in-memory crypt.
	v []byte
	k [16]byte
}

// TODO(dij): Also encode the Profile data like in Rust
func (c *container) Wrap() {
	for i := 0; i < 16; i++ {
		c.k[i] = byte(util.FastRand())
	}
	if c.k[0] == 0 {
		c.k[0] = 1
	}
	subtle.XorOp(c.v, c.k[:])
}
func (c *container) Unwrap() {
	if c.k[0] == 0 {
		return
	}
	subtle.XorOp(c.v, c.k[:])
}
func (c *container) Set(s string) {
	if c.k[0] = 0; len(c.v) == 0 {
		c.v = []byte(s)
		return
	}
	if i := len(s) - len(c.v); i > 0 {
		c.v = append(c.v, make([]byte, i)...)
	}
	n := copy(c.v, s)
	c.v = c.v[:n]
}
func (c container) String() string {
	// NOTE(dij): No allocs baby!
	return *(*string)(unsafe.Pointer((*header)(unsafe.Pointer(&c.v))))
}
