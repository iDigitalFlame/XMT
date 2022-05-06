//go:build !implant && nokeyset

package c2

import "github.com/iDigitalFlame/xmt/com"

func (*Session) sessionKeyInit(_ string, _ *com.Packet)           {}
func (*Session) sessionKeyUpdate(_ string, _ *com.Packet, _ bool) {}
