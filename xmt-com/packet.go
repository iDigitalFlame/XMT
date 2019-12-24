package com

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	data "github.com/iDigitalFlame/xmt/xmt-data"
	device "github.com/iDigitalFlame/xmt/xmt-device"
	util "github.com/iDigitalFlame/xmt/xmt-util"
)

const (
	// PacketHeaderSize is the length of the Packet header in bytes.
	PacketHeaderSize = 44
)

var (
	// ErrMismatchedID is an error that occurs when attempting to combine a Packet with a Packet that does
	// not match the ID of the parent Packet.
	ErrMismatchedID = errors.New("packet ID does not match combining packet ID")
)

// Packet is a struct that is a Reader and Writer that can
// be generated to be sent, or received from a Connection.
type Packet struct {
	ID, Job uint16
	Flags   Flag
	Device  device.ID

	buf        []byte
	rpos, wpos int
	stream     *Stream
}

// Reset resets the payload buffer to be empty,
// but it retains the underlying storage for use by future writes.
func (p *Packet) Reset() {
	if p.stream != nil {
		p.stream.Clear()
		p.stream = nil
	}
	p.ID = 0
	p.wpos = 0
	p.rpos = 0
	p.Flags = 0
	p.Device = nil
	p.buf = p.buf[:0]
}

// Clear is similar to Reset, but discards the buffer,
// which must be allocated again. If using the buffer the Reset
// function is preferable.
func (p *Packet) Clear() {
	if p.stream != nil {
		p.stream.Clear()
		p.stream = nil
	}
	p.Reset()
	p.buf = nil
}

// Len returns the total size of this Packet payload.
func (p Packet) Len() int {
	if p.stream != nil {
		return p.stream.Len()
	}
	return p.Size()
}

// Size returns the total size of this Packet payload.
func (p *Packet) Size() int {
	if p.stream != nil {
		return p.stream.Size()
	}
	if p.buf == nil {
		return 0
	}
	return len(p.buf) - p.wpos
}

// IsEmpty returns true if this packet is nil
// or does not have any value or data associated with it.
func (p *Packet) IsEmpty() bool {
	return p == nil || (p.buf == nil && p.ID <= 0)
}

