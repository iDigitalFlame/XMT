//go:build !windows && !js
// +build !windows,!js

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
	"strconv"

	"github.com/iDigitalFlame/xmt/data"
)

func readProcStats(s string) (string, uint32) {
	var (
		n string
		p uint32
	)
	for _, e := range data.ReadSplit(s, "\n") {
		if len(n) > 0 && p > 0 {
			return n, p
		}
		if len(e) > 6 && e[0] == 'N' && e[2] == 'm' && e[4] == ':' {
			for i := 5; i < len(e); i++ {
				if e[i] == ' ' || e[i] == 9 || e[i] == '\t' {
					continue
				}
				n = e[i:]
				break
			}
		}
		if len(e) > 5 && e[0] == 'P' && e[1] == 'P' && e[3] == 'd' && e[4] == ':' {
			for i := 5; i < len(e); i++ {
				if e[i] == ' ' || e[i] == 9 || e[i] == '\t' {
					continue
				}
				if v, err := strconv.ParseUint(e[i:], 10, 32); err == nil {
					p = uint32(v)
					break
				}
			}
		}
	}
	return n, p
}
