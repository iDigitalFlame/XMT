//go:build nokeyset

package c2

import "github.com/iDigitalFlame/xmt/com"

func (s *Session) keyCheck() {}
func (s *Session) doNextKeySwap() bool {
	return false
}
func (s *Session) generateSessionKey(n *com.Packet) {
	n.Write(s.key[:])
}
func (*Session) sessionKeyInit(_ *Listener, _ *com.Packet)   {}
func (*Session) sessionKeyUpdate(_ *Listener, _ *com.Packet) {}
