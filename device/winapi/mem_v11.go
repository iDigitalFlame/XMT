//go:build windows && cgo && go1.11 && !go1.12
// +build windows,cgo,go1.11,!go1.12

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

const (
	x64         = 1 << (^uintptr(0) >> 63) / 2
	arenaL1Bits = 1 << (6 * x64)
	arenaL2Bits = 1 << ((x64*48 + (1-x64)*32) - 22 - (6 * x64))
)

//go:linkname gcBitsArenas runtime.gcBitsArenas
var gcBitsArenas [4]uintptr

type mheap struct {
	_         uintptr
	_         [128][2]uintptr
	freeLarge *treapNode
	_         [128][2]uintptr
	_, _      uintptr
	_, _, _   uint32
	allspans  []*mspan
	_         [2]struct {
		_, _, _, _ uintptr
		_          uint32
	}
	_, _, _, _     uint64
	_              float64
	_, _, _, _     uint64
	_              [67]uint64
	arenas         [arenaL1Bits]*[arenaL2Bits]uintptr
	heapArenaAlloc linearAlloc
	arenaHints     *arenaHint
	area           linearAlloc
	_              [134]struct {
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
	_                     uintptr
	arenaHintAlloc        fixalloc
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
type arenaHint struct {
	addr uintptr
	_    bool
	next *arenaHint
}
type linearAlloc struct {
	next      uintptr
	mapped, _ uintptr
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
	for i := range h.arenas {
		if h.arenas[i] == nil {
			continue
		}
		if m.add(uintptr(unsafe.Pointer(h.arenas[i]))); x64 == 0 {
			continue
		}
		for z := range h.arenas[i] {
			if h.arenas[i][z] == 0 {
				continue
			}
			m.add(uintptr(unsafe.Pointer(h.arenas[i][z])))
		}
	}
	if m.add(h.area.next); h.area.mapped > 2 {
		m.add(h.area.mapped - 2)
	}
	if m.add(h.heapArenaAlloc.next); h.heapArenaAlloc.mapped > 2 {
		m.add(h.heapArenaAlloc.mapped - 2)
	}
	for x := h.arenaHints; x != nil; x = x.next {
		m.add(x.addr)
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
	m.add(h.arenaHintAlloc.chunk)
	m.add(h.arenaHintAlloc.inuse)
}
