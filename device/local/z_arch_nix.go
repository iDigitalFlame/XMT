//go:build !windows && !js && !plan9 && (386 || arm)

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
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/device/arch"
	"golang.org/x/sys/unix"
)

func systemType() uint8 {
	// NOTE(dij): Check if we're running under a 64bit kernel and report the /actual/
	//            system arch, since we only know what we've been built as.
	var (
		u   unix.Utsname
		err = unix.Uname(&u)
	)
	if err != nil {
		return uint8(uint8(device.OS)<<4 | uint8(arch.Current))
	}
	switch {
	case device.Arch == arch.ARM && (u.Machine[10] == 0 && u.Machine[0] == 'a' && u.Machine[6] == '4' && u.Machine[5] == '6' && u.Machine[9] == 'e'): // Match aarch64_be
		fallthrough
	case device.Arch == arch.ARM && (u.Machine[7] == 0 && u.Machine[0] == 'a' && u.Machine[6] == '4' && u.Machine[5] == '6'): // Match aarch64
		fallthrough
	case device.Arch == arch.ARM && (u.Machine[6] == 0 && u.Machine[0] == 'a' && u.Machine[4] == '8' && u.Machine[3] == 'v'): // Match armv8l and armv8b
		return uint8(uint8(device.OS)<<4 | uint8(arch.ARMOnARM64))
	case device.Arch == arch.X86 && (u.Machine[6] == 0 && u.Machine[0] == 'x' && u.Machine[5] == '4' && u.Machine[4] == '6'): // Match x86_64
		return uint8(uint8(device.OS)<<4 | uint8(arch.X86OnX64))
	}
	return uint8(uint8(device.OS)<<4 | uint8(arch.Current))
}
