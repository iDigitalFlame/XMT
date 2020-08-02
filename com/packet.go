package com

import (
	"bytes"
	"errors"
	"strconv"

	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// PacketHeaderSize is the length of the Packet header in bytes.
const PacketHeaderSize = 45

// Packet is a struct that is a Reader and Writer that can
// be generated to be sent, or received from a Connection.
type Packet struct {
	ID     uint8
	Job    uint16
	Tags   []uint32
	Flags  Flag
	Device device.ID
	data.Chunk
}

// Size returns the amount of bytes written or contained in this Packet with the header size added.
func (p Packet) Size() int {
	if p.Empty() {
		return PacketHeaderSize
	}
	s := uint64(p.Chunk.Size() + PacketHeaderSize + (4 * len(p.Tags)))
	switch {
	case s < data.DataLimitSmall:
		return int(s) + 1
	case s < data.DataLimitMedium:
		return int(s) + 2
	case s < data.DataLimitLarge:
		return int(s) + 4
	default:
		return int(s) + 8
	}
}

// String returns a string descriptor of the Packet struct.
func (p Packet) String() string {
	switch {
	case p.Empty() && p.Flags == 0 && p.Job == 0:
		return "0x" + strconv.FormatUint(uint64(p.ID), 16)
	case p.Empty() && p.Flags == 0:
		return "0x" + strconv.FormatUint(uint64(p.ID), 16) + "/" + strconv.Itoa(int(p.Job))
	case p.Empty() && p.Job == 0:
		return "0x" + strconv.FormatUint(uint64(p.ID), 16) + " " + p.Flags.String()
	case p.Empty():
		return "0x" + strconv.FormatUint(uint64(p.ID), 16) + "/" + strconv.Itoa(int(p.Job)) + " " + p.Flags.String()
	case p.Flags == 0 && p.Job == 0:
		return "0x" + strconv.FormatUint(uint64(p.ID), 16) + ": " + strconv.Itoa(p.Size()) + "B"
	case p.Flags == 0:
		return "0x" + strconv.FormatUint(uint64(p.ID), 16) + "/" + strconv.Itoa(int(p.Job)) + ": " + strconv.Itoa(p.Size()) + "B"
	case p.Job == 0:
		return "0x" + strconv.FormatUint(uint64(p.ID), 16) + " " + p.Flags.String() + ": " + strconv.Itoa(p.Size()) + "B"
	}
	return "0x" + strconv.FormatUint(uint64(p.ID), 16) + "/" + strconv.Itoa(int(p.Job)) + " " + p.Flags.String() + ": " + strconv.Itoa(p.Size()) + "B"
}

// Add attempts to combine the data and properties the supplied Packet with the existsing Packet. This
// function will return an error if the ID's have a mismatch or there was an error during the write operation.
func (p *Packet) Add(n *Packet) error {
	if n == nil || n.Empty() {
		return nil
	}
	if p.ID != n.ID {
		return errors.New("Packet ID " + strconv.FormatUint(uint64(n.ID), 16) + " does not match combining Packet ID " + strconv.FormatUint(uint64(p.ID), 16))
	}
	if _, err := n.WriteTo(p); err != nil {
		return xerr.Wrap("unable to write to Packet", err)
	}
	return nil
}

// Belongs returns true if the specified Packet is a Frag that was a part of the split Chunks of this as the
// original packet.
func (p Packet) Belongs(n *Packet) bool {
	return n != nil && p.Flags >= FlagFrag && n.Flags >= FlagFrag && p.ID == n.ID && p.Job == n.Job && p.Flags.Group() == n.Flags.Group() && !n.Empty()
}

// Verify is a function that will set any missing Job or Device parameters. This function will return true if
// the Device is nil or matches the specified host ID, false if otherwise.
func (p *Packet) Verify(i device.ID) bool {
	if p.Job == 0 && p.Flags&FlagProxy == 0 {
		p.Job = uint16(util.FastRand())
	}
	if p.Device == nil {
		p.Device = i
		return true
	}
	return bytes.Equal(p.Device, i)
}

// MarshalStream writes the data of this Packet to the supplied Writer.
func (p Packet) MarshalStream(w data.Writer) error {
	if p.Device == nil {
		p.Device = device.UUID
	}
	if err := w.WriteUint8(p.ID); err != nil {
		return err
	}
	if err := w.WriteUint16(p.Job); err != nil {
		return err
	}
	if err := w.WriteUint8(uint8(len(p.Tags))); err != nil {
		return err
	}
	if err := p.Flags.MarshalStream(w); err != nil {
		return err
	}
	if err := p.Device.MarshalStream(w); err != nil {
		return err
	}
	for i := 0; i < len(p.Tags) && i < 256; i++ {
		if err := w.WriteUint32(p.Tags[i]); err != nil {
			return err
		}
	}
	if err := p.Chunk.MarshalStream(w); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream reads the data of this Packet from the supplied Reader.
func (p *Packet) UnmarshalStream(r data.Reader) error {
	if err := r.ReadUint8(&p.ID); err != nil {
		return err
	}
	if err := r.ReadUint16(&p.Job); err != nil {
		return err
	}
	t, err := r.Uint8()
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
		for i := uint8(0); i < t; i++ {
			if err := r.ReadUint32(&p.Tags[i]); err != nil {
				return err
			}
		}
	}
	if err := p.Chunk.UnmarshalStream(r); err != nil {
		return err
	}
	return nil
}