// String returns the contents of the unread portion of the buffer
// as a string.
func (p *Packet) String() string {
	switch {
	case p == nil:
		return empty
	case (p.buf == nil || len(p.buf) == 0) && p.Flags == 0 && p.Job == 0:
		return fmt.Sprintf("0x%X", p.ID)
	case (p.buf == nil || len(p.buf) == 0) && p.Flags == 0:
		return fmt.Sprintf("0x%X/%d", p.ID, p.Job)
	case (p.buf == nil || len(p.buf) == 0) && p.Job == 0:
		return fmt.Sprintf("0x%X %s", p.ID, p.Flags)
	case (p.buf == nil || len(p.buf) == 0):
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

// NewStream converts a Packet into a Packet Stream and sets
// the underlying buffer to a Packet Stream.
func NewStream(p *Packet) *Packet {
	s := &Stream{
		ID:     p.ID,
		Job:    p.Job,
		Flags:  p.Flags,
		Device: p.Device,
	}
	s.Add(p)
	return &Packet{
		ID:     p.ID,
		buf:    make([]byte, 8),
		Job:    p.Job,
		Flags:  s.Flags,
		stream: s,
		Device: p.Device,
	}
}

// Payload returns a slice of length p.Len() holding the unread portion of the
// Packet payload buffer.
func (p *Packet) Payload() []byte {
	if p.buf == nil {
		return nil
	}
	return p.buf[p.wpos:]
}

// Check is a function that will set any missing Job or device
// parameters. This function will return true if the Device is nil or
// matches the specified host ID, false if otherwise.
func (p *Packet) Check(i device.ID) bool {
	if p.Job == 0 && p.Flags&FlagProxy == 0 {
		p.Job = uint16(util.Rand.Uint32())
	}
	if p.Device == nil {
		p.Device = i
		return true
	}
	return bytes.Equal(p.Device, i)
}

// Combine attempts to append the Payload of the supplied Packet to this Packet's
// current buffer. Combine fails with the 'ErrMismatchedID' if the ID values are not the same.
// This function keeps the original Flag, ID and Job values of the parent Packet.
func (p *Packet) Combine(o *Packet) error {
	if p == nil || o == nil {
		return nil
	}
	if o.buf == nil || o.Len() == 0 {
		return nil
	}
	if p.ID != o.ID {
		return ErrMismatchedID
	}
	p.Flags.SetFragTotal(o.Flags.FragTotal())
	if p.Flags.FragPosition() < o.Flags.FragPosition() {
		p.Flags.SetFragPosition(o.Flags.FragPosition())
	}
	if p.stream != nil {
		return p.stream.Add(o)
	}
	if _, err := p.Write(o.buf[o.wpos:]); err != nil {
		return fmt.Errorf("unable to write to Packet: %w", err)
	}
	return nil
}

// MarshalJSON writes the data of this Packet into JSON format.
func (p *Packet) MarshalJSON() ([]byte, error) {
	if p.Device == nil {
		p.Device = device.Local.ID
	}
	return json.Marshal(
		map[string]interface{}{
			"id":      p.ID,
			"job":     p.Job,
			"flags":   p.Flags,
			"device":  p.Device,
			"payload": p.buf,
		},
	)
}

// UnmarshalJSON read the data of this Packet from JSON format.
func (p *Packet) UnmarshalJSON(b []byte) error {
	m := make(map[string]json.RawMessage)
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	i, ok := m["id"]
	if !ok {
		return fmt.Errorf("json: missing \"id\" property")
	}
	if err := json.Unmarshal(i, &(p.ID)); err != nil {
		return fmt.Errorf("unable to unmarshal \"id\" value: %w", err)
	}
	j, ok := m["job"]
	if !ok {
		return fmt.Errorf("json: missing \"job\" property")
	}
	if err := json.Unmarshal(j, &(p.Job)); err != nil {
		return fmt.Errorf("unable to unmarshal \"job\" value: %w", err)
	}
	q, ok := m["flags"]
	if !ok {
		return fmt.Errorf("json: missing \"flags\" property")
	}
	if err := json.Unmarshal(q, &(p.Flags)); err != nil {
		return fmt.Errorf("unable to unmarshal \"flags\" value: %w", err)
	}
	d, ok := m["device"]
	if err := json.Unmarshal(d, &(p.Device)); err != nil {
		return fmt.Errorf("unable to unmarshal \"device\" value: %w", err)
	}
	if !ok {
		return fmt.Errorf("json: missing \"device\" property")
	}
	if o, ok := m["payload"]; ok {
		if err := json.Unmarshal(o, &(p.buf)); err != nil {
			return fmt.Errorf("unable to unmarshal \"payload\" value: %w", err)
		}
	}
	return nil
}

// MarshalStream writes the data of this Packet to the supplied Writer.
func (p *Packet) MarshalStream(w data.Writer) error {
	if p.Device == nil {
		p.Device = device.Local.ID
	}
	if err := w.WriteUint16(p.ID); err != nil {
		return err
	}
	if err := w.WriteUint16(p.Job); err != nil {
		return err
	}
	if err := p.Flags.MarshalStream(w); err != nil {
		return err
	}
	if err := p.Device.MarshalStream(w); err != nil {
		return err
	}
	if err := w.WriteBytes(p.buf); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream reads the data of this Packet from the supplied Reader.
func (p *Packet) UnmarshalStream(r data.Reader) error {
	if err := r.ReadUint16(&(p.ID)); err != nil {
		return err
	}
	if err := r.ReadUint16(&(p.Job)); err != nil {
		return err
	}
	if err := p.Flags.UnmarshalStream(r); err != nil {
		return err
	}
	if err := p.Device.UnmarshalStream(r); err != nil {
		return err
	}
	var err error
	if p.buf, err = r.Bytes(); err != nil {
		return err
	}
	return nil
}
