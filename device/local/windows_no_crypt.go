//go:build windows && !crypt
// +build windows,!crypt

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
	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/device/winapi/registry"
	"github.com/iDigitalFlame/xmt/util"
)

func sysID() []byte {
	if s, err := winapi.GetSystemSID(); err == nil {
		return []byte(s.String())
	}
	// 0x101 - KEY_WOW64_64KEY | KEY_QUERY_VALUE
	k, err := registry.Open(registry.KeyLocalMachine, `Software\Microsoft\Cryptography`, 0x101)
	if err != nil {
		return nil
	}
	v, _, err := k.String("MachineGuid")
	if k.Close(); err == nil {
		return []byte(v)
	}
	return nil
}
func version() string {
	var n string
	// 0x101 - KEY_WOW64_64KEY | KEY_QUERY_VALUE
	if k, err := registry.Open(registry.KeyLocalMachine, `Software\Microsoft\Windows NT\CurrentVersion`, 0x101); err == nil {
		n, _, _ = k.String("ProductName")
		k.Close()
	}
	var (
		j, y, h = winapi.GetVersionNumbers()
		b       = util.Uitoa(uint64(h))
		v       string
	)
	if y > 0 {
		v = util.Uitoa(uint64(j)) + "." + util.Uitoa(uint64(y))
	} else {
		v = util.Uitoa(uint64(j))
	}
	switch {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return "Windows"
	case len(n) == 0 && len(b) > 0 && len(v) > 0:
		return "Windows (" + v + ", " + b + ")"
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return "Windows (" + v + ")"
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return "Windows (" + b + ")"
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return n + " (" + v + ", " + b + ")"
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return n + " (" + v + ")"
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return n + " (" + b + ")"
	}
	return "Windows"
}
