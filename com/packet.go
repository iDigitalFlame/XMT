package com

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util"
)

// PacketHeaderSize is the length of the Packet header in bytes.
const PacketHeaderSize = 46

// ErrMismatchedID is an error that occurs when attempting to combine a Packet with a Packet that does
// not match the ID of the parent Packet.
var ErrMismatchedID = errors.New("packet ID does not match combining packet ID")

// Packet is a struct that is a Reader and Writer that can
// be generated to be sent, or received from a Connection.
type Packet struct {
	ID     uint16
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
	s := p.Chunk.Size() + PacketHeaderSize + (4 * len(p.Tags))
	switch {
	case s < data.DataLimitSmall:
		return s + 1
	case s < data.DataLimitMedium:
		return s + 2
	case s < data.DataLimitLarge:
		return s + 4
	default:
		return s + 8
	}
}

// String returns a string descriptor of the Packet struct.
func (p Packet) String() string {
	switch {
	case p.Empty() && p.Flags == 0 && p.Job == 0:
		return fmt.Sprintf("0x%X", p.ID)
	case p.Empty() && p.Flags == 0:
		return fmt.Sprintf("0x%X/%d", p.ID, p.Job)
	case p.Empty() && p.Job == 0:
		return fmt.Sprintf("0x%X %s", p.ID, p.Flags)
	case p.Empty():
		return fmt.Sprintf("0x%X/%d %s", p.ID, p.Job, p.Flags)
	case p.Flags == 0 && p.Job == 0:
		return fmt.Sprintf("0x%X: %dB", p.ID, p.Size())
	case p.Flags == 0:
		return fmt.Sprintf("0x%X/%d: %dB", p.ID, p.Job, p.Size())
	case p.Job == 0:
		return fmt.Sprintf("0x%X %s: %dB", p.ID, p.Flags, p.Size())
	}
	return fmt.Sprintf("0x%X/%d %s: %dB", p.ID, p.Job, p.Flags, p.Size())
}

// Add attempts to combine the data and properties the supplied Packet with the existsing Packet. This
// function will return an error if the ID's have a mismatch or there was an error during the write operation.
func (p *Packet) Add(n *Packet) error {
	if n == nil || n.Empty() {
		return nil
	}
	if p.ID != n.ID {
		return ErrMismatchedID
	}
	if _, err := n.WriteTo(p); err != nil {
		return fmt.Errorf("unable to write to Packet: %w", err)
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
		p.Job = uint16(util.Rand.Uint32())
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
		p.Device = device.Local.ID
	}
	if err := w.WriteUint16(p.ID); err != nil {
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
	if err := r.ReadUint16(&p.ID); err != nil {
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
