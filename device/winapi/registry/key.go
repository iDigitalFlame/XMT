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

package registry

import (
	"io"
	"runtime"
	"syscall"
	"time"

	"github.com/iDigitalFlame/xmt/device/winapi"
)

// Windows defines some predefined root keys that are always open.
//
// An application can use these keys as entry points to the registry. Normally
// these keys are used in OpenKey to open new keys, but they can also be used
// anywhere a Key is required.
const (
	KeyClassesRoot     = Key(syscall.HKEY_CLASSES_ROOT)
	KeyCurrentUser     = Key(syscall.HKEY_CURRENT_USER)
	KeyLocalMachine    = Key(syscall.HKEY_LOCAL_MACHINE)
	KeyUsers           = Key(syscall.HKEY_USERS)
	KeyCurrentConfig   = Key(syscall.HKEY_CURRENT_CONFIG)
	KeyPerformanceData = Key(syscall.HKEY_PERFORMANCE_DATA)
)

const errNoMoreItems syscall.Errno = 259

// Key is a handle to an open Windows registry key.
//
// Keys can be obtained by calling OpenKey.
type Key uintptr

// KeyInfo describes the statistics of a key.
//
// It is returned by a call to Stat.
type KeyInfo struct {
	SubKeyCount     uint32
	MaxSubKeyLen    uint32
	ValueCount      uint32
	MaxValueNameLen uint32
	MaxValueLen     uint32
	lastWriteTime   syscall.Filetime
}

// Close closes the open key.
func (k Key) Close() error {
	return syscall.RegCloseKey(syscall.Handle(k))
}

// Flush calls 'winapi.RegFlushKey' on this key to sync the data with the
// underlying filesystem.
func (k Key) Flush() error {
	return winapi.RegFlushKey(uintptr(k))
}

// ModTime returns the key's last write time.
func (i *KeyInfo) ModTime() time.Time {
	return time.Unix(0, i.lastWriteTime.Nanoseconds())
}

// Stat retrieves information about the open key.
func (k Key) Stat() (*KeyInfo, error) {
	var (
		i   KeyInfo
		err = syscall.RegQueryInfoKey(
			syscall.Handle(k), nil, nil, nil,
			&i.SubKeyCount, &i.MaxSubKeyLen, nil, &i.ValueCount,
			&i.MaxValueNameLen, &i.MaxValueLen, nil, &i.lastWriteTime,
		)
	)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

// Values returns the value names in the key. This function is similar to
// 'ValueNames' but returns ALL names instead.
func (k Key) Values() ([]string, error) {
	return k.ValueNames(-1)
}

// DeleteKey deletes the subkey path of the key and its values.
//
// Convince function added directly to the Key alias.
//
// This function fails with "invalid argument" if the key has subkeys or
// values. Use 'DeleteTreeKey' instead to delete the full non-empty key.
func (k Key) DeleteKey(s string) error {
	return DeleteEx(k, s, 0)
}

// DeleteValue removes a named value from the key.
func (k Key) DeleteValue(n string) error {
	return winapi.RegDeleteValue(uintptr(k), n)
}

// SubKeys returns the names of subkeys of key. This function is similar to
// 'SubKeyNames' but returns ALL names instead.
func (k Key) SubKeys() ([]string, error) {
	return k.SubKeyNames(-1)
}

// ValueNames returns the value names in the key.
//
// The parameter controls the number of returned names, analogous to the way
// 'os.File.Readdirnames' works.
func (k Key) ValueNames(n int) ([]string, error) {
	x, err := k.Stat()
	if err != nil {
		return nil, err
	}
	if x.ValueCount == 0 || x.MaxValueNameLen == 0 {
		return nil, nil
	}
	var (
		o = make([]string, 0, x.ValueCount)
		b = make([]uint16, x.MaxValueNameLen+1)
	)
loop:
	for i := uint32(0); i < x.ValueCount; i++ {
		if n > 0 {
			if len(o) == n {
				return o, nil
			}
		}
		l := uint32(len(b))
		for {
			if err = winapi.RegEnumValue(uintptr(k), i, &b[0], &l, nil, nil, nil); err == nil {
				break
			}
			if err == syscall.ERROR_MORE_DATA {
				l = uint32(2 * len(b))
				b = make([]uint16, l)
				continue
			}
			if err == errNoMoreItems {
				break loop
			}
			return o, err
		}
		o = append(o, winapi.UTF16ToString(b[:l]))
	}
	if n > len(o) {
		return o, io.EOF
	}
	return o, nil
}

// SubKeyNames returns the names of subkeys of key.
//
// The parameter controls the number of returned names, analogous to the way
// 'os.File.Readdirnames' works.
func (k Key) SubKeyNames(n int) ([]string, error) {
	runtime.LockOSThread()
	var (
		o   = make([]string, 0)
		b   = make([]uint16, 256)
		err error
	)
loop:
	for i := uint32(0); ; i++ {
		if n > 0 {
			if len(o) == n {
				runtime.UnlockOSThread()
				return o, nil
			}
		}
		l := uint32(len(b))
		for {
			if err = syscall.RegEnumKeyEx(syscall.Handle(k), i, &b[0], &l, nil, nil, nil, nil); err == nil {
				break
			}
			if err == syscall.ERROR_MORE_DATA {
				l = uint32(2 * len(b))
				b = make([]uint16, l)
				continue
			}
			break loop
		}
		o = append(o, winapi.UTF16ToString(b[:l]))
	}
	if runtime.UnlockOSThread(); err == errNoMoreItems {
		err = nil
	}
	if n > len(o) {
		return o, io.EOF
	}
	return o, err
}

// Open opens a new key with path name relative to the key.
//
// Convince function added directly to the Key alias.
//
// It accepts any open root key, including CurrentUser for example, and returns
// the new key and an any errors that may occur during opening.
//
// The access parameter specifies desired access rights to the key to be opened.
func (k Key) Open(s string, a uint32) (Key, error) {
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

// DeleteKeyEx deletes the subkey path of the key and its values.
//
// Convince function added directly to the Key alias.
//
// This function fails with "invalid argument" if the key has subkeys or
// values. Use 'DeleteTreeKey' instead to delete the full non-empty key.
//
// This function allows for specifying which WOW endpoint to delete from
// leave the flags value as zero to use the current WOW/non-WOW registry point.
//
// NOTE(dij): WOW64_32 is 0x200 and WOW64_64 is 0x100
func (k Key) DeleteKeyEx(s string, flags uint32) error {
	return winapi.RegDeleteKeyEx(uintptr(k), s, flags)
}

// Create creates a key named path under the open key.
//
// Convince function added directly to the Key alias.
//
// CreateKey returns the new key and a boolean flag that reports whether the key
// already existed.
//
// The access parameter specifies the access rights for the key to be created.
func (k Key) Create(s string, a uint32) (Key, bool, error) {
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
