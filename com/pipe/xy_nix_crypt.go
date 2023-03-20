//go:build !windows && crypt
// +build !windows,crypt

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

import (
	"os"

	"github.com/iDigitalFlame/xmt/util/crypt"
)

// PermEveryone is the Linux permission string used in sockets to allow anyone to write and read
// to the listening socket. This can be used for socket communication between privilege boundaries.
// This can be applied to the ListenPerm function.
var PermEveryone = crypt.Get(19) // 0766

// Format will ensure the path for this Pipe socket fits the proper OS based pathname. Valid path names will be
// returned without any changes.
func Format(s string) string {
	if len(s) > 0 && s[0] != '/' {
		var (
			p      = crypt.Get(20) + s // /var/run/
			f, err = os.OpenFile(p, 0x242, 0600)
			// 0x242 - CREATE | TRUNCATE | RDWR
		)
		if err != nil {
			return crypt.Get(21) + s // /tmp/
		}
		f.Close()
		os.Remove(p)
		return p
	}
	return s
}
