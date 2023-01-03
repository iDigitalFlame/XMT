//go:build (darwin || ios) && !crypt

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
	"os/exec"
	"strings"
)

func sysID() []byte {
	o, err := exec.Command("/usr/sbin/ioreg", "-rd1", "-c", "IOPlatformExpertDevice").CombinedOutput()
	if err != nil || len(o) == 0 {
		return nil
	}
	for _, v := range strings.Split(string(o), "\n") {
		if !strings.Contains(v, `"IOPlatformUUID"`) {
			continue
		}
		x := strings.IndexByte(v, '=')
		if x < 14 || len(v)-x < 2 {
			continue
		}
		c, s := len(v)-1, x+1
		for ; c > x && (v[c] == '"' || v[c] == ' ' || v[c] == 9); c-- {
		}
		for ; s < c && (v[s] == '"' || v[s] == ' ' || v[s] == 9); s++ {
		}
		if s == c || s > len(v) || s > c {
			continue
		}
		return []byte(v[s : c+1])
	}
	return nil
}
func version() string {
	var b, n, v string
	if o, err := exec.Command("/usr/bin/sw_vers").CombinedOutput(); err == nil {
		m := make(map[string]string)
		for _, v := range strings.Split(string(o), "\n") {
			x := strings.IndexByte(v, ':')
			if x < 1 || len(v)-x < 2 {
				continue
			}
			c, s := len(v)-1, x+1
			for ; c > x && (v[c] == ' ' || v[c] == 9); c-- {
			}
			for ; s < c && (v[s] == ' ' || v[s] == 9); s++ {
			}
			m[strings.ToUpper(v[:x])] = v[s : c+1]
		}
		n = m["PRODUCTNAME"]
		b = m["BUILDVERSION"]
		v = m["PRODUCTVERSION"]
	}
	if len(v) == 0 {
		v = uname()
	}
	switch {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return "MacOS"
	case len(n) == 0 && len(b) > 0 && len(v) > 0:
		return "MacOS (" + v + ", " + b + ")"
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return "MacOS (" + v + ")"
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return "MacOS (" + b + ")"
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return n + " (" + v + ", " + b + ")"
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return n + " (" + v + ")"
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return n + " (" + b + ")"
	}
	return "MacOS"
}
