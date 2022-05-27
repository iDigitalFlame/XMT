//go:build crypt

package device

import "github.com/iDigitalFlame/xmt/util/crypt"

var emptyIP = crypt.Get(60) // 0.0.0.0
