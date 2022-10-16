//go:build windows && (386 || s390x || arm)

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
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/device/arch"
	"github.com/iDigitalFlame/xmt/device/winapi"
)

func systemType() uint8 {
	// NOTE(dij): Check if we're running under WOW64 and report the /actual/
	//            system arch, since we only know what we've been built as.
	//            Apparently applies to x86 AND ARM!
	// TODO(dij): Is there a thing like this for *nix?
	//            Might have to look into it.
	switch r, _ := winapi.IsWow64Process(); {
	case r && device.Arch == arch.X86:
		return uint8(uint8(device.OS)<<4 | uint8(arch.X86OnX64))
	case r && device.Arch == arch.ARM:
		return uint8(uint8(device.OS)<<4 | uint8(arch.ARMOnARM64))
	}
	return uint8(uint8(device.OS)<<4 | uint8(arch.Current))
}
