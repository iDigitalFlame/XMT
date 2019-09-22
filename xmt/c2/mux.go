package c2

import (
	"github.com/iDigitalFlame/xmt/xmt/com"
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
