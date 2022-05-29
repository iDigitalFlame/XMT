//go:build windows && (altload || crypt)

package winapi

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

type lazyDLL struct {
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
type imageNtHeader struct {
	Signature uint32
	File      imageFileHeader
	Optional  imageOptionalHeader
}
type imageExportDir struct {
	_, _                  uint32
	_, _                  uint16
	Name                  uint32
	Base                  uint32
	NumberOfFunctions     uint32
	NumberOfNames         uint32
	AddressOfFunctions    uint32
	AddressOfNames        uint32
	AddressOfNameOrdinals uint32
}
type imageDosHeader struct {
	magic               uint16
	_, _, _, _, _, _, _ uint16
	_, _, _, _, _, _    uint16
	_                   [4]uint16
	_                   uint16
	_                   [10]uint16
	pos                 int32
}
type imageFileHeader struct {
	_, _            uint16
	_, _, _         uint32
	_               uint16
	Characteristics uint16
}
type imageDataDirectory struct {
	VirtualAddress uint32
	Size           uint32
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
	if err != nil {
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
func (d *lazyDLL) initFunctions(h uintptr) error {
	b := (*imageDosHeader)(unsafe.Pointer(h))
	if b.magic != 0x5A4D {
		return xerr.Sub("base is not a valid DOS header", 0x19)
	}
	n := (*imageNtHeader)(unsafe.Pointer(h + uintptr(b.pos)))
	if n.Signature != 0x00004550 {
		return xerr.Sub("offset base is not a valid NT header", 0x1A)
	}
	if n.File.Characteristics&0x2000 == 0 {
		return xerr.Sub("header does not represent a DLL", 0x1B)
	}
	if n.Optional.Directory[0].Size == 0 || n.Optional.Directory[0].VirtualAddress == 0 {
		return xerr.Sub("header has an invalid first entry point", 0x1C)
	}
	var (
		i = (*imageExportDir)(unsafe.Pointer(h + uintptr(n.Optional.Directory[0].VirtualAddress)))
		f = h + uintptr(i.AddressOfFunctions)
		s = h + uintptr(i.AddressOfNames)
		o = h + uintptr(i.AddressOfNameOrdinals)
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
		p.addr = a
		delete(d.funcs, k)
	}
	d.funcs = nil
	return nil
}
