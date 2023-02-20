//go:build dragonfly
// +build dragonfly

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

	_ "unsafe"
)

const errMemory = syscall.Errno(0xC)

type utsName struct {
	Sysname  [32]byte
	Nodename [32]byte
	Release  [32]byte
	Version  [32]byte
	Machine  [32]byte
}

func uname(u *utsName) error {
	var (
		m = []int32{0x1, 0x1}
		c = uintptr(32)
	)
	if err := sysctl(m, &u.Sysname[0], &c, nil, 0); err != nil && err != errMemory {
		return err
	}
	m[1], c = 0xA, 32
	if err := sysctl(m, &u.Nodename[0], &c, nil, 0); err != nil && err != errMemory {
		return err
	}
	m[1], c = 0x2, 32
	if err := sysctl(m, &u.Release[0], &c, nil, 0); err != nil && err != errMemory {
		return err
	}
	m[1], c = 0x4, 32
	if err := sysctl(m, &u.Version[0], &c, nil, 0); err != nil && err != errMemory {
		return err
	}
	m[0], m[1], c = 0x6, 0x1, 32
	if err := sysctl(m, &u.Machine[0], &c, nil, 0); err != nil && err != errMemory {
		return err
	}
	u.Sysname[31], u.Nodename[31] = 0, 0
	u.Release[31], u.Machine[31] = 0, 0
	for i := range u.Version {
		if u.Version[i] == 0 {
			break
		}
		if u.Version[i] == '\n' || u.Version[i] == '\t' {
			if i == 31 {
				u.Version[i] = 0
				break
			}
			u.Version[i] = ' '
		}
	}
	return nil
}

//go:linkname sysctl syscall.sysctl
func sysctl(mib []int32, old *byte, oldlen *uintptr, new *byte, newlen uintptr) error
