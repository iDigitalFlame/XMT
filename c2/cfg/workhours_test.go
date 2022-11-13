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

package cfg

import (
	"testing"
	"time"
)

func TestWorkhours(t *testing.T) {
	var (
		w = time.Now().Weekday()
		h = &WorkHours{Days: (127) &^ (1 << w)}
	)
	if d := h.Work(); d == 0 {
		t.Fatalf("TestWorkhours(): Work time should not be zero!")
	}
	n := byte('R')
	if w != time.Thursday {
		n = w.String()[0]
	}
	c := make([]rune, 0, 7)
	for i, v := range "SMTWRFS" {
		if byte(v) != n {
			c = append(c, v)
			continue
		}
		if n == 'S' && ((w != time.Sunday && i == 0) || (w != time.Saturday && i > 0)) {
			c = append(c, 'S')
		}
	}
	if r := string(c); r != h.String() {
		t.Fatalf(`TestWorkhours(): Work String() should be "%s" not "%s"!`, r, h.String())
	}
}
