//go:build windows && crypt

package cmd

import "github.com/iDigitalFlame/xmt/util/crypt"

var sysRoot = crypt.Get(66) // SYSTEMROOT
