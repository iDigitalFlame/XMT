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

// TransformB64 is a Setting that enables the Base64 Transform for the generated
// Profile.
const TransformB64 = cBit(0xE0)

const (
	valDNS      = cBit(0xE1)
	valB64Shift = cBit(0xE2)
)

// TransformDNS returns a Setting that will apply the DNS Transform to the
// generated Profile. If any DNS Domains are specified, they will be used in the
// Transform.
//
// If a Transform Setting is already contained in the current Config Group, a
// 'ErrMultipleTransforms' error will be returned when the 'Profile' function
// is called.
func TransformDNS(n ...string) Setting {
	s := cBytes{byte(valDNS), 0}
	if len(s) == 0 {
		return s
	}
	if len(n) > 0xFF {
		s[1] = 0xFF
	} else {
		s[1] = byte(len(n))
	}
	for i, c := 0, 2; i < len(n) && i < 0xFF; i++ {
		v := n[i]
		if len(v) > 0xFF {
			v = v[:0xFF]
		}
		s = append(s, make([]byte, len(v)+1)...)
		s[c] = byte(len(v))
		c += copy(s[c+1:], v) + 1
	}
	return s
}

// TransformB64Shift returns a Setting that will apply the Base64 Shift Transform
// to the generated Profile. The specified number will be the shift index of the
// Transform.
//
// If a Transform Setting is already contained in the current Config Group, a
// 'ErrMultipleTransforms' error will be returned when the 'Profile' function is
// called.
func TransformB64Shift(s int) Setting {
	return cBytes{byte(valB64Shift), byte(s)}
}
