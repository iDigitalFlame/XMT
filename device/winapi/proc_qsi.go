//go:build windows && !snap
// +build windows,!snap

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
	"unsafe"
)

const (
	procSize   = unsafe.Sizeof(procInfo{})
	threadSize = unsafe.Sizeof(threadInfo{})
)

type procInfo struct {
	// DO NOT REORDER
	NextEntryOffset              uint32
	NumberOfThreads              uint32
	_                            [6]int64
	ImageName                    lsaString
	_                            int32
	UniqueProcessID              uintptr
	InheritedFromUniqueProcessID uintptr
	_                            uint32
	SessionID                    uint32
	_                            [(ptrSize * 13) + 48]byte
}
type threadInfo struct {
	// DO NOT REORDER
	_            [28]byte
	StartAddress uintptr
	ClientID     clientID
	_            [12]byte
	ThreadState  uint32
	WaitReason   uint32
	_            uint32
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
	var (
		s       uint64
		r, _, _ = syscallN(funcNtQuerySystemInformation.address(), 0x5, 0, 0, uintptr(unsafe.Pointer(&s)))
	)
	if s == 0 {
		return formatNtError(r)
	}
	// NOTE(dij): Doubling this to ensure we get the correct amount.
	b := make([]byte, s*2)
	if r, _, _ = syscallN(funcNtQuerySystemInformation.address(), 0x5, uintptr(unsafe.Pointer(&b[0])), uintptr(s), uintptr(unsafe.Pointer(&s))); r > 0 {
		return formatNtError(r)
	}
	var err error
	for x, i := (*procInfo)(unsafe.Pointer(&b[0])), uint32(0); ; x = (*procInfo)(unsafe.Pointer(&b[i])) {
		if x.UniqueProcessID == 0 {
			i += x.NextEntryOffset
			continue
		}
		err = f(ProcessEntry{
			PID:     uint32(x.UniqueProcessID),
			PPID:    uint32(x.InheritedFromUniqueProcessID),
			Name:    UTF16PtrToString(x.ImageName.Buffer),
			Threads: x.NumberOfThreads,
			session: int32(x.SessionID),
		})
		if i += x.NextEntryOffset; err != nil {
			break
		}
		if x.NextEntryOffset == 0 {
			break
		}
	}
	if err == ErrNoMoreFiles {
		return nil
	}
	return err
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
	var (
		s       uint64
		r, _, _ = syscallN(funcNtQuerySystemInformation.address(), 0x5, 0, 0, uintptr(unsafe.Pointer(&s)))
	)
	if s == 0 {
		return formatNtError(r)
	}
	// NOTE(dij): Doubling this to ensure we get the correct amount.
	b := make([]byte, s*2)
	if r, _, _ = syscallN(funcNtQuerySystemInformation.address(), 0x5, uintptr(unsafe.Pointer(&b[0])), uintptr(s), uintptr(unsafe.Pointer(&s))); r > 0 {
		return formatNtError(r)
	}
	var err error
outer:
	for x, i := (*procInfo)(unsafe.Pointer(&b[0])), uint32(0); ; x = (*procInfo)(unsafe.Pointer(&b[i])) {
		if uint32(x.UniqueProcessID) != pid {
			if i += x.NextEntryOffset; x.NextEntryOffset == 0 {
				break
			}
			continue
		}
		for z, n := i+uint32(procSize), uint32(0); n < x.NumberOfThreads; n++ {
			var (
				v = (*threadInfo)(unsafe.Pointer(&b[z+(n*uint32(threadSize))]))
				t = ThreadEntry{TID: uint32(v.ClientID.Thread), PID: uint32(v.ClientID.Process), sus: 1}
			)
			if v.ThreadState == 5 && v.WaitReason == 5 {
				t.sus = 2
			}
			if err = f(t); err != nil {
				break outer
			}
		}
		break // Should be fine doing this
	}
	if err == ErrNoMoreFiles {
		return nil
	}
	return err
}
