//go:build windows && snap
// +build windows,snap

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

// ThreadEntry32 matches the THREADENTRY32 struct
//
//	https://docs.microsoft.com/en-us/windows/win32/api/tlhelp32/ns-tlhelp32-threadentry32
//
//	typedef struct tagTHREADENTRY32 {
//	  DWORD dwSize;
//	  DWORD cntUsage;
//	  DWORD th32ThreadID;
//	  DWORD th32OwnerProcessID;
//	  LONG  tpBasePri;
//	  LONG  tpDeltaPri;
//	  DWORD dwFlags;
//	} THREADENTRY32;
//
// DO NOT REORDER
type ThreadEntry32 struct {
	Size           uint32
	Usage          uint32
	ThreadID       uint32
	OwnerProcessID uint32
	BasePri        int32
	DeltaPri       int32
	Flags          uint32
}

// ProcessEntry32 matches the PROCESSENTRY32 struct
//
//	https://docs.microsoft.com/en-us/windows/win32/api/tlhelp32/ns-tlhelp32-processentry32
//
//	typedef struct tagPROCESSENTRY32 {
//	  DWORD     dwSize;
//	  DWORD     cntUsage;
//	  DWORD     th32ProcessID;
//	  ULONG_PTR th32DefaultHeapID;
//	  DWORD     th32ModuleID;
//	  DWORD     cntThreads;
//	  DWORD     th32ParentProcessID;
//	  LONG      pcPriClassBase;
//	  DWORD     dwFlags;
//	  CHAR      szExeFile[MAX_PATH];
//	} PROCESSENTRY32;
//
// DO NOT REORDER
type ProcessEntry32 struct {
	Size            uint32
	Usage           uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	Threads         uint32
	ParentProcessID uint32
	PriClassBase    int32
	Flags           uint32
	ExeFile         [260]uint16
}

// EnumProcesses attempts to reterive the list of currently running Processes
// and will call the supplied function with an entry for each Process.
//
// The user supplied function can return an error that if non-nil, will stop
// Process iteration immediately and will be returned by this function.
//
// Callers can return the special 'winapi.ErrNoMoreFiles' error that will stop
// iteration but will cause this function to return nil. This can be used to
// stop iteration without errors if needed.
//
// This function is affected by the 'snap' buildtag, which if supplied will use
// the 'CreateToolhelp32Snapshot' API function instead of the default
// 'NtQuerySystemInformation' API function.
func EnumProcesses(f func(ProcessEntry) error) error {
	// 0x2 - TH32CS_SNAPPROCESS
	h, err := CreateToolhelp32Snapshot(0x2, 0)
	if err != nil {
		return err
	}
	var e ProcessEntry32
	e.Size = uint32(unsafe.Sizeof(e))
	for err = Process32First(h, &e); err == nil; err = Process32Next(h, &e) {
		err = f(ProcessEntry{
			Name:    UTF16ToString(e.ExeFile[:]),
			PID:     e.ProcessID,
			PPID:    e.ParentProcessID,
			Threads: e.Threads,
			session: -1,
		})
		if err != nil {
			break
		}
	}
	if CloseHandle(h); err != nil && err != ErrNoMoreFiles {
		return err
	}
	return nil
}

// Thread32Next Windows API Call
//
//	Retrieves information about the next thread of any process encountered in
//	the system memory snapshot.
//
// https://docs.microsoft.com/en-us/windows/win32/api/tlhelp32/nf-tlhelp32-thread32next
//
// NOTE: This function is only avaliable if the "snap" build tag is used. To work
// around this restriction, use the higher level 'EnumThreads' function.
func Thread32Next(h uintptr, e *ThreadEntry32) error {
	r, _, err := syscallN(funcThread32Next.address(), h, uintptr(unsafe.Pointer(e)))
	if r == 0 {
		return unboxError(err)
	}
	return nil
}

