package c2

import (
	"context"
	"fmt"
	"time"

	"github.com/iDigitalFlame/xmt/xmt/c2/control"
	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/util"
)

// Job is a struct that is used to track and manage
// Tasks given to Session Clients. This struct has function
// callbacks that can be used to watch for completion and also
// offers a Wait function to pause execution until a response is received.
type Job struct {
	ID              uint16
	Done, Error     func(*Session, *com.Packet)
	Result, Initial *com.Packet
	Start, Complete time.Time

	ctx    context.Context
	cancel context.CancelFunc
}
type muxClient bool
type muxServer struct {
	active     map[uint16]*Job
	controller *Server
}

// Wait will block until the Job is completed or the
// parent Controller is shutdown.
func (j *Job) Wait() {
	if j.ctx.Err() != nil {
		return
	}
	<-j.ctx.Done()
}

// IsDone returns true when the Job has received a response.
func (j *Job) IsDone() bool {
	return j.ctx.Err() != nil
}
func (muxClient) Handle(s *Session, p *com.Packet) {
	s.controller.Log.Debug("[%s:Mux] Received packet \"%s\"...", s.ID, p.String())
	switch p.ID {
	case MsgUpload:
		Process(control.Upload, s, p)
	case MsgRefresh:
		Process(control.Refresh, s, p)
	case MsgExecute:
		Process(control.Execute, s, p)
	case MsgDownload:
		Process(control.Download, s, p)
	case MsgProcesses:
		Process(control.ProcessList, s, p)
	case MsgHello, MsgPing, MsgSleep, MsgRegistered:
		return
	}
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
	j.Complete = time.Now()
	j.cancel()
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
	j := &Job{
		ID:      p.Job,
		Start:   time.Now(),
		Initial: p,
	}
	j.ctx, j.cancel = context.WithCancel(m.controller.ctx)
	m.active[p.Job] = j
	return j, nil
}

/*
// NewJobContext creates a new Job struct and initializes the context and
// cancel objects inside. The provided uint16 is the Job ID and the context
// is the parent context, which will be used to cancel the Job if the parent
// cancels.
func NewJobContext(x context.Context, i uint16) *Job {
	j := &Job{ID: i}
	j.ctx, j.cancel = context.WithCancel(x)
	return j
} */
/*
// NewJob creates a new Job struct and initializes the context and
// cancel objects inside. The provided uint16 is the Job ID.
func NewJob(i uint16) *Job {
	return NewJobContext(context.Background(), i)
}*/
