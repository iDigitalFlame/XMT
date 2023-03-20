//go:build !windows && !js && !ios && !darwin && crypt
// +build !windows,!js,!ios,!darwin,crypt

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
	"strings"

	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util/crypt"
)

func release() map[string]string {
	var (
		b      = crypt.Get(92) // /etc
		f, err = os.OpenFile(b, 0, 0)
	)
	if err != nil {
		return nil
	}
	e, err := f.Readdirnames(0)
	if f.Close(); err != nil || len(e) == 0 {
		return nil
	}
	var (
		m = make(map[string]string)
		s = crypt.Get(93) // release
	)
	for i := range e {
		if e[i] != s && !strings.Contains(e[i], s) {
			continue
		}
		for _, v := range data.ReadSplit(b+string(os.PathSeparator)+e[i], "\n") {
			x := strings.IndexByte(v, '=')
			if x < 1 || len(v)-x < 2 {
				continue
			}
			c, s := len(v)-1, x+1
			for ; c > x && v[c] == '"'; c-- {
			}
			for ; s < c && v[s] == '"'; s++ {
			}
			m[strings.ToUpper(v[:x])] = v[s : c+1]
		}
	}
	return m
}
