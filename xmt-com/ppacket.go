package com

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	data "github.com/iDigitalFlame/xmt/xmt-data"
	device "github.com/iDigitalFlame/xmt/xmt-device"
	util "github.com/iDigitalFlame/xmt/xmt-util"
)

var (
	bufs = &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

type packet struct {
	ID     uint16
	Job    uint16
	Flags  Flag
	Device device.ID

	*data.Chunk
}

// Clear discards the underlying buffers, which must be regenerated again.
func (p *packet) Clear() {
	for i := range p.b {
		p.b[i].buf = nil
	}
	p.b = nil
	p.l, p.w = 0, 0
}

// Len returns the number of underlying Chunks contained in the Packet.
func (p packet) Len() int {
	return len(p.b)
}

// Size returns the total size of this Packet's payload.
func (p packet) Size() int {
	if p.b == nil || len(p.b) == 0 {
		return 0
	}
	if len(p.b) == 1 {
		return p.b[0].Size()
	}
	var n int
	for i := range p.b {
		n += p.b[i].Size()
	}
	return n
}

// Empty returns true if the Packet's underlying Chunks are empty or nil.
func (p packet) Empty() bool {
	if p.b == nil || len(p.b) == 0 {
		return true
	}
	if len(p.b) == 1 {
		return p.b[0].buf == nil || len(p.b[0].buf) == 0
	}
	return false
}

// Close will close all the underlying Chunks and will assign the correct
// frag index and grouping.
func (p *packet) Close() error {
	for i := range p.b {
		p.b[i].Close()
		//p.b[i].fpos = uint16(i)
	}
	if len(p.b) == 1 {
		p.Flags.ClearFrag()
	} else {
		p.Flags.SetFragTotal(uint16(len(p.b)))
		if p.Flags.FragGroup() == 0 {
			p.Flags.SetFragGroup(uint16(util.Rand.Int()))
		}
	}
	//if s.writer != nil {
	//	s.last++
	//	s.flushPackets(true)
	//}
	return nil
}
func (p *packet) next() *Chunk {
	if p.b == nil {
		p.b = make([]*Chunk, 0, 1)
	} else if len(p.b) > 0 {
		p.b[p.l].Close()
		p.l++
	}
	//if p.Flags.FragGroup() == 0 {
	//	s.Flags.SetFragGroup(uint16(util.Rand.Int()))
	//}
	//p.Flags.SetFragPosition(s.last)
	n := &Chunk{} //fpos: uint16(p.l)}
	p.b = append(p.b, n)
	return n
}

// String returns a string descriptor of the Packet struct.
func (p packet) String() string {
	switch {
	case (p.b == nil || len(p.b) == 0) && p.Flags == 0 && p.Job == 0:
		return fmt.Sprintf("0x%X", p.ID)
	case (p.b == nil || len(p.b) == 0) && p.Flags == 0:
		return fmt.Sprintf("0x%X/%d", p.ID, p.Job)
	case (p.b == nil || len(p.b) == 0) && p.Job == 0:
		return fmt.Sprintf("0x%X %s", p.ID, p.Flags)
	case (p.b == nil || len(p.b) == 0):
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

// Swap swaps the Packets in the array in the current
// supplied positions.
func (p *packet) Swap(i, j int) {
	p.b[i], p.b[j] = p.b[j], p.b[i]
}

// Less returns true if the position of the Chunk at i
// is less than the supplied Chunk at position j.
func (p packet) Less(i, j int) bool {
	return false //p.b[i].fpos < p.b[j].fpos
}

// Belongs returns true if the specified Packet is a Frag that was
// a part of the split Chunks of the original packet.
func (p packet) Belongs(n *packet) bool {
	return n != nil && p.Flags > 0 && n.Flags > 0 && n.b != nil && len(n.b) > 0 && p.ID == n.ID && p.Flags.FragGroup() == n.Flags.FragGroup()
}

func (p *packet) Add(n *packet) error {
	return nil
}

func (p packet) Chunk(i int) (*Chunk, error) {
	if i < 0 {
		return nil, nil
	}
	if uint16(i) >= p.l {
		return nil, nil
	}
	return p.b[uint16(i)], nil
}

// Verify is a function that will set any missing Job or device
// parameters. This function will return true if the Device is nil or
// matches the specified host ID, false if otherwise.
func (p *packet) Verify(i device.ID) bool {
	if p.Job == 0 && p.Flags&FlagProxy == 0 {
		p.Job = uint16(util.Rand.Uint32())
	}
	if p.Device == nil {
		p.Device = i
		return true
	}
	return bytes.Equal(p.Device, i)
}

// MarshalJSON transform this Packet into a JSON text representation.
func (p packet) MarshalJSON() ([]byte, error) {
	if p.Device == nil {
		p.Device = device.Local.ID
	}
	return json.Marshal(
		map[string]interface{}{
			"id":      p.ID,
			"job":     p.Job,
			"flags":   p.Flags,
			"device":  p.Device,
			"payload": p.Payload(),
		},
	)
}

// UnmarshalJSON transform the Packet from a JSON from JSON text representation.
func (p *packet) UnmarshalJSON(b []byte) error {
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
		n := p.next()
		if err := json.Unmarshal(o, &(n.buf)); err != nil {
			return fmt.Errorf("unable to unmarshal \"payload\" value: %w", err)
		}
		n.wpos = len(n.buf)
	}
	return nil
}

// MarshalStream writes the data of this Packet to the supplied Writer.
func (p packet) MarshalStream(w data.Writer) error {
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
	//if err := w.WriteBytes(p.buf); err != nil {
	//	return err
	//}
	return nil
}

// UnmarshalStream reads the data of this Packet from the supplied Reader.
func (p *packet) UnmarshalStream(r data.Reader) error {
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
	//var err error
	//if p.buf, err = r.Bytes(); err != nil {
	//	return err
	//}
	return nil
}

// Payload returns a slice of length p.Size() holding the unread portion of the
// Packet payload buffer.
func (p packet) Payload() []byte {
	if p.b == nil || len(p.b) == 0 {
		return nil
	}
	if len(p.b) == 1 {
		return p.b[0].buf[p.b[0].wpos:]
	}
	b := bufs.Get().(*bytes.Buffer)
	for i := range p.b {
		b.Write(p.b[i].buf[p.b[i].wpos:])
	}
	o := make([]byte, b.Len())
	copy(o, b.Bytes())
	b.Reset()
	bufs.Put(b)
	return o
}
