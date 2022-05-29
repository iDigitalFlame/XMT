//go:build !implant

package c2

import (
	"time"

	"github.com/iDigitalFlame/xmt/com"
)

// These are status values that indicate the general status of the Job.
const (
	StatusWaiting  status = 0
	StatusAccepted status = iota
	StatusReceiving
	StatusCompleted
	StatusError
	StatusCanceled
)

// Job is a struct that is used to track and manage Tasks given to Session
// Clients.
//
// This struct has function callbacks that can be used to watch for completion
// and offers a Wait function to pause execution until a response is received.
//
// This struct is always empty for implants.
type Job struct {
	Start, Complete time.Time

	Update func(*Job)
	Result *com.Packet
	done   chan struct{}
	s      *Session

	Error              string
	ID, Frags, Current uint16

	Type   uint8
	Status status
}
type status uint8

// Wait will block until the Job is completed or the parent Server is shutdown.
func (j *Job) Wait() {
	if j == nil {
		return
	}
	if j.done == nil {
		return
	}
	<-j.done
}

// Cancel will stop the current Job in-flight and will remove it from the Task
// queue. Any threads waiting on this Job will return once this function completes.
//
// This does NOT prevent the client Session from running it, but will close
// out all receiving channels and any received data will marked as an un-tracked
// Job.
//
// This is the only method that results in a Status of Canceled.
func (j *Job) Cancel() {
	if j == nil || j.done == nil {
		return
	}
	if j.Status >= StatusCompleted {
		// Something happened and didn't close done.
		if j.done != nil {
			// NOTE(dij): I don't think this will panic, but I need too test to
			//            be 100% sure.
			close(j.done)
		}
		return
	}
	j.s.lock.Lock()
	if j.s.jobs == nil || len(j.s.jobs) == 0 {
		close(j.done)
		j.Status, j.done = StatusCanceled, nil
		// NOTE(dij): We're using the Session Mutex to protect all Jobs since it's
		//            the only non-OOB place we'd cancel em at.
		j.s.lock.Unlock()
		return
	}
	if _, ok := j.s.jobs[j.ID]; !ok {
		close(j.done)
		j.Status, j.done = StatusCanceled, nil
		j.s.lock.Unlock()
		// NOTE(dij): I know this does a lot of work while the Mutex is spinning
		//            but it stays in sync.
		return
	}
	j.s.jobs[j.ID] = nil
	delete(j.s.jobs, j.ID)
	close(j.done)
	j.done = nil
	if j.s.lock.Unlock(); j.Update == nil {
		return
	}
	j.s.m.queue(event{j: j, jf: j.Update})
}

// IsDone returns true when the Job has received a response, has error out or
// was canceled. Use the Status field to determine the state of the Job.
func (j *Job) IsDone() bool {
	if j == nil || j.done == nil {
		return true
	}
	select {
	case <-j.done:
		return true
	default:
	}
	return false
}

// IsError returns true when the Job has received a response, but the response
// is an error.
func (j *Job) IsError() bool {
	if j == nil {
		return false
	}
	if j.IsDone() {
		return len(j.Error) > 0
	}
	return false
}

// Session returns the Session that is associated with this Job.
func (j *Job) Session() *Session {
	return j.s
}
