//go:build windows && !crypt

package local

import (
	"strconv"

	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/device/winapi/registry"
)

func sysID() []byte {
	if s, err := winapi.GetSystemSID(); err == nil {
		return []byte(s.String())
	}
	// 0x101 - KEY_WOW64_64KEY | KEY_QUERY_VALUE
	k, err := registry.Open(registry.KeyLocalMachine, `SOFTWARE\Microsoft\Cryptography`, 0x101)
	if err != nil {
		return nil
	}
	v, _, err := k.String("MachineGuid")
	if k.Close(); err == nil {
		return []byte(v)
	}
	return nil
}
func version() string {
	// 0x101 - KEY_WOW64_64KEY | KEY_QUERY_VALUE
	k, err := registry.Open(registry.KeyLocalMachine, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, 0x101)
	if err != nil {
		return "Windows (?)"
	}
	var (
		b, v    string
		n, _, _ = k.String("ProductName")
	)
	if s, _, err := k.String("CurrentBuild"); err == nil {
		b = s
	} else if s, _, err := k.String("ReleaseId"); err == nil {
		b = s
	}
	if i, _, err := k.Integer("CurrentMajorVersionNumber"); err == nil {
		if x, _, err := k.Integer("CurrentMinorVersionNumber"); err == nil {
			v = strconv.FormatUint(i, 10) + "." + strconv.FormatUint(x, 10)
		} else {
			v = strconv.FormatUint(i, 10)
		}
	} else {
		v, _, _ = k.String("CurrentVersion")
	}
	switch k.Close(); {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return "Windows"
	case len(n) == 0 && len(b) > 0 && len(v) > 0:
		return "Windows (" + v + ", " + b + ")"
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return "Windows (" + v + ")"
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return "Windows (" + b + ")"
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return n + " (" + v + ", " + b + ")"
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return n + " (" + v + ")"
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return n + " (" + b + ")"
	}
	return "Windows"
}
