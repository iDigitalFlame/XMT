//go:build windows && !crypt
// +build windows,!crypt

package evade

import "github.com/iDigitalFlame/xmt/device/winapi"

const sect = ".text"

func fullPath(n string) string {
	if !isBaseName(n) {
		return n
	}
	d, err := winapi.GetSystemDirectory()
	if err != nil {
		d = `C:\Windows\System32`
	}
	return d + "\\" + n
}
