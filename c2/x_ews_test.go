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

package c2

import (
	"strconv"
	"testing"
)

func TestContainer(t *testing.T) {
	var c container
	for i := 0; i < 32; i++ {
		v := "test-container-" + strconv.Itoa(i)
		c.Set(v)
		c.Wrap()
		c.Unwrap()
		if s := c.String(); s != v {
			t.Fatalf(`TestContainer(): Unwrapped string "%s" should equal test string "%s"!`, s, v)
		}
	}
}
