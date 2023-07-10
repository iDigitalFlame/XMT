//go:build windows && cgo && go1.15 && !go1.16
// +build windows,cgo,go1.15,!go1.16

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
	summarySize = 4 + (1 * x64)
	chunkL1Bits = 1 << (13 * x64)
	chunkL2Bits = 1 << ((x64*48 + (1-x64)*32) - 22 - (13 * x64))
	arenaL1Bits = 1 << (6 * x64)
	arenaL2Bits = 1 << ((x64*48 + (1-x64)*32) - 22 - (6 * x64))
)

//go:linkname gcBitsArenas runtime.gcBitsArenas
var gcBitsArenas [4]uintptr

type mheap struct {
	_     uintptr
	pages struct {
		summary [summarySize][]uint64
		chunks  [chunkL1Bits]*[chunkL2Bits][16]uint64
		_       uintptr
		_, _    uint
		inUse   struct {
			ranges [][2]uintptr
			_, _   uintptr
		}
		scav struct {
			_ struct {
				_    [][2]uintptr
				_, _ uintptr
			}
			_          uint32
			_, _, _, _ uintptr
		}
		_, _ uintptr
		_    bool
	}
	_, _, _  uint32
	allspans []*mspan
	_        [2]struct {
		_, _, _, _ uintptr
		_          uint32
	}
	_              uint32
	_, _, _, _     uint64
	_              float64
	_, _           uint64
	_              uintptr
	_, _, _, _     uint64
	_              [67]uint64
	arenas         [arenaL1Bits]*[arenaL2Bits]uintptr
	heapArenaAlloc linearAlloc
	arenaHints     *arenaHint
	area           linearAlloc
	_, _, _        []uint
	base           uintptr
	_              uintptr
	_              [134]struct {
		_          uintptr
		_          uint8
		_, _, _, _ uintptr
		_          [4]struct {
			_, _, _, _ uintptr
			_          uint64
		}
		_ uint64
		_ [cacheLineSize - ((5*ptrSize)+9+(2*((4*ptrSize)+8)))%cacheLineSize]byte
	}
	spanalloc             fixalloc
	cachealloc            fixalloc
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
	for i := range h.pages.summary {
		m.add(uintptr(unsafe.Pointer(&h.pages.summary[i][0])))
	}
	for i := range h.pages.chunks {
		if h.pages.chunks[i] == nil {
			continue
		}
		m.add(uintptr(unsafe.Pointer(&h.pages.chunks[i])))
	}
	for i := range h.pages.inUse.ranges {
		if h.pages.inUse.ranges[i][0] == 0 {
			continue
		}
		m.add(h.pages.inUse.ranges[i][0])
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
	m.add(h.base)
	m.add(h.spanalloc.chunk)
	m.add(h.spanalloc.inuse)
	m.add(h.cachealloc.chunk)
	m.add(h.cachealloc.inuse)
	m.add(h.specialfinalizeralloc.chunk)
	m.add(h.specialfinalizeralloc.inuse)
	m.add(h.specialprofilealloc.chunk)
	m.add(h.specialprofilealloc.inuse)
	m.add(h.arenaHintAlloc.chunk)
	m.add(h.arenaHintAlloc.inuse)
}
