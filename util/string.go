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

package util

import "github.com/iDigitalFlame/xmt/data/crypto/subtle"

// Decode is used to un-encode a string written in a XOR byte array "encrypted"
// by the specified key.
//
// This function returns the string value of the result but also modifies the
// input array, which can be used to re-use the resulting string.
// NOTE(dij): Is this still used?
func Decode(k, d []byte) string {
	if len(k) == 0 || len(d) == 0 {
		return ""
	}
	subtle.XorOp(d, k)
	return string(d)
}
