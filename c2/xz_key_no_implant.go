//go:build !implant && !nokeyset

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

import (
	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
)

func (s *Session) sessionKeyInit(l string, n *com.Packet) {
	if v, err := n.Read(s.key[:]); v != data.KeySize || err != nil {
		if cout.Enabled {
			s.log.Warning("[%s:%s/Crypt] Error (Re)Generating key set!", l, s.ID)
		}
		return
	}
	if generateKeys(&s.key, s.ID); cout.Enabled {
		s.log.Debug("[%s:%s/Crypt] (Re)Generated key set!", l, s.ID)
	}
	if bugtrack.Enabled {
		bugtrack.Track("c2.Session.sessionKeyInit(): %s Key details [%v].", s.ID, s.key)
	}
}
func (s *Session) sessionKeyUpdate(l string, n *com.Packet, d bool) {
	if d {
		n.Crypt(&s.key)
	}
	if n.Flags&com.FlagCrypt == 0 || n.Empty() {
		return
	}
	s.sessionKeyInit(l, n)
}
