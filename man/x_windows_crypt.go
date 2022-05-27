//go:build windows && crypt

package man

import "github.com/iDigitalFlame/xmt/util/crypt"

var (
	slot   = crypt.Get(106) // \\.\mailslot\
	prefix = crypt.Get(107) // Global\
)
