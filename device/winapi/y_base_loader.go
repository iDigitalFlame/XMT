//go:build windows && !altload && !crypt
// +build windows,!altload,!crypt

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
	"sync/atomic"
	"syscall"
)

// LazyDLL is a struct that can be used to load a DLL into the current Process.
// The DLL is not loaded until a function is executed.
//
// If the 'crypt' build tag is present, function names MUST be FNV32 hashes.
type LazyDLL struct {
	_    [0]func()
	name string
	sync.Mutex
	addr uintptr
}

// LazyProc is a struct returned from a LazyDLL struct. This represents a function
// that can be called from the target DLL. This struct does not load the function
// address until called or the 'Load' function is called.
//
// If the 'crypt' or 'altload' build tag is present, function names MUST be FNV32 hashes.
type LazyProc struct {
	_    [0]func()
	dll  *LazyDLL
	name string
	sync.Mutex
	addr uintptr
}

// Free will call tne 'FreeLibrary' function on the DLL (if loaded) and release
// it's resources. After being free'd, it is recommended to NOT call any functions
// loaded from it, as it may cause undefined behavior.
//
// Extra calls to Free do nothing.
func (d *LazyDLL) Free() error {
	if d.addr == 0 {
		return nil
	}
	d.Lock()
	err := syscall.FreeLibrary(syscall.Handle(d.addr))
	atomic.StoreUintptr(&d.addr, 0)
	d.Unlock()
	return err
}

// Load will force the DLL and all functions to be loaded, if not already.
//
// If the DLL is already loaded, this function does nothing.
//
// It is recommended to NOT call this directly until all functions are retrieved
// as newly generated LazyProc functions may not map.
func (d *LazyDLL) Load() error {
	if atomic.LoadUintptr(&d.addr) > 0 {
		return nil
	}
	d.Lock()
	var (
		h   uintptr
		err error
	)
	if (len(d.name) == 12 || len(d.name) == 14) && d.name[0] == 'k' && d.name[2] == 'r' && d.name[3] == 'n' {
		if h, err = loadDLL(d.name); fallbackLoad {
			if h == 0 && len(d.name) == 14 {
				// NOTE(dij): The "kernelbase.dll" file was not avaliable before
				//            Windows 7 so we'll redirect all KernelBase calls to
				//            Kernel32. We can tell this is "kernelbase.dll" fails
				//            to load.
				d.name = dllKernel32.name
				h, err = loadDLL(dllKernel32.name)
			}
		}
	} else {
		h, err = loadLibraryEx(d.name)
	}
	if h > 0 {
		atomic.StoreUintptr(&d.addr, h)
	}
	d.Unlock()
	return err
}

// Load will force the DLL that owns this function and all functions to be loaded,
// if not already.
//
// If the DLL is already loaded, this function does nothing.
//
// If the function does not exist, this call returns an error.
//
// It is recommended to NOT call this directly until all functions are retrieved
// as newly generated LazyProc functions may not map.
func (p *LazyProc) Load() error {
	if atomic.LoadUintptr(&p.addr) > 0 {
		return nil
	}
	var err error
	p.Lock()
	if err = p.dll.Load(); err != nil {
		p.Unlock()
		return err
	}
	var h uintptr
	if h, err = findProc(p.dll.addr, p.name, p.dll.name); err == nil {
		atomic.StoreUintptr(&p.addr, h)
	}
	p.Unlock()
	return err
}

// Proc will return a LazyProc that links the specified function name or hash.
// The DLL or function won't be loaded until called or the 'LazyProc.Load()'
// function is called.
//
// If the 'crypt' or 'altload' build tag is present, function names MUST be FNV32 hashes.
func (d *LazyDLL) Proc(n string) *LazyProc {
	return &LazyProc{name: n, dll: d}
}
func (d *LazyDLL) sysProc(n string) *LazyProc {
	if len(d.name) != 9 && d.name[0] != 'n' && d.name[1] != 't' {
		return d.Proc(n)
	}
	p := d.Proc(n)
	registerSyscall(p, n, 0)
	return p
}
