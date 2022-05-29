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
	k, err := registry.Open(registry.KeyLocalMachine, crypt.Get(61), 0x101) // SOFTWARE\Microsoft\Cryptography
	if err != nil {
		return nil
	}
	v, _, err := k.String(crypt.Get(62)) // MachineGuid
	if k.Close(); err == nil {
		return []byte(v)
	}
	return nil
}
func version() string {
	// 0x101 - KEY_WOW64_64KEY | KEY_QUERY_VALUE
	k, err := registry.Open(registry.KeyLocalMachine, crypt.Get(63), 0x101) // SOFTWARE\Microsoft\Windows NT\CurrentVersion
	if err != nil {
		return crypt.Get(64) // Windows
	}
	var (
		b, v    string
		n, _, _ = k.String(crypt.Get(65)) // ProductName
	)
	if s, _, err := k.String(crypt.Get(66)); err == nil { // CurrentBuild
		b = s
	} else if s, _, err := k.String(crypt.Get(67)); err == nil { // ReleaseId
		b = s
	}
	if i, _, err := k.Integer(crypt.Get(68)); err == nil { // CurrentMajorVersionNumber
		if x, _, err := k.Integer(crypt.Get(69)); err == nil { // CurrentMinorVersionNumber
			v = strconv.FormatUint(i, 10) + "." + strconv.FormatUint(x, 10)
		} else {
			v = strconv.FormatUint(i, 10)
		}
	} else {
		v, _, _ = k.String(crypt.Get(70)) // CurrentVersion
	}
	switch k.Close(); {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return crypt.Get(64) // Windows
	case len(n) == 0 && len(b) > 0 && len(v) > 0:
		return crypt.Get(64) + " (" + v + ", " + b + ")" // Windows
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return crypt.Get(64) + " (" + v + ")" // Windows
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return crypt.Get(64) + " (" + b + ")" // Windows
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return n + " (" + v + ", " + b + ")"
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return n + " (" + v + ")"
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return n + " (" + b + ")"
	}
	return crypt.Get(64) // Windows
}
