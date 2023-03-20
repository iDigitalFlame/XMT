//go:build windows
// +build windows

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

package local

import (
	"syscall"
	"unsafe"

	"github.com/iDigitalFlame/xmt/device/winapi"
)

func getPPID() uint32 {
	return winapi.Getppid()
}
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
		e |= 0x80
	}
	return e
}
func getUsername() string {
	if u, err := winapi.GetLocalUser(); err == nil && len(u) > 0 {
		return u
	}
	return "?"
}
func checkElevatedToken() bool {
	if !winapi.IsWindowsVista() {
		return winapi.UserInAdminGroup()
	}
	var t uintptr
	// 0x8 - TOKEN_QUERY
	if err := winapi.OpenThreadToken(winapi.CurrentThread, 0x8, true, &t); err != nil {
		if err = winapi.OpenProcessToken(winapi.CurrentProcess, 0x8, &t); err != nil {
			return false
		}
	}
	var (
		n uint32 = 32
		b [32]byte
	)
	// 0x19 - TokenIntegrityLevel
	if err := winapi.GetTokenInformation(t, 0x19, &b[0], n, &n); err != nil {
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
