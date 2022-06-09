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
	"unsafe"

	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/util"
)

var (
	// ErrNotExist is returned when a registry key or value does not exist.
	ErrNotExist = syscall.ERROR_FILE_NOT_FOUND
	// ErrShortBuffer is returned when the buffer was too short for the operation.
	ErrShortBuffer = syscall.ERROR_MORE_DATA
)

// SetString sets the data and type of a named value under the key to the
// supplied value and SZ.
//
// The value must not contain a zero byte.
func (k Key) SetString(n, v string) error {
	return k.setStringValue(n, TypeString, v)
}

// SetExpandString sets the data and type of a named value under the key
// to the supplied value and EXPAND_SZ.
//
// The value must not contain a zero byte.
func (k Key) SetExpandString(n, v string) error {
	return k.setStringValue(n, TypeExpandString, v)
}

// SetDWord sets the data and type of a named value under the key to the
// supplied value and DWORD.
func (k Key) SetDWord(n string, v uint32) error {
	return k.setValue(n, TypeDword, (*[4]byte)(unsafe.Pointer(&v))[:])
}

// SetQWord sets the data and type of a named value under the key to the
// supplied value and QWORD.
func (k Key) SetQWord(n string, v uint64) error {
	return k.setValue(n, TypeQword, (*[8]byte)(unsafe.Pointer(&v))[:])
}

// SetBinary sets the data and type of a name value under key k to value and
// BINARY.
func (k Key) SetBinary(n string, v []byte) error {
	return k.setValue(n, TypeBinary, v)
}

// SetStrings sets the data and type of a named value under the key to the
// supplied value and MULTI_SZ.
//
// The value strings must not contain a zero byte.
func (k Key) SetStrings(n string, v []string) error {
	var b util.Builder
	for i := range v {
		for x := range v[i] {
			if v[i][x] == 0 {
				return syscall.EINVAL
			}
		}
		b.WriteString(v[i] + "\x00")
	}
	var (
		r = winapi.UTF16EncodeStd([]rune(b.Output() + "\x00"))
		o = (*[1 << 29]byte)(unsafe.Pointer(&r[0]))[: len(r)*2 : len(r)*2]
	)
	return k.setValue(n, TypeStringList, o)
}

// String retrieves the string value for the specified value name
// associated with the open key. It also returns the value's type.
//
// If value does not exist, String returns ErrNotExist.
//
// If value is not SZ or EXPAND_SZ, it will return the correct value type and
// ErrUnexpectedType.
func (k Key) String(s string) (string, uint32, error) {
	d, t, err := k.getValue(s, make([]byte, 64))
	if err != nil {
		return "", t, err
	}
	switch t {
	case TypeString, TypeExpandString:
	default:
		return "", t, ErrUnexpectedType
	}
	if len(d) == 0 {
		return "", t, nil
	}
	u := (*[1 << 29]uint16)(unsafe.Pointer(&d[0]))[: len(d)/2 : len(d)/2]
	return winapi.UTF16ToString(u), t, nil
}

// Binary retrieves the binary value for the specified value name
// associated with the open key. It also returns the value's type.
//
// If value does not exist, Binary returns ErrNotExist.
//
// If value is not BINARY, it will return the correct value type and
// ErrUnexpectedType.
func (k Key) Binary(s string) ([]byte, uint32, error) {
	d, t, err := k.getValue(s, make([]byte, 64))
	if err != nil {
		return nil, t, err
	}
	if t != TypeBinary {
		return nil, t, ErrUnexpectedType
	}
	return d, t, nil
}

