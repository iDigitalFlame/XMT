package devtools

import (
	"io"
	"os"
	"time"
	"unicode/utf16"
	"unsafe"

	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Registry constant types ripped from
//  https://cs.opensource.google/go/x/sys/+/0f9fa26a:windows/registry/value.go;l=17
// to remove the dependency for *nix systems to use this package.
const (
	TypeString         = 1
	TypeExpandString   = 2
	TypeBinary         = 3
	TypeDWORD          = 4
	TypeDWORDBigEndian = 5
	TypeMultiString    = 7
	TypeQWORD          = 11
)

// ErrUnexpectedType is returned by the value retriving functions when the value's type was not the
// requested underlying type.
var ErrUnexpectedType = xerr.New("unexpected key value type")

// RegistryFile is a struct that is returned from a Registry function call on Windows devices.
// This interface is a combinaton of the io.Reader and os.FileInfo interfaces.
type RegistryFile struct {
	_    [0]func()
	m    time.Time
	k, v string
	b    []byte
	pos  int
	t    byte
}

// Len returns the number of bytes of the unread portion of the RegistryFile.
func (r *RegistryFile) Len() int {
	if r.pos >= len(r.b) {
		return 0
	}
	return len(r.b) - r.pos
}

// Type returns the Registry Value type, expressed as an integer. This value will be 0 (NONE) for Keys.
func (r *RegistryFile) Type() int {
	return int(r.t)
}

// Size returns the size of the data enclosed in this RegistryFile. This function returns 0 if the
// path is to a registry key or there is no data to read.
func (r *RegistryFile) Size() int64 {
	return int64(len(r.b))
}

// IsDir returns true if the specified registry path represents a key.
func (r *RegistryFile) IsDir() bool {
	return len(r.b) > 0
}

// Name returns the full path of this RegistryFile.
func (r *RegistryFile) Name() string {
	if len(r.b) == 0 {
		return r.k
	}
	return r.k + ":" + r.v
}

// Close fulfills the io.Closer interface. For this struct, this function clears any internal buffers
// and always returns nil.
func (r *RegistryFile) Close() error {
	r.b, r.pos = nil, 0
	return nil
}

// Similar to the Name function, this returns the full path of this RegistryFile.
func (r *RegistryFile) String() string {
	return r.Name()
}

// Sys will return a pointer to the underlying buffer if the RegistryFile represents a value.
func (r RegistryFile) Sys() interface{} {
	return r.b
}

// Mode returns the file mode of this RegistryFile. This will return a ModeDir is this represents a key.
func (r *RegistryFile) Mode() os.FileMode {
	if len(r.b) == 0 {
		return os.ModeDir | os.ModeExclusive | os.ModeIrregular
	}
	return os.ModeExclusive | os.ModeIrregular
}

// ModTime returns the RegistryFile's last modified time, if avaliable.
func (r *RegistryFile) ModTime() time.Time {
	return r.m
}

// Int retrieves the integer value for the specified RegistryFile value.
// If value is not DWORD (TypeDWORD), QWORD (TypeQWORD) or DWORD_BIG_ENDIAN (TypeDWORDBigEndian), it will return ErrUnexpectedType.
// If the buffer does not contain enough space to read the requested type size, it will return an error.
//
// This function will advance the buffer 4 bytes (DWORD) or 8 bytes (QWORD) and may continue to have leftover data.
func (r *RegistryFile) Int() (uint64, error) {
	switch r.t {
	case TypeDWORD:
		if r.Len() < 4 {
			return 0, xerr.New("DWORD value is not 4 bytes long")
		}
		_ = r.b[r.pos+3]
		v := uint64(r.b[r.pos]) | uint64(r.b[r.pos+1])<<8 | uint64(r.b[r.pos+2])<<16 | uint64(r.b[r.pos+3])<<24
		r.pos += 4
		return v, nil
	case TypeQWORD:
		if r.Len() < 8 {
			return 0, xerr.New("QWORD value is not 8 bytes long")
		}
		_ = r.b[r.pos+7]
		v := uint64(r.b[r.pos]) | uint64(r.b[r.pos+1])<<8 | uint64(r.b[r.pos+2])<<16 | uint64(r.b[r.pos+3])<<24 |
			uint64(r.b[r.pos+4])<<32 | uint64(r.b[r.pos+5])<<40 | uint64(r.b[r.pos+6])<<48 | uint64(r.b[r.pos+7])<<56
		r.pos += 8
		return v, nil
	case TypeDWORDBigEndian:
		if r.Len() < 4 {
			return 0, xerr.New("DWORD value is not 4 bytes long")
		}
		_ = r.b[r.pos+3]
		v := uint64(r.b[r.pos+3]) | uint64(r.b[r.pos+2])<<8 | uint64(r.b[r.pos+1])<<16 | uint64(r.b[r.pos])<<24
		r.pos += 4
		return v, nil
	}
	return 0, ErrUnexpectedType
}

// Bytes retrieves the binary value for the specified RegistryFile value.
// This function does not verify the underlying type, which allows for direct access to the raw Registry byte values.
//
// This function will empty the underlying buffer. Future calls to 'Read' will return 'io.EOF'.
func (r *RegistryFile) Bytes() ([]byte, error) {
	if r.Len() == 0 {
		return nil, io.EOF
	}
	b := r.b[r.pos:]
	r.pos += len(b)
	return b, nil
}

// Strings retrieves the []string value for the specified RegistryFile value.
// If value is not MULTI_SZ (TypeMultiString), it will return ErrUnexpectedType.
//
// This function will empty the underlying buffer. Future calls to 'Read' will return 'io.EOF'.
func (r *RegistryFile) Strings() ([]string, error) {
	if r.t != TypeMultiString {
		return nil, ErrUnexpectedType
	}
	if r.Len() == 0 {
		return nil, io.EOF
	}
	b := r.b[r.pos:]
	r.pos += len(b)
	u := (*[1 << 29]uint16)(unsafe.Pointer(&b[0]))[: len(b)/2 : len(b)/2]
	if len(u) == 0 {
		return nil, nil
	}
	if u[len(u)-1] == 0 {
		u = u[:len(u)-1]
	}
	o := make([]string, 0, 5)
	for i, x := 0, 0; i < len(u); i++ {
		if u[i] != 0 {
			continue
		}
		o = append(o, string(utf16.Decode(u[x:i])))
		x = i + 1
	}
	return o, nil
}

// StringVal retrieves the string value for the specified RegistryFile value.
// If value is not SZ (TypeString) or EXPAND_SZ (TypeExpandString), it will return ErrUnexpectedType.
//
// This function will empty the underlying buffer. Future calls to 'Read' will return 'io.EOF'.
func (r *RegistryFile) StringVal() (string, error) {
	switch r.t {
	case TypeString, TypeExpandString:
	default:
		return "", ErrUnexpectedType
	}
	if r.Len() == 0 {
		return "", io.EOF
	}
	b := r.b[r.pos:]
	r.pos += len(b)
	u := (*[1 << 29]uint16)(unsafe.Pointer(&b[0]))[: len(b)/2 : len(b)/2]
	if len(u) == 0 {
		return "", nil
	}
	for i := range u {
		if u[i] > 0 {
			continue
		}
		return string(utf16.Decode(u[:i])), nil
	}
	return string(utf16.Decode(u)), nil
}

// Read will attempt to read the data from this RegistryFile into the supplied buffer. This will return
// io.EOF if this struct represents a key or there is no data left to read.
func (r *RegistryFile) Read(b []byte) (int, error) {
	if len(r.b) == 0 || r.pos >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(b, r.b[r.pos:])
	r.pos += n
	return n, nil
}

// Seek will attempt to seek to the provided offset index and whence. This function will return the new offset
// if successful and will return an error if the offset and/or whence are invalid.
func (r *RegistryFile) Seek(o int64, w int) (int64, error) {
	switch w {
	case io.SeekStart:
		if o < 0 {
			return 0, data.ErrInvalidIndex
		}
	case io.SeekCurrent:
		o += int64(r.pos)
	case io.SeekEnd:
		o += int64(len(r.b))
	default:
		return 0, xerr.New("seek whence is invalid")
	}
	if o < 0 || int(o) > len(r.b) {
		return 0, data.ErrInvalidIndex
	}
	r.pos = int(o)
	return o, nil
}

// WriteTo writes data to the supplied Writer until there's no more data to write or when an error occurs. The return
// value is the number of bytes written. Any error encountered during the write is also returned.
func (r *RegistryFile) WriteTo(w io.Writer) (int64, error) {
	if len(r.b) == 0 || r.pos >= len(r.b) {
		return 0, io.EOF
	}
	n, err := w.Write(r.b[r.pos:])
	return int64(n), err
}
