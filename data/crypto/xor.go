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

package crypto

import "github.com/iDigitalFlame/xmt/data/crypto/subtle"

// XOR is an alias for a byte array that acts as the XOR key data buffer.
type XOR []byte

// BlockSize returns the cipher's block size.
func (x XOR) BlockSize() int {
	return len(x)
}

// Operate preforms the XOR operation on the specified byte
// array using the cipher as the key.
func (x XOR) Operate(b []byte) {
	if len(x) == 0 {
		return
	}
	subtle.XorOp(b, x)
}

// Decrypt preforms the XOR operation on the specified byte array using the cipher
// as the key.
func (x XOR) Decrypt(dst, src []byte) {
	subtle.XorBytes(dst, x, src)
}

// Encrypt preforms the XOR operation on the specified byte array using the cipher
// as the key.
func (x XOR) Encrypt(dst, src []byte) {
	subtle.XorBytes(dst, x, src)
}
