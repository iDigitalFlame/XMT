//go:build aix && ppc
// +build aix,ppc

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

/*
#include <stdint.h>
#include <stddef.h>
int uname(uintptr_t);
*/
import "C"
import "unsafe"

type utsName struct {
	Sysname  [32]byte
	Nodename [32]byte
	Release  [32]byte
	Version  [32]byte
	Machine  [32]byte
}

func uname(u *utsName) error {
	if r, err := C.uname(C.uintptr_t(uintptr(unsafe.Pointer(u)))); r == -1 && err != nil {
		return err
	}
	return nil
}
