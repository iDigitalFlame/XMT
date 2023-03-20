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

type lazyDLL struct {
	_    [0]func()
	name string
	sync.Mutex
	addr uintptr
}
type lazyProc struct {
	_    [0]func()
	dll  *lazyDLL
	name string
	sync.Mutex
	addr uintptr
}

func (d *lazyDLL) free() error {
	if d.addr == 0 {
		return nil
	}
	d.Lock()
	err := syscall.FreeLibrary(syscall.Handle(d.addr))
	atomic.StoreUintptr(&d.addr, 0)
	d.Unlock()
	return err
}
func (d *lazyDLL) load() error {
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
func (p *lazyProc) find() error {
	if atomic.LoadUintptr(&p.addr) > 0 {
		return nil
	}
	var err error
	p.Lock()
	if err = p.dll.load(); err != nil {
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
func (d *lazyDLL) proc(n string) *lazyProc {
	return &lazyProc{name: n, dll: d}
}
func (d *lazyDLL) sysProc(n string) *lazyProc {
	if len(d.name) != 9 && d.name[0] != 'n' && d.name[1] != 't' {
		return d.proc(n)
	}
	p := d.proc(n)
	registerSyscall(p, n, 0)
	return p
}
