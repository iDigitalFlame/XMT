//go:build regexp && crypt

package text

import (
	"regexp"

	"github.com/iDigitalFlame/xmt/util/crypt"
)

var regxBuild = regexp.MustCompile(crypt.Get(1)) // (\%(\d+f?)?[dhcsuln])
