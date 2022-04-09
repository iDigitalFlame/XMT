//go:build windows && crypt

package local

import (
	"strconv"
	"syscall"
	"unsafe"

	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/device/winapi/registry"
	"github.com/iDigitalFlame/xmt/util/crypt"
)

func sysID() []byte {
	if s, err := winapi.GetSystemSID(); err == nil {
		return []byte(s.String())
	}
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
func isElevated() uint8 {
	var e uint8
	if checkElevatedToken() {
		e = 1
	}
	var (
		d *uint16
		s uint32
	)
	if err := syscall.NetGetJoinInformation(nil, &d, &s); err != nil {
		return e
	}
	if syscall.NetApiBufferFree((*byte)(unsafe.Pointer(d))); s == 3 {
		e += 128
	}
	return e
}
func checkElevatedToken() bool {
	var (
		t   uintptr
		err = winapi.OpenProcessToken(winapi.CurrentProcess, 0x8, &t)
	)
	if err != nil {
		return false
	}
	var n uint32
	if winapi.GetTokenInformation(t, 0x19, nil, 0, &n); n < 4 {
		winapi.CloseHandle(t)
		return false
	}
	b := make([]byte, n)
	_ = b[n-1]
	if err = winapi.GetTokenInformation(t, 0x19, &b[0], n, &n); err != nil {
		winapi.CloseHandle(t)
		return false
	}
	var (
		p = uint32(b[n-4]) | uint32(b[n-3])<<8 | uint32(b[n-2])<<16 | uint32(b[n-1])<<24
		r = p >= 0x3000
	)
	if !r {
		r = winapi.IsTokenElevated(t)
	}
	winapi.CloseHandle(t)
	return r
}
