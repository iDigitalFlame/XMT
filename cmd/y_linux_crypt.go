//go:build !windows && !js && crypt
// +build !windows,!js,crypt

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
	"sort"
	"strconv"

	"github.com/iDigitalFlame/xmt/util/crypt"
)

// Processes attempts to gather the current running Processes and returns them
// as a slice of ProcessInfo structs, otherwise any errors are returned.
func Processes() ([]ProcessInfo, error) {
	f, err := os.OpenFile(crypt.Get(11), 0, 0) // /proc/
	if err != nil {
		return nil, err
	}
	l, err := f.Readdir(0)
	if f.Close(); err != nil {
		return nil, err
	}
	if len(l) == 0 {
		return nil, nil
	}
	var (
		v uint64
		p uint32
		n string
		r = make(processList, 0, len(l)/2)
	)
	for i := range l {
		if !l[i].IsDir() {
			continue
		}
		if n = l[i].Name(); len(n) < 1 || n[0] < 48 || n[0] > 57 {
			continue
		}
		if v, err = strconv.ParseUint(n, 10, 32); err != nil {
			continue
		}
		n, p = readProcStats(
			crypt.Get(11) + // /proc/
				n +
				crypt.Get(12), // /status
		)
		if len(n) == 0 {
			continue
		}
		r = append(r, ProcessInfo{
			PID:  uint32(v),
			PPID: p,
			Name: n,
			User: getProcUser(crypt.Get(11) + n), // /proc/
		})
	}
	sort.Sort(r)
	return r, nil
}
