//go:build !windows
// +build !windows

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

package data

import "strings"

// ReadSplit is a *nix specific helper function that can be used to loop over
// a file delimited by the supplied string.
//
// This function replaces the open/read/split loop and will directly return the
// result to be used in a loop.
//
// If the file does not exist or an error occurs, the loop will be a NOP as this
// function returns nil.
func ReadSplit(s, sep string) []string {
	b, err := ReadFile(s)
	if err != nil {
		return nil
	}
	return strings.Split(string(b), sep)
}
