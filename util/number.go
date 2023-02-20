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

package util

// HexTable is a static string hex mapping constant string. This can be used
// multiple places to prevent reuse.
const HexTable = "0123456789ABCDEF"

// Itoa converts val to a decimal string.
//
// Similar to the "strconv" variant.
// Taken from the "internal/itoa" package.
func Itoa(v int64) string {
	if v < 0 {
		return "-" + Uitoa(uint64(-v))
	}
	return Uitoa(uint64(v))
}

// Uitoa converts val to a decimal string.
//
// Similar to the "strconv" variant.
// Taken from the "internal/itoa" package.
func Uitoa(v uint64) string {
	if v == 0 {
		return "0"
	}
	var (
		i = 0x13
		b [20]byte
	)
	for v >= 0xA {
		n := v / 0xA
		b[i] = byte(0x30 + v - n*0xA)
		i--
		v = n
	}
	b[i] = byte(0x30 + v)
	return string(b[i:])
}

// Uitoa16 converts val to a hexadecimal string.
func Uitoa16(v uint64) string {
	if v == 0 {
		return "0"
	}
	var (
		i = 0x13
		b [20]byte
	)
	for {
		n := (v >> (4 * uint(0x13-i)))
		b[i] = HexTable[n&0xF]
		if i--; n <= 0xF {
			break
		}
	}
	return string(b[i+1:])
}
