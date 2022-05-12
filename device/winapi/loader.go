//go:build windows

package winapi

import (
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

const errPending = syscall.Errno(997)

var searchSystem32 struct {
	sync.Once
	v bool
}

type lazyDLL struct {
	_    [0]func()
	Name string
	lock sync.Mutex
	addr uintptr
}
type lazyProc struct {
	_    [0]func()
	lock sync.Mutex
	dll  *lazyDLL
	Name string
	addr uintptr
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
func (d *lazyDLL) load() error {
	if atomic.LoadUintptr(&d.addr) > 0 {
		return nil
	}
	if d.lock.Lock(); d.addr > 0 {
		d.lock.Unlock()
		return nil
	}
	if len(d.Name) == 0 {
		return xerr.Sub("empty DLL name", 0x93)
	}
	var (
		h   uintptr
		err error
	)
	if len(d.Name) == 12 && d.Name[0] == 'k' && d.Name[2] == 'r' && d.Name[3] == 'n' {
		h, err = loadDLL(d.Name)
	} else {
		h, err = loadLibraryEx(d.Name)
	}
	if err == nil {
		atomic.StoreUintptr(&d.addr, h)
	}
	d.lock.Unlock()
	return err
}
func (p *lazyProc) find() error {
	if atomic.LoadUintptr(&p.addr) == 0 {
		var err error
		if p.lock.Lock(); p.addr == 0 {
			if err = p.dll.load(); err != nil {
				return err
			}
			var h uintptr
			if h, err = findProc(p.dll.addr, p.Name, p.dll.Name); err == nil {
				atomic.StoreUintptr(&p.addr, h)
			}
		}
		p.lock.Unlock()
		return err
	}
	return nil
}
func (p *lazyProc) address() uintptr {
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
		if xerr.Concat {
			return 0, xerr.Wrap(`cannot load DLL "`+s+`"`, err)
		}
		return 0, xerr.Wrap("cannot load DLL", err)
	}
	return h, nil
}
func byteSlicePtr(s string) (*byte, error) {
	if strings.IndexByte(s, 0) != -1 {
		return nil, syscall.EINVAL
	}
	a := make([]byte, len(s)+1)
	copy(a, s)
	return &a[0], nil
}
func (d *lazyDLL) proc(n string) *lazyProc {
	return &lazyProc{Name: n, dll: d}
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
	v, err := byteSlicePtr(s)
	if err != nil {
		return 0, err
	}
	h, err2 := syscallGetProcAddress(h, v)
	if err2 != 0 {
		if xerr.Concat {
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
