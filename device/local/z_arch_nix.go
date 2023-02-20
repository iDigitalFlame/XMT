//go:build !windows && !js && !plan9 && (386 || arm)
// +build !windows
// +build !js
// +build !plan9
// +build 386 arm

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
	"github.com/iDigitalFlame/xmt/device/unix"
)

func systemType() uint8 {
	if arch.Current != arch.ARM && arch.Current != arch.X86 {
		return uint8(uint8(device.OS)<<4 | uint8(arch.Current))
	}
	// NOTE(dij): Check if we're running under a 64bit kernel and report the /actual/
	//            system arch, since we only know what we've been built as.
	switch {
	case !unix.IsMachine64():
	case arch.Current == arch.ARM:
		return uint8(uint8(device.OS)<<4 | uint8(arch.ARMOnARM64))
	case arch.Current == arch.X86:
		return uint8(uint8(device.OS)<<4 | uint8(arch.X86OnX64))
	}
	return uint8(uint8(device.OS)<<4 | uint8(arch.Current))
}
