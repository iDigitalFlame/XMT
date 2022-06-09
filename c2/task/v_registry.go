//go:build !implant

// Copyright (C) 2020 - 2022 iDigitalFlame
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
	"github.com/iDigitalFlame/xmt/data"
)

// RegLs returns a list registry keys/values Packet. This can be used to instruct
// the client to return a list of Registry entries for the specified registry
// path.
//
// C2 Details:
//  ID: TvRegistry
//
//  Input:
//      uint8            // Operation
//      string           // Key Path
//  Output:
//      uint8            // Operation
//      uint32           // Count
//      []Entry struct { // List of Entries
//          string       // Name
//          uint32       // Type
//          []byte       // Data
//      }
func RegLs(s string) *com.Packet {
	n := &com.Packet{ID: TvRegistry}
	n.WriteUint8(regOpLs)
	n.WriteString(s)
	return n
}

// RegMakeKey returns a make registry key Packet. This can be used to instruct
// the client to make a key at specified registry path.
//
// C2 Details:
//  ID: TvRegistry
//
//  Input:
//      uint8  // Operation
//      string // Key Path
//  Output:
//      uint8  // Operation
func RegMakeKey(key string) *com.Packet {
	n := &com.Packet{ID: TvRegistry}
	n.WriteUint8(regOpMake)
	n.WriteString(key)
	return n
}

// RegGet returns a get key/value Packet. This can be used to instruct the client
// to return a entry details for the specified registry path.
//
// C2 Details:
//  ID: TvRegistry
//
//  Input:
//      uint8          // Operation
//      string         // Key Path
//      string         // Value Name
//  Output:
//      uint8          // Operation
//      Entry struct { // Entry
//          string     // Name
//          uint32     // Type
//          []byte     // Data
//      }
func RegGet(key, value string) *com.Packet {
	n := &com.Packet{ID: TvRegistry}
	n.WriteUint8(regOpGet)
	n.WriteString(key)
	n.WriteString(value)
	return n
}

// RegSetString returns a set as string key/value Packet. This can be used to
// instruct the client to set the value content to the supplied string for the
// specified registry path.
//
// C2 Details:
//  ID: TvRegistry
//
//  Input:
//      uint8  // Operation
//      string // Key Path
//      string // Value Name
//      string // Content
//  Output:
//      uint8  // Operation
func RegSetString(key, value, v string) *com.Packet {
	n := &com.Packet{ID: TvRegistry}
	n.WriteUint8(regOpSetString)
	n.WriteString(key)
	n.WriteString(value)
	n.WriteString(v)
	return n
}

// RegDeleteKey returns a delete key Packet. This can be used to instruct the
// client to delete a key at the specified registry path.
//
// C2 Details:
//  ID: TvRegistry
//
//  Input:
//      uint8  // Operation
//      string // Key Path
//      bool   // Delete Recursively or Delete non-empty Keys
//  Output:
//      uint8  // Operation
func RegDeleteKey(key string, force bool) *com.Packet {
	n := &com.Packet{ID: TvRegistry}
	n.WriteUint8(regOpDeleteKey)
	n.WriteString(key)
	n.WriteBool(force)
	return n
}

// RegDelete returns a delete key/value Packet. This can be used to instruct the
// client to delete a key or value at the specified registry path.
//
// C2 Details:
//  ID: TvRegistry
//
//  Input:
//      uint8  // Operation
//      string // Key Path
//      string // Value Name
//      bool   // Delete Recursively or Delete non-empty Keys
//  Output:
//      uint8  // Operation
func RegDelete(key, value string, force bool) *com.Packet {
	n := &com.Packet{ID: TvRegistry}
	n.WriteUint8(regOpDelete)
	n.WriteString(key)
	n.WriteString(value)
	n.WriteBool(force)
	return n
}

// RegSetDword returns a set as a DWORD (uint32) key/value Packet. This can be
// used to instruct the client to set the value content to the supplied DWORD
// for the specified registry path.
//
// C2 Details:
//  ID: TvRegistry
//
//  Input:
//      uint8  // Operation
//      string // Key Path
//      string // Value Name
//      uint32 // Content
//  Output:
//      uint8  // Operation
func RegSetDword(key, value string, v uint32) *com.Packet {
	n := &com.Packet{ID: TvRegistry}
	n.WriteUint8(regOpSetDword)
	n.WriteString(key)
	n.WriteString(value)
	n.WriteUint32(v)
	return n
}

// RegSetQword returns a set as QWORD (uint64) key/value Packet. This can be
// used to instruct the client to set the value content to the supplied QWORD
// for the specified registry path.
//
// C2 Details:
//  ID: TvRegistry
//
//  Input:
//      uint8  // Operation
//      string // Key Path
//      string // Value Name
//      uint64 // Content
//  Output:
//      uint8  // Operation
func RegSetQword(key, value string, v uint64) *com.Packet {
	n := &com.Packet{ID: TvRegistry}
	n.WriteUint8(regOpSetQword)
	n.WriteString(key)
	n.WriteString(value)
	n.WriteUint64(v)
	return n
}

// RegSetBytes returns a set as a BINARY (bytes) key/value Packet. This can be
// used to instruct the client to set the value content to the supplied bytes
// for the specified registry path.
//
// C2 Details:
//  ID: TvRegistry
//
//  Input:
//      uint8  // Operation
//      string // Key Path
//      string // Value Name
//      []byte // Content
//  Output:
//      uint8  // Operation
func RegSetBytes(key, value string, b []byte) *com.Packet {
	n := &com.Packet{ID: TvRegistry}
	n.WriteUint8(regOpSetBytes)
	n.WriteString(key)
	n.WriteString(value)
	n.WriteBytes(b)
	return n
}

// RegSetExpandString returns a set as expand string key/value Packet. This can
// be used to instruct the client to set the value content to the supplied
// string for the specified registry path.
//
// C2 Details:
//  ID: TvRegistry
//
//  Input:
//      uint8  // Operation
//      string // Key Path
//      string // Value Name
//      string // Content
//  Output:
//      uint8  // Operation
func RegSetExpandString(key, value, v string) *com.Packet {
	n := &com.Packet{ID: TvRegistry}
	n.WriteUint8(regOpSetExpandString)
	n.WriteString(key)
	n.WriteString(value)
	n.WriteString(v)
	return n
}

// RegSet returns a set content key/value Packet. This can be used to instruct
// the client to set the raw value content to the supplied raw bytes for the
// specified registry path along with the type.
//
// C2 Details:
//  ID: TvRegistry
//
//  Input:
//      uint8   // Operation
//      string  // Key Path
//      string  // Value Name
//      uint32  // Type
//      []byte  // Content
//  Output:
//      uint8  // Operation
func RegSet(key, value string, t uint32, b []byte) *com.Packet {
	n := &com.Packet{ID: TvRegistry}
	n.WriteUint8(regOpSet)
	n.WriteString(key)
	n.WriteString(value)
	n.WriteUint32(t)
	n.WriteBytes(b)
	return n
}

// RegSetStringList returns a set as multi string key/value Packet. This can
// be used to instruct the client to set the value content to the supplied
// strings for the specified registry path.
//
// C2 Details:
//  ID: TvRegistry
//
//  Input:
//      uint8    // Operation
//      string   // Key Path
//      string   // Value Name
//      []string // Content
//  Output:
//      uint8  // Operation
func RegSetStringList(key, value string, v []string) *com.Packet {
	n := &com.Packet{ID: TvRegistry}
	n.WriteUint8(regOpSetStringList)
	n.WriteString(key)
	n.WriteString(value)
	data.WriteStringList(n, v)
	return n
}
