//go:build windows

// Package registry
//
// Optimized copy from sys/windows/registry to work with Crypt.
package registry

import (
	"io"
	"runtime"
	"syscall"

	"github.com/iDigitalFlame/xmt/device/winapi"
)

// Key is a handle to an open Windows registry key.
//
// Keys can be obtained by calling OpenKey.
type Key uintptr

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

	errNoMoreItems syscall.Errno = 259
)

// Close closes the open key.
func (k Key) Close() error {
	return syscall.RegCloseKey(syscall.Handle(k))
}

// DeleteKey deletes the subkey path of the key and its values.
func DeleteKey(k Key, s string) error {
	return winapi.RegDeleteKey(uintptr(k), s)
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

// OpenKey opens a new key with path name relative to the key.
//
// It accepts any open key, including CURRENT_USER and others, and returns the
// new key and an error.
//
// The access parameter specifies desired access rights to the key to be opened.
func OpenKey(k Key, s string, a uint32) (Key, error) {
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

// ReadSubKeyNames returns the names of subkeys of key k.
//
// The parameter controls the number of returned names, analogous to the way
// 'os.File.Readdirnames' works.
func (k Key) ReadSubKeyNames(n int) ([]string, error) {
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
	if err == errNoMoreItems {
		err = nil
	}
	if n > len(o) {
		return o, io.EOF
	}
	return o, nil
}

// CreateKey creates a key named path under the open key.
//
// CreateKey returns the new key and a boolean flag that reports whether the key
// already existed.
//
// The access parameter specifies the access rights for the key to be created.
func CreateKey(k Key, s string, a uint32) (Key, bool, error) {
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
