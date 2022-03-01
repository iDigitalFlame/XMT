//go:build crypt

package com

import "github.com/iDigitalFlame/xmt/util/crypt"

func (udpErr) Error() string {
	return crypt.Get(43) // deadline exceeded
}
