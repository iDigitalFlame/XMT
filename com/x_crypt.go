//go:build crypt

package com

import "github.com/iDigitalFlame/xmt/util/crypt"

// Named Network Constants
var (
	NameIP   = crypt.Get(2)   // ip
	NameTCP  = crypt.Get(3)   // tcp
	NameUDP  = crypt.Get(4)   // udp
	NameUnix = crypt.Get(5)   // unix
	NamePipe = crypt.Get(6)   // pipe
	NameHTTP = crypt.Get(116) // http
)

func (udpErr) Error() string {
	return crypt.Get(43) // deadline exceeded
}
