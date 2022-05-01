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
