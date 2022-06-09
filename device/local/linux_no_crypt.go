//go:build (android || linux) && !crypt

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

import "os"

func sysID() []byte {
	b, err := os.ReadFile("/var/lib/dbus/machine-id")
	if err == nil {
		return b
	}
	b, _ = os.ReadFile("/etc/machine-id")
	return b
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
		v = uname()
	}
	switch {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return "Linux"
	case len(n) == 0 && len(b) > 0 && len(v) > 0:
		return "Linux (" + v + ", " + b + ")"
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return "Linux (" + v + ")"
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return "Linux (" + b + ")"
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return n + " (" + v + ", " + b + ")"
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return n + " (" + v + ")"
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return n + " (" + b + ")"
	}
	return "Linux"
}
