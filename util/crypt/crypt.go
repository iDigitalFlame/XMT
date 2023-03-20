//go:build crypt
// +build crypt

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

// Package crypt is a builtin package that provides compile-time encoded string
// values to be decoded and used when first starting up.
//
// This package should only be used with the "crypt" tag, which is auto compiled
// during build.
package crypt

import (
	"encoding/base64"
	"os"

	"github.com/iDigitalFlame/xmt/data/crypto/subtle"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const cryptMax = 0xFF

var (
	key     string
	values  [cryptMax]string
	payload string
)

func init() {
	if len(payload) == 0 || len(key) == 0 {
		if xerr.ExtendedInfo {
			panic("crypt: no data supplied during build")
		}
		os.Exit(2)
	}
	var (
		b      = make([]byte, base64.URLEncoding.DecodedLen(len(payload)))
		v, err = base64.URLEncoding.Decode(b, []byte(payload))
	)
	if err != nil || len(b) == 0 || v == 0 {
		if xerr.ExtendedInfo {
			panic("crypt: cannot read supplied data")
		}
		os.Exit(2)
	}
	var (
		k = make([]byte, base64.URLEncoding.DecodedLen(len(key)))
		c int
	)
	if c, err = base64.URLEncoding.Decode(k, []byte(key)); err != nil || len(k) == 0 || c == 0 {
		if xerr.ExtendedInfo {
			panic("crypt: empty or invalid crypt encoded data")
		}
		os.Exit(2)
	}
	subtle.XorOp(b[:v], k[:c])
	for s, e, n := 0, 0, 0; e < v && n < cryptMax; e++ {
		if b[e] != 0 {
			continue
		}
		if e-s > 0 {
			values[n] = string(b[s:e])
		}
		s, n = e+1, n+1
	}
	key, payload = "", ""
}

// Get returns the crypt value at the provided string index.
func Get(i uint8) string {
	return values[i]
}
