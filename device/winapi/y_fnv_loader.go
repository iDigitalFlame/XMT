//go:build windows && (altload || crypt)
// +build windows
// +build altload crypt

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
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

type lazyDLL struct {
	_ [0]func()
	sync.Mutex
	funcs map[uint32]*lazyProc
	name  string
	addr  uintptr
}
type lazyProc struct {
	_    [0]func()
	dll  *lazyDLL
	addr uintptr
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
	if h == 0 {
		d.Unlock()
		return err
	}
	atomic.StoreUintptr(&d.addr, h)
	err = d.initFunctions(h)
	d.Unlock()
	return err
}
func (p *lazyProc) find() error {
	if atomic.LoadUintptr(&p.addr) > 0 {
		return nil
	}
	if err := p.dll.load(); err != nil {
		return err
	}
	if atomic.LoadUintptr(&p.addr) > 0 {
		return nil
	}
	return xerr.Sub("cannot load DLL function", 0x18)
}
func fnvHash(b [256]byte) uint32 {
	h := uint32(2166136261)
	for i := range b {
		if b[i] == 0 {
			break
		}
		h *= 16777619
		h ^= uint32(b[i])
	}
	return h
}
func (d *lazyDLL) proc(h uint32) *lazyProc {
	if d.funcs == nil {
		d.funcs = make(map[uint32]*lazyProc)
	}
	p := &lazyProc{dll: d}
	d.funcs[h] = p
	return p
}
func (d *lazyDLL) sysProc(h uint32) *lazyProc {
	if len(d.name) != 9 && d.name[0] != 'n' && d.name[1] != 't' {
		return d.proc(h)
	}
	p := d.proc(h)
	registerSyscall(p, "", h)
	return p
}
func (d *lazyDLL) initFunctions(h uintptr) error {
	b := (*imageDosHeader)(unsafe.Pointer(h))
	if b.magic != 0x5A4D {
		return xerr.Sub("base is not a valid DOS header", 0x19)
	}
	n := *(*imageNtHeader)(unsafe.Pointer(h + uintptr(b.pos)))
	if n.Signature != 0x00004550 {
		return xerr.Sub("offset base is not a valid NT header", 0x1A)
	}
	if n.File.Characteristics&0x2000 == 0 {
		return xerr.Sub("header does not represent a DLL", 0x1B)
	}
	switch n.File.Machine {
	case 0, 0x14C, 0x1C4, 0xAA64, 0x8664:
	default:
		return xerr.Sub("header does not represent a DLL", 0x1B)
	}
	var (
		p = b.pos + int32(unsafe.Sizeof(n))
		v [16]imageDataDirectory
	)
	if *(*uint16)(unsafe.Pointer(h + uintptr(p))) == 0x20B {
		v = (*imageOptionalHeader64)(unsafe.Pointer(h + uintptr(p))).Directory
	} else {
		v = (*imageOptionalHeader32)(unsafe.Pointer(h + uintptr(p))).Directory
	}
	if v[0].Size == 0 || v[0].VirtualAddress == 0 {
		return xerr.Sub("header has an invalid first entry point", 0x1C)
	}
	var (
		i = (*imageExportDir)(unsafe.Pointer(h + uintptr(v[0].VirtualAddress)))
		f = h + uintptr(i.AddressOfFunctions)
		s = h + uintptr(i.AddressOfNames)
		o = h + uintptr(i.AddressOfNameOrdinals)
		m = h + uintptr(v[0].VirtualAddress) + uintptr(v[0].Size)
	)
	for x, k, a := uint32(0), uint32(0), uintptr(0); x < i.NumberOfNames; x++ {
		k = fnvHash(*(*[256]byte)(unsafe.Pointer(
			h + uintptr(*(*uint32)(unsafe.Pointer(s + uintptr(x*4)))),
		)))
		a = h + uintptr(
			*(*uint32)(unsafe.Pointer(f + uintptr(
				*(*uint16)(unsafe.Pointer(o + uintptr(x*2))),
			)*4)),
		)
		p, ok := d.funcs[k]
		if !ok {
			continue
		}
		if a < m && a > f {
			var err error
			if p.addr, err = loadForwardFunc((*[256]byte)(unsafe.Pointer(a))); err != nil {
				return err
			}
		} else {
			p.addr = a
		}
		delete(d.funcs, k)
	}
	d.funcs = nil
	return nil
}
func loadForwardFunc(b *[256]byte) (uintptr, error) {
	var n int
	for n < 256 {
		if (*b)[n] == 0 {
			break
		}
		n++
	}
	var (
		v = string((*b)[:n])
		i = strings.LastIndexByte(v, '.')
	)
	if i == -1 {
		return 0, syscall.EINVAL
	}
	d, f := v[0:i], v[i+1:]
	if i < 5 || v[i-4] != '.' {
		d = d + dllExt
	}
	if bugtrack.Enabled {
		bugtrack.Track(`winapi.loadForwardFunc(): Loading forwarded function "%s" from "%s".`, f, d)
	}
	x, err := loadDLL(d)
	if err != nil {
		return 0, err
	}
	return findProc(x, f, d)
}
