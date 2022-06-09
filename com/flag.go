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

package com

import "github.com/iDigitalFlame/xmt/data"

const (
	// FlagFrag is a flag used to indicate that the packet is part of a
	// fragment group and the server should re-assemble the packet before
	// preforming actions on it.
	FlagFrag = 1 << iota
	// FlagMulti is a flag used to indicate that the packet is a container
	// for multiple packets, auto added by a processing agent. This Flag also
	// carries the 'FlagFrag' flag.
	FlagMulti
	// FlagProxy is a flag used to indicate that the packet was sent from another
	// client acting as a forwarding proxy.
	FlagProxy
	// FlagError is a flag used to indicate that the packet indicates that an error
	// condition has occurred. The contents of the Packet can be used to
	// understand the error cause.
	FlagError
	// FlagChannel is a flag used to signify that the connection should be converted
	// into/from a single channel connection. This means that the connection is
	// kept alive and the client will not poll the server.
	//
	// This flag will be present on the top level multi-packet if included in a
	// single packet inside. This flag will take affect on each hop that it goes
	// through.
	//
	// Incompatible with 'FlagOneshot'. Can only be used once per single connection.
	FlagChannel
	// FlagChannelEnd is a flag used to signify that a Channel connection should
	// be terminated. Unlike the 'FlagChannel' option, this will only affect the
	// targeted hop.
	//
	// Incompatible with 'FlagOneshot'. Can only be used once per single connection.
	FlagChannelEnd
	// FlagOneshot is used to signal that the Packet contains information and
	// should not be used to create or re-establish a session.
	FlagOneshot
	// FlagMultiDevice is used to determine if the Multi packet contains Packets
	// with separate device IDs. This is used to speed up processing and allows
	// packets that are all destined for the same host to be batch processed.
	FlagMultiDevice
	// FlagCrypt is used to indicate that the Packet is carrying Crypt related
	// information or a side of the conversation is asking for a re-key.
	FlagCrypt
)

// Flag is a bitwise integer that represents important
// information about the packet that its assigned to.
//
// Mapping
//
//  64        56        48        40        32        24        16         8         0
//   | 8 4 2 1 | 8 4 2 1 | 8 4 2 1 | 8 4 2 1 | 8 4 2 1 | 8 4 2 1 | 8 4 2 1 | 8 4 2 1 |
//   |    Frag Total     |   Frag Position   |   Frag Group ID   |       Flags       |
//   |                         Frag Data                         |                   |
//
type Flag uint64

// Clear clears all Frag and Multi related data values.
func (f *Flag) Clear() {
	*f = Flag(uint16(*f)) ^ FlagFrag
}

// Set appends the Flag value to this current Flag value.
func (f *Flag) Set(n Flag) {
	*f = *f | n
}

// Len returns the count of fragmented packets that make up this fragment group.
func (f Flag) Len() uint16 {
	return uint16(f >> 48)
}

// Unset removes the Flag value to this current Flag value.
func (f *Flag) Unset(n Flag) {
	*f = *f &^ n
}

// Group returns the fragment group ID that this packet is part of.
func (f Flag) Group() uint16 {
	return uint16(f >> 16)
}

// Position represents position of this packet in a fragment group.
func (f Flag) Position() uint16 {
	return uint16(f >> 32)
}

// SetLen sets the total count of packets in the fragment group.
func (f *Flag) SetLen(n uint16) {
	*f = Flag(n)<<48 | Flag(f.Position())<<32 | Flag(uint32(*f)) | FlagFrag
}

// SetGroup sets the group ID of the fragment group this packet is part of.
func (f *Flag) SetGroup(n uint16) {
	*f = Flag((*f>>32)<<32) | Flag(n)<<16 | Flag(uint16(*f)) | FlagFrag
}

// SetPosition sets the position this packet is located in the fragment group.
func (f *Flag) SetPosition(n uint16) {
	*f = Flag(f.Len())<<48 | Flag(n)<<32 | Flag(uint32(*f)) | FlagFrag
}

// MarshalStream writes the data of this Flag to the supplied Writer.
func (f Flag) MarshalStream(w data.Writer) error {
	return w.WriteUint64(uint64(f))
}

// UnmarshalStream reads the data of this Flag from the supplied Reader.
func (f *Flag) UnmarshalStream(r data.Reader) error {
	return r.ReadUint64((*uint64)(f))
}
