//go:build !implant

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

package arch

// String returns the name of the Architecture type.
func (a Architecture) String() string {
	switch a {
	case X86:
		return "32bit"
	case X64:
		return "64bit"
	case ARM:
		return "ARM"
	case WASM:
		return "WASM"
	case Risc:
		return "RiscV"
	case Mips:
		return "MIPS"
	case ARM64:
		return "ARM64"
	case PowerPC:
		return "PowerPC"
	case Loong64:
		return "Loong64"
	case X86OnX64:
		return "32bit [64bit]"
	case ARMOnARM64:
		return "ARM [ARM64]"
	}
	return "Unknown"
}
