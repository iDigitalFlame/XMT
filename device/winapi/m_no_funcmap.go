//go:build windows && !funcmap

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
	"syscall"

	"github.com/iDigitalFlame/xmt/data"
)

// FuncEntry is a simple struct that is used to describe the current status of
// function mappings. This struct is returned by a call to 'FuncRemaps' in a
// slice of current remaps.
type FuncEntry struct{}

// FuncUnmapAll attempts to call 'FuncUnmap' on all currently mapped functions.
// If any error occurs during unmapping, this function will stop and return an
// error. Errors will stop any pending unmap calls from occuring.
func FuncUnmapAll() error {
	return nil
}

// FuncUnmap will attempt to unmap the ntdll.dll function by name. If successful
// all calls to the affected function will work normally and the allocated memory
// region will be freed.
//
// This function returns ErrNotExist if the function name is not a recognized
// ntdll.dll function that does a direct syscall.
//
// This function returns nil even if the function was not previously remapped.
//
// If this function returns any errors do not assume the call site was fixed
// to behave normally.
func FuncUnmap(_ string) error {
	return nil
}

// FuncRemapList returns a list of all current remapped functions. This includes
// the old and new addresses and the function name hash.
//
// If no functions are remapped, this function returns nil.
func FuncRemapList() []FuncEntry {
	return nil
}

// FuncUnmapHash will attempt to unmap the ntdll.dll by its function hash. If
// successful all calls to the affected function will work normally and the
// allocated memory region will be freed.
//
// This function returns ErrNotExist if the function name is not a recognized
// ntdll.dll function that does a direct syscall.
//
// This function returns nil even if the function was not previously remapped.
//
// If this function returns any errors do not assume the call site was fixed
// to behave normally.
func FuncUnmapHash(_ uint32) error {
	return nil
}

// FuncRemap attempts to remap the raw ntdll.dll function name with the supplied
// machine-code bytes. If successful, this will point all function calls in the
// runtime to that allocated byte array in memory, bypassing any hooked calls
// without overriting any existing memory.
//
// This function returns EINVAL if the byte slice is empty or ErrNotExist if the
// function name is not a recognized ntdll.dll function that does a direct syscall.
//
// It is recommended to call 'FuncUnmap(name)' or 'FuncUnmapAll' once complete
// to release the memory space.
//
// The 'Func*' functions only work of the build tag "funcmap" is used during
// buildtime, otherwise these functions return EINVAL.
func FuncRemap(_ string, _ []byte) error {
	return syscall.EINVAL
}

// FuncRemapHash attempts to remap the raw ntdll.dll function hash with the supplied
// machine-code bytes. If successful, this will point all function calls in the
// runtime to that allocated byte array in memory, bypassing any hooked calls
// without overriting any existing memory.
//
// This function returns EINVAL if the byte slice is empty or ErrNotExist if the
// function hash is not a recognized ntdll.dll function that does a direct syscall.
//
// It is recommended to call 'FuncUnmap(name)' or 'FuncUnmapAll' once complete
// to release the memory space.
//
// The 'Func*' functions only work of the build tag "funcmap" is used during
// buildtime, otherwise these functions return EINVAL.
func FuncRemapHash(_ uint32, _ []byte) error {
	return syscall.EINVAL
}

// MarshalStream transforms this struct into a binary format and writes to the
// supplied data.Writer.
func (FuncEntry) MarshalStream(_ data.Writer) error {
	return nil
}
func registerSyscall(_ *lazyProc, _ string, _ uint32) {}
