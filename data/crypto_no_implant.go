//go:build !implant
// +build !implant

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

package data

import (
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// String returns a colon ':' seperated version of the PublicKey's hex value.
func (p PublicKey) String() string {
	var b [(publicKeySize * 3) - 1]byte
	for i, n := 0, 0; i < len(p); i, n = i+1, n+2 {
		if n > 0 {
			b[n] = ':'
			n++
		}
		if p[i] < 16 {
			b[n] = '0'
			b[n+1] = util.HexTable[p[i]&0x0F]
		} else {
			b[n] = util.HexTable[p[i]>>4]
			b[n+1] = util.HexTable[p[i]&0x0F]
		}
	}
	return string(b[:])
}

// String returns a colon ':' seperated version of the PrivateKey's hex value.
func (p PrivateKey) String() string {
	var b [(privateKeySize * 3) - 1]byte
	for i, n := 0, 0; i < len(p); i, n = i+1, n+2 {
		if n > 0 {
			b[n] = ':'
			n++
		}
		if p[i] < 16 {
			b[n] = '0'
			b[n+1] = util.HexTable[p[i]&0x0F]
		} else {
			b[n] = util.HexTable[p[i]>>4]
			b[n+1] = util.HexTable[p[i]&0x0F]
		}
	}
	return string(b[:])
}
func hexByteToInt(v byte) (uint8, error) {
	// Quick way to convert hex bytes to real numbers.
	switch {
	case v >= 0x61 && v <= 0x66: // a - f
		return (v - 0x61) + 0xA, nil
	case v >= 0x41 && v <= 0x46: // A - F
		return (v - 0x41) + 0xA, nil
	case v >= 0x30 && v <= 0x39: // 0 - 9
		return v - 0x30, nil
	}
	return 0, xerr.Sub("invalid non-hex byte", 0x76)
}

// Parse will attempt to fill the data of this PublicKey from the supplied PublicKey
// colon-seperated hex string.
//
// This function will only overrite the PublicKey data if the entire parsing
// process succeeds.
//
// Any errors occurred during parsing will be returned.
func (p *PublicKey) Parse(v string) error {
	if len(v) != (publicKeySize*3)-1 {
		return xerr.Sub("invalid hex string length", 0x74)
	}
	var (
		b    [publicKeySize]byte
		f, s byte
		err  error
	)
	for i, n := 0, 0; i < (publicKeySize*3)-1; i, n = i+2, n+1 {
		if i > 0 {
			if v[i] != ':' {
				return xerr.Sub("invalid non-colon character", 0x75)
			}
			i++
		}
		if f, err = hexByteToInt(v[i]); err != nil {
			return err
		}
		if s, err = hexByteToInt(v[i+1]); err != nil {
			return err
		}
		b[n] = (f << 0x4) | s
	}
	copy((*p)[:], b[:])
	return nil
}

// Parse will attempt to fill the data of this PrivateKey from the supplied PublicKey
// colon-seperated hex string.
//
// This function will only overrite the PrivateKey data if the entire parsing
// process succeeds.
//
// Any errors occurred during parsing will be returned.
func (p *PrivateKey) Parse(v string) error {
	if len(v) != (privateKeySize*3)-1 {
		return xerr.Sub("invalid hex string length", 0x74)
	}
	var (
		b    [privateKeySize]byte
		f, s byte
		err  error
	)
	for i, n := 0, 0; i < (privateKeySize*3)-1; i, n = i+2, n+1 {
		if i > 0 {
			if v[i] != ':' {
				return xerr.Sub("invalid non-colon character", 0x75)
			}
			i++
		}
		if f, err = hexByteToInt(v[i]); err != nil {
			return err
		}
		if s, err = hexByteToInt(v[i+1]); err != nil {
			return err
		}
		b[n] = (f << 0x4) | s
	}
	copy((*p)[:], b[:])
	return nil
}
