//go:build plan9 && !crypt

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
	"os"
	"os/exec"
)

func sysID() []byte {
	if b, err := os.ReadFile("/etc/hostid"); err == nil {
		return b
	}
	o, _ := exec.Command("kenv", "-q", "smbios.system.uuid").CombinedOutput()
	return o
}
func version() string {
	var (
		ok      bool
		b, n, v string
	)
	if m := release(); len(m) > 0 {
		b = m["ID"]
		if n, ok = m["PRETTY_NAME"]; !ok {
			n = m["NAME"]
		}
		if v, ok = m["VERSION_ID"]; !ok {
			v = m["VERSION"]
		}
	}
	if len(v) == 0 {
		v = "plan9"
	}
	switch {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return "plan9"
	case len(n) == 0 && len(b) > 0 && len(v) > 0:
		return "plan9 (" + v + ", " + b + ")"
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return "plan9 (" + v + ")"
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return "plan9 (" + b + ")"
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return n + " (" + v + ", " + b + ")"
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return n + " (" + v + ")"
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return n + " (" + b + ")"
	}
	return "plan9"
}
