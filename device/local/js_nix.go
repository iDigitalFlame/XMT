//go:build !windows
// +build !windows

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
	"os"
	"os/user"
)

func getPPID() uint32 {
	return uint32(os.Getppid())
}
func getUsername() string {
	if u, err := user.Current(); err == nil {
		switch {
		case len(u.Username) > 0:
			return u.Username
		case len(u.Uid) > 0:
			return u.Uid
		case len(u.Name) > 0:
			return u.Name
		}
	}
	return "?"
}
