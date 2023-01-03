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

package cfg

import (
	"testing"
)

func TestTransform(t *testing.T) {
	c := Pack(
		TransformB64,
		TransformB64Shift(10),
		TransformDNS("test.com"),
	)

	if _, err := c.Build(); err == nil {
		t.Fatalf("TestTransform(): Invalid build should have failed!")
	}

	if n := c.Len(); n != 14 {
		t.Fatalf(`TestTransform(): Len returned invalid size "%d" should ne "14"!`, n)
	}
	if c[0] != byte(TransformB64) {
		t.Fatalf(`TestTransform(): Invalid byte at position "0"!`)
	}
	if c[1] != byte(valB64Shift) || c[2] != 10 {
		t.Fatalf(`TestTransform(): Invalid byte at position "1"!`)
	}
	if c[3] != byte(valDNS) || c[6] != 't' {
		t.Fatalf(`TestTransform(): Invalid byte at position "3:6"!`)
	}
}
