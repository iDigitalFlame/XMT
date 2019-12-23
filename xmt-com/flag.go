package com

import (
	"fmt"
	"strings"
	"sync"

	data "github.com/iDigitalFlame/xmt/xmt-data"
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
	// FlagIgnore is used to signal to the server or client that this Packet should be dropped and
	// not processed. Can be used for connectivity checks.
	FlagIgnore
	// FlagOneshot is used to signal that the Packet contains information and should not be used to
	// create or re-establish a session.
	FlagOneshot
	// FlagMultiDevice is used to determine if the Multi packet contains Packets with separate device IDs.
	// This is used to speed up processing and allows packets that are all destined for the same host to be
	// batch processed.
	FlagMultiDevice
)

var (
	sbufs = &sync.Pool{
		New: func() interface{} {
			return new(strings.Builder)
		},
	}
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

// ClearFrag clears all Frag and Multi related flags and data.
func (f *Flag) ClearFrag() {
	*f = Flag(uint16(*f)) ^ FlagFrag
}

// FragLen is similar to the Total function. This returns
// the total amount of fragmentented packets to expect.
func (f Flag) FragLen() int {
	return int(f.FragTotal())
}

// String returns a character representation of this Flag
// integer.
func (f Flag) String() string {
	b := sbufs.Get().(*strings.Builder)
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
	if f&FlagIgnore != 0 {
		b.WriteRune('I')
	}
	if f&FlagOneshot != 0 {
		b.WriteRune('O')
	}
	if f&FlagMultiDevice != 0 {
		b.WriteRune('X')
	}
	if b.Len() == 0 {
		b.WriteString(fmt.Sprintf("V%X", int64(f)))
	}
	if f&FlagMulti != 0 {
		b.WriteString(fmt.Sprintf("[%d]", f.FragTotal()))
	} else if f&FlagFrag != 0 {
		b.WriteString(fmt.Sprintf("[%X:%d/%d]", f.FragGroup(), f.FragPosition()+1, f.FragTotal()))
	}
	s := b.String()
	b.Reset()
	sbufs.Put(b)
	return s
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
	*f = Flag((*f>>32)<<32) | Flag(n)<<16 | Flag(uint16(*f)) | FlagFrag
}

// SetFragTotal sets the total count of packets
// in the fragment group.
func (f *Flag) SetFragTotal(n uint16) {
	*f = Flag(n)<<48 | Flag(f.FragPosition())<<32 | Flag(uint32(*f)) | FlagFrag
}

// SetFragPosition sets the position this packet is located
// in the fragment group.
func (f *Flag) SetFragPosition(n uint16) {
	*f = Flag(f.FragTotal())<<48 | Flag(n)<<32 | Flag(uint32(*f)) | FlagFrag
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