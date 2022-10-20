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

// Package arch contains platform architecture constants and a specific type
// that correlates to the current system architecture.
//
package arch

const (
	// X64 represents the 64-bit chipset family.
	X64 Architecture = 0x0
	// X86 represents the 32-bit chipset family.
	X86 Architecture = 0x1
	// ARM represents the ARM chipset family.
	ARM Architecture = 0x2
	// PowerPC represents the PowerPC chipset family.
	PowerPC Architecture = 0x3
	// Mips represents the MIPS chipset family.
	Mips Architecture = 0x4
	// Risc represents the RiscV chipset family.
	Risc Architecture = 0x5
	// ARM64 represents the ARM64 chipset family.
	ARM64 Architecture = 0x6
	// WASM represents the WASM/JavaScript software family.
	WASM Architecture = 0x7
	// Loong64 represents the LoongArch64 chipset family.
	Loong64 Architecture = 0x8
	// X86OnX64 represents the 64-bit chipset family, but the executable
	// is compiled for X86.
	X86OnX64 Architecture = 0x9
	// ARMOnARM64 represents the ARM64 chipset family, but the executable
	// is compiled for ARM.
	ARMOnARM64 Architecture = 0xA
	// Unknown represents an unknown chipset family.
	Unknown Architecture = 0xF
)

// Architecture is a number representation of the chipset architecture.
type Architecture uint8
