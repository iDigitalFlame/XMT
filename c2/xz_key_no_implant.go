//go:build !implant && !nokeyset
// +build !implant,!nokeyset

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

package c2

import (
	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
)

func (s *Session) keyListenerRegenerate(l string, n *com.Packet) error {
	if err := s.keys.Read(n); err != nil {
		if cout.Enabled {
			s.log.Error("[%s:%s/Crypt] Reading KeyPair failed: %s!", l, s.ID, err)
		}
		return err
	}
	if err := s.keys.Sync(); err != nil {
		if cout.Enabled {
			s.log.Error("[%s:%s/Crypt] Syncing KeyPair failed: %s!", l, s.ID, err)
		}
		return err
	}
	if cout.Enabled {
		bugtrack.Track("c2.(*Session).keyListenerRegenerate(): %s KeyPair details updated! [Public %s, Shared: %v]", s.ID, s.keys.Public, s.keys.Shared())
	}
	return nil
}
func (s *Session) keyCryptAndUpdate(l string, n *com.Packet, d bool) error {
	if d {
		n.KeyCrypt(s.keys)
	}
	if n.Flags&com.FlagCrypt == 0 || n.Empty() {
		return nil
	}
	return s.keyListenerRegenerate(l, n)
}
func (s *Session) keyListenerInit(k data.PrivateKey, l string, n *com.Packet) error {
	if err := s.keys.Read(n); err != nil {
		if cout.Enabled {
			s.log.Error("[%s:%s/Crypt] Generating KeyPair failed: %s!", l, s.ID, err)
		}
		return err
	}
	if err := s.keys.FillPrivate(k); err != nil {
		if cout.Enabled {
			s.log.Error("[%s:%s/Crypt] Syncing KeyPair failed: %s!", l, s.ID, err)
		}
		return err
	}
	if cout.Enabled {
		bugtrack.Track("c2.(*Session).keyListenerInit(): %s KeyPair details updated! [Public: %s, Shared: %v]", s.ID, s.keys.Public, s.keys.Shared())
	}
	return nil
}
