//go:build !nokeyset

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
		s.log.Info("[%s] Regenerated KeyCrypt key set!", s.ID)
	}
	copy(s.key[:], (*s.keyNew)[:])
	if generateKeys(&s.key, s.ID); bugtrack.Enabled {
		bugtrack.Track("c2.Listener.talk(): %s KeyCrypt details [%v].", s.ID, s.key)
	}
	s.keyNew = nil
}
func (s *Session) keyRevert() {
	s.keyNew = nil
}
func (s *Session) doNextKeySwap() bool {
	if s.s != nil || s.keyNew != nil {
		return false
	}
	if util.FastRandN(125) != 0 {
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
		(*k)[i] = ((*k)[i] ^ d[i]) + byte(i)
	}
}
func (s *Session) generateSessionKey(n *com.Packet) {
	rand.Read(s.key[:])
	n.Write(s.key[:])
	n.Flags |= com.FlagCrypt
	generateKeys(&s.key, s.ID)
}
