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

package winapi

import (
	"sync"
	"syscall"

	// Needed to use "linkname"
	_ "unsafe"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

const errPending = syscall.Errno(997)

var searchSystem32 struct {
	_ [0]func()
	sync.Once
	v bool
}

//go:linkname systemDirectoryPrefix syscall.systemDirectoryPrefix
var systemDirectoryPrefix string

func isBaseName(s string) bool {
	for i := range s {
		switch s[i] {
		case ':', '/', '\\':
			return false
		}
	}
	return true
}
func byteSlicePtr(s string) *byte {
	a := make([]byte, len(s)+1)
	copy(a, s)
	return &a[0]
}

// IsLoaded returns true if this function has been loaded into memory.
func (p *LazyProc) IsLoaded() bool {
	return p.addr != 0
}

// NewLazyDLL returns a new LazyDLL struct bound to the supplied DLL.
//
// This function DOES NOT make any loads or calls.
func NewLazyDLL(s string) *LazyDLL {
	return &LazyDLL{name: s}
}
func (p *LazyProc) address() uintptr {
	return p.MustAddress()
}
func unboxError(e syscall.Errno) error {
	switch e {
	case 0:
		return syscall.EINVAL
	case 997:
		return errPending
	}
	return e
}

// LoadDLL loads DLL file into memory.
//
// This function will attempt to load non-absolute paths from the system
// dependent DLL directory (usually system32).
func LoadDLL(s string) (uintptr, error) {
	return loadLibraryEx(s)
}
func loadDLL(s string) (uintptr, error) {
	n, err := UTF16PtrFromString(s)
	if err != nil {
		return 0, err
	}
	h, err2 := syscallLoadLibrary(n)
	if err2 != 0 {
		if xerr.ExtendedInfo {
			return 0, xerr.Wrap(`cannot load DLL "`+s+`"`, err)
		}
		return 0, xerr.Wrap("cannot load DLL", err)
	}
	return h, nil
}

// MustAddress is like the Address function returns the memory addess of this function.
// If it's not loaded, it will attempt to be loaded. If the load fails, this function
// will panic instead of returning an error.
func (p *LazyProc) MustAddress() uintptr {
	if p.addr > 0 {
		// NOTE(dij): Might be racy, but will catch most of the re-used calls
		//            that are already populated without an additional alloc and
		//            call to "find".
		return p.addr
	}
	if err := p.Load(); err != nil {
		if !canPanic {
			syscall.Exit(2)
			return 0
		}
		panic(err.Error())
	}
	return p.addr
}
func loadLibraryEx(s string) (uintptr, error) {
	var (
		n = s
		f uintptr
	)
	if doSearchSystem32() {
		// 0x800 - LOAD_LIBRARY_SEARCH_SYSTEM32
		f = 0x800
	} else if isBaseName(s) {
		n = systemDirectoryPrefix + s
	}
	return LoadLibraryEx(n, f)
}

// Address returns the memory addess of this function. If it's not loaded, it
// will attempt to be loaded. If the load fails, this function will return an error.
func (p *LazyProc) Address() (uintptr, error) {
	if p.addr > 0 {
		// NOTE(dij): Might be racy, but will catch most of the re-used calls
		//            that are already populated without an additional alloc and
		//            call to "find".
		return p.addr, nil
	}
	if err := p.Load(); err != nil {
		return 0, err
	}
	return p.addr, nil
}

// Call will call this function with the specified arguments.
//
// This is the same as the 'syscall.SyscallN' call.
//
// If the function could not be loaded, this function will return an error.
//
// The return of this function is the function result and any errors that may
// be returned.
//
// It is recommened to check the result of the first result as the error result
// may be non-nil but still indicate success.
func (p *LazyProc) Call(a ...uintptr) (uintptr, error) {
	if err := p.Load(); err != nil {
		return 0, err
	}
	// p.addr should be loaded by here.
	r, _, err := syscallN(p.addr, a...)
	return r, unboxError(err)
}
func findProc(h uintptr, s, n string) (uintptr, error) {
	h, err := syscallGetProcAddress(h, byteSlicePtr(s))
	if err != 0 {
		if xerr.ExtendedInfo {
			return 0, xerr.Wrap(`cannot load DLL "`+n+`" function "`+s+`"`, err)
		}
		return 0, xerr.Wrap("cannot load DLL function", err)
	}
	return h, nil
}

//go:linkname syscallLoadLibrary syscall.loadlibrary
func syscallLoadLibrary(n *uint16) (uintptr, syscall.Errno)

//go:linkname syscallGetProcAddress syscall.getprocaddress
func syscallGetProcAddress(h uintptr, n *uint8) (uintptr, syscall.Errno)
