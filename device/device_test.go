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

package device

import "testing"

func TestDebug(t *testing.T) {
	t.Logf("TestDebug(): IsDebugged returned: %t", IsDebugged())
}
func TestExpand(t *testing.T) {
	v := [...]string{
		"${PWD}-1",
		"$PWD",
		"%PWD%",
	}
	for i := range v {
		if r := Expand(v[i]); v[i] == r {
			t.Fatalf(`TestExpand(): Expanded string "%s" equals non-expanded string "%s"!`, r, v)
		}
	}
}
