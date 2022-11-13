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

package subtle

import (
	"bytes"
	"testing"

	"github.com/iDigitalFlame/xmt/util"
)

func TestXOR(t *testing.T) {
	b, m := make([]byte, 64), make([]byte, 64)
	util.Rand.Read(b)
	copy(m, b)
	XorOp(b, []byte("this is my key!"))
	if XorOp(b, []byte("this is my key!")); !bytes.Equal(m, b) {
		t.Fatalf("TestXOR(): Xor bytes does not match the initial byte slice!")
	}
}
