package com

import (
	"errors"
	"io"
	"sort"

	"github.com/iDigitalFlame/xmt/xmt/data"
	"github.com/iDigitalFlame/xmt/xmt/device"
	"github.com/iDigitalFlame/xmt/xmt/util"
)

var (
	// ErrWriterAttached is an error returned by the Read function
	// when a Packet Writer is not nil. When a Writer is attached, the Stream
	// is considered a 'write-only' stream.
	ErrWriterAttached = errors.New("cannot read while a Packet Writer is attached")
	// ErrAlreadyAttached is an error returned by the Writer function. This
	// error is only returned if a Packet Writer is already attached.
	ErrAlreadyAttached = errors.New("a Packet Writer is already attached")
)

// Stream is a struct that can be used to write and
// fragment large Packets transparently. This struct can be
// used to write-on-demand the Packet buffer if a Packet Writer
// is supplied via the Writer function.
type Stream struct {
	ID     uint16
	Job    uint16
	Max    int
	Flags  Flag
	Device device.ID

	last    uint16
	wrote   uint16
	writer  writer
	packets []*Packet
}

// Clear discards the underlying buffers, which must be regenerated
// again.
func (s *Stream) Clear() {
	for i := range s.packets {
		s.packets[i].Clear()
	}
	s.packets = nil
	s.last, s.wrote = 0, 0
}

// Len returns the number of Packets in the
// underlying Packet array.
func (s *Stream) Len() int {
	return len(s.packets)
}

// Size returns the total size of this Stream payload.
func (s *Stream) Size() int {
	if s.packets == nil || len(s.packets) == 0 {
		return 0
	}
	var n int
	for i := range s.packets {
		n += s.packets[i].Size()
	}
	return n
}

// New creates a new Packet to be added to the underlying
// Stream. This packet will inherit the ID, Job and Flags
// of the Parent stream.
func (s *Stream) New() *Packet {
	if s.packets == nil {
		s.packets = make([]*Packet, 0, 1)
	} else {
		s.packets[s.last].Close()
		s.last++
	}
	if s.Flags.FragGroup() == 0 {
		s.Flags.SetFragGroup(uint16(util.Rand.Int()))
	}
	p := &Packet{
		ID:     s.ID,
		Job:    s.Job,
		Flags:  s.Flags,
		Device: s.Device,
	}
	p.Flags.SetFragPosition(s.last)
	s.packets = append(s.packets, p)
	return p
}

// Close fulfills the io.Closer interface and will commit all
// the packets to the assigned Packet Writer (if not nil) and
// will set the correct position and total Packet Flags.
func (s *Stream) Close() error {
	for i := range s.packets {
		s.packets[i].Close()
		s.packets[i].ID = s.ID
		s.packets[i].Job = s.Job
		s.packets[i].Flags = s.Flags
		s.packets[i].Device = s.Device
		s.packets[i].Flags.SetFragPosition(uint16(i))
		s.packets[i].Flags.SetFragTotal(uint16(len(s.packets)))
	}
	if s.writer != nil {
		s.last++
		s.flushPackets()
	}
	return nil
}
func (s *Stream) flushPackets() {
	if s.writer == nil {
		return
	}
	for s.wrote < s.last {
		if len(s.packets) == 1 {
			s.packets[s.wrote].Flags = 0
		}
		s.writer.WriteWait(s.packets[s.wrote])
		s.wrote++
	}
	return
}

// Swap swaps the Packets in the array in the current
// supplied positions.
func (s *Stream) Swap(i, j int) {
	s.packets[i], s.packets[j] = s.packets[j], s.packets[i]
}

// Less returns true if the Frag Position of
// the Packet is less than the other supplied Packet
// position.
func (s *Stream) Less(i, j int) bool {
	return s.packets[i].Flags.FragPosition() < s.packets[j].Flags.FragPosition()
}

// Add adds the supplied Packet to the stream array. This
// function triggers a sort of the array to ensure the Frags
// are in the correct position.
func (s *Stream) Add(p *Packet) error {
	if s.ID != s.ID {
		return ErrMismatchedID
	}
	p.Close()
	if s.packets == nil {
		s.packets = make([]*Packet, 0, 1)
	}
	s.packets = append(s.packets, p)
	s.last++
	sort.Sort(s)
	return nil
}

// Consolidate combines the data in the underling Packet
// array into a single Packet.
func (s *Stream) Consolidate() *Packet {
	var p *Packet
	for i := range s.packets {
		if p == nil {
			p = s.packets[i]
			continue
		}
		p.Combine(s.packets[i])
	}
	return p
}

// Writer attaches the Packet Writer to this stream.
// This enables the write-on-demand function and will attempt
// to write all Packets in the backlog on subsequent writes.
func (s *Stream) Writer(w writer) error {
	if s.writer != nil {
		return ErrAlreadyAttached
	}
	s.writer = w
	return nil
}

// Read fulfills the io.Reader interface. This function will
// return ErrWriterAttached if a Packet Writer is attached.
func (s *Stream) Read(b []byte) (int, error) {
	if s.writer != nil {
		return 0, ErrWriterAttached
	}
	if s.packets == nil || len(s.packets) == 0 || s.wrote >= uint16(len(s.packets)) {
		return 0, io.EOF
	}
	var n int
	var p *Packet
	for ; n < len(b); s.wrote++ {
		if s.wrote >= uint16(len(s.packets)) {
			return n, nil
		}
		p = s.packets[s.wrote]
		p.rpos = copy(b[n:], p.buf[p.rpos:])
		n += p.rpos
	}
	return n, nil
}

// Write fulfills the io.Writer interface. This will generate
// and allocate new Packets on the fly. This will also send
// Packets out, if a Packet Session Writer is set.
func (s *Stream) Write(b []byte) (int, error) {
	var p *Packet
	if s.packets == nil {
		p = s.New()
	} else {
		p = s.packets[s.last]
	}
	if s.Max <= -1 {
		return p.Write(b)
	} else if s.Max == 0 {
		s.Max = data.DataLimitMedium
	}
	if p.Size() >= s.Max {
		p = s.New()
		s.flushPackets()
	}
	if len(b) < s.Max {
		return p.Write(b)
	}
	var n int
	for r := len(b) - n; n < len(b); r = len(b) - n {
		if r <= 0 {
			break
		}
		if r > s.Max {
			r = s.Max
		}
		v, err := p.Write(b[n : n+int(r)])
		if err != nil {
			return n, err
		}
		n += v
		if n < len(b) {
			p = s.New()
			s.flushPackets()
		}
	}
	return n, nil
}
