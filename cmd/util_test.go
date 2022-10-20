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

package cmd

import "testing"

func TestSplit(t *testing.T) {
	v := [...]struct {
		Cmd    string
		Result []string
	}{
		{"cmd.exe /c", []string{"cmd.exe", "/c"}},
		{`notepad.exe "derp"`, []string{"notepad.exe", "derp"}},
		{`C:\Windows\system32\calc.exe "open1" "open 2" open 3`, []string{`C:\Windows\system32\calc.exe`, "open1", "open 2", "open", "3"}},
		{`test1 /test2 -test3 'test 3' "test 4"`, []string{"test1", "/test2", "-test3", "test3", "test 4"}},
		{`test1 /test2 -test3 'test 3' "test 4" "test '5'"`, []string{"test1", "/test2", "-test3", "test3", "test 4", "test '5'"}},
	}
	for i := range v {
		r := Split(v[i].Cmd)
		if len(r) != len(v[i].Result) {
			t.Fatalf(`Split result %v does not match expected %v for "%s"!`, r, v[i].Result, v[i].Cmd)
		}
	}
}
