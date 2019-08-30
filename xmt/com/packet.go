package com

import (
	"encoding/json"
	"fmt"

	"github.com/iDigitalFlame/xmt/xmt/data"
	"github.com/iDigitalFlame/xmt/xmt/device"
	"github.com/iDigitalFlame/xmt/xmt/device/local"
	"golang.org/x/xerrors"
)

var (
	// ErrTooLarge is raised if memory cannot be allocated to store data in a buffer.
	ErrTooLarge = xerrors.New("buffer size is too large")
	// ErrInvalidIndex is raised if a specified Grow or index function is supplied with an
	// negative or out of bounds number.
	ErrInvalidIndex = xerrors.New("buffer index provided is not valid")
)

// Packet is a strust that is a Reader and Writer that can
// be generated to be sent, or received from a Connection.
type Packet struct {
	ID     uint16
	Frag   Fragment
	Flags  Flag
	Device device.ID

	buf  []byte
	rpos int
	wpos int
}

// Reset resets the payload buffer to be empty,
// but it retains the underlying storage for use by future writes.
func (p *Packet) Reset() {
	p.buf = p.buf[:0]
	p.wpos = 0
	p.rpos = 0
}

// Len returns the number of bytes of the unread portion of the
// Packet payload buffer.
func (p *Packet) Len() int {
	return len(p.buf) - p.wpos
}

// Empty returns true if this packet is nil
// or does not have any value or data associated with it.
func (p *Packet) Empty() bool {
	return p == nil || (p.buf == nil && p.ID <= 0)
}

// String returns the contents of the unread portion of the buffer
// as a string.
func (p *Packet) String() string {
	if p == nil {
		return "0x0 <empty>"
	}
	return fmt.Sprintf("0x%X %d bytes", p.ID, p.Len())
}

// Payload returns a slice of length p.Len() holding the unread portion of the
// Packet payload buffer.
func (p *Packet) Payload() []byte {
	return p.buf[p.wpos:]
}

// MarshalJSON writes the data of this Packet into JSON format.
func (p *Packet) MarshalJSON() ([]byte, error) {
	if p.Device == nil {
		p.Device = local.ID()
	}
	return json.Marshal(
		map[string]interface{}{
			"id":      p.ID,
			"frag":    p.Frag,
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
		return xerrors.Errorf("json: missing \"id\" property")
	}
	if err := json.Unmarshal(i, &(p.ID)); err != nil {
		return err
	}
	f, ok := m["frag"]
	if !ok {
		return xerrors.Errorf("json: missing \"frag\" property")
	}
	if err := json.Unmarshal(f, &(p.Frag)); err != nil {
		return err
	}
	q, ok := m["flags"]
	if !ok {
		return xerrors.Errorf("json: missing \"flags\" property")
	}
	if err := json.Unmarshal(q, &(p.Flags)); err != nil {
		return err
	}
	d, ok := m["device"]
	if err := json.Unmarshal(d, &(p.Device)); err != nil {
		return err
	}
	if !ok {
		return xerrors.Errorf("json: missing \"device\" property")
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
	if err := w.WriteUint32(uint32(p.Frag)); err != nil {
		return err
	}
	if err := w.WriteUint32(uint32(p.Flags)); err != nil {
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
	var i, f uint32
	if p.ID, err = r.Uint16(); err != nil {
		return err
	}
	if i, err = r.Uint32(); err != nil {
		return err
	}
	if f, err = r.Uint32(); err != nil {
		return err
	}
	p.Flags = Flag(f)
	p.Frag = Fragment(i)
	if err := p.Device.UnmarshalStream(r); err != nil {
		return err
	}
	if p.buf, err = r.Bytes(); err != nil {
		return err
	}
	return nil
}
