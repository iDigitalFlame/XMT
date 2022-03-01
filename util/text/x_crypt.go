//go:build crypt
// +build crypt

package text

import (
	"regexp"

	"github.com/iDigitalFlame/xmt/util/crypt"
)

var (
	alpha = crypt.Get(0) // abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789

	regxBuild = regexp.MustCompile(crypt.Get(1)) // (\%(\d+f?)?[dhcsuln])
)
