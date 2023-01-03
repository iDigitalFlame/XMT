//go:build !windows && !js && !plan9

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

import "golang.org/x/sys/unix"

func uname() string {
	var (
		u   unix.Utsname
		err = unix.Uname(&u)
	)
	if err != nil {
		return ""
	}
	var (
		v = make([]byte, 65)
		i int
	)
	for ; i < 65; i++ {
		if u.Release[i] == 0 {
			break
		}
		v[i] = u.Release[i]
	}
	return string(v[:i])
}
