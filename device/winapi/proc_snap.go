//go:build windows && snap

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
