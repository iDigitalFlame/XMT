//go:build plan9 && crypt

package local

import "github.com/iDigitalFlame/xmt/util/crypt"

func uname() string {
	return crypt.Get(81) // plan9
}
