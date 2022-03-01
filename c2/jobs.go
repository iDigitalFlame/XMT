package c2

import (
	"context"
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
)

// Job is a struct that is used to track and manage Tasks given to Session
// Clients.
//
// This struct has function callbacks that can be used to watch for completion
// and offers a Wait function to pause execution until a response is received.
type Job struct {
	Start, Complete time.Time
	ctx             context.Context

	Result  *com.Packet
	Session *Session

	Update func(*Job)
	cancel context.CancelFunc

	Error              string
	ID, Frags, Current uint16

	Type   uint8
	Status status
}
type status uint8

// Wait will block until the Job is completed or the parent Server is shutdown.
func (j *Job) Wait() {
	if j != nil {
		<-j.ctx.Done()
	}
}

// IsDone returns true when the Job has received a response.
func (j *Job) IsDone() bool {
	if j == nil {
		return true
	}
	select {
	case <-j.ctx.Done():
		return true
	default:
		return false
	}
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
