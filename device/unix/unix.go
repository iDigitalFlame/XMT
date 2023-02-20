//go:build !js && !plan9 && !windows
// +build !js,!plan9,!windows

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

// Package unix is a nix* specific package that assists with calling Unix/Linux/BSD
// specific functions and data gathering.
package unix

// Release returns the system "uname" release version as a string. The underlying
// system call is depdent on the underlying system.
//
// If any errors occur, this functions returns an empty string.
func Release() string {
	var (
		u   utsName
		err = uname(&u)
	)
	if err != nil {
		return ""
	}
	var (
		v [257]byte
		i int
	)
	for ; i < len(u.Release); i++ {
		if u.Release[i] == 0 {
			break
		}
		v[i] = byte(u.Release[i])
	}
	return string(v[:i])
}

// IsMachine64 returns true if the underlying kernel reports the machine type as
// one that runs on a 64bit CPU (in 64bit mode). Otherwise, it returns false.
//
// If any errors occur, this functions returns false.
func IsMachine64() bool {
	var (
		u   utsName
		err = uname(&u)
	)
	if err != nil {
		return false
	}
	switch {
	case u.Machine[10] == 0 && u.Machine[0] == 'a' && u.Machine[6] == '4' && u.Machine[5] == '6' && u.Machine[9] == 'e': // Match aarch64_be
	case u.Machine[7] == 0 && u.Machine[0] == 'a' && u.Machine[6] == '4' && u.Machine[5] == '6': // Match aarch64
	case u.Machine[6] == 0 && u.Machine[0] == 'a' && u.Machine[4] == '8' && u.Machine[3] == 'v': // Match armv8l and armv8b
	case u.Machine[6] == 0 && u.Machine[0] == 'x' && u.Machine[5] == '4' && u.Machine[4] == '6': // Match x86_64
	default:
		return false
	}
	return true
}
