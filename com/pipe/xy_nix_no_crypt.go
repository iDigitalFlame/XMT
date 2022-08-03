//go:build !windows && !crypt

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

package pipe

import (
	"os"
	"path/filepath"
)

// PermEveryone is the Linux permission string used in sockets to allow anyone to write and read
// to the listening socket. This can be used for socket communcation between privilege boundaries.
// This can be applied to the ListenPerm function.
var PermEveryone = "0766"

// Format will ensure the path for this Pipe socket fits the proper OS based pathname. Valid pathnames will be
// returned without any changes.
func Format(s string) string {
	if !filepath.IsAbs(s) {
		var (
			p      = "/var/run/" + s
			f, err = os.OpenFile(p, 0x242, 0o400)
			// 0x242 - CREATE | TRUNCATE | RDWR
		)
		if err != nil {
			return "/tmp/" + s
		}
		f.Close()
		os.Remove(p)
		return p
	}
	return s
}
