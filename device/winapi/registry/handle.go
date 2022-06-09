//go:build windows

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

package registry

import (
	"syscall"

	"github.com/iDigitalFlame/xmt/device/winapi"
)

const (
	// WoW64to32 is a bitmask used in any registry access flags that indicates
	// to the kernel that you want to use the 32bit registry hives instead of
	// the 64bit ones. Has no effect on 32bit systems.
	WoW64to32 uint32 = 0x200
	// WoW64to64 is a bitmask used in any registry access flags that indicates
	// to the kernel that you want to use the 64bit registry hives instead of
	// the 32bit ones. Has no effect on 32bit systems.
	//
	// NOTE(dij): I think this is the /default/ for 64bit systems tho?!
	//            SO you might be able to safely ignore this flag if necessary.
	WoW64to64 uint32 = 0x100
)

// Delete deletes the subkey path of the key and its values.
//
// This function fails with "invalid argument" if the key has subkeys or
// values. Use 'DeleteTreeKey' instead to delete the full non-empty key.
func Delete(k Key, s string) error {
	return DeleteEx(k, s, 0)
}

// DeleteTree deletes the subkey path of the key and its values recursively.
func DeleteTree(k Key, s string) error {
	return winapi.RegDeleteTree(uintptr(k), s)
}

// Open opens a new key with path name relative to the key.
//
// It accepts any open root key, including CurrentUser for example, and returns
// the new key and an any errors that may occur during opening.
//
// The access parameter specifies desired access rights to the key to be opened.
func Open(k Key, s string, a uint32) (Key, error) {
	v, err := winapi.UTF16PtrFromString(s)
	if err != nil {
		return 0, err
	}
	var h syscall.Handle
	if err = syscall.RegOpenKeyEx(syscall.Handle(k), v, 0, a, &h); err != nil {
		return 0, err
	}
	return Key(h), nil
}

// DeleteEx deletes the subkey path of the key and its values.
//
// This function fails with "invalid argument" if the key has subkeys or
// values. Use 'DeleteTreeKey' instead to delete the full non-empty key.
//
// This function allows for specifying which WOW endpoint to delete from
// leave the flags value as zero to use the current WOW/non-WOW registry point.
//
// NOTE(dij): WOW64_32 is 0x200 and WOW64_64 is 0x100
func DeleteEx(k Key, s string, flags uint32) error {
	return winapi.RegDeleteKeyEx(uintptr(k), s, flags)
}

// Create creates a key named path under the open key.
//
// CreateKey returns the new key and a boolean flag that reports whether the key
// already existed.
//
// The access parameter specifies the access rights for the key to be created.
func Create(k Key, s string, a uint32) (Key, bool, error) {
	var (
		h   uintptr
		d   uint32
		err = winapi.RegCreateKeyEx(uintptr(k), s, "", 0, a, nil, &h, &d)
	)
	if err != nil {
		return 0, false, err
	}
	return Key(h), d == 2, nil
}
