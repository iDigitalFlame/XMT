//go:build windows
// +build windows

package registry

import (
	"io"
	"syscall"
	"unsafe"

	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Registry value types.
const (
	TypeString         = 1
	TypeExpandString   = 2
	TypeBinary         = 3
	TypeDword          = 4
	TypeDwordBigEndian = 5
	TypeStringList     = 7
	TypeQword          = 11
)

var (
	// ErrNotExist is returned when a registry key or value does not exist.
	ErrNotExist = syscall.ERROR_FILE_NOT_FOUND
	// ErrShortBuffer is returned when the buffer was too short for the operation.
	ErrShortBuffer = syscall.ERROR_MORE_DATA
	// ErrUnexpectedSize is returned when the key data size was unexpected.
	ErrUnexpectedSize = xerr.Sub("unexpected key size", 0x10)
	// ErrUnexpectedType is returned by Get*Value when the value's type was
	// unexpected.
	ErrUnexpectedType = xerr.Sub("unexpected key type", 0xD)
)

// DeleteValue removes a named value from the key.
func (k Key) DeleteValue(n string) error {
	return winapi.RegDeleteValue(uintptr(k), n)
}

// SetStringValue sets the data and type of a named value under the key to the
// supplied value and SZ.
//
// The value must not contain a zero byte.
func (k Key) SetStringValue(n, v string) error {
	return k.setStringValue(n, TypeString, v)
}

// SetExpandStringValue sets the data and type of a named value under the key
// to the supplied value and EXPAND_SZ.
//
// The value must not contain a zero byte.
func (k Key) SetExpandStringValue(n, v string) error {
	return k.setStringValue(n, TypeExpandString, v)
}

// ReadValueNames returns the value names in the key.
//
// The parameter controls the number of returned names, analogous to the way
// 'os.File.Readdirnames' works.
func (k Key) ReadValueNames(n int) ([]string, error) {
	x, err := k.Stat()
	if err != nil {
		return nil, err
	}
	var (
		o = make([]string, 0, x.ValueCount)
		b = make([]uint16, x.MaxValueNameLen+1)
	)
loop:
	for i := uint32(0); ; i++ {
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

// SetDWordValue sets the data and type of a named value under the key to the
// supplied value and DWORD.
func (k Key) SetDWordValue(n string, v uint32) error {
	return k.setValue(n, TypeDword, (*[4]byte)(unsafe.Pointer(&v))[:])
}

// SetQWordValue sets the data and type of a named value under the key to the
// supplied value and QWORD.
func (k Key) SetQWordValue(n string, v uint64) error {
	return k.setValue(n, TypeQword, (*[8]byte)(unsafe.Pointer(&v))[:])
}

// SetBinaryValue sets the data and type of a name value
// under key k to value and BINARY.
func (k Key) SetBinaryValue(n string, v []byte) error {
	return k.setValue(n, TypeBinary, v)
}

// SetStringsValue sets the data and type of a named value under the key to the
// supplied value and MULTI_SZ.
//
// The value strings must not contain a zero byte.
func (k Key) SetStringsValue(n string, v []string) error {
	var b util.Builder
	for i := range v {
		for x := 0; x < len(v[i]); i++ {
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
func (k Key) setValue(s string, t uint32, d []byte) error {
	if len(d) == 0 {
		return winapi.RegSetValueEx(uintptr(k), s, t, nil, 0)
	}
	return winapi.RegSetValueEx(uintptr(k), s, t, &d[0], uint32(len(d)))
}

// GetStringValue retrieves the string value for the specified value name
// associated with the open key. It also returns the value's type.
//
// If value does not exist, GetStringValue returns ErrNotExist.
//
// If value is not SZ or EXPAND_SZ, it will return the correct value type and
// ErrUnexpectedType.
func (k Key) GetStringValue(s string) (string, uint32, error) {
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

// GetBinaryValue retrieves the binary value for the specified value name
// associated with the open key. It also returns the value's type.
//
// If value does not exist, GetBinaryValue returns ErrNotExist.
//
// If value is not BINARY, it will return the correct value type and
// ErrUnexpectedType.
func (k Key) GetBinaryValue(s string) ([]byte, uint32, error) {
	d, t, err := k.getValue(s, make([]byte, 64))
	if err != nil {
		return nil, t, err
	}
	if t != TypeBinary {
		return nil, t, ErrUnexpectedType
	}
	return d, t, nil
}

// GetValue retrieves the type and data for the specified value associated with
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
// appropriate Get*Value function instead.
func (k Key) GetValue(s string, b []byte) (int, uint32, error) {
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

// GetIntegerValue retrieves the integer value for the specified value name
// associated with the open key. It also returns the value's type.
//
// If value does not exist, GetIntegerValue returns ErrNotExist.
//
// If value is not DWORD or QWORD, it will return the correct value type and
// ErrUnexpectedType.
func (k Key) GetIntegerValue(s string) (uint64, uint32, error) {
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
	case TypeDwordBigEndian:
		if len(d) != 4 {
			return 0, t, ErrUnexpectedSize
		}
		_ = d[3]
		return uint64(d[3]) | uint64(d[2])<<8 | uint64(d[1])<<16 | uint64(d[0])<<24, t, nil
	case TypeQword:
		if len(d) != 8 {
			return 0, t, ErrUnexpectedSize
		}
		return uint64(d[0]) | uint64(d[1])<<8 | uint64(d[2])<<16 | uint64(d[3])<<24 |
			uint64(d[4])<<32 | uint64(d[5])<<40 | uint64(d[6])<<48 | uint64(d[7])<<56, t, nil
	default:
	}
	return 0, t, ErrUnexpectedType
}
func (k Key) setStringValue(n string, t uint32, v string) error {
	p, err := winapi.UTF16FromString(v)
	if err != nil {
		return err
	}
	return k.setValue(n, t, (*[1 << 29]byte)(unsafe.Pointer(&p[0]))[:len(p)*2:len(p)*2])
}

// GetStringsValue retrieves the []string value for the specified value name
// associated with an open key k. It also returns the value's type.
//
// If value does not exist, GetStringsValue returns ErrNotExist.
//
// If value is not MULTI_SZ, it will return the correct value type and
// ErrUnexpectedType.
func (k Key) GetStringsValue(s string) ([]string, uint32, error) {
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
