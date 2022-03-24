//go:build regexp && !crypt

package text

import "regexp"

var regxBuild = regexp.MustCompile(`(\%(\d+f?)?[dhcsuln])`)
