//go:build !amd64 && !386 && windows && !crypt
// +build !amd64,!386,windows,!crypt

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

import "github.com/iDigitalFlame/xmt/device/winapi/registry"

func isVirtual() bool {
	// 0x1 - KEY_QUERY_VALUE
	k, err := registry.Open(registry.KeyLocalMachine, `Hardware\Description\System\BIOS`, 0x1)
	if err != nil {
		return false
	}
	r := checkVendorKey(k, "BaseBoardManufacturer") ||
		checkVendorKey(k, "BaseBoardProduct") ||
		checkVendorKey(k, "BIOSVendor") ||
		checkVendorKey(k, "SystemManufacturer") ||
		checkVendorKey(k, "SystemFamily") ||
		checkVendorKey(k, "SystemProductName") ||
		checkVendorKey(k, "SystemVersion")
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
