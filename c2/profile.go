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

import "github.com/iDigitalFlame/xmt/c2/cfg"

// ProfileParser is a package level constant to be used when performing Profile
// loads from a raw binary source. This function will take the resulting byte
// array and Marshal it into a working Profile interface.
//
// This function starts out as nil, which in case the "c2/cfg" binary profile
// parser will be used.
//
// This is used to apply custom profiles to be used. Custom profiles can be Marshaled
// by exposing the 'MarshalBinary() ([]byte, error)' function.
var ProfileParser func(b []byte) (cfg.Profile, error)

func parseProfile(b []byte) (cfg.Profile, error) {
	if ProfileParser == nil {
		return cfg.Raw(b)
	}
	return ProfileParser(b)
}
