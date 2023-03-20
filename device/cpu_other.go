//go:build !amd64 && !386
// +build !amd64,!386

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
	return isVirtual()
}
func isKnownVendor(b []byte) bool {
	switch len(b) {
	case 0:
		return false
	case 3: // KVM, VMware, Xen
		if (b[0] == 'K' || b[0] == 'k') && (b[1] == 'V' || b[1] == 'v') && (b[2] == 'M' || b[2] == 'm') {
			return true
		}
		if (b[0] == 'V' || b[0] == 'v') && (b[1] == 'M' || b[1] == 'm') && (b[2] == 'W' || b[2] == 'w') {
			return true
		}
		if (b[0] == 'X' || b[0] == 'x') && (b[1] == 'E' || b[1] == 'e') && (b[2] == 'N' || b[2] == 'n') {
			return true
		}
	case 4: // Qemu
		if (b[0] == 'Q' || b[0] == 'q') && (b[1] == 'E' || b[1] == 'e') && (b[2] == 'M' || b[2] == 'm') && (b[3] == 'U' || b[3] == 'u') {
			return true
		}
	case 5: // Bochs, Bhyve
		if (b[0] == 'B' || b[0] == 'b') && (b[1] == 'O' || b[1] == 'o') && (b[3] == 'H' || b[3] == 'h') && (b[4] == 'S' || b[4] == 's') {
			return true
		}
		if (b[0] == 'B' || b[0] == 'b') && (b[1] == 'H' || b[1] == 'h') && (b[3] == 'V' || b[3] == 'v') && (b[4] == 'E' || b[4] == 'e') {
			return true
		}
	case 6: // VMware
		if (b[0] == 'V' || b[0] == 'v') && (b[1] == 'M' || b[1] == 'm') && (b[2] == 'W' || b[2] == 'w') && (b[5] == 'E' || b[5] == 'e') {
			return true
		}
	case 7: // Hyper-V
		if (b[0] == 'H' || b[0] == 'h') && (b[1] == 'Y' || b[1] == 'y') && b[5] == '-' && (b[6] == 'V' || b[6] == 'v') {
			return true
		}
	case 8: // KubeVirt
		if (b[0] == 'K' || b[0] == 'k') && (b[1] == 'U' || b[1] == 'u') && (b[4] == 'V' || b[4] == 'v') && (b[5] == 'I' || b[5] == 'i') {
			return true
		}
	case 9: // OpenStack, Parallels
		if (b[0] == 'O' || b[0] == 'o') && (b[1] == 'P' || b[2] == 'p') && (b[4] == 'S' || b[4] == 's') && (b[7] == 'C' || b[7] == 'c') {
			return true
		}
		if (b[0] == 'P' || b[0] == 'p') && (b[2] == 'R' || b[2] == 'r') && (b[5] == 'L' || b[5] == 'l') && (b[7] == 'L' || b[7] == 'l') {
			return true
		}
	case 10: // VirtualBox
		if (b[0] == 'V' || b[0] == 'v') && (b[3] == 'T' || b[3] == 't') && (b[7] == 'B' || b[7] == 'b') && (b[9] == 'X' || b[9] == 'x') {
			return true
		}
	}
	// Non-fixed tests
	// Amazon *
	if len(b) > 7 && (b[0] == 'A' || b[0] == 'a') && (b[1] == 'M' || b[1] == 'm') && (b[2] == 'A' || b[2] == 'a') && (b[3] == 'Z' || b[3] == 'z') {
		return true
	}
	// innotek *
	if len(b) > 8 && (b[0] == 'I' || b[0] == 'i') && (b[2] == 'N' || b[2] == 'n') && (b[3] == 'O' || b[3] == 'o') && (b[6] == 'K' || b[6] == 'k') {
		return true
	}
	// Apple Virt*
	if len(b) > 10 && (b[0] == 'A' || b[0] == 'a') && (b[2] == 'P' || b[2] == 'p') && (b[3] == 'L' || b[3] == 'l') && (b[6] == 'V' || b[6] == 'v') {
		return true
	}
	return false
}
