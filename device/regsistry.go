package device

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const (
	modeDir  = os.ModeDir | os.ModeExclusive | os.ModeIrregular
	modeFile = os.ModeExclusive | os.ModeIrregular
)

// RegistryFile is a struct that is returned from a Registry function call on Windows devices.
// This interface is a combinaton of the io.Reader and os.FileInfo interfaces.
type RegistryFile struct {
	_    [0]func()
	r    io.Reader
	m    time.Time
	k, v string
}

// Close fulfills the io.Closer interface. For this struct, this function does nothing and always returns nil.
func (RegistryFile) Close() error {
	return nil
}

// Size returns the size of the data enclosed in this RegistryFile. This function returns 0 if the
// path is to a registry key or there is not data to read.
func (r RegistryFile) Size() int64 {
	if r.r == nil {
		return 0
	}
	if i, ok := r.r.(*strings.Reader); ok {
		return int64(i.Len())
	}
	if i, ok := r.r.(*bytes.Reader); ok {
		return int64(i.Len())
	}
	return 0
}

// IsDir returns true if the specified registry path represents a key.
func (r RegistryFile) IsDir() bool {
	return r.r == nil
}

// Name returns the full path of this RegistryFile.
func (r RegistryFile) Name() string {
	if r.r == nil {
		return r.k
	}
	return fmt.Sprintf("%s:%s", r.k, r.v)
}

// Similar to the Name function, this returns the full path of this RegistryFile.
func (r RegistryFile) String() string {
	return r.Name()
}

// Sys will return a pointer to the underlying buffer if the RegistryFile represents a value.
func (r RegistryFile) Sys() interface{} {
	return r.r
}

// Mode returns the file mode of this RegistryFile. This will return a ModeDir is this represents a key.
func (r RegistryFile) Mode() os.FileMode {
	if r.r == nil {
		return modeDir
	}
	return modeFile
}

// ModTime returns the RegistryFile's last modified time, if avaliable.
func (r RegistryFile) ModTime() time.Time {
	return r.m
}

// Read will attempt to read the data from this RegistryFile into the supplied buffer. This will return
// io.EOF if this struct represents a key.
func (r *RegistryFile) Read(b []byte) (int, error) {
	if r.r == nil {
		return 0, io.EOF
	}
	return r.r.Read(b)
}
