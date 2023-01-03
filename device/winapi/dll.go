//go:build windows || (!windows && !implant)

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
	"io"
	"os"
	"strings"
	"unsafe"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

const sectionSize = unsafe.Sizeof(imageSectionHeader{})

type imageNtHeader struct {
	Signature uint32
	File      imageFileHeader
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
	magic uint16
	_     [56]byte
	pos   int32
}
type imageFileHeader struct {
	Machine              uint16
	NumberOfSections     uint16
	_, _, _              uint32
	SizeOfOptionalHeader uint16
	Characteristics      uint16
}
type imageSectionHeader struct {
	Name             [8]uint8
	VirtualSize      uint32
	VirtualAddress   uint32
	SizeOfRawData    uint32
	PointerToRawData uint32
	_, _             uint32
	_, _             uint16
	Characteristics  uint32
}
type imageDataDirectory struct {
	VirtualAddress uint32
	Size           uint32
}
type imageOptionalHeader32 struct {
	_                   [92]byte
	NumberOfRvaAndSizes uint32
	Directory           [16]imageDataDirectory
}
type imageOptionalHeader64 struct {
	_                   [108]byte
	NumberOfRvaAndSizes uint32
	Directory           [16]imageDataDirectory
}

func byteString(b [256]byte) string {
	var n int
	for i := range b {
		if b[i] == 0 {
			break
		}
		n++
	}
	return string(b[:n])
}

// ExtractDLLBase will extract the '.text' (executable) section of the supplied
// DLL file path or basename (Windows-only) and return the '.text' base address
// and raw bytes to be used in calls to 'winapi.Patch*' or 'winapi.Check*'
//
// This function returns any errors that may occur during reading.
//
// Non-Windows devices may use this function to extract DLL data.
func ExtractDLLBase(dll string) (uint32, []byte, error) {
	b, err := os.ReadFile(fullPath(dll))
	if err != nil {
		return 0, nil, err
	}
	return ExtractDLLBaseRaw(b)
}

// ExtractDLLBaseRaw will extract the '.text' (executable) section of the supplied
// DLL raw bytes and return the '.text' base address and raw bytes to be used in
// calls to 'winapi.Patch*' or 'winapi.Check*'
//
// This function returns any errors that may occur during reading.
//
// Non-Windows devices may use this function to extract DLL data.
func ExtractDLLBaseRaw(v []byte) (uint32, []byte, error) {
	_, s, _, b, err := extractDLLBase(v)
	return s.VirtualAddress, b[s.PointerToRawData:s.SizeOfRawData], err
}

// ExtractDLLFunction will extract 'count' bytes from the supplied DLL file path
// or basename (Windows-only) at the base of the supplied function name.
//
// If 'count' is zero, this defaults to 16 bytes.
//
// This function returns any errors that may occur during reading. Forwarded
// functions also return an error that indicates where the forward points to.
//
// Non-Windows devices may use this function to extract DLL data.
func ExtractDLLFunction(dll string, name string, count uint32) ([]byte, error) {
	b, err := os.ReadFile(fullPath(dll))
	if err != nil {
		return nil, err
	}
	return ExtractDLLFunctionRaw(b, name, count)
}

