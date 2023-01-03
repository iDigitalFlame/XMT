//go:build windows && !map

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

package cmd

import "github.com/iDigitalFlame/xmt/device/winapi"

func freeMemory(h, addr uintptr) error {
	return winapi.NtFreeVirtualMemory(h, addr)
}
func writeMemory(h uintptr, protect uint32, n uint64, b []byte) (uintptr, error) {
	// 0x4 - PAGE_READWRITE
	a, err := winapi.NtAllocateVirtualMemory(h, uint32(n), 0x4)
	if err != nil {
		return 0, err
	}
	if _, err := winapi.NtWriteVirtualMemory(h, a, b); err != nil {
		winapi.NtFreeVirtualMemory(h, a)
		return 0, err
	}
	if _, err := winapi.NtProtectVirtualMemory(h, a, uint32(n), protect); err != nil {
		winapi.NtFreeVirtualMemory(h, a)
		return 0, err
	}
	return a, nil
}
