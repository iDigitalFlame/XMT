//go:build windows

package regedit

import (
	"syscall"
	"unsafe"

	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/device/winapi/registry"
	"github.com/iDigitalFlame/xmt/util"
)

// SetString will attempt to set the data of the value in the supplied key path
// to the supplied string as a REG_SZ value type.
//
// This will create the key/value if it does not exist.
//
// The key path can either be a "reg" style path (ex: HKLM\System or
// HKCU\Software) or PowerShell style (ex: HKLM:\System or HKCU:\Software).
//
// Any errors writing the data will be returned.
//
// Returns device.ErrNoWindows on non-Windows devices.
func SetString(key, value, v string) error {
	return setString(key, value, registry.TypeString, v)
}

// SetBytes will attempt to set the data of the value in the supplied key path
// to the supplied byte array as a REG_BINARY value type.
//
// This will create the key/value if it does not exist.
//
// The key path can either be a "reg" style path (ex: HKLM\System or
// HKCU\Software) or PowerShell style (ex: HKLM:\System or HKCU:\Software).
//
// Any errors writing the data will be returned.
//
// Returns device.ErrNoWindows on non-Windows devices.
func SetBytes(key, value string, b []byte) error {
	return Set(key, value, registry.TypeBinary, b)
}

// SetDword will attempt to set the data of the value in the supplied key path
// to the supplied uint32 integer as a REG_DWORD value type.
//
// This will create the key/value if it does not exist.
//
// The key path can either be a "reg" style path (ex: HKLM\System or
// HKCU\Software) or PowerShell style (ex: HKLM:\System or HKCU:\Software).
//
// Any errors writing the data will be returned.
//
// Returns device.ErrNoWindows on non-Windows devices.
func SetDword(key, value string, v uint32) error {
	return Set(key, value, registry.TypeDword, (*[4]byte)(unsafe.Pointer(&v))[:])
}

// SetQword will attempt to set the data of the value in the supplied key path
// to the supplied uint64 integer as a REG_QWORD value type.
//
// This will create the key/value if it does not exist.
//
// The key path can either be a "reg" style path (ex: HKLM\System or
// HKCU\Software) or PowerShell style (ex: HKLM:\System or HKCU:\Software).
//
// Any errors writing the data will be returned.
//
// Returns device.ErrNoWindows on non-Windows devices.
func SetQword(key, value string, v uint64) error {
	return Set(key, value, registry.TypeQword, (*[8]byte)(unsafe.Pointer(&v))[:])
}

// SetExpandString will attempt to set the data of the value in the supplied key
// path to the supplied string as a REG_EXPAND_SZ value type.
//
// This will create the key/value if it does not exist.
//
// The key path can either be a "reg" style path (ex: HKLM\System or
// HKCU\Software) or PowerShell style (ex: HKLM:\System or HKCU:\Software).
//
// Any errors writing the data will be returned.
//
// Returns device.ErrNoWindows on non-Windows devices.
func SetExpandString(key, value, v string) error {
	return setString(key, value, registry.TypeExpandString, v)
}

// SetStrings will attempt to set the data of the value in the supplied key path
// to the supplied string list as a REG_MULTI_SZ value type.
//
// This will create the key/value if it does not exist.
//
// The key path can either be a "reg" style path (ex: HKLM\System or
// HKCU\Software) or PowerShell style (ex: HKLM:\System or HKCU:\Software).
//
// Any errors writing the data will be returned.
//
// Returns device.ErrNoWindows on non-Windows devices.
func SetStrings(key, value string, v []string) error {
	k, err := read(key, true)
	if err != nil {
		return err
	}
	var b util.Builder
	for i := range v {
		for x := range v[i] {
			if v[i][x] == 0 {
				k.Close()
				return syscall.EINVAL
			}
		}
		b.WriteString(v[i] + "\x00")
	}
	var (
		r = winapi.UTF16EncodeStd([]rune(b.Output() + "\x00"))
		o = (*[1 << 29]byte)(unsafe.Pointer(&r[0]))[: len(r)*2 : len(r)*2]
	)
	err = winapi.RegSetValueEx(uintptr(k), value, registry.TypeStringList, &o[0], uint32(len(o)))
	k.Close()
	return err
}

// Set will attempt to set the data of the value in the supplied key path to the
// supplied value type indicated with the type flag and will use the provided
// raw bytes for the value data.
//
// This function is a "lower-level" function but allows more control of HOW the
// data is used.
//
// This will create the key/value if it does not exist.
//
// The key path can either be a "reg" style path (ex: HKLM\System or
// HKCU\Software) or PowerShell style (ex: HKLM:\System or HKCU:\Software).
//
// Any errors writing the data will be returned.
//
// Returns device.ErrNoWindows on non-Windows devices.
func Set(key, value string, t uint32, b []byte) error {
	k, err := read(key, true)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		err = winapi.RegSetValueEx(uintptr(k), value, t, nil, 0)
	} else {
		err = winapi.RegSetValueEx(uintptr(k), value, t, &b[0], uint32(len(b)))
	}
	k.Close()
	return err
}
func setString(key, value string, t uint32, v string) error {
	k, err := read(key, true)
	if err != nil {
		return err
	}
	p, err := winapi.UTF16FromString(v)
	if err != nil {
		k.Close()
		return err
	}
	b := (*[1 << 29]byte)(unsafe.Pointer(&p[0]))[: len(p)*2 : len(p)*2]
	err = winapi.RegSetValueEx(uintptr(k), value, t, &b[0], uint32(len(b)))
	k.Close()
	return err
}
