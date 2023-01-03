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

package regedit

import (
	"testing"

	"github.com/iDigitalFlame/xmt/device/winapi/registry"
)

func TestEntry(t *testing.T) {
	e := Entry{
		Data: []byte{0xFF, 0x0, 0x10, 0x20},
		Type: registry.TypeDword,
	}
	v, err := e.ToInteger()
	if err != nil {
		t.Fatalf("TestEntry(): ToInteger returned an error: %s", err.Error())
	}
	if v != 0x201000FF {
		t.Fatalf(`TestEntry(): ToInteger result "0x%X" does not match expected "0x201000FF"!`, v)
	}
	z := Entry{
		Data: []byte{0xFF, 0, 0x10, 0x20, 0x40, 0x40, 0x40, 0x50},
		Type: registry.TypeQword,
	}
	if v, err = z.ToInteger(); err != nil {
		t.Fatalf("TestEntry(): ToInteger returned an error: %s", err.Error())
	}
	if v != 0x50404040201000FF {
		t.Fatalf(`TestEntry(): ToInteger result "0x%X" does not match expected "0x50404040201000FF"!"`, v)
	}
}
