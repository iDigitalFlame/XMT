//go:build crypt

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

package man

import "github.com/iDigitalFlame/xmt/util/crypt"

var (
	local = crypt.Get(108) // localhost:
	execA = crypt.Get(1)   // *.so
	execB = crypt.Get(2)   // *.dll
	execC = crypt.Get(3)   // *.exe
)

func (o objSync) String() string {
	switch o {
	case Mutex:
		return crypt.Get(109) // mutex
	case Event:
		return crypt.Get(110) // event
	case Mailslot:
		return crypt.Get(111) // mailslot
	case Semaphore:
		return crypt.Get(112) // semaphore
	}
	return crypt.Get(109) // mutex
}
