//go:build windows && map

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

import (
	"unsafe"

	"github.com/iDigitalFlame/xmt/device/winapi"
)

func freeMemory(h, addr uintptr) error {
	return winapi.NtUnmapViewOfSection(h, addr)
}
func writeMemory(h uintptr, protect uint32, n uint64, b []byte) (uintptr, error) {
	// 0xE       - SECTION_MAP_READ | SECTION_MAP_WRITE | SECTION_MAP_EXECUTE
	// 0x40      - PAGE_EXECUTE_READWRITE
	// 0x8000000 - SEC_COMMIT
	s, err := winapi.NtCreateSection(0xE, n, 0x40, 0x8000000, 0)
	if err != nil {
		return 0, err
	}
	// 0x4 - PAGE_READWRITE
	// 0x2 - ViewUnmap
	a, err := winapi.NtMapViewOfSection(s, winapi.CurrentProcess, 0, n, 2, 0, 0x4)
	if err != nil {
		winapi.CloseHandle(s)
		return 0, err
	}
	// 0x2 - ViewUnmap
	r, err := winapi.NtMapViewOfSection(s, h, 0, n, 0x2, 0, protect)
	if err != nil {
		winapi.NtUnmapViewOfSection(winapi.CurrentProcess, a)
		winapi.CloseHandle(s)
		return 0, err
	}
	for i := range b {
		(*(*[1]byte)(unsafe.Pointer(a + uintptr(i))))[0] = b[i]
	}
	winapi.NtUnmapViewOfSection(winapi.CurrentProcess, a)
	winapi.CloseHandle(s)
	return r, nil
}
