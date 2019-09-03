package com

import (
	"github.com/iDigitalFlame/xmt/xmt/data"
)

const (
	// FlagData is a flag used to indicate that Hello/Echo/Sleep packets
	// also include additional data to be read.
	FlagData Flag = 1 << iota
	// FlagFrag is a flag used to indicate that the packet is part of a
	// fragment group (PacketGroup) and the server should re-assemble the Packet
	// before preforming actions on it.
	FlagFrag
	// FlagMulti is a flag used to indicate that the packet is a container
	// for multiple packets, auto addded by a processing agent.
	FlagMulti
	FlagProxy
	FlagIgnore
	FlagOneshot
	FlagMultiDevice
)

// Flag is a bitwise integer that represents important
// information about the packet that its assigned to.
//
// Mapping
// 64        56        48        40        32        24        16         8         0
//  | 8 4 2 1 | 8 4 2 1 | 8 4 2 1 | 8 4 2 1 | 8 4 2 1 | 8 4 2 1 | 8 4 2 1 | 8 4 2 1 |
//  |    Frag Total     |   Frag Position   |   Frag Group ID   |       Flags       |
//  |                         Frag Data                         |                   |
type Flag uint64

// Add appends the Flag to this current
// flag value using the Bitwise OR operator.
func (f *Flag) Add(n Flag) {
	*f = *f | n
}

// FragLen is similar to the Total function. This returns
// the total amount of fragmentented packets to expect.
func (f Flag) FragLen() int {
	return int(f.FragTotal())
}

// FragGroup returns the fragment group ID that this packet
// is part of.
func (f Flag) FragGroup() uint16 {
	return uint16(f >> 16)
}

// FragTotal returns the count of fragmented packets
// that make up this fragment group.
func (f Flag) FragTotal() uint16 {
	return uint16(f >> 48)
}

// FragPosition represents position of this packet in a fragment group.
func (f Flag) FragPosition() uint16 {
	return uint16(f >> 32)
}

// SetFragGroup sets the group ID of the fragment group this packet
// is part of.
func (f *Flag) SetFragGroup(n uint16) {
	*f = Flag(*f<<32) | Flag(n)<<16 | Flag(uint16(*f))
}

// SetFragTotal sets the total count of packets
// in the fragment group.
func (f *Flag) SetFragTotal(n uint16) {
	*f = Flag(n)<<48 | Flag(f.FragPosition())<<32 | Flag(uint32(*f))
}

// SetFragPosition sets the position this packet is located
// in the fragment group.
func (f *Flag) SetFragPosition(n uint16) {
	*f = Flag(f.FragTotal())<<48 | Flag(n)<<32 | Flag(uint32(*f))
}

// MarshalStream writes the data of this Flag to the supplied Writer.
func (f *Flag) MarshalStream(w data.Writer) error {
	return w.WriteUint64(uint64(*f))
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
