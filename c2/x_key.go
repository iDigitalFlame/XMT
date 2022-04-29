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
func (s *Session) doNextKeySwap() bool {
	if s.s != nil {
		return false
	}
	if util.FastRandN(5) != 0 {
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
func (s *Session) sessionKeyInit(l *Listener, n *com.Packet) {
	if v, err := n.Read(s.key[:]); v != data.KeySize || err != nil {
		if cout.Enabled {
			l.log.Warning("[%s:%s/Crypt] Error (Re)Generating key set!", l.name, s.ID)
		}
		return
	}
	if generateKeys(&s.key, s.ID); cout.Enabled {
		l.log.Info("[%s:%s/Crypt] (Re)Generated key set!", l.name, s.ID)
	}
	if bugtrack.Enabled {
		bugtrack.Track("c2.Session.sessionKeyInit(): %s Key details [%v].", s.ID, s.key)
	}
}
func (s *Session) sessionKeyUpdate(l *Listener, n *com.Packet) {
	if n.Crypt(&s.key); n.Flags&com.FlagCrypt == 0 || n.Empty() {
		return
	}
	s.sessionKeyInit(l, n)
}