// Thread32First Windows API Call
//
//	Retrieves information about the first thread of any process encountered in
//	a system snapshot.
//
// https://docs.microsoft.com/en-us/windows/win32/api/tlhelp32/nf-tlhelp32-thread32first
//
// NOTE: This function is only avaliable if the "snap" build tag is used. To work
// around this restriction, use the higher level 'EnumThreads' function.
func Thread32First(h uintptr, e *ThreadEntry32) error {
	r, _, err := syscallN(funcThread32First.address(), h, uintptr(unsafe.Pointer(e)))
	if r == 0 {
		return unboxError(err)
	}
	return nil
}

// Process32Next Windows API Call
//
//	Retrieves information about the next process recorded in a system snapshot.
//
// https://docs.microsoft.com/en-us/windows/win32/api/tlhelp32/nf-tlhelp32-process32nextw
//
// NOTE: This function is only avaliable if the "snap" build tag is used. To work
// around this restriction, use the higher level 'EnumProcesses' function.
func Process32Next(h uintptr, e *ProcessEntry32) error {
	r, _, err := syscallN(funcProcess32Next.address(), h, uintptr(unsafe.Pointer(e)))
	if r == 0 {
		return unboxError(err)
	}
	return nil
}

// Process32First Windows API Call
//
//	Retrieves information about the next process recorded in a system snapshot.
//
// https://docs.microsoft.com/en-us/windows/win32/api/tlhelp32/nf-tlhelp32-process32next
//
// NOTE: This function is only avaliable if the "snap" build tag is used. To work
// around this restriction, use the higher level 'EnumProcesses' function.
func Process32First(h uintptr, e *ProcessEntry32) error {
	r, _, err := syscallN(funcProcess32First.address(), h, uintptr(unsafe.Pointer(e)))
	if r == 0 {
		return unboxError(err)
	}
	return nil
}

// EnumThreads attempts to reterive the list of currently running Process Threads
// and will call the supplied function with an entry for each Thread that matches
// the supplied Process ID.
//
// The user supplied function can return an error that if non-nil, will stop
// Thread iteration immediately and will be returned by this function.
//
// Callers can return the special 'winapi.ErrNoMoreFiles' error that will stop
// iteration but will cause this function to return nil. This can be used to
// stop iteration without errors if needed.
//
// This function is affected by the 'snap' buildtag, which if supplied will use
// the 'CreateToolhelp32Snapshot' API function instead of the default
// 'NtQuerySystemInformation' API function.
func EnumThreads(pid uint32, f func(ThreadEntry) error) error {
	// 0x4 - TH32CS_SNAPTHREAD
	h, err := CreateToolhelp32Snapshot(0x4, 0)
	if err != nil {
		return err
	}
	var t ThreadEntry32
	t.Size = uint32(unsafe.Sizeof(t))
	for err = Thread32First(h, &t); err == nil; err = Thread32Next(h, &t) {
		if t.OwnerProcessID != pid {
			continue
		}
		if err = f(ThreadEntry{TID: t.ThreadID, PID: t.OwnerProcessID}); err != nil {
			break
		}
	}
	if CloseHandle(h); err != nil && err != ErrNoMoreFiles {
		return err
	}
	return nil
}

// CreateToolhelp32Snapshot Windows API Call
//
//	Takes a snapshot of the specified processes, as well as the heaps, modules,
//	and threads used by these processes.
//
// https://docs.microsoft.com/en-us/windows/win32/api/tlhelp32/nf-tlhelp32-createtoolhelp32snapshot
//
// NOTE: This function is only avaliable if the "snap" build tag is used. To work
// around this restriction, use the higher level 'EnumProcesses' or 'EnumThreads'
// functions.
func CreateToolhelp32Snapshot(flags, pid uint32) (uintptr, error) {
	r, _, err := syscallN(funcCreateToolhelp32Snapshot.address(), uintptr(flags), uintptr(pid))
	if r == invalid {
		return 0, unboxError(err)
	}
	return r, nil
}
