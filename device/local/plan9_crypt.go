//go:build plan9 && crypt
// +build plan9,crypt

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
	"github.com/iDigitalFlame/xmt/util/crypt"
)

func sysID() []byte {
	if b, err := data.ReadFile(crypt.Get(73)); err == nil { // /etc/hostid
		return b
	}
	o, _ := output(crypt.Get(74)) // kenv -q smbios.system.uuid
	return o
}
func version() string {
	var (
		ok      bool
		b, n, v string
	)
	if m := release(); len(m) > 0 {
		b = m[crypt.Get(1)]                // ID
		if n, ok = m[crypt.Get(75)]; !ok { // PRETTY_NAME
			n = m[crypt.Get(75)[7:]] // PRETTY_NAME
		}
		if v, ok = m[crypt.Get(76)]; !ok { // VERSION_ID
			v = m[crypt.Get(76)[0:7]] // VERSION_ID
		}
	}
	if len(v) == 0 {
		v = crypt.Get(91) // plan9
	}
	switch {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return crypt.Get(91) // plan9
	case len(n) == 0 && len(b) > 0 && len(v) > 0:
		return crypt.Get(91) + " (" + v + ", " + b + ")" // plan9
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return crypt.Get(91) + " (" + v + ")" // plan9
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return crypt.Get(91) + " (" + b + ")" // plan9
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return n + " (" + v + ", " + b + ")"
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return n + " (" + v + ")"
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return n + " (" + b + ")"
	}
	return crypt.Get(91) // plan9
}
