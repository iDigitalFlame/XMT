//go:build !amd64 && !386 && windows && crypt
// +build !amd64,!386,windows,crypt

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

import (
	"github.com/iDigitalFlame/xmt/device/winapi/registry"
	"github.com/iDigitalFlame/xmt/util/crypt"
)

func isVirtual() bool {
	// 0x1 - KEY_QUERY_VALUE
	k, err := registry.Open(registry.KeyLocalMachine, crypt.Get(63), 0x1) // Hardware\Description\System\BIOS
	if err != nil {
		return false
	}
	r := checkVendorKey(k, crypt.Get(64)) || // BaseBoardManufacturer
		checkVendorKey(k, crypt.Get(65)) || // BaseBoardProduct
		checkVendorKey(k, crypt.Get(66)) || // BIOSVendor
		checkVendorKey(k, crypt.Get(67)) || // SystemManufacturer
		checkVendorKey(k, crypt.Get(68)) || // SystemFamily
		checkVendorKey(k, crypt.Get(69)) || // SystemProductName
		checkVendorKey(k, crypt.Get(70)) // SystemVersion
	k.Close()
	return r
}
func checkVendorKey(k registry.Key, s string) bool {
	v, _, err := k.String(s)
	if err != nil || len(v) == 0 {
		return false
	}
	return isKnownVendor([]byte(v))
}
