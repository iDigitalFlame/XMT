//go:build zos && s390x
// +build zos,s390x

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

package unix

import (
	"syscall"
	"unsafe"
)

type utsName struct {
	Sysname    [65]byte
	Nodename   [65]byte
	Release    [65]byte
	Version    [65]byte
	Machine    [65]byte
	Domainname [65]byte
}

func uname(u *utsName) error {
	if _, _, err := syscall.RawSyscall(0x77E, uintptr(unsafe.Pointer(u)), 0, 0); err != 0 {
		return err
	}
	return nil
}
