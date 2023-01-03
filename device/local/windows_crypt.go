//go:build windows && crypt

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
	"strconv"

	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/device/winapi/registry"
	"github.com/iDigitalFlame/xmt/util/crypt"
)

func sysID() []byte {
	if s, err := winapi.GetSystemSID(); err == nil {
		return []byte(s.String())
	}
	// 0x101 - KEY_WOW64_64KEY | KEY_QUERY_VALUE
	k, err := registry.Open(registry.KeyLocalMachine, crypt.Get(51), 0x101) // Software\Microsoft\Cryptography
	if err != nil {
		return nil
	}
	v, _, err := k.String(crypt.Get(52)) // MachineGuid
	if k.Close(); err == nil {
		return []byte(v)
	}
	return nil
}
func version() string {
	var n string
	// 0x101 - KEY_WOW64_64KEY | KEY_QUERY_VALUE
	if k, err := registry.Open(registry.KeyLocalMachine, crypt.Get(53), 0x101); err == nil { // Software\Microsoft\Windows NT\CurrentVersion
		n, _, _ = k.String(crypt.Get(54)) // ProductName
		k.Close()
	}
	var (
		j, y, h = winapi.GetVersionNumbers()
		b       = strconv.FormatUint(uint64(h), 10)
		v       string
	)
	if y > 0 {
		v = strconv.FormatUint(uint64(j), 10) + "." + strconv.FormatUint(uint64(y), 10)
	} else {
		v = strconv.FormatUint(uint64(j), 10)
	}
	switch {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return crypt.Get(55) // Windows
	case len(n) == 0 && len(b) > 0 && len(v) > 0:
		return crypt.Get(55) + " (" + v + ", " + b + ")" // Windows
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return crypt.Get(55) + " (" + v + ")" // Windows
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return crypt.Get(55) + " (" + b + ")" // Windows
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return n + " (" + v + ", " + b + ")"
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return n + " (" + v + ")"
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return n + " (" + b + ")"
	}
	return crypt.Get(55) // Windows
}
