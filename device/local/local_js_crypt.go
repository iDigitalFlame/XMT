//go:build (wasm || js) && crypt
// +build wasm js
// +build crypt

package local

import "github.com/iDigitalFlame/xmt/util/crypt"

func sysID() []byte {
	return nil
}
func version() string {
	return crypt.Get(97) // JavaScript
}
func isElevated() bool {
	return false
}
