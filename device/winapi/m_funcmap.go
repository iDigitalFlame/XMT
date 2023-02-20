//go:build windows && funcmap
// +build windows,funcmap

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
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/iDigitalFlame/xmt/data"
)

var (
	funcLock   sync.Mutex
	funcMapper map[uint32]*funcMap
)

type funcMap struct {
	_    [0]func()
	proc *lazyProc
	bak  uintptr
	swap uintptr
	len  uint32
}

// FuncEntry is a simple struct that is used to describe the current status of
// function mappings. This struct is returned by a call to 'FuncRemaps' in a
// slice of current remaps.
type FuncEntry struct {
	_        [0]func()
	Hash     uint32
	Swapped  uintptr
	Original uintptr
}

// FuncUnmapAll attempts to call 'FuncUnmap' on all currently mapped functions.
// If any error occurs during unmapping, this function will stop and return an
// error. Errors will stop any pending unmap calls from occuring.
func FuncUnmapAll() error {
	funcLock.Lock()
	for _, v := range funcMapper {
		if err := v.unmap(); err != nil {
			funcLock.Unlock()
			return err
		}
	}
	funcLock.Unlock()
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
func FuncUnmap(f string) error {
	if len(f) == 0 {
		return os.ErrNotExist
	}
	return FuncUnmapHash(FnvHash(f))
}
func (v *funcMap) unmap() error {
	if v.bak == 0 || v.swap == 0 || v.len == 0 {
		return nil
	}
	atomic.SwapUintptr(&v.proc.addr, v.bak)
	err := NtFreeVirtualMemory(CurrentProcess, v.swap)
	v.swap, v.len = 0, 0
	return err
}

// FuncRemapList returns a list of all current remapped functions. This includes
// the old and new addresses and the function name hash.
//
// If no functions are remapped, this function returns nil.
func FuncRemapList() []FuncEntry {
	if funcLock.Lock(); len(funcMapper) == 0 {
		funcLock.Unlock()
		return nil
	}
	r := make([]FuncEntry, 0, len(funcMapper))
	for k, v := range funcMapper {
		if v.bak == 0 || v.swap == 0 {
			continue
		}
		r = append(r, FuncEntry{Hash: k, Swapped: v.swap, Original: v.bak})
	}
	funcLock.Unlock()
	return r
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
func FuncUnmapHash(h uint32) error {
	if h == 0 {
		return os.ErrNotExist
	}
	funcLock.Lock()
	v, ok := funcMapper[h]
	if funcLock.Unlock(); !ok {
		return os.ErrNotExist
	}
	return v.unmap()
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
func FuncRemap(f string, b []byte) error {
	if len(f) == 0 {
		return os.ErrNotExist
	}
	return FuncRemapHash(FnvHash(f), b)
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
func FuncRemapHash(h uint32, b []byte) error {
	if h == 0 {
		return os.ErrNotExist
	}
	if len(b) == 0 {
		return syscall.EINVAL
	}
	funcLock.Lock()
	v, ok := funcMapper[h]
	if funcLock.Unlock(); !ok {
		return os.ErrNotExist
	}
	if err := v.proc.find(); err != nil {
		return err
	}
	atomic.CompareAndSwapUintptr(&v.bak, 0, v.proc.addr) // Single swap to prevent lock
	var (
		n      = uint32(len(b))
		a, err = NtAllocateVirtualMemory(CurrentProcess, n, 0x4)
		// 0x4 - PAGE_READWRITE
	)
	if err != nil {
		return err
	}
	for i := range b {
		(*(*[1]byte)(unsafe.Pointer(a + uintptr(i))))[0] = b[i]
	}
	// 0x20 - PAGE_EXECUTE_READ
	if _, err = NtProtectVirtualMemory(CurrentProcess, a, n, 0x20); err == nil {
		v.swap, v.len = a, n
		atomic.SwapUintptr(&v.proc.addr, a)
		return nil
	}
	NtFreeVirtualMemory(CurrentProcess, a)
	return err
}
func registerSyscall(p *lazyProc, n string, h uint32) {
	if funcLock.Lock(); funcMapper == nil {
		funcMapper = make(map[uint32]*funcMap, 8)
	}
	switch {
	case h == 0 && len(n) == 0:
	case h > 0:
		funcMapper[h] = &funcMap{proc: p}
	case len(n) > 0:
		funcMapper[FnvHash(n)] = &funcMap{proc: p}
	}
	funcLock.Unlock()
}

// MarshalStream transforms this struct into a binary format and writes to the
// supplied data.Writer.
func (e FuncEntry) MarshalStream(w data.Writer) error {
	if err := w.WriteUint32(e.Hash); err != nil {
		return err
	}
	if err := w.WriteUint64(uint64(e.Original)); err != nil {
		return err
	}
	return w.WriteUint64(uint64(e.Swapped))
}
