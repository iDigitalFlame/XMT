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

package cfg

import "crypto/rand"

const (
	// WrapHex is a Setting that enables the Hex Wrapper for the generated Profile.
	WrapHex = cBit(0xD0)
	// WrapZlib is a Setting that enables the ZLIB Wrapper for the generated Profile.
	WrapZlib = cBit(0xD1)
	// WrapGzip is a Setting that enables the GZIP Wrapper for the generated Profile.
	WrapGzip = cBit(0xD2)
	// WrapBase64 is a Setting that enables the Base64 Wrapper for the generated
	// Profile.
	WrapBase64 = cBit(0xD3)
)

const (
	valXOR = cBit(0xD4)
	valCBK = cBit(0xD5)
	valAES = cBit(0xD6)
)

// WrapXOR returns a Setting that will apply the XOR Wrapper to the generated
// Profile. The specified key will be the XOR key used.
func WrapXOR(k []byte) Setting {
	n := len(k)
	if n > 0xFFFF {
		n = 0xFFFF
	}
	return append(cBytes{byte(valXOR), byte(n >> 8), byte(n)}, k[:n]...)
}

// WrapAES returns a Setting that will apply the AES Wrapper to the generated
// Profile. The specified key and IV will be the AES Key and IV used.
func WrapAES(k, iv []byte) Setting {
	n, v := len(k), len(iv)
	if n > 0xFF {
		n = 0xFF
	}
	i := iv
	if n > 0 && v == 0 {
		i, v = make([]byte, 16), 16
		rand.Read(i)
	}
	c := make(cBytes, 3+n+v)
	c[0] = byte(valAES)
	c[1], c[2] = byte(n), byte(v)
	n = copy(c[3:], k) + 3
	copy(c[n:], i)
	return c
}

// WrapCBK returns a Setting that will apply the CBK Wrapper to the generated
// Profile. The specified ABC and Type values are the CBK letters used.
//
// To specify the CBK buffer size, use the 'WrapCBKSize' function instead.
func WrapCBK(a, b, c, d byte) Setting {
	return cBytes{byte(valCBK), 128, a, b, c, d}
}

// WrapCBKSize returns a Setting that will apply the CBK Wrapper to the generated
// Profile. The specified Size, ABC and Type values and the CBK size and letters
// used.
func WrapCBKSize(s, a, b, c, d byte) Setting {
	return cBytes{byte(valCBK), s, a, b, c, d}
}
