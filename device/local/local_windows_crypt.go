//go:build windows && crypt
// +build windows,crypt

package local

import (
	"os"
	"strconv"

	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/device/winapi/registry"
	"github.com/iDigitalFlame/xmt/util/crypt"
)

func sysID() []byte {
	if s, err := winapi.GetSystemSID(); err == nil {
		return []byte(s.String())
	}
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, crypt.Get(68), 0x101) // SOFTWARE\Microsoft\Cryptography
	if err != nil {
		return nil
	}
	v, _, err := k.GetStringValue(crypt.Get(69)) // MachineGuid
	if k.Close(); err == nil {
		return []byte(v)
	}
	return nil
}
func version() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, crypt.Get(70), 0x101) // SOFTWARE\Microsoft\Windows NT\CurrentVersion
	if err != nil {
		return crypt.Get(71) // Windows (?)
	}
	var (
		b, v    string
		n, _, _ = k.GetStringValue(crypt.Get(72)) // ProductName
	)
	if s, _, err := k.GetStringValue(crypt.Get(73)); err == nil { // CurrentBuild
		b = s
	} else if s, _, err := k.GetStringValue(crypt.Get(74)); err == nil { // ReleaseId
		b = s
	}
	if i, _, err := k.GetIntegerValue(crypt.Get(75)); err == nil { // CurrentMajorVersionNumber
		if x, _, err := k.GetIntegerValue(crypt.Get(76)); err == nil { // CurrentMinorVersionNumber
			v = strconv.Itoa(int(i)) + "." + strconv.Itoa(int(x))
		} else {
			v = strconv.Itoa(int(i))
		}
	} else {
		v, _, _ = k.GetStringValue(crypt.Get(77)) // CurrentVersion
	}
	switch k.Close(); {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return crypt.Get(20) // Windows
	case len(n) == 0 && len(b) > 0 && len(v) > 0:
		return crypt.Get(20) + " (" + v + ", " + b + ")" // Windows
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return crypt.Get(20) + " (" + v + ")" // Windows
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return crypt.Get(20) + " (" + b + ")" // Windows
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return n + " (" + v + ", " + b + ")"
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return n + " (" + v + ")"
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return n + " (" + b + ")"
	}
	return crypt.Get(20) // Windows
}
func isElevated() bool {
	if p, err := os.OpenFile(crypt.Get(78), 0, 0); err == nil { // \\.\PHYSICALDRIVE0
		p.Close()
		return true
	}
	return false
}