// Integer retrieves the integer value for the specified value name
// associated with the open key. It also returns the value's type.
//
// If value does not exist, Integer returns ErrNotExist.
//
// If value is not DWORD or QWORD, it will return the correct value type and
// ErrUnexpectedType.
func (k Key) Integer(s string) (uint64, uint32, error) {
	d, t, err := k.getValue(s, make([]byte, 8))
	if err != nil {
		return 0, t, err
	}
	switch t {
	case TypeDword:
		if len(d) != 4 {
			return 0, t, ErrUnexpectedSize
		}
		_ = d[3]
		return uint64(d[0]) | uint64(d[1])<<8 | uint64(d[2])<<16 | uint64(d[3])<<24, t, nil
	case TypeQword:
		if len(d) != 8 {
			return 0, t, ErrUnexpectedSize
		}
		_ = d[7]
		return uint64(d[0]) | uint64(d[1])<<8 | uint64(d[2])<<16 | uint64(d[3])<<24 |
			uint64(d[4])<<32 | uint64(d[5])<<40 | uint64(d[6])<<48 | uint64(d[7])<<56, t, nil
	default:
	}
	return 0, t, ErrUnexpectedType
}

// Strings retrieves the []string value for the specified value name
// associated with an open key k. It also returns the value's type.
//
// If value does not exist, Strings returns ErrNotExist.
//
// If value is not MULTI_SZ, it will return the correct value type and
// ErrUnexpectedType.
func (k Key) Strings(s string) ([]string, uint32, error) {
	d, t, err := k.getValue(s, make([]byte, 64))
	if err != nil {
		return nil, t, err
	}
	if t != TypeStringList {
		return nil, t, ErrUnexpectedType
	}
	if len(d) == 0 {
		return nil, t, nil
	}
	p := (*[1 << 29]uint16)(unsafe.Pointer(&d[0]))[: len(d)/2 : len(d)/2]
	if len(p) == 0 {
		return nil, t, nil
	}
	if p[len(p)-1] == 0 {
		p = p[:len(p)-1]
	}
	r := make([]string, 0, 5)
	for i, n := 0, 0; i < len(p); i++ {
		if p[i] > 0 {
			continue
		}
		r = append(r, string(winapi.UTF16Decode(p[n:i])))
		n = i + 1
	}
	return r, t, nil
}
func (k Key) setValue(s string, t uint32, d []byte) error {
	if len(d) == 0 {
		return winapi.RegSetValueEx(uintptr(k), s, t, nil, 0)
	}
	return winapi.RegSetValueEx(uintptr(k), s, t, &d[0], uint32(len(d)))
}

// Value retrieves the type and data for the specified value associated with
// the open key.
//
// It fills up buffer buf and returns the retrieved byte count n. If buf is too
// small to fit the stored value it returns an ErrShortBuffer error along with
// the required buffer size n.
//
// If no buffer is provided, it returns true and actual buffer size and the
// value's type only.
//
// If the value does not exist, the error returned is ErrNotExist.
//
// GetValue is a low level function. If value's type is known, use the
// appropriate "Value" function instead.
func (k Key) Value(s string, b []byte) (int, uint32, error) {
	n, err := winapi.UTF16PtrFromString(s)
	if err != nil {
		return 0, 0, err
	}
	var (
		l = uint32(len(b))
		t uint32
		o *byte
	)
	if l > 0 {
		o = (*byte)(unsafe.Pointer(&b[0]))
	}
	if err = syscall.RegQueryValueEx(syscall.Handle(k), n, nil, &t, o, &l); err != nil {
		return int(l), t, err
	}
	return int(l), t, nil
}
func (k Key) setStringValue(n string, t uint32, v string) error {
	p, err := winapi.UTF16FromString(v)
	if err != nil {
		return err
	}
	return k.setValue(n, t, (*[1 << 29]byte)(unsafe.Pointer(&p[0]))[:len(p)*2:len(p)*2])
}
func (k Key) getValue(s string, b []byte) ([]byte, uint32, error) {
	p, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		return nil, 0, err
	}
	var (
		n = uint32(len(b))
		t uint32
	)
	for {
		err = syscall.RegQueryValueEx(syscall.Handle(k), p, nil, &t, (*byte)(unsafe.Pointer(&b[0])), &n)
		if err == nil {
			return b[:n], t, nil
		}
		if err != syscall.ERROR_MORE_DATA {
			return nil, 0, err
		}
		if n <= uint32(len(b)) {
			return nil, 0, err
		}
		b = make([]byte, n)
	}
}
