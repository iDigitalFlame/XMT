//go:build aix && ppc64 && gc
// +build aix,ppc64,gc

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

//go:cgo_import_dynamic libc_uname uname "libc.a/shr_64.o"
//go:linkname funcUname libc_uname
var funcUname uintptr

type utsName struct {
	Sysname  [32]byte
	Nodename [32]byte
	Release  [32]byte
	Version  [32]byte
	Machine  [32]byte
}

func uname(u *utsName) error {
	if _, _, err := syscall.RawSyscall6(uintptr(unsafe.Pointer(&funcUname)), uintptr(unsafe.Pointer(u)), 0, 0, 0, 0, 0); err != 0 {
		return err
	}
	return nil
}
