//go:build windows && !crypt
// +build windows,!crypt

package local

import (
	"os"
	"strconv"

	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/device/winapi/registry"
)

func sysID() []byte {
	if s, err := winapi.GetSystemSID(); err == nil {
		return []byte(s.String())
	}
	k, err := registry.OpenKey(registry.KeyLocalMachine, `SOFTWARE\Microsoft\Cryptography`, 0x101)
	if err != nil {
		return nil
	}
	v, _, err := k.GetStringValue("MachineGuid")
	if k.Close(); err == nil {
		return []byte(v)
	}
	return nil
}
func version() string {
	k, err := registry.OpenKey(registry.KeyLocalMachine, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, 0x101)
	if err != nil {
		return "Windows (?)"
	}
	var (
		b, v    string
		n, _, _ = k.GetStringValue("ProductName")
	)
	if s, _, err := k.GetStringValue("CurrentBuild"); err == nil {
		b = s
	} else if s, _, err := k.GetStringValue("ReleaseId"); err == nil {
		b = s
	}
	if i, _, err := k.GetIntegerValue("CurrentMajorVersionNumber"); err == nil {
		if x, _, err := k.GetIntegerValue("CurrentMinorVersionNumber"); err == nil {
			v = strconv.Itoa(int(i)) + "." + strconv.Itoa(int(x))
		} else {
			v = strconv.Itoa(int(i))
		}
	} else {
		v, _, _ = k.GetStringValue("CurrentVersion")
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
func isElevated() bool {
	if p, err := os.OpenFile(`\\.\PHYSICALDRIVE0`, 0, 0); err == nil {
		p.Close()
		return true
	}
	return false
}
