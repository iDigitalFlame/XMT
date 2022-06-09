//go:build windows

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

package local

import (
	"syscall"
	"unsafe"

	"github.com/iDigitalFlame/xmt/device/winapi"
)

func isElevated() uint8 {
	var e uint8
	if checkElevatedToken() {
		e = 1
	}
	var (
		d *uint16
		s uint32
	)
	if err := syscall.NetGetJoinInformation(nil, &d, &s); err != nil {
		return e
	}
	if syscall.NetApiBufferFree((*byte)(unsafe.Pointer(d))); s == 3 {
		e += 128
	}
	return e
}
func checkElevatedToken() bool {
	var (
		t   uintptr
		err = winapi.OpenProcessToken(winapi.CurrentProcess, 0x8, &t)
		// 0x8 - TOKEN_QUERY
	)
	if err != nil {
		return false
	}
	var n uint32
	// 0x19 - TokenIntegrityLevel
	if winapi.GetTokenInformation(t, 0x19, nil, 0, &n); n < 4 {
		winapi.CloseHandle(t)
		return false
	}
	b := make([]byte, n)
	_ = b[n-1]
	// 0x19 - TokenIntegrityLevel
	if err = winapi.GetTokenInformation(t, 0x19, &b[0], n, &n); err != nil {
		winapi.CloseHandle(t)
		return false
	}
	var (
		p = uint32(b[n-4]) | uint32(b[n-3])<<8 | uint32(b[n-2])<<16 | uint32(b[n-1])<<24
		r = p >= 0x3000
	)
	if !r {
		r = winapi.IsTokenElevated(t)
	}
	winapi.CloseHandle(t)
	return r
}
