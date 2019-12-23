package c2

import (
	"github.com/iDigitalFlame/xmt/xmt/c2/control"
	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/com/limits"
)

// Mux is an interface that handles Packets when the arrive.
type Mux interface {
	Handle(*Session, *com.Packet)
}

// MuxMap is a Mux handler that allows for mapping
// Packet ID values to a function or a specific default.
type MuxMap struct {
	Default MuxFunc

	m map[uint16]MuxFunc
}

// Scheduler is a type of Mux that allows for Scheduling Packets
// to be sent to a client and tracked by the Server.
type Scheduler interface {
	Schedule(*Session, *com.Packet) (*Job, error)
	Mux
}

// MuxFunc is the definition of a Mux Handler func. Once wrapped as a 'MuxFunc'..,
// these function aliases can be also used in place of the Mux interface.
type MuxFunc func(*Session, *com.Packet)

// Delete removes the MuxFun assigned to the Packet ID supplied.
// This will complete even if the ID value does not exist.
func (m *MuxMap) Delete(i uint16) {
	if m.m == nil {
		return
	}
	delete(m.m, i)
}

// Add appends the supplied MuxFunc to the mapping for the specified Packet ID.
func (m *MuxMap) Add(i uint16, h MuxFunc) {
	if m.m == nil {
		m.m = make(map[uint16]MuxFunc)
	}
	m.m[i] = h
}

// Handle satisfies the Mux interface requirement and will process the received
// Packet based on the Packet ID.
func (m *MuxMap) Handle(s *Session, p *com.Packet) {
	if p == nil {
		return
	}
	if m.m == nil {
		if m.Default != nil {
			m.Default(s, p)
		}
		return
	}
	h, ok := m.m[p.ID]
	if !ok {
		if m.Default != nil {
			m.Default(s, p)
		}
		return
	}
	h(s, p)
}

// Handle satisfies the Mux interface requirement and will process the received
// Packet. This function allows Wrapped MuxFunc objects to be used directly in place
// of more complex Mux definitions.
func (m MuxFunc) Handle(s *Session, p *com.Packet) {
	m(s, p)
}

// Process runs the Execute function of the supplied Action and sends the
// results back to the provided Session via Packet Stream.
func Process(a control.Action, s *Session, p *com.Packet) {
	if a.Thread() {
		go processAction(a, s, p)
	} else {
		processAction(a, s, p)
	}
}
func processAction(a control.Action, s *Session, p *com.Packet) {
	o := &com.Stream{
		ID:     MsgResult,
		Job:    p.Job,
		Max:    limits.FragLimit(),
		Device: p.Device,
	}
	o.Writer(s)
	if err := a.Execute(s, p, o); err != nil {
		o.Clear()
		o.Flags |= com.FlagError
		o.WriteString(err.Error())
	}
	o.Close()
}
