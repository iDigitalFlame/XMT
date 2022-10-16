//go:build !windows && !js && !ios && !darwin && crypt

// Copyright (C) 2020 - 2022 iDigitalFlame
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
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/iDigitalFlame/xmt/util/crypt"
)

func release() map[string]string {
	var (
		f      = os.DirFS(crypt.Get(79))   // /etc
		e, err = fs.Glob(f, crypt.Get(80)) // *release*
	)
	if err != nil || len(e) == 0 {
		return nil
	}
	m := make(map[string]string)
	for i := range e {
		d, err := f.Open(e[i])
		if err != nil {
			continue
		}
		b, err := io.ReadAll(d)
		if d.Close(); err != nil || len(b) == 0 {
			continue
		}
		for _, v := range strings.Split(string(b), "\n") {
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
