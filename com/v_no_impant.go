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

package com

import "github.com/iDigitalFlame/xmt/util"

// String returns a character representation of this Flag.
func (f Flag) String() string {
	var (
		b [27]byte
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
	if f&FlagCrypt != 0 {
		b[n] = 'Z'
		n++
	}
	if n == 0 {
		n += copy(b[:], "V"+util.Uitoa16(uint64(f)))
	}
	switch {
	case f&FlagMulti != 0 && f.Len() > 0:
		n += copy(b[n:], "["+util.Uitoa(uint64(f.Len()))+"]")
	case f&FlagFrag != 0 && f&FlagMulti == 0:
		if f.Len() == 0 {
			n += copy(b[n:], "["+util.Uitoa16(uint64(f.Group()))+"]")
		} else {
			n += copy(b[n:], "["+util.Uitoa16(uint64(f.Group()))+":"+util.Uitoa(uint64(f.Position())+1)+"/"+util.Uitoa(uint64(f.Len()))+"]")
		}
	}
	return string(b[:n])
}
func byteHexStr(b byte) string {
	if b < 16 {
		return util.HexTable[b&0x0F : (b&0x0F)+1]
	}
	return util.HexTable[b>>4:(b>>4)+1] + util.HexTable[b&0x0F:(b&0x0F)+1]
}

// String returns a string descriptor of the Packet struct.
func (p Packet) String() string {
	switch {
	case p.Empty() && p.Flags == 0 && p.Job == 0 && p.ID == 0:
		return "NoP"
	case p.Empty() && p.Flags == 0 && p.Job == 0:
		return "0x" + byteHexStr(p.ID)
	case p.Empty() && p.Flags == 0 && p.ID == 0:
		return "<invalid>NoP/" + util.Uitoa(uint64(p.Job))
	case p.Empty() && p.Flags == 0:
		return "0x" + byteHexStr(p.ID) + "/" + util.Uitoa(uint64(p.Job))
	case p.Empty() && p.Job == 0 && p.ID == 0:
		return p.Flags.String()
	case p.Empty() && p.Job == 0:
		return "0x" + byteHexStr(p.ID) + " " + p.Flags.String()
	case p.Empty() && p.ID == 0:
		return "NoP/" + util.Uitoa(uint64(p.Job)) + " " + p.Flags.String()
	case p.Empty():
		return "0x" + byteHexStr(p.ID) + "/" + util.Uitoa(uint64(p.Job)) + " " + p.Flags.String()
	case p.Flags == 0 && p.Job == 0 && p.ID == 0:
		return "<invalid>NoP: " + util.Uitoa(uint64(p.Size())) + "B"
	case p.Flags == 0 && p.Job == 0:
		return "0x" + byteHexStr(p.ID) + ": " + util.Uitoa(uint64(p.Size())) + "B"
	case p.Flags == 0 && p.ID == 0:
		return "<invalid>NoP/" + util.Uitoa(uint64(p.Job)) + ": " + util.Uitoa(uint64(p.Size())) + "B"
	case p.Flags == 0:
		return "0x" + byteHexStr(p.ID) + "/" + util.Uitoa(uint64(p.Job)) + ": " + util.Uitoa(uint64(p.Size())) + "B"
	case p.Job == 0 && p.ID == 0:
		return p.Flags.String() + ": " + util.Uitoa(uint64(p.Size())) + "B"
	case p.Job == 0:
		return "0x" + byteHexStr(p.ID) + " " + p.Flags.String() + ": " + util.Uitoa(uint64(p.Size())) + "B"
	case p.ID == 0:
		return "<invalid>NoP/" + util.Uitoa(uint64(p.Job)) + " " + p.Flags.String() + ": " + util.Uitoa(uint64(p.Size())) + "B"
	}
	return "0x" + byteHexStr(p.ID) + "/" + util.Uitoa(uint64(p.Job)) + " " + p.Flags.String() + ": " + util.Uitoa(uint64(p.Size())) + "B"
}
func (i *ipListener) String() string {
	return "IP:" + util.Uitoa(uint64(i.proto)) + "/" + i.Addr().String()
}
func (t *tcpListener) String() string {
	return "TCP/" + t.Addr().String()
}
func (l *udpListener) String() string {
	return "UDP/" + l.sock.LocalAddr().String()
}
