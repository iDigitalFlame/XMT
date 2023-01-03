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

package util

import "testing"

func TestBuilder(t *testing.T) {
	var b Builder
	for i := 0; i < 32; i++ {
		b.WriteString("hello ")
		b.Write([]byte("world!"))
		b.InsertByte('!')
		if n := b.Len(); n > b.Cap() {
			t.Fatalf(`TestBuilder(): Builder capacity value "%d" is not greater than the expected size!`, n)
		}
		if n := b.Len(); n != len("!hello world!") {
			t.Fatalf(`TestBuilder(): Builder length value "%d" does not match the expected size!`, n)
		}
		if v := b.String(); v != "!hello world!" {
			t.Fatalf(`TestBuilder(): Builder string output value "%s" does not match the expected value!`, v)
		}
		if v := b.Output(); v != "!hello world!" {
			t.Fatalf(`TestBuilder(): Builder output value "%s" does not match the expected value!`, v)
		}
		if n := b.Len(); n != 0 {
			t.Fatalf(`TestBuilder(): Builder size value "%d" does not match the expected value of zero!`, n)
		}
	}
}
