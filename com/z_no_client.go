//go:build !client
// +build !client

package com

import (
	"strconv"

	"github.com/iDigitalFlame/xmt/util"
)

// String returns a character representation of this Flag.
func (f Flag) String() string {
	var (
		b [26]byte
		n int
	)
	if f&FlagFrag != 0 {
		b[n] = 'F'
		n++
	}
	if f&FlagMulti != 0 {
		b[n] = 'M'
		n++
	}
	if f&FlagProxy != 0 {
		b[n] = 'P'
		n++
	}
	if f&FlagError != 0 {
		b[n] = 'E'
		n++
	}
	if f&FlagChannel != 0 {
		b[n] = 'C'
		n++
	}
	if f&FlagChannelEnd != 0 {
		b[n] = 'K'
		n++
	}
	if f&FlagOneshot != 0 {
		b[n] = 'O'
		n++
	}
	if f&FlagMultiDevice != 0 {
		b[n] = 'X'
		n++
	}
	if n == 0 {
		n += copy(b[:], "V"+strconv.FormatUint(uint64(f), 16))
	}
	switch {
	case f&FlagMulti != 0 && f.Len() > 0:
		n += copy(b[n:], "["+strconv.Itoa(int(f.Len()))+"]")
	case f&FlagFrag != 0 && f&FlagMulti == 0:
		if f.Len() == 0 {
			n += copy(b[n:], "["+strconv.FormatUint(uint64(f.Group()), 16)+"]")
		} else {
			n += copy(b[n:], "["+strconv.FormatUint(uint64(f.Group()), 16)+":"+strconv.Itoa(int(f.Position()+1))+"/"+strconv.Itoa(int(f.Len()))+"]")
		}
	}
	return string(b[:n])
}

// String returns a string descriptor of the Packet struct.
func (p *Packet) String() string {
	if p == nil {
		return "<nil>"
	}
	switch {
	case p.Empty() && p.Flags == 0 && p.Job == 0 && p.ID == 0:
		return "NoP"
	case p.Empty() && p.Flags == 0 && p.Job == 0:
		return "0x" + util.ByteHexString(p.ID)
	case p.Empty() && p.Flags == 0 && p.ID == 0:
		return "<invalid>NoP/" + strconv.Itoa(int(p.Job))
	case p.Empty() && p.Flags == 0:
		return "0x" + util.ByteHexString(p.ID) + "/" + strconv.Itoa(int(p.Job))
	case p.Empty() && p.Job == 0 && p.ID == 0:
		return p.Flags.String()
	case p.Empty() && p.Job == 0:
		return "0x" + util.ByteHexString(p.ID) + " " + p.Flags.String()
	case p.Empty() && p.ID == 0:
		return "NoP/" + strconv.Itoa(int(p.Job)) + " " + p.Flags.String()
	case p.Empty():
		return "0x" + util.ByteHexString(p.ID) + "/" + strconv.Itoa(int(p.Job)) + " " + p.Flags.String()
	case p.Flags == 0 && p.Job == 0 && p.ID == 0:
		return "<invalid>NoP: " + strconv.Itoa(p.Size()) + "B"
	case p.Flags == 0 && p.Job == 0:
		return "0x" + util.ByteHexString(p.ID) + ": " + strconv.Itoa(p.Size()) + "B"
	case p.Flags == 0 && p.ID == 0:
		return "<invalid>NoP/" + strconv.Itoa(int(p.Job)) + ": " + strconv.Itoa(p.Size()) + "B"
	case p.Flags == 0:
		return "0x" + util.ByteHexString(p.ID) + "/" + strconv.Itoa(int(p.Job)) + ": " + strconv.Itoa(p.Size()) + "B"
	case p.Job == 0 && p.ID == 0:
		return p.Flags.String() + ": " + strconv.Itoa(p.Size()) + "B"
	case p.Job == 0:
		return "0x" + util.ByteHexString(p.ID) + " " + p.Flags.String() + ": " + strconv.Itoa(p.Size()) + "B"
	case p.ID == 0:
		return "<invalid>NoP/" + strconv.Itoa(int(p.Job)) + " " + p.Flags.String() + ": " + strconv.Itoa(p.Size()) + "B"
	}
	return "0x" + util.ByteHexString(p.ID) + "/" + strconv.Itoa(int(p.Job)) + " " + p.Flags.String() + ": " + strconv.Itoa(p.Size()) + "B"
}
