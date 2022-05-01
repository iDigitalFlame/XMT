//go:build !implant && !nokeyset

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
		s.log.Info("[%s:%s/Crypt] (Re)Generated key set!", l, s.ID)
	}
	if bugtrack.Enabled {
		bugtrack.Track("c2.Session.sessionKeyInit(): %s Key details [%v].", s.ID, s.key)
	}
}
func (s *Session) sessionKeyUpdate(l string, n *com.Packet) {
	if n.Crypt(&s.key); n.Flags&com.FlagCrypt == 0 || n.Empty() {
		return
	}
	s.sessionKeyInit(l, n)
}
