//go:build amd64 || 386
// +build amd64 386

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

package device

// IsVirtual attempts to determine if the underlying device is inside a container
// or is running in a virtual machine.
//
// If this returns true, it is suspected that a non-physical device is present.
//
// Different versions of this function are used depending on CPU type.
//   - For x86/x64/amd64 this function uses the CPUID instruction.
//     See https://en.wikipedia.org/wiki/CPUID for more info.
func IsVirtual() bool {
	_, _, c, d := cpuid(0x1, 0) // CPU Features
	if c&(1<<31) != 0 {         // Hypervisor flag
		return true
	}
	t := d&(1<<29) == 0 // Thermal monitor
	// Save this one for later, most VMs /don't/ have thermal monitoring but we
	// will check all other options first.
	_, b, c, d := cpuid(0x40000000, 0) // Hypervisor CPUID
	if d == 0 || b == 0 || c == 0 {
		if !t {
			// Fall back to thermal if CPUID returns '0' as this can be overridden
			// in some hypervisors.
			a, b, c, _ := cpuid(0x6, 0) // Thermal and power management
			if t = a <= 4 && (b == 0 || c == 0); !t {
				return false
			}
		}
	}
	// We don't need to check the name /but/ we're just making sure that the
	// name is not all zero's first.
	n := [16]byte{
		byte((b >> 0)), byte((b >> 8)), byte((b >> 16)), byte((b >> 24)),
		byte((c >> 0)), byte((c >> 8)), byte((c >> 16)), byte((c >> 24)),
		byte((d >> 0)), byte((d >> 8)), byte((d >> 16)), byte((d >> 24)),
	}
	var z int
	for i := range n {
		if n[i] > 0 {
			continue
		}
		z++
	}
	// Less than 10 chars are zeros.
	if z < 10 {
		return true
	}
	// Fallback to temp.
	return t
}

// cpuid is implemented in ASM file "cpu.s"
func cpuid(arg1 uint32, arg2 uint32) (uint32, uint32, uint32, uint32)
