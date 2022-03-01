//go:build windows && crypt
// +build windows,crypt

package man

import "github.com/iDigitalFlame/xmt/util/crypt"

var (
	slot   = crypt.Get(15) // \\.\mailslot\
	prefix = crypt.Get(16) // Global\
)
