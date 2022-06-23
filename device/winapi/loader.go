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

package winapi

import (
	"sync"
	"syscall"
	"unsafe"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

const errPending = syscall.Errno(997)

var searchSystem32 struct {
	sync.Once
	v bool
}

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
func (p *lazyProc) address() uintptr {
	if p.addr > 0 {
		// NOTE(dij): Might be racy, but will catch most of the re-used calls
		//            that are already populated without an additional alloc and
		//            call to "find".
		return p.addr
	}
	if err := p.find(); err != nil {
		if !canPanic {
			syscall.Exit(2)
			return 0
		}
		panic(err.Error())
	}
	return p.addr
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
func loadLibraryEx(s string) (uintptr, error) {
	var (
		n = s
		f uintptr
	)
	if doSearchSystem32() {
		// 0x800 - LOAD_LIBRARY_SEARCH_SYSTEM32
		f = 0x800
	} else if isBaseName(s) {
		d, err := GetSystemDirectory()
		if err != nil {
			return 0, err
		}
		n = d + "\\" + s
	}
	return LoadLibraryEx(n, f)
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
func getSystemDirectory(s *uint16, n uint32) (uint32, error) {
	r, _, e := syscall.SyscallN(funcGetSystemDirectory.address(), uintptr(unsafe.Pointer(s)), uintptr(n))
	if r == 0 {
		return 0, unboxError(e)
	}
	return uint32(r), nil
}

//go:linkname syscallGetProcAddress syscall.getprocaddress
func syscallGetProcAddress(h uintptr, n *uint8) (uintptr, syscall.Errno)
