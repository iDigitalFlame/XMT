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

import "testing"

func TestWrap(t *testing.T) {
	c := Pack(
		WrapAES([]byte("0123456789ABCDEF0123456789ABCDEF"), []byte("ABCDEF9876543210")),
		WrapCBK(10, 20, 30, 40),
		WrapCBKSize(64, 10, 20, 30, 40),
	)

	if _, err := c.Build(); err != nil {
		t.Fatalf("TestWrap(): Build failed with error: %s!", err.Error())
	}

	if n := c.Len(); n != 63 {
		t.Fatalf(`TestWrap(): Len returned invalid size "%d" should ne "63"!`, n)
	}
	if c[51] != byte(valCBK) {
		t.Fatalf(`TestWrap(): Invalid byte at position "51"!`)
	}
	if c[53] != 10 {
		t.Fatalf(`TestWrap(): Invalid byte at position "52"!`)
	}
	if c[57] != byte(valCBK) {
		t.Fatalf(`TestWrap(): Invalid byte at position "57"!`)
	}
	if c[58] != 64 || c[60] != 20 {
		t.Fatalf(`TestWrap(): Invalid byte at position "58:60"!`)
	}
}
