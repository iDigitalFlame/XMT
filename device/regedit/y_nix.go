//go:build !windows

// Copyright (C) 2020 - 2022 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package regedit

import "github.com/iDigitalFlame/xmt/device"

// MakeKey will attempt to create an empty key for the supplied path.
//
// The key path can either be a "reg" style path (ex: HKLM\System or
// HKCU\Software) or PowerShell style (ex: HKLM:\System or HKCU:\Software).
//
// Any errors creating the key will be returned.
//
// Returns device.ErrNoWindows on non-Windows devices.
func MakeKey(_ string) error {
	return device.ErrNoWindows
}

// Delete will attempt to delete the specified key or value name specified.
//
// The value will be probed for a type and if it is a key, it will delete the
// key ONLY if it is empty. (Use 'DeleteEx' for recursive deletion).
//
// The key path can either be a "reg" style path (ex: HKLM\System or
// HKCU\Software) or PowerShell style (ex: HKLM:\System or HKCU:\Software).
//
// Any errors deleting the key or value will be returned.
//
// Returns device.ErrNoWindows on non-Windows devices.
func Delete(_, _ string) error {
	return device.ErrNoWindows
}

// Dir returns a list of registry entries for the supplied key or an error if
// the path does not exist.
//
// The key path can either be a "reg" style path (ex: HKLM\System or
// HKCU\Software) or PowerShell style (ex: HKLM:\System or HKCU:\Software).
//
// Returns device.ErrNoWindows on non-Windows devices.
func Dir(_ string) ([]Entry, error) {
	return nil, device.ErrNoWindows
}

// Get returns a single registry entry for the supplied value name under the
// key path specified or an error if any of the paths do not exist.
//
// The key path can either be a "reg" style path (ex: HKLM\System or
// HKCU\Software) or PowerShell style (ex: HKLM:\System or HKCU:\Software).
//
// Returns device.ErrNoWindows on non-Windows devices.
func Get(_, _ string) (Entry, error) {
	return Entry{}, device.ErrNoWindows
}

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
func SetString(_, _, _ string) error {
	return device.ErrNoWindows
}

// DeleteKey will attempt to delete the specified subkey name specified.
//
// If force is specified, this will recursively delete a subkey and will delete
// non-empty subkeys. Otherwise, if force is false, non-empty subkeys will NOT
// be deleted.
//
// The key path can either be a "reg" style path (ex: HKLM\System or
// HKCU\Software) or PowerShell style (ex: HKLM:\System or HKCU:\Software).
//
// Any errors creating the key will be returned.
//
// Returns device.ErrNoWindows on non-Windows devices.
func DeleteKey(_ string, _ bool) error {
	return device.ErrNoWindows
}

// DeleteEx will attempt to delete the specified key or value name specified.
//
// The value will be probed for a type and if it is a key, it will delete the
// key ONLY if it is empty or force is true (which will recursively delete the
// key)
//
// The key path can either be a "reg" style path (ex: HKLM\System or
// HKCU\Software) or PowerShell style (ex: HKLM:\System or HKCU:\Software).
//
// Any errors deleting the key or value will be returned.
//
// Returns device.ErrNoWindows on non-Windows devices.
func DeleteEx(_, _ string, _ bool) error {
	return device.ErrNoWindows
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
func SetBytes(_, _ string, _ []byte) error {
	return device.ErrNoWindows
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
func SetDword(_, _ string, _ uint32) error {
	return device.ErrNoWindows
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
func SetQword(_, _ string, _ uint64) error {
	return device.ErrNoWindows
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
func SetExpandString(_, _, _ string) error {
	return device.ErrNoWindows
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
func SetStrings(_, _ string, _ []string) error {
	return device.ErrNoWindows
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
func Set(_, _ string, _ uint32, _ []byte) error {
	return device.ErrNoWindows
}

// SetFromString will attempt to set the data of the value in the supplied key
// path to the supplied value type indicated with the type flag and will interpret
// the supplied string using the type value to properly set the value data.
//
// - Dword or Qword values will attempt to parse the string as a base10 integer that may be negative.
// - String and ExpandString values will be interrupted as is.
// - Binary will attempt to decode the value using the StdEncoding base64 decoder.
// - StringLists will be split using the NEWLINE ("\n") character into an array to be used.
//
// This will create the key/value if it does not exist.
//
// The key path can either be a "reg" style path (ex: HKLM\System or
// HKCU\Software) or PowerShell style (ex: HKLM:\System or HKCU:\Software).
//
// Any errors writing the data will be returned.
//
// Returns device.ErrNoWindows on non-Windows devices.
func SetFromString(_, _ string, _ uint32, _ string) error {
	return device.ErrNoWindows
}
