package com

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iDigitalFlame/xmt/xmt/data"
	"github.com/iDigitalFlame/xmt/xmt/device"
	"github.com/iDigitalFlame/xmt/xmt/device/local"
)

const (
	// MaxAutoMultiSize is the max amount of packets that
	// are automatically converted into a group before sending from
	// a client or server. Groups of larger multi-packets can be created manually.
	MaxAutoMultiSize = 64

	emptyPacket = "0xNULL"
)

var (
	// ErrTooLarge is raised if memory cannot be allocated to store data in a buffer.
	ErrTooLarge = errors.New("buffer size is too large")
	// ErrInvalidIndex is raised if a specified Grow or index function is supplied with an
	// negative or out of bounds number.
	ErrInvalidIndex = errors.New("buffer index provided is not valid")
	// ErrMismatchedID is an error that occurs when attempting to combine a Packet with a Packet that does
	// not match the ID of the parent Packet.
	ErrMismatchedID = errors.New("packet ID does not match combining packet ID")
)

// Packet is a struct that is a Reader and Writer that can
// be generated to be sent, or received from a Connection.
type Packet struct {
	ID     uint16
	Job    uint16
	Flags  Flag
	Device device.ID

	buf  []byte
	rpos int
	wpos int
}

// Reset resets the payload buffer to be empty,
// but it retains the underlying storage for use by future writes.
func (p *Packet) Reset() {
	p.ID = 0
	p.wpos = 0
	p.rpos = 0
	p.Flags = 0
	p.Device = nil
	p.buf = p.buf[:0]
}

// Len returns the number of bytes of the unread portion of the
// Packet payload buffer.
func (p *Packet) Len() int {
	if p.buf == nil {
		return 0
	}
	return len(p.buf) - p.wpos
}

// ResetFull is similar to Reset, but discards the buffer,
// which must be allocated again. If using the buffer, Reset()
// is preferable.
func (p *Packet) ResetFull() {
	p.Reset()
	p.buf = nil
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
		return emptyPacket
	case (p.buf == nil || len(p.buf) == 0) && p.Flags == 0 && p.Job == 0:
		return fmt.Sprintf("0x%X", p.ID)
	case (p.buf == nil || len(p.buf) == 0) && p.Flags == 0:
		return fmt.Sprintf("0x%X/%d", p.ID, p.Job)
	case (p.buf == nil || len(p.buf) == 0) && p.Job == 0:
		return fmt.Sprintf("0x%X %s", p.ID, p.Flags)
	case (p.buf == nil || len(p.buf) == 0):
		return fmt.Sprintf("0x%X/%d %s", p.ID, p.Job, p.Flags)
	case p.Flags == 0 && p.Job == 0:
		return fmt.Sprintf("0x%X: %dB", p.ID, p.Len())
	case p.Flags == 0:
		return fmt.Sprintf("0x%X/%d: %dB", p.ID, p.Job, p.Len())
	case p.Job == 0:
		return fmt.Sprintf("0x%X %s: %dB", p.ID, p.Flags, p.Len())
	}
	return fmt.Sprintf("0x%X/%d %s: %dB", p.ID, p.Job, p.Flags, p.Len())
}

// Payload returns a slice of length p.Len() holding the unread portion of the
// Packet payload buffer.
func (p *Packet) Payload() []byte {
	if p.buf == nil {
		return nil
	}
	return p.buf[p.wpos:]
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
	_, err := p.Write(o.buf[o.wpos:])
	return err
}

// MarshalJSON writes the data of this Packet into JSON format.
func (p *Packet) MarshalJSON() ([]byte, error) {
	if p.Device == nil {
		p.Device = local.ID()
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
		return err
	}
	j, ok := m["job"]
	if !ok {
		return fmt.Errorf("json: missing \"job\" property")
	}
	if err := json.Unmarshal(j, &(p.Job)); err != nil {
		return err
	}
	q, ok := m["flags"]
	if !ok {
		return fmt.Errorf("json: missing \"flags\" property")
	}
	if err := json.Unmarshal(q, &(p.Flags)); err != nil {
		return err
	}
	d, ok := m["device"]
	if err := json.Unmarshal(d, &(p.Device)); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("json: missing \"device\" property")
	}
	if o, ok := m["payload"]; ok {
		if err := json.Unmarshal(o, &(p.buf)); err != nil {
			return err
		}
	}
	return nil
}

// MarshalStream writes the data of this Packet to the supplied Writer.
func (p *Packet) MarshalStream(w data.Writer) error {
	if p.Device == nil {
		p.Device = local.ID()
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
	var err error
	if p.ID, err = r.Uint16(); err != nil {
		return err
	}
	if p.Job, err = r.Uint16(); err != nil {
		return err
	}
	if err := p.Flags.UnmarshalStream(r); err != nil {
		return err
	}
	if err := p.Device.UnmarshalStream(r); err != nil {
		return err
	}
	if p.buf, err = r.Bytes(); err != nil {
		return err
	}
	return nil
}
