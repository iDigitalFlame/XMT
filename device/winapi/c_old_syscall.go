//go:build windows && !go1.18
// +build windows,!go1.18

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

package winapi

import "syscall"

func syscallN(t uintptr, a ...uintptr) (uintptr, uintptr, syscall.Errno) {
	switch len(a) {
	case 0:
		return syscall.Syscall(t, 0, 0, 0, 0)
	case 1:
		return syscall.Syscall(t, 1, a[0], 0, 0)
	case 2:
		return syscall.Syscall(t, 2, a[0], a[1], 0)
	case 3:
		return syscall.Syscall(t, 3, a[0], a[1], a[2])
	case 4:
		return syscall.Syscall6(t, 4, a[0], a[1], a[2], a[3], 0, 0)
	case 5:
		return syscall.Syscall6(t, 5, a[0], a[1], a[2], a[3], a[4], 0)
	case 6:
		return syscall.Syscall6(t, 6, a[0], a[1], a[2], a[3], a[4], a[5])
	case 7:
		return syscall.Syscall9(t, 7, a[0], a[1], a[2], a[3], a[4], a[5], a[6], 0, 0)
	case 8:
		return syscall.Syscall9(t, 8, a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], 0)
	case 9:
		return syscall.Syscall9(t, 9, a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8])
	case 10:
		return syscall.Syscall12(t, 10, a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8], a[9], 0, 0)
	case 11:
		return syscall.Syscall12(t, 11, a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8], a[9], a[10], 0)
	case 12:
		return syscall.Syscall12(t, 12, a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8], a[9], a[10], a[11])
	case 13:
		return syscall.Syscall15(t, 13, a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8], a[9], a[10], a[11], a[12], 0, 0)
	case 14:
		return syscall.Syscall15(t, 14, a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8], a[9], a[10], a[11], a[12], a[13], 0)
	case 15:
		return syscall.Syscall15(t, 15, a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8], a[9], a[10], a[11], a[12], a[13], a[14])
	}
	return 0, 0, syscall.EINVAL
}
