//go:build !nokeyset

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
	"crypto/rand"

	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
)

func (s *Session) keyCheck() {
	if s.keyNew == nil {
		return
	}
	if cout.Enabled {
		s.log.Debug("[%s] Regenerated KeyCrypt key set!", s.ID)
	}
	copy(s.key[:], (*s.keyNew)[:])
	if generateKeys(&s.key, s.ID); bugtrack.Enabled {
		bugtrack.Track("c2.(*Session).keyCheck(): %s KeyCrypt details %v.", s.ID, s.key)
	}
	s.keyNew = nil
}
func (s *Session) keyRevert() {
	s.keyNew = nil
}
func (s *Session) doNextKeySwap() bool {
	if !s.IsClient() || s.keyNew != nil {
		return false
	}
	if util.FastRandN(100) != 0 {
		return false
	}
	if cout.Enabled {
		s.log.Info("[%s] Generating new KeysSet!", s.ID)
	}
	var b data.Key
	rand.Read(b[:])
	s.keyNew = &b
	return true
}
func generateKeys(k *data.Key, d device.ID) {
	for i := 0; i < data.KeySize; i++ {
		// NOTE(dij): Since the pos index is added here we can't use the "subtle" package.
		(*k)[i] = ((*k)[i] ^ d[i]) + byte(i)
	}
}
func (s *Session) generateSessionKey(n *com.Packet) {
	rand.Read(s.key[:])
	n.Write(s.key[:])
	n.Flags |= com.FlagCrypt
	generateKeys(&s.key, s.ID)
}
