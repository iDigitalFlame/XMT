//go:build !crypt
// +build !crypt

package text

import "regexp"

const alpha = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var regxBuild = regexp.MustCompile(`(\%(\d+f?)?[dhcsuln])`)
