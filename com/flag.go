package com

import (
	"strconv"
	"strings"
	"sync"

	"github.com/iDigitalFlame/xmt/data"
)

const (
	// FlagData is a flag used to indicate that Hello/Echo/Sleep packets
	// also include additional data to be read.
	FlagData Flag = 1 << iota
	// FlagFrag is a flag used to indicate that the packet is part of a
	// fragment group (PacketGroup) and the server should re-assemble the packet
	// before preforming actions on it.
	FlagFrag
	// FlagMulti is a flag used to indicate that the packet is a container
	// for multiple packets, auto addded by a processing agent.
	FlagMulti
	// FlagProxy is a flag used to indicate that the packet was sent from another client
	// acting as a forwarding proxy.
	FlagProxy
	// FlagError is a flag used to indicate that the packet indicates that an error
	// condition has occurred. The contents of the Packet can be used to understand the error cause.
	FlagError
	// FlagChannel is a flag used to signify that the connection should be converted into/from a single
	// channel connection. This means that the connection is kept alive and the client will not poll the server.
	// This flag will be present on the top level multi-packet if included in a single packet inside. This flag will
	// take affect on each hop that it goes through. Incompatible with 'FlagOneshot'. Has to be used once per connection.
	FlagChannel
	// FlagOneshot is used to signal that the Packet contains information and should not be used to
	// create or re-establish a session.
	FlagOneshot
	// FlagMultiDevice is used to determine if the Multi packet contains Packets with separate device IDs.
	// This is used to speed up processing and allows packets that are all destined for the same host to be
	// batch processed.
	FlagMultiDevice
)

var stringBuf = sync.Pool{
	New: func() interface{} {
		return new(strings.Builder)
	},
}

// Flag is a bitwise integer that represents important
// information about the packet that its assigned to.
//
// Mapping
// 64        56        48        40        32        24        16         8         0
//  | 8 4 2 1 | 8 4 2 1 | 8 4 2 1 | 8 4 2 1 | 8 4 2 1 | 8 4 2 1 | 8 4 2 1 | 8 4 2 1 |
//  |    Frag Total     |   Frag Position   |   Frag Group ID   |       Flags       |
//  |                         Frag Data                         |                   |
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

// String returns a character representation of this Flag.
func (f Flag) String() string {
	b := stringBuf.Get().(*strings.Builder)
	if f&FlagData != 0 {
		b.WriteRune('D')
	}
	if f&FlagFrag != 0 {
		b.WriteRune('F')
	}
	if f&FlagMulti != 0 {
		b.WriteRune('M')
	}
	if f&FlagProxy != 0 {
		b.WriteRune('P')
	}
	if f&FlagError != 0 {
		b.WriteRune('E')
	}
	if f&FlagChannel != 0 {
		b.WriteRune('C')
	}
	if f&FlagOneshot != 0 {
		b.WriteRune('O')
	}
	if f&FlagMultiDevice != 0 {
		b.WriteRune('X')
	}
	if b.Len() == 0 {
		b.WriteString("V" + strconv.FormatUint(uint64(f), 16))
	}
	if f&FlagMulti != 0 {
		b.WriteString("[" + strconv.Itoa(int(f.Len())) + "]")
	} else if f&FlagFrag != 0 {
		b.WriteString("[" + strconv.FormatUint(uint64(f.Group()), 16) + ":" + strconv.Itoa(int(f.Position()+1)) + "/" + strconv.Itoa(int(f.Len()+1)) + "]")
	}
	s := b.String()
	b.Reset()
	stringBuf.Put(b)
	return s
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
	n, err := r.Uint64()
	if err != nil {
		return err
	}
	*f = Flag(n)
	return nil
}
