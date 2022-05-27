//go:build crypt

package com

import "github.com/iDigitalFlame/xmt/util/crypt"

// Named Network Constants
var (
	NameIP   = crypt.Get(26) // ip
	NameTCP  = crypt.Get(27) // tcp
	NameUDP  = crypt.Get(28) // udp
	NameUnix = crypt.Get(29) // unix
	NamePipe = crypt.Get(30) // pipe
	NameHTTP = crypt.Get(31) // http
)
