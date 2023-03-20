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

import (
	"io"

	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const (
	// PacketMaxTags is the max amount of tags that are allowed on a specific
	// Packet. If the amount of tags exceed this limit, an error will occur
	// doing writing.
	PacketMaxTags = 2 << 14
	// PacketHeaderSize is the length of the Packet header in bytes.
	PacketHeaderSize = 46
)

// ErrMalformedTag is an error returned when a read on a Packet Tag returns
// an empty (zero) tag value.
var ErrMalformedTag = xerr.Sub("malformed Tag", 0x2A)

// Packet is a struct that is a Reader and Writer that can be generated to be
// sent, or received from a Connection.
//
// Acts as a data buffer and 'parent' of 'data.Chunk'.
type Packet struct {
	Tags []uint32
	data.Chunk

	Flags Flag
	Job   uint16

	Device device.ID
	ID     uint8
	len    uint64
}

// Size returns the amount of bytes written or contained in this Packet with the
// header size added.
func (p *Packet) Size() int {
	if p.Empty() {
		return PacketHeaderSize
	}
	switch s := uint64(p.Chunk.Size() + PacketHeaderSize + (4 * len(p.Tags))); {
	case s < data.LimitSmall:
		return int(s) + 1
	case s < data.LimitMedium:
		return int(s) + 2
	case s < data.LimitLarge:
		return int(s) + 4
	default:
		return int(s) + 8
	}
}

// Add attempts to combine the data and properties the supplied Packet with the
// existing Packet. This function will return an error if the ID's have a
// mismatch or there was an error during the write operation.
func (p *Packet) Add(n *Packet) error {
	if n == nil || n.Empty() {
		return nil
	}
	if p.ID != n.ID {
		return xerr.Sub("packet ID does not match the supplied ID", 0x2C)
	}
	if _, err := n.WriteTo(p); err != nil {
		return xerr.Wrap("unable to write to Packet", err)
	}
	// NOTE(dij): Preserve frag flags.
	p.Flags |= Flag(uint16(n.Flags))
	return nil
}

// Belongs returns true if the specified Packet is a Frag that was a part of the
// split Chunks of this as the original packet.
func (p *Packet) Belongs(n *Packet) bool {
	return n != nil && p.Flags >= FlagFrag && n.Flags >= FlagFrag && p.ID == n.ID && p.Job == n.Job && p.Flags.Group() == n.Flags.Group()
}

// Marshal will attempt to write this Packet's data and headers to the specified
// Writer. This function will return any errors that have occurred during writing.
func (p *Packet) Marshal(w io.Writer) error {
	if err := p.writeHeader(w); err != nil {
		return xerr.Wrap("marshal header", err)
	}
	if err := p.writeBody(w); err != nil {
		return xerr.Wrap("marshal body", err)
	}
	return nil
}
func (p *Packet) readBody(r io.Reader) error {
	if len(p.Tags) > 0 {
		if bugtrack.Enabled {
			bugtrack.Track("com.(*Packet).readBody(): len(p.Tags)=%d", len(p.Tags))
		}
		var b [4]byte
		for i := range p.Tags {
			n, err := io.ReadFull(r, b[:])
			if err != nil {
				return err
			}
			if n != 4 {
				return io.ErrUnexpectedEOF
			}
			if p.Tags[i] = uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24; p.Tags[i] == 0 {
				return ErrMalformedTag
			}
		}
	}
	if bugtrack.Enabled {
		bugtrack.Track("com.(*Packet).readBody(): p.len=%d", p.len)
	}
	if p.len == 0 {
		return nil
	}
	p.Limit = int(p.len)
	var (
		t   uint64
		err error
	)
	for n := int64(0); t < p.len && err == nil; {
		n, err = p.ReadFrom(r)
		if t += uint64(n); err != nil || n == 0 {
			break
		}
	}
	if p.Limit = 0; bugtrack.Enabled {
		bugtrack.Track("com.(*Packet).readBody(): p.len=%d, t=%d, err=%s", p.len, t, err)
	}
	if t < p.len {
		return io.ErrUnexpectedEOF
	}
	if t == p.len {
		err = nil
	}
	p.len = 0
	return err
}

// Unmarshal will attempt to read Packet data and headers from the specified Reader.
// This function will return any errors that have occurred during reading.
func (p *Packet) Unmarshal(r io.Reader) error {
	if err := p.readHeader(r); err != nil {
		return xerr.Wrap("unmarshal header", err)
	}
	if err := p.readBody(r); err != nil {
		return xerr.Wrap("unmarshal body", err)
	}
	return nil
}
func (p *Packet) writeBody(w io.Writer) error {
	if len(p.Tags) > 0 {
		if bugtrack.Enabled {
			bugtrack.Track("com.(*Packet).writeBody(): len(p.Tags)=%d", len(p.Tags))
		}
		var b [4]byte
		for _, t := range p.Tags {
			if t == 0 {
				return ErrMalformedTag
			}
			b[0], b[1], b[2], b[3] = byte(t>>24), byte(t>>16), byte(t>>8), byte(t)
			n, err := w.Write(b[0:4])
			if err != nil {
				return err
			}
			if n != 4 {
				return io.ErrShortWrite
			}
		}
	}
	if p.Seek(0, 0); p.Chunk.Size() == 0 {
		return nil
	}
	n, err := p.WriteTo(w)
	if err != nil {
		return err
	}
	if n != int64(p.Chunk.Size()) {
		return io.ErrShortWrite
	}
	if bugtrack.Enabled {
		bugtrack.Track("com.(*Packet).writeBody(): p.Chunk.Size()=%d, n=%d, err=%s", p.Chunk.Size(), n, err)
	}
	return nil
}
func (p *Packet) readHeader(r io.Reader) error {
	if err := p.Device.Read(r); err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("com.(*Packet).readHeader(): Read Device failed err=%s", err)
		}
		return err
	}
	var (
		b      [14]byte
		n, err = io.ReadFull(r, b[:])
	)
	if bugtrack.Enabled {
		bugtrack.Track("com.(*Packet).readHeader(): n=%d, err=%s", n, err)
	}
	if n != 14 {
		if err != nil {
			return err
		}
		return io.ErrUnexpectedEOF
	}
	_ = b[13]
	if bugtrack.Enabled {
		bugtrack.Track(
			"com.(*Packet).readHeader(): b=[%d, %d, %d, %d, %d, %d, %d, %d, %d, %d, %d, %d, %d, %d]",
			b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11], b[12], b[13],
		)
	}
	p.ID, p.Job = b[0], uint16(b[2])|uint16(b[1])<<8
	p.Flags = Flag(b[10]) | Flag(b[9])<<8 | Flag(b[8])<<16 | Flag(b[7])<<24 |
		Flag(b[6])<<32 | Flag(b[5])<<40 | Flag(b[4])<<48 | Flag(b[3])<<56
	if l := int(b[12]) | int(b[11])<<8; l > 0 {
		p.Tags = make([]uint32, l)
	}
	switch b[13] {
	case 0:
		p.len, err = 0, nil
	case 1:
		if n, err = io.ReadFull(r, b[0:1]); n != 1 {
			if err == nil {
				err = io.ErrUnexpectedEOF
			}
			break
		}
		if bugtrack.Enabled {
			bugtrack.Track("com.(*Packet).readHeader(): 1, n=%d, b=[%d]", n, b[0])
		}
		p.len, err = uint64(b[0]), nil
	case 3:
		if n, err = io.ReadFull(r, b[0:2]); n != 2 {
			if err == nil {
				err = io.ErrUnexpectedEOF
			}
			break
		}
		if bugtrack.Enabled {
			bugtrack.Track("com.(*Packet).readHeader(): 3, n=%d, b=[%d, %d]", n, b[0], b[1])
		}
		p.len, err = uint64(b[1])|uint64(b[0])<<8, nil
	case 5:
		if n, err = io.ReadFull(r, b[0:4]); n != 4 {
			if err == nil {
				err = io.ErrUnexpectedEOF
			}
			break
		}
		if bugtrack.Enabled {
			bugtrack.Track("com.(*Packet).readHeader(): 5, n=%d, b=[%d, %d, %d, %d]", n, b[0], b[1], b[2], b[3])
		}
		p.len, err = uint64(b[3])|uint64(b[2])<<8|uint64(b[1])<<16|uint64(b[0])<<24, nil
	case 7:
		if n, err = io.ReadFull(r, b[0:8]); n != 8 {
			if err == nil {
				err = io.ErrUnexpectedEOF
			}
			break
		}
		if bugtrack.Enabled {
			bugtrack.Track(
				"com.(*Packet).readHeader(): 7, n=%d, b=[%d, %d, %d, %d, %d, %d, %d, %d]",
				n, b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7],
			)
		}
		p.len, err = uint64(b[7])|uint64(b[6])<<8|uint64(b[5])<<16|uint64(b[4])<<24|
			uint64(b[3])<<32|uint64(b[2])<<40|uint64(b[1])<<48|uint64(b[0])<<56, nil
	default:
		return data.ErrInvalidType
	}
	if bugtrack.Enabled {
		bugtrack.Track("com.(*Packet).readHeader(): p.ID=%d, p.len=%d, err=%s", p.ID, p.len, err)
	}
	return err
}
func (p *Packet) writeHeader(w io.Writer) error {
	t := len(p.Tags)
	if t > PacketMaxTags {
		return xerr.Sub("tags list is too large", 0x2B)
	}
	if bugtrack.Enabled {
		if p.Device.Empty() {
			bugtrack.Track("com.(*Packet).writeHeader(): Calling writeHeader with empty Device, p.ID=%d!", p.ID)
		}
	}
	if err := p.Device.Write(w); err != nil {
		return err
	}
	var (
		b [22]byte
		c int
	)
	_ = b[21]
	b[0], b[1], b[2] = p.ID, byte(p.Job>>8), byte(p.Job)
	b[3], b[4], b[5], b[6] = byte(p.Flags>>56), byte(p.Flags>>48), byte(p.Flags>>40), byte(p.Flags>>32)
	b[7], b[8], b[9], b[10] = byte(p.Flags>>24), byte(p.Flags>>16), byte(p.Flags>>8), byte(p.Flags)
	b[11], b[12] = byte(t>>8), byte(t)
	switch l := uint64(p.Chunk.Size()); {
	case l == 0:
		b[13] = 0
	case l < data.LimitSmall:
		b[13], b[14], c = 1, byte(l), 1
	case l < data.LimitMedium:
		b[13], b[14], b[15], c = 3, byte(l>>8), byte(l), 2
	case l < data.LimitLarge:
		b[13], c = 5, 4
		b[14], b[15], b[16], b[17] = byte(l>>24), byte(l>>16), byte(l>>8), byte(l)
	default:
		b[13], c = 7, 8
		b[14], b[15], b[16], b[17] = byte(l>>56), byte(l>>48), byte(l>>40), byte(l>>32)
		b[18], b[19], b[20], b[21] = byte(l>>24), byte(l>>16), byte(l>>8), byte(l)
	}
	// NOTE(dij): This write is split into two writes as some stateful writes (XOR)
	//             require writes and reads to re-constructed in the same way.
	n, err := w.Write(b[0:14])
	if err != nil {
		return err
	}
	if n != 14 {
		return io.ErrShortWrite
	}
	if n, err = w.Write(b[14 : 14+c]); err != nil {
		return err
	}
	if n != c {
		return io.ErrShortWrite
	}
	if bugtrack.Enabled {
		bugtrack.Track("com.(*Packet).writeHeader(): p.ID=%d, p.len=%d, n=%d", p.ID, p.Chunk.Size(), c+14+device.IDSize)
	}
	return nil
}

