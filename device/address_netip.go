//go:build go1.18
// +build go1.18

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

package device

import "net/netip"

// ToAddr will return this Address as a netip.Addr struct. This will choose the
// type based on the underlying address size.
func (a *Address) ToAddr() netip.Addr {
	if a.Is4() {
		return netip.AddrFrom4([4]byte{byte(a.low >> 24), byte(a.low >> 16), byte(a.low >> 8), byte(a.low)})
	}
	return netip.AddrFrom16([16]byte{
		byte(a.hi >> 56), byte(a.hi >> 48), byte(a.hi >> 40), byte(a.hi >> 32),
		byte(a.hi >> 24), byte(a.hi >> 16), byte(a.hi >> 8), byte(a.hi),
		byte(a.low >> 56), byte(a.low >> 48), byte(a.low >> 40), byte(a.low >> 32),
		byte(a.low >> 24), byte(a.low >> 16), byte(a.low >> 8), byte(a.low),
	})
}
