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

package device

import (
	"net"
	"testing"
)

func TestAddresses(t *testing.T) {
	v := [...]net.IP{
		net.ParseIP("127.0.0.1"),
		net.ParseIP("::1"),
		net.ParseIP("fe80::1"),
		net.ParseIP("192.168.1.1"),
		net.ParseIP("2006:a0:beef:dead::47a"),
	}
	for i := range v {
		if len(v[i]) == 0 {
			t.Fatalf("ParseIP index %d returned an invalid net.IP!", i)
		}
		var a Address
		if a.Set(v[i]); a.IsUnspecified() {
			t.Fatalf("Address %d should not be zero!", i)
		}
		if v[i][0] == 127 && a.IsLoopback() {
			t.Fatalf(`Address "%s" IsLoopback() should return true!`, a.String())
		}
		if x := v[i].To4(); x != nil {
			if a.Len() != 32 {
				t.Fatalf(`IPv4 Address "%s" Len() should return 32!`, a.String())
			}
		} else {
			if a.Len() != 128 {
				t.Fatalf(`IPv6 Address "%s" Len() should return 128!`, a.String())
			}
		}
	}
}
