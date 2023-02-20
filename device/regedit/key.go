//go:build windows
// +build windows

// Copyright (C) 2020 - 2023 iDigitalFlame
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

import "github.com/iDigitalFlame/xmt/device/winapi/registry"

// MakeKey will attempt to create an empty key for the supplied path.
//
// The key path can either be a "reg" style path (ex: HKLM\System or
// HKCU\Software) or PowerShell style (ex: HKLM:\System or HKCU:\Software).
//
// Any errors creating the key will be returned.
//
// Returns device.ErrNoWindows on non-Windows devices.
func MakeKey(key string) error {
	k, err := read(key, true)
	if err != nil {
		return err
	}
	k.Close()
	return nil
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
func Delete(key, value string) error {
	return DeleteEx(key, value, false)
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
func DeleteKey(key string, force bool) error {
	h, d, err := translateRootKey(key)
	if err != nil {
		return err
	}
	if d >= len(key) || h == 0 {
		return registry.ErrNotExist
	}
	if force {
		return registry.DeleteTree(h, key[d:])
	}
	return registry.DeleteEx(h, key[d:], 0)
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
func DeleteEx(key, value string, force bool) error {
	k, err := read(key, true)
	if err != nil {
		return err
	}
	_, t, err := k.Value(value, nil)
	if err == registry.ErrNotExist {
		// 0x2001F - KEY_READ | KEY_WRITE
		s, err1 := k.Open(value, 0x2001F)
		if err1 != nil {
			k.Close()
			return err
		}
		s.Close()
		if force {
			err = registry.DeleteTree(k, value)
		} else {
			err = k.DeleteKey(value)
		}
		k.Close()
		return err
	}
	if err == nil && t > 0 {
		err = k.DeleteValue(value)
	}
	k.Close()
	return err
}
