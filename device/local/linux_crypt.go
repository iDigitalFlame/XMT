//go:build (android || linux) && crypt
// +build android linux
// +build crypt

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
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device/unix"
	"github.com/iDigitalFlame/xmt/util/crypt"
)

func sysID() []byte {
	b, err := data.ReadFile(crypt.Get(63)) // /var/lib/dbus/machine-id
	if err == nil {
		return b
	}
	b, _ = data.ReadFile(crypt.Get(64)) // /etc/machine-id
	return b
}
func version() string {
	var (
		ok      bool
		b, n, v string
	)
	if m := release(); len(m) > 0 {
		b = m[crypt.Get(2)]                // ID
		if n, ok = m[crypt.Get(65)]; !ok { // PRETTY_NAME
			n = m[crypt.Get(66)] // NAME
		}
		if v, ok = m[crypt.Get(67)]; !ok { // VERSION_ID
			v = m[crypt.Get(68)] // VERSION
		}
	}
	if len(v) == 0 {
		v = unix.Release()
	}
	switch {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return crypt.Get(69) // Linux
	case len(n) == 0 && len(b) > 0 && len(v) > 0:
		return crypt.Get(69) + " (" + v + ", " + b + ")" // Linux
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return crypt.Get(69) + " (" + v + ")" // Linux
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return crypt.Get(69) + " (" + b + ")" // Linux
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return n + " (" + v + ", " + b + ")"
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return n + " (" + v + ")"
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return n + " (" + b + ")"
	}
	return crypt.Get(69) // Linux
}
