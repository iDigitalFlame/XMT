//go:build windows && crypt

package local

import (
	"strconv"

	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/device/winapi/registry"
	"github.com/iDigitalFlame/xmt/util/crypt"
)

func sysID() []byte {
	if s, err := winapi.GetSystemSID(); err == nil {
		return []byte(s.String())
	}
	// 0x101 - KEY_WOW64_64KEY | KEY_QUERY_VALUE
	k, err := registry.Open(registry.KeyLocalMachine, crypt.Get(68), 0x101) // SOFTWARE\Microsoft\Cryptography
	if err != nil {
		return nil
	}
	v, _, err := k.String(crypt.Get(69)) // MachineGuid
	if k.Close(); err == nil {
		return []byte(v)
	}
	return nil
}
func version() string {
	// 0x101 - KEY_WOW64_64KEY | KEY_QUERY_VALUE
	k, err := registry.Open(registry.KeyLocalMachine, crypt.Get(70), 0x101) // SOFTWARE\Microsoft\Windows NT\CurrentVersion
	if err != nil {
		return crypt.Get(71) // Windows (?)
	}
	var (
		b, v    string
		n, _, _ = k.String(crypt.Get(72)) // ProductName
	)
	if s, _, err := k.String(crypt.Get(73)); err == nil { // CurrentBuild
		b = s
	} else if s, _, err := k.String(crypt.Get(74)); err == nil { // ReleaseId
		b = s
	}
	if i, _, err := k.Integer(crypt.Get(75)); err == nil { // CurrentMajorVersionNumber
		if x, _, err := k.Integer(crypt.Get(76)); err == nil { // CurrentMinorVersionNumber
			v = strconv.Itoa(int(i)) + "." + strconv.Itoa(int(x))
		} else {
			v = strconv.Itoa(int(i))
		}
	} else {
		v, _, _ = k.String(crypt.Get(77)) // CurrentVersion
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
