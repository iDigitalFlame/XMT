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

package text

import "testing"

func TestRandomString(t *testing.T) {
	if v, c := Rand.String(10), Rand.String(10); v == c {
		t.Fatalf(`TestRandomString(): Random string 2 "%s" should not equal random string 2 "%s"!`, v, c)
	}
}
func TestRandomSetUpper(t *testing.T) {
	for _, v := range Upper.String(16) {
		if v < 'A' || v > 'Z' {
			t.Fatalf(`TestRandomString(): Non-upper character "%c" found in 'Upper' generator!`, v)
		}
	}
}
func TestRandomSetLower(t *testing.T) {
	for _, v := range Lower.String(16) {
		if v < 'a' || v > 'z' {
			t.Fatalf(`TestRandomString(): Non-lower character "%c" found in 'Lower' generator!`, v)
		}
	}
}
func TestRandomSetNumbers(t *testing.T) {
	for _, v := range Numbers.String(16) {
		if v < '0' || v > '9' {
			t.Fatalf(`TestRandomString(): Non-number character "%c" found in 'Numbers' generator!`, v)
		}
	}
}
