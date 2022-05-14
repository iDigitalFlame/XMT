//go:build crypt

package device

import "github.com/iDigitalFlame/xmt/util/crypt"

var emptyIP = crypt.Get(18) // "0.0.0.0"
