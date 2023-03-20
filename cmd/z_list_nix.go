//go:build !windows && !plan9 && !js
// +build !windows,!plan9,!js

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

package cmd

import (
	"os"
	"os/user"
	"syscall"

	"github.com/iDigitalFlame/xmt/util"
)

func getProcUser(p string) string {
	f, err := os.Lstat(p)
	if err != nil {
		return ""
	}
	s, ok := f.Sys().(*syscall.Stat_t)
	if !ok {
		return ""
	}
	u, err := user.LookupId(util.Uitoa(uint64(s.Uid)))
	if err != nil {
		return ""
	}
	return u.Username
}
