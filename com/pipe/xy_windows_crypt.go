//go:build windows && crypt

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

package pipe

import "github.com/iDigitalFlame/xmt/util/crypt"

var (
	// PermEveryone is the SDDL string used in Windows Pipes to allow anyone to
	// write and read to the listening Pipe
	//
	// This can be used for Pipe communication between privilege boundaries.
	//
	// Can be applied with the ListenPerm function.
	PermEveryone = crypt.Get(20) // D:PAI(A;;FA;;;WD)(A;;FA;;;SY)

	// ErrTimeout is an error returned by the 'Dial*' functions when the
	// specified timeout was reached when attempting to connect to a Pipe.
	ErrTimeout = &errno{m: crypt.Get(21), t: true} // connection timeout
	// ErrEmptyConn is an error received when the 'Listen' function receives a
	// shortly lived Pipe connection.
	//
	// This error is only temporary and does not indicate any Pipe server failures.
	ErrEmptyConn = &errno{m: crypt.Get(22), t: true} // empty connection
)

// Format will ensure the path for this Pipe socket fits the proper OS based
// pathname. Valid path names will be returned without any changes.
func Format(s string) string {
	if len(s) > 2 && s[0] == '\\' && s[1] == '\\' {
		return s
	}
	return crypt.Get(23) + s // \\.\pipe\
}
