//go:build windows && go1.16
// +build windows,go1.16

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

import "unsafe"

func destoryAllM() {
	for i := uintptr(allm); ; {
		mdestroy(i)
		n := (*uintptr)(unsafe.Pointer(i + ptrNext))
		if n == nil || *n == 0 {
			break // Reached bottom of linked list
		}
		i = *n
	}
}

//go:linkname mdestroy runtime.mdestroy
func mdestroy(uintptr)
