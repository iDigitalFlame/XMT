//go:build windows && !altload && !crypt

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
	"strings"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/iDigitalFlame/xmt/util/xerr"

	// Required to link "syscallGetProcAddress"
	_ "unsafe"
)

type lazyDLL struct {
	_    [0]func()
	name string
	sync.Mutex
	addr uintptr
}
type lazyProc struct {
	_ [0]func()
	sync.Mutex
	dll  *lazyDLL
	name string
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
	if len(d.name) == 12 && d.name[0] == 'k' && d.name[2] == 'r' && d.name[3] == 'n' {
		h, err = loadDLL(d.name)
	} else {
		h, err = loadLibraryEx(d.name)
	}
	if err == nil {
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
func byteSlicePtr(s string) (*byte, error) {
	if strings.IndexByte(s, 0) != -1 {
		return nil, syscall.EINVAL
	}
	a := make([]byte, len(s)+1)
	copy(a, s)
	return &a[0], nil
}
func findProc(h uintptr, s, n string) (uintptr, error) {
	v, err := byteSlicePtr(s)
	if err != nil {
		return 0, err
	}
	h, err2 := syscallGetProcAddress(h, v)
	if err2 != 0 {
		if xerr.ExtendedInfo {
			return 0, xerr.Wrap(`cannot load DLL "`+n+`" function "`+s+`"`, err)
		}
		return 0, xerr.Wrap("cannot load DLL function", err)
	}
	return h, nil
}

//go:linkname syscallGetProcAddress syscall.getprocaddress
func syscallGetProcAddress(h uintptr, n *uint8) (uintptr, syscall.Errno)
