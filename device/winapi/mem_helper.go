//go:build windows && cgo && freemem
// +build windows,cgo,freemem

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

/*
#include <windows.h>

// Free and release memory. This will remove the base memory allocations from Go.
// This has to be ran in CGo so we can prevent the runtime from panicing when
// we pull the rug (memory) from under it.
//
// This should be called from the last thread in "killRuntime".
//
// This function is passed the address of the "NtFreeVirtualMemory" function and
// a variable array of memory address regions collected from various sources.
void releaseMemory(void *func, int n, PVOID *regions) {
	for (int i = 0; i < n; i++) {
		#ifdef _AMD64_
			SIZE_T s = 0;
			((NTSTATUS (*)(HANDLE, PVOID*, PSIZE_T, ULONG))func)((HANDLE)-1, &(regions[i]), &s, 0x8000);
		#else
			// When we are in WOW64 or in 32bit mode, it seems that NtFreeVirtualMemory
			// seems to cause an 0xC0000005 exception for some reason, but VirtualFree
			// works perfectly fine, so we'll use that.
			VirtualFree(regions[i], 0, 0x8000);
		#endif
	}
	free(regions);
	ExitThread(0);
}
*/
import "C"
import (
	"unsafe"

	"github.com/iDigitalFlame/xmt/util/bugtrack"
)

const memoryBasicInfoSize = unsafe.Sizeof(memoryBasicInfo{})

//go:linkname heapMemory runtime.mheap_
var heapMemory mheap

type memoryMap map[uintptr]struct{}

func freeRuntimeMemory() {
	var (
		h = &heapMemory
		x = *h
		v = memoryBase(uintptr(unsafe.Pointer(&x)))
		m = make(memoryMap, 64)
	)
	enumRuntimeMemory(h, m)
	if m.add(v); bugtrack.Enabled {
		bugtrack.Track("winapi.freeRuntimeMemory(): Found %d runtime memory regions to free.", len(m))
	}
	t := make([]uintptr, 0, len(m))
	for r := range m {
		if bugtrack.Enabled {
			bugtrack.Track("winapi.freeRuntimeMemory(): Free target %d: 0x%X", len(t), r)
		}
		t = append(t, r)
	}
	l := len(m)
	if m = nil; l > 255 {
		l = 255
	}
	var (
		b = C.malloc(C.size_t(l) * C.size_t(ptrSize))
		d = (*[255]uintptr)(b)
	)
	for i := 0; i < l; i++ {
		d[i] = t[i]
	}
	C.releaseMemory(unsafe.Pointer(funcNtFreeVirtualMemory.address()), C.int(l), (*C.PVOID)(b))
}
func (m memoryMap) add(h uintptr) {
	if h == 0 {
		return
	}
	v := memoryBase(h)
	if v == 0 {
		return
	}
	if _, ok := m[v]; ok {
		return
	}
	m[v] = caught
}
func memoryBase(h uintptr) uintptr {
	if h == 0 {
		return 0
	}
	var (
		m       memoryBasicInfo
		s       uint32
		r, _, _ = syscallN(funcNtQueryVirtualMemory.address(), CurrentProcess, h, 0, uintptr(unsafe.Pointer(&m)), memoryBasicInfoSize, uintptr(unsafe.Pointer(&s)))
	)
	if r != 0 {
		return h
	}
	// We can't free mapped or already free memory regions.
	// 0x10000 - MEM_FREE
	// 0x40000 - MEM_MAPPED
	if m.State == 0x10000 || m.Type == 0x40000 {
		return 0
	}
	return uintptr(m.AllocationBase) // return uintptr(m.BaseAddress)
}
