//go:build windows && cgo && go1.10 && !go1.11
// +build windows,cgo,go1.10,!go1.11

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

import "unsafe"

//go:linkname gcBitsArenas runtime.gcBitsArenas
var gcBitsArenas [4]uintptr

type mheap struct {
	_         uintptr
	_         [256]uintptr
	freeLarge *treapNode
	_         [256]uintptr
	_, _      uintptr
	_, _, _   uint32
	allspans  []*mspan
	spans     []*mspan
	_         [2]struct {
		_, _, _, _ uintptr
		_          uint32
	}
	_          uint32
	_, _, _, _ uint64
	_          float64
	_, _, _, _ uint64
	_          [67]uint64
	_, _       uintptr
	base       uintptr
	_, _       uintptr
	end        uintptr
	_          bool
	_          [134]struct {
		_          uintptr
		_          uint8
		_, _, _, _ uintptr
		_          uint64
		_          [cacheLineSize - ((5*ptrSize)+9)%cacheLineSize]byte
	}
	spanalloc             fixalloc
	cachealloc            fixalloc
	treapalloc            fixalloc
	specialfinalizeralloc fixalloc
	specialprofilealloc   fixalloc
}
type mspan struct {
	_, _      *mspan
	_         uintptr
	startAddr uintptr
}
type fixalloc struct {
	_, _, _, _ uintptr
	chunk      uintptr
	_          uint32
	inuse      uintptr
	_          uintptr
	_          bool
}
type treapNode struct {
	_, _    uintptr
	parent  *treapNode
	_       uintptr
	spanKey uintptr
}

func enumRuntimeMemory(h *mheap, m memoryMap) {
	for x := h.freeLarge; x != nil; x = x.parent {
		m.add(x.spanKey)
	}
	for i := 1; i < len(gcBitsArenas); i++ {
		m.add(gcBitsArenas[i])
	}
	if len(h.allspans) > 0 {
		for i := range h.allspans {
			if h.allspans[i] != nil {
				m.add(h.allspans[i].startAddr)
			}
		}
		m.add(uintptr(unsafe.Pointer(&h.allspans[0])))
	}
	for i := range h.spans {
		if h.spans[i] != nil {
			m.add(h.spans[i].startAddr)
		}
	}
	if m.add(h.base); h.end > 1 {
		m.add(h.end - 1)
	}
	m.add(h.spanalloc.chunk)
	m.add(h.spanalloc.inuse)
	m.add(h.cachealloc.chunk)
	m.add(h.cachealloc.inuse)
	m.add(h.treapalloc.chunk)
	m.add(h.treapalloc.inuse)
	m.add(h.specialfinalizeralloc.chunk)
	m.add(h.specialfinalizeralloc.inuse)
	m.add(h.specialprofilealloc.chunk)
	m.add(h.specialprofilealloc.inuse)
}
