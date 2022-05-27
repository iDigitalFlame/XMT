//go:build js && crypt

package local

import "github.com/iDigitalFlame/xmt/util/crypt"

func sysID() []byte {
	return nil
}
func version() string {
	return crypt.Get(91) // JavaScript
}
func isElevated() uint8 {
	return 0
}
