package c2

import (
	"context"
	"io"
	"strconv"
	"time"

	"github.com/PurpleSec/escape"
	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
)

// These are status values that indicate the general status of the Job.
const (
	StatusWaiting  status = 0
	StatusAccepted status = iota
	StatusReceiving
	StatusCompleted
	StatusError
)

// Job is a struct that is used to track and manage Tasks given to Session Clients. This struct has function callbacks
// that can be used to watch for completion and also offers a Wait function to pause execution until a response is received.
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
	if j.IsDone() {
		return len(j.Error) > 0
	}
	return false
}
func (s status) String() string {
	if !cout.Enabled {
		return ""
	}
	switch s {
	case StatusError:
		return "error"
	case StatusWaiting:
		return "waiting"
	case StatusAccepted:
		return "accepted"
	case StatusReceiving:
		return "receiving"
	case StatusCompleted:
		return "completed"
	}
	return "invalid"
}

// JSON returns the data of this Job as a JSON blob.
func (j *Job) JSON(w io.Writer) error {
	if !cout.Enabled {
		return nil
	}
	if _, err := w.Write([]byte(`{"id":` + strconv.Itoa(int(j.ID)) + `,` +
		`"type":` + strconv.Itoa(int(j.Type)) + `,` +
		`"error":` + escape.JSON(j.Error) + `,` +
		`"status":"` + j.Status.String() + `",` +
		`"start":"` + j.Start.Format(time.RFC3339) + `"`,
	)); err != nil {
		return err
	}
	if j.Session != nil && !j.Session.ID.Empty() {
		if _, err := w.Write([]byte(`,"host":"` + j.Session.ID.String() + `"`)); err != nil {
			return err
		}
	}
	if !j.Complete.IsZero() {
		if _, err := w.Write([]byte(`,"complete":"` + j.Complete.Format(time.RFC3339) + `"`)); err != nil {
			return err
		}
	}
	if j.Result != nil {
		if _, err := w.Write([]byte(`,"result":` + strconv.Itoa(j.Result.Size()))); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte{'}'})
	return err
}

// MarshalJSON fulfils the JSON Marshaler interface.
func (j *Job) MarshalJSON() ([]byte, error) {
	if !cout.Enabled {
		return nil, nil
	}
	b := buffers.Get().(*data.Chunk)
	j.JSON(b)
	d := b.Payload()
	returnBuffer(b)
	return d, nil
}