// ExtractDLLFunctionRaw will extract 'count' bytes from the supplied DLL raw bytes
// at the base of the supplied function name.
//
// If 'count' is zero, this defaults to 16 bytes.
//
// This function returns any errors that may occur during reading. Forwarded
// functions also return an error that indicates where the forward points to.
//
// Non-Windows devices may use this function to extract DLL data.
func ExtractDLLFunctionRaw(v []byte, name string, count uint32) ([]byte, error) {
	a, q, e, b, err := extractDLLBase(v)
	if err != nil {
		return nil, err
	}
	u := uint32(len(b))
	if u < a {
		return nil, xerr.Sub("cannot find data section", 0x1D)
	}
	if u < e.PointerToRawData+a+e.VirtualSize+40 {
		return nil, io.ErrUnexpectedEOF
	}
	if count == 0 {
		count = 16
	}
	var (
		i = (*imageExportDir)(unsafe.Pointer(&b[e.PointerToRawData+a]))
		h = e.PointerToRawData - e.VirtualAddress
		f = h + i.AddressOfFunctions
		s = h + i.AddressOfNames
		o = h + i.AddressOfNameOrdinals
		m = q.VirtualAddress + q.VirtualSize
		r = make([]byte, 0, count)
	)
	if u < f || u < s || u < o || u < m {
		return nil, io.ErrUnexpectedEOF
	}
	for x, k, c := uint32(0), "", uint32(0); x < i.NumberOfNames; x++ {
		k = byteString(*(*[256]byte)(unsafe.Pointer(
			&b[h+*(*uint32)(unsafe.Pointer(&b[s+(x*4)]))],
		)))
		if !strings.EqualFold(k, name) {
			continue
		}
		// Grab ASM from '.text' section
		c = (q.PointerToRawData - q.VirtualAddress) + *(*uint32)(unsafe.Pointer(
			&b[f+uint32(*(*uint16)(unsafe.Pointer(&b[o+(x*2)]))*4)],
		))
		if c < m && c > f {
			if xerr.ExtendedInfo {
				return nil, xerr.Sub(`function is a forward to "`+byteString(*(*[256]byte)(unsafe.Pointer(&b[c])))+`"`, 0x70)
			}
			return nil, xerr.Sub("function is a forward", 0x70)
		}
		for z := uint32(0); z < count; z++ {
			r = append(r, b[c+z])
		}
	}
	if len(r) == 0 {
		return nil, xerr.Sub("cannot find function", 0x6F)
	}
	return r, nil
}
func extractDLLBase(b []byte) (uint32, *imageSectionHeader, *imageSectionHeader, []byte, error) {
	if len(b) < 62 {
		return 0, nil, nil, nil, xerr.Sub("base is not a valid DOS header", 0x19)
	}
	var (
		u = int32(len(b))
		d = (*imageDosHeader)(unsafe.Pointer(&b[0]))
	)
	if d.magic != 0x5A4D {
		return 0, nil, nil, nil, xerr.Sub("base is not a valid DOS header", 0x19)
	}
	if u < d.pos+24 {
		return 0, nil, nil, nil, xerr.Sub("offset base is not a valid NT header", 0x1A)
	}
	n := *(*imageNtHeader)(unsafe.Pointer(&b[d.pos]))
	if n.Signature != 0x00004550 {
		return 0, nil, nil, nil, xerr.Sub("offset base is not a valid NT header", 0x1A)
	}
	if n.File.Characteristics&0x2000 == 0 {
		return 0, nil, nil, nil, xerr.Sub("header does not represent a DLL", 0x1B)
	}
	switch n.File.Machine {
	case 0, 0x14C, 0x1C4, 0xAA64, 0x8664:
	default:
		return 0, nil, nil, nil, xerr.Sub("header does not represent a DLL", 0x1B)
	}
	var (
		p = d.pos + int32(unsafe.Sizeof(n))
		v [16]imageDataDirectory
	)
	if u < p+104 {
		return 0, nil, nil, nil, io.ErrUnexpectedEOF
	}
	if *(*uint16)(unsafe.Pointer(&b[p])) == 0x20B {
		v = (*imageOptionalHeader64)(unsafe.Pointer(&b[p])).Directory
	} else {
		v = (*imageOptionalHeader32)(unsafe.Pointer(&b[p])).Directory
	}
	if p = d.pos + int32(unsafe.Sizeof(n.File)) + int32(n.File.SizeOfOptionalHeader) + 4; u < p {
		return 0, nil, nil, nil, io.ErrUnexpectedEOF
	}
	// NOTE(dij): For clarity 's' is our '.text' section, it CAN be our entry
	//            points section, but it might not. 'e' will store the entry
	//            points section.
	var s, e *imageSectionHeader
	for i := uint16(0); i < n.File.NumberOfSections; i++ {
		k := p + (int32(sectionSize) * int32(i))
		if u < k+40 {
			return 0, nil, nil, nil, io.ErrUnexpectedEOF
		}
		x := (*imageSectionHeader)(unsafe.Pointer(&b[k]))
		// Find the '.text' section
		if x.Name[0] == 0x2E && x.Name[1] == 0x74 && x.Name[3] == 0x78 {
			s = x
		}
		// Find the entry point table
		if x.VirtualAddress < v[0].VirtualAddress && v[0].VirtualAddress < (x.VirtualAddress+x.VirtualSize) {
			e = x
		}
		if e != nil && s != nil {
			break
		}
	}
	if s == nil || len(b) < int(s.PointerToRawData) {
		return 0, nil, nil, nil, xerr.Sub("cannot find data section", 0x1D)
	}
	if e == nil {
		e = s // Make sure 'e' is never nil
	}
	return v[0].VirtualAddress - e.VirtualAddress, s, e, b, nil
}