// MarshalStream writes the data of this Packet to the supplied Writer.
func (p *Packet) MarshalStream(w data.Writer) error {
	if bugtrack.Enabled {
		if p.Device.Empty() {
			bugtrack.Track("com.(*Packet).writeHeader(): Calling writeHeader with empty Device, p.ID=%d!", p.ID)
		}
	}
	if err := w.WriteUint8(p.ID); err != nil {
		return err
	}
	if err := w.WriteUint16(p.Job); err != nil {
		return err
	}
	if err := w.WriteUint16(uint16(len(p.Tags))); err != nil {
		return err
	}
	if err := p.Flags.MarshalStream(w); err != nil {
		return err
	}
	if err := p.Device.MarshalStream(w); err != nil {
		return err
	}
	for i := 0; i < len(p.Tags) && i < PacketMaxTags; i++ {
		if err := w.WriteUint32(p.Tags[i]); err != nil {
			return err
		}
	}
	return p.Chunk.MarshalStream(w)
}

// UnmarshalStream reads the data of this Packet from the supplied Reader.
func (p *Packet) UnmarshalStream(r data.Reader) error {
	if err := r.ReadUint8(&p.ID); err != nil {
		return err
	}
	if err := r.ReadUint16(&p.Job); err != nil {
		return err
	}
	t, err := r.Uint16()
	if err != nil {
		return err
	}
	if err := p.Flags.UnmarshalStream(r); err != nil {
		return err
	}
	if err := p.Device.UnmarshalStream(r); err != nil {
		return err
	}
	if t > 0 {
		p.Tags = make([]uint32, t)
		for i := uint16(0); i < t && i < PacketMaxTags; i++ {
			if err := r.ReadUint32(&p.Tags[i]); err != nil {
				return err
			}
			if p.Tags[i] == 0 {
				return ErrMalformedTag
			}
		}
	}
	return p.Chunk.UnmarshalStream(r)
}
