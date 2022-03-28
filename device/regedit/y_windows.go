//go:build windows

package regedit

import (
	"strings"
	"syscall"

	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/device/winapi/registry"
)

// Dir returns an list of registry entries for the supplied key or an error if
// the path does not exist.
//
// The key path can either be a "reg" style path (ex: HKLM\System or
// HKCU\Software) or PowerShell style (ex: HKLM:\System or HKCU:\Software).
//
// Returns device.ErrNoWindows on non-Windows devices.
func Dir(key string) ([]Entry, error) {
	k, err := read(key, false)
	if err != nil {
		return nil, err
	}
	s, err := k.SubKeys()
	if err != nil {
		k.Close()
		return nil, err
	}
	v, err := k.Values()
	if err != nil {
		k.Close()
		return nil, err
	}
	n := len(v) + len(s)
	if n == 0 {
		k.Close()
		return nil, nil
	}
	var (
		r = make([]Entry, n+1)
		c int
	)
	for i := range s {
		r[c].Name = s[i]
		c++
	}
	r[c].Name = "(Default)"
	r[c].Data, r[c].Type, _ = readFullValueData(k, "")
	// NOTE(dij):         ^ Ignoring since not-set (Default) will return an
	//                    error even though regedit displays it.
	c++
	for i := range v {
		r[c].Name = v[i]
		r[c].Data, r[c].Type, err = readFullValueData(k, v[i])
		if c++; err != nil {
			break
		}
	}
	k.Close()
	return r, err
}
func increaseSlash(i int, s string) int {
	if len(s) <= i {
		return i
	}
	if s[i] == '\\' {
		return i + 1
	}
	return i
}

// Get returns a single registry entry for the supplied value name under the
// key path specified or an error if any of the paths do not exist.
//
// The key path can either be a "reg" style path (ex: HKLM\System or
// HKCU\Software) or PowerShell style (ex: HKLM:\System or HKCU:\Software).
//
// Returns device.ErrNoWindows on non-Windows devices.
func Get(key, value string) (Entry, error) {
	var (
		k, err = read(key, false)
		e      Entry
	)
	if err != nil {
		return e, err
	}
	e.Name = value
	e.Data, e.Type, err = readFullValueData(k, value)
	k.Close()
	return e, err
}
func read(k string, w bool) (registry.Key, error) {
	h, d, err := translateRootKey(k)
	if err != nil {
		return 0, err
	}
	if d >= len(k) || h == 0 {
		return 0, registry.ErrNotExist
	}
	if w {
		x, _, err := registry.Create(h, k[d:], 0x2001F)
		return x, err
	}
	return registry.Open(h, k[d:], 0x20019)
}
func translateRootKey(v string) (registry.Key, int, error) {
	if len(v) < 4 || (v[0] != 'H' && v[0] != 'h') {
		return 0, 0, registry.ErrNotExist
	}
	i := strings.IndexByte(v, ':')
	if i == -1 {
		if i = strings.IndexByte(v, '\\'); i == -1 {
			return 0, 0, registry.ErrNotExist
		}
	}
	if len(v) > 6 && v[4] == '_' {
		if i < 5 {
			return 0, 0, registry.ErrNotExist
		}
		switch v[i-1] {
		case 'E', 'e':
			return registry.KeyLocalMachine, increaseSlash(i+1, v), nil
		case 'R', 'r':
			return registry.KeyCurrentUser, increaseSlash(i+1, v), nil
		case 'S', 's':
			return registry.KeyUsers, increaseSlash(i+1, v), nil
		case 'G', 'g':
			return registry.KeyCurrentConfig, increaseSlash(i+1, v), nil
		case 'A', 'a':
			return registry.KeyPerformanceData, increaseSlash(i+1, v), nil
		case 'T', 't':
			return registry.KeyClassesRoot, increaseSlash(i+1, v), nil
		}
		return 0, 0, registry.ErrNotExist
	}
	if i == 3 && (v[2] == 'U' || v[2] == 'u') {
		return registry.KeyUsers, increaseSlash(4, v), nil
	}
	if i < 4 {
		return 0, 0, registry.ErrNotExist
	}
	switch v[i-1] {
	case 'M', 'm':
		return registry.KeyLocalMachine, increaseSlash(i+1, v), nil
	case 'U', 'u':
		return registry.KeyCurrentUser, increaseSlash(i+1, v), nil
	case 'C', 'c':
		return registry.KeyCurrentConfig, increaseSlash(i+1, v), nil
	case 'D', 'd':
		return registry.KeyPerformanceData, increaseSlash(i+1, v), nil
	case 'R', 'r':
		return registry.KeyClassesRoot, increaseSlash(i+1, v), nil
	}
	return 0, 0, registry.ErrNotExist
}
func readFullValueData(k registry.Key, n string) ([]byte, uint32, error) {
	v, err := winapi.UTF16PtrFromString(n)
	if err != nil {
		return nil, 0, err
	}
	var t, s uint32
	if err = syscall.RegQueryValueEx(syscall.Handle(k), v, nil, &t, nil, &s); err != nil {
		return nil, 0, err
	}
	b := make([]byte, s)
	if err = syscall.RegQueryValueEx(syscall.Handle(k), v, nil, &t, &b[0], &s); err != nil {
		return nil, t, err
	}
	return b, t, err
}
