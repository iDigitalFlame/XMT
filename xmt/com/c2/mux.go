package c2

import (
	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/device"
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
type clientMux struct{}
type serverMux struct{}

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
func (m *clientMux) Handle(s *Session, p *com.Packet) {
	if p == nil || s == nil {
		return
	}
	s.controller.Log.Trace("[%s] Received packet \"%s\"...", s.ID, p.String())
	switch p.ID {
	case MsgHello, MsgPing, MsgSleep, MsgRegister:
		return
	case MsgRefresh:
		n := &com.Packet{ID: MsgRefresh, Job: p.Job}
		if err := device.Local.Refresh(); err != nil {
			n.Flags |= com.FlagError
			n.WriteString(err.Error())
			n.Close()
			s.WritePacket(n)
			return
		}
		if err := device.Local.MarshalStream(n); err != nil {
			s.controller.Log.Warning("[%s] Received an error when attempting to write data to a Packet! (%s)", s.ID, err.Error())
			n.Flags |= com.FlagError
			n.WriteString(err.Error())
			n.Close()
			s.WritePacket(n)
			return
		}
		s.WritePacket(n)
	}
}
func (m *serverMux) Handle(s *Session, p *com.Packet) {
	if p == nil {
		return
	}
	if s != nil {
		s.controller.Log.Trace("[%s] Received packet \"%s\"...", s.ID, p.String())
	} else {
		s.controller.Log.Trace("[000000] Received packet \"%s\"...", s.ID, p.String())
	}
}
