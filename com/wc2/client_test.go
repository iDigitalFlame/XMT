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

package wc2

import "testing"

func TestRawParse(t *testing.T) {
	if _, err := rawParse("google.com"); err != nil {
		t.Fatalf(`Raw URL Parse "google.com" failed with error: %s!`, err.Error())
	}
	if _, err := rawParse("https://google.com"); err != nil {
		t.Fatalf(`Raw URL Parse "https://google.com" failed with error: %s!`, err.Error())
	}
	if _, err := rawParse("/google.com"); err != nil {
		t.Fatalf(`Raw URL Parse "/google.com" failed with error: %s!`, err.Error())
	}
	if _, err := rawParse("\\\\google.com"); err == nil {
		t.Fatalf(`Raw URL Parse "\\google.com" should have failed!`)
	}
	if _, err := rawParse("\\google.com"); err == nil {
		t.Fatalf(`Raw URL Parse "\google.com" should have failed!`)
	}
	if _, err := rawParse("derp:google.com"); err == nil {
		t.Fatalf(`Raw URL Parse "\google.com" should have failed!`)
	}
}
