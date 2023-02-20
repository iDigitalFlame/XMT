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

package regedit

import (
	"encoding/hex"
	"unsafe"

	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/device/winapi/registry"
	"github.com/iDigitalFlame/xmt/util"
)

// String returns the string representation of the data held in the Data buffer.
// Invalid values of keys return an empty string.
func (e Entry) String() string {
	if e.Type == 0 {
		if len(e.Name) == 0 {
			return "<invalid>"
		}
		return ""
	}
	switch e.Type {
	case registry.TypeDword:
		if len(e.Data) != 4 {
			return ""
		}
		_ = e.Data[3]
		return util.Uitoa(uint64(e.Data[0]) | uint64(e.Data[1])<<8 | uint64(e.Data[2])<<16 | uint64(e.Data[3])<<24)
	case registry.TypeQword:
		if len(e.Data) != 8 {
			return ""
		}
		_ = e.Data[7]
		return util.Uitoa(
			uint64(e.Data[0]) | uint64(e.Data[1])<<8 | uint64(e.Data[2])<<16 | uint64(e.Data[3])<<24 |
				uint64(e.Data[4])<<32 | uint64(e.Data[5])<<40 | uint64(e.Data[6])<<48 | uint64(e.Data[7])<<56,
		)
	case registry.TypeBinary:
		return hex.EncodeToString(e.Data)
	case registry.TypeStringList:
		if len(e.Data) < 3 {
			return ""
		}
		var (
			b util.Builder
			v = (*[1 << 29]uint16)(unsafe.Pointer(&e.Data[0]))[: len(e.Data)/2 : len(e.Data)/2]
		)
		if len(v) == 0 {
			return ""
		}
		if v[len(v)-1] == 0 {
			v = v[:len(v)-1]
		}
		for i, n := 0, 0; i < len(v); i++ {
			if v[i] > 0 {
				continue
			}
			if n > 0 {
				b.WriteByte(',')
				b.WriteByte(' ')
			}
			b.WriteString(string(winapi.UTF16Decode(v[n:i])))
			n = i + 1
		}
		return b.Output()
	case registry.TypeString, registry.TypeExpandString:
		if len(e.Data) < 3 {
			return ""
		}
		return winapi.UTF16ToString((*[1 << 29]uint16)(unsafe.Pointer(&e.Data[0]))[: len(e.Data)/2 : len(e.Data)/2])
	}
	return ""
}

// TypeName returns a string representation of the Type value, which represents
// the value data type.
func (e Entry) TypeName() string {
	if e.Type == 0 {
		return "KEY"
	}
	switch e.Type {
	case registry.TypeDword:
		return "DWORD"
	case registry.TypeQword:
		return "QWORD"
	case registry.TypeBinary:
		return "BINARY"
	case registry.TypeStringList:
		return "MULTI_STRING"
	case registry.TypeString, registry.TypeExpandString:
		return "STRING"
	}
	return ""
}
