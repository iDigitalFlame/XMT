//go:build !implant

// Copyright (C) 2020 - 2022 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package c2

import (
	"strconv"
	"sync"
	"time"

	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/c2/task"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// ErrNoTask is returned from some functions that return Jobs. This will
// be returned when the Job object will be nil due to the fact the function
// was called on the client-side instead of the server-side.
//
// This is more of an informational message than an error, as this does NOT
// indicate that the function failed, but that the Job object should NOT be
// used as it is nil. (In case the Job object is not checked.)
var ErrNoTask = xerr.Sub("no Job created for client Session", 0x58)

// Session is a struct that represents a connection between the client and the
// Listener.
//
// This struct does some automatic handling and acts as the communication
// channel between the client and server.
type Session struct {
	lock   sync.RWMutex
	keyNew *data.Key

	Last    time.Time
	Created time.Time
	connection

	swap            Profile
	ch, wake        chan struct{}
	parent          *Listener
	send, recv, chn chan *com.Packet
	frags           map[uint16]*cluster
	jobs            map[uint16]*Job

	Shutdown func(*Session)
	Receive  func(*Session, *com.Packet)
	proxy    *proxyBase
	tick     *time.Ticker
	peek     *com.Packet
	host     container
	proxies  []proxyData

	Device device.Machine
	sleep  time.Duration
	state  state
	key    data.Key

	ID             device.ID
	jitter, errors uint8
}

// Jobs returns all current Jobs for this Session.
//
// This returns nil if there are no Jobs or this Session does not have the
// ability to schedule them.
func (s *Session) Jobs() []*Job {
	if s.jobs == nil || len(s.jobs) == 0 {
		return nil
	}
	s.lock.RLock()
	r := make([]*Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		r = append(r, j)
	}
	s.lock.RUnlock()
	return r
}

// IsClient returns true when this Session is not associated to a Listener on
// this end, which signifies that this session is Client initiated, or we are
// on a client device.
func (s *Session) IsClient() bool {
	return s.parent == nil && s.s == nil
}
func (s *Session) accept(i uint16) {
	if i < 2 || s.parent == nil || s.jobs == nil || len(s.jobs) == 0 {
		return
	}
	s.lock.RLock()
	j, ok := s.jobs[i]
	if s.lock.RUnlock(); !ok {
		return
	}
	if j.Status = StatusAccepted; j.Update != nil {
		s.m.queue(event{j: j, jf: j.Update})
	}
	if cout.Enabled {
		s.log.Trace("[%s] Set JobID %d to accepted.", s.ID, i)
	}
}
func (s *Session) newJobID() uint16 {
	var (
		ok   bool
		i, c uint16
	)
	s.lock.RLock()
	for ; c < 512; c++ {
		i = uint16(util.FastRand())
		if _, ok = s.jobs[i]; !ok && i > 1 {
			s.lock.RUnlock()
			return i
		}
	}
	s.lock.RUnlock()
	return 0
}

// Job returns a Job with the associated ID, if it exists. It returns nil
// otherwise.
func (s *Session) Job(i uint16) *Job {
	if i < 2 || s.jobs == nil || len(s.jobs) == 0 {
		return nil
	}
	s.lock.RLock()
	j := s.jobs[i]
	s.lock.RUnlock()
	return j
}

// Listener will return the Listener that created the Session. This will return
// nil if the session is not on the server side.
func (s *Session) Listener() *Listener {
	return s.parent
}
func (s *Session) handle(p *com.Packet) bool {
	if p == nil || p.Device.Empty() || (p.ID != RvResult && p.ID != RvMigrate) || p.Job < 2 {
		return false
	}
	if s.jobs == nil || len(s.jobs) == 0 {
		if cout.Enabled {
			s.log.Warning("[%s/ShC] Received an un-tracked Job %d!", s.ID, p.Job)
		}
		return false
	}
	if s.state.Moving() {
		if cout.Enabled {
			s.log.Warning("[%s/ShC] Dropping Job %d as Session is being Migrated!", s.ID, p.Job)
		}
		return true
	}
	s.lock.RLock()
	j, ok := s.jobs[p.Job]
	if s.lock.RUnlock(); !ok {
		if cout.Enabled {
			s.log.Warning("[%s/ShC] Received an un-tracked Job %d!", s.ID, p.Job)
		}
		return false
	}
	if cout.Enabled {
		s.log.Info("[%s/ShC] Received response for Job %d.", s.ID, j.ID)
	}
	if j.Result, j.Complete, j.Status = p, time.Now(), StatusCompleted; p.Flags&com.FlagError != 0 {
		j.Status = StatusError
		if err := p.ReadString(&j.Error); err != nil {
			j.Error = err.Error()
		}
	} else if j.Result != nil {
		s.handleInfoResult(j.ID, j.Type, j.Result)
	}
	s.lock.Lock()
	delete(s.jobs, j.ID)
	if s.lock.Unlock(); j.done != nil {
		close(j.done)
		j.done = nil
	}
	if j.Update != nil {
		s.m.queue(event{j: j, jf: j.Update})
	}
	return true
}
func (s *Session) frag(i, id, max, cur uint16) {
	if i < 2 || s.parent == nil || s.jobs == nil || len(s.jobs) == 0 {
		return
	}
	s.lock.RLock()
	j, ok := s.jobs[i]
	if s.lock.RUnlock(); !ok {
		return
	}
	if j.Frags == 0 {
		j.Status = StatusReceiving
	}
	if j.Frags, j.Current = max, cur; j.Update != nil {
		s.m.queue(event{j: j, jf: j.Update})
	}
	if cout.Enabled {
		s.log.Trace("[%s/Frag] Tracking Job %d Frag Group %X, Current %d of %d.", s.ID, i, id, cur+1, max)
	}
}

// SetJitter sets Jitter percentage of the Session's wake interval. This is a 0
// to 100 percentage (inclusive) that will determine any +/- time is added to
// the waiting period. This assists in evading IDS/NDS devices/systems.
//
// A value of 0 will disable Jitter and any value over 100 will set the value to
// 100, which represents using Jitter 100% of the time.
//
// If this is a Server-side Session, the new value will be sent to the Client in
// a MvTime Packet.
func (s *Session) SetJitter(j int) (*Job, error) {
	return s.SetDuration(0, j)
}

// Task is a function that will attach a JobID to the specified Packet (if
// empty) and wil return a Job promise that can be used to internally keep track
// of a response Packet with a matching Job ID.
//
// Errors will be returned if Task is attempted on an invalid Packet, this
// Session is a client-side Session, Job ID is already used or the scheduler is
// full.
func (s *Session) Task(n *com.Packet) (*Job, error) {
	if n == nil {
		return nil, xerr.Sub("empty or nil Job", 0x59)
	}
	if s.parent == nil || s.jobs == nil {
		return nil, xerr.Sub("cannot be a client session", 0x4E)
	}
	if s.isMoving() {
		return nil, xerr.Sub("migration in progress", 0x4F)
	}
	if n.Job == 0 {
		if n.Job = s.newJobID(); n.Job == 0 {
			return nil, xerr.Sub("cannot assign a Job ID", 0x5A)
		}
	}
	if n.Device.Empty() {
		n.Device = s.Device.ID
	}
	s.lock.RLock()
	_, ok := s.jobs[n.Job]
	if s.lock.RUnlock(); ok {
		if xerr.ExtendedInfo {
			return nil, xerr.Sub("job "+strconv.FormatUint(uint64(n.Job), 10)+" already registered", 0x5B)
		}
		return nil, xerr.Sub("job already registered", 0x5B)
	}
	if err := s.write(false, n); err != nil {
		return nil, err
	}
	j := &Job{ID: n.Job, Type: n.ID, Start: time.Now(), s: s, done: make(chan struct{})}
	s.lock.Lock()
	s.jobs[n.Job] = j
	if s.lock.Unlock(); cout.Enabled {
		s.log.Info("[%s/ShC] Added JobID %d to Track!", s.ID, n.Job)
	}
	return j, nil
}
func (s *Session) setProfile(b []byte) (*Job, error) {
	if s.parent == nil {
		return nil, ErrNoTask
	}
	n := &com.Packet{ID: task.MvProfile, Device: s.Device.ID}
	n.WriteBytes(b)
	return s.Task(n)
}

// SetProfile will set the Profile used by this Session. This function will
// ensure that the profile is marshalable before setting and will then pass it
// to be set by the client Session (if this isn't one already).
//
// If this is a server-side Session, this will trigger the sending of a MvProfile
// Packet to update the client-side instance, which will update on it's next
// wakeup cycle.
//
// If this is a client-side session the error 'ErrNoTask' will be returned AFTER
// setting the Profile and indicates that no Packet will be sent and that the
// Job object result is nil.
func (s *Session) SetProfile(p Profile) (*Job, error) {
	if p == nil {
		return nil, ErrInvalidProfile
	}
	m, ok := p.(marshaler)
	if !ok {
		return nil, xerr.Sub("cannot marshal Profile", 0x50)
	}
	b, err := m.MarshalBinary()
	if err != nil {
		return nil, xerr.Wrap("cannot marshal Profile", err)
	}
	s.p = p
	return s.setProfile(b)
}

// Tasklet is a function similar to Task and will attach a JobID to the specified
// Packet created by the supplied Tasklet and wil return a Job promise that can be
// used to internally keep track of a response Packet with a matching Job ID.
//
// If the Tasklet has an issue generating the payload, it will return an error
// before scheduling.
//
// Errors will be returned if Task is attempted on an invalid Packet, this Session
// is a client-side Session, Job ID is already or the scheduler is full.
func (s *Session) Tasklet(t task.Tasklet) (*Job, error) {
	if t == nil {
		return nil, xerr.Sub("empty or nil Tasklet", 0x5C)
	}
	n, err := t.Packet()
	if err != nil {
		return nil, err
	}
	return s.Task(n)
}

// SetSleep sets the wake interval period for this Session. This is the time value
// between connections to the C2 Server.
//
// If this is a Server-side Session, the new value will be sent to the Client in
// a MvTime Packet. This setting does not affect Jitter.
func (s *Session) SetSleep(t time.Duration) (*Job, error) {
	return s.SetDuration(t, -1)
}

// SetProfileBytes will set the Profile used by this Session. This function will
// unmarshal and set the server-side before setting and will then pass it to be
// set by the client Session (if this isn't one already).
//
// If this is a server-side Session, this will trigger the sending of a MvProfile
// Packet to update the client-side instance, which will update on it's next
// wakeup cycle.
//
// This function will fail if no ProfileParser is set.
//
// If this is a client-side session the error 'ErrNoTask' will be returned AFTER
// setting the Profile and indicates that no Packet will be sent and that the
// Job object result is nil.
func (s *Session) SetProfileBytes(b []byte) (*Job, error) {
	if ProfileParser == nil {
		return nil, xerr.Sub("no Profile parser loaded", 0x44)
	}
	p, err := ProfileParser(b)
	if err != nil {
		return nil, xerr.Wrap("parse Profile", err)
	}
	s.p = p
	return s.setProfile(b)
}

// SetDuration sets the wake interval period and Jitter for this Session. This is
// the time value between connections to the C2 Server.
//
// Jitter is a 0 to 100 percentage (inclusive) that will determine any +/- time
// is added to the waiting period. This assists in evading IDS/NDS devices/systems.
//
// A value of 0 will disable Jitter and any value over 100 will set the value to
// 100, which represents using Jitter 100% of the time.
//
// If this is a Server-side Session, the new value will be sent to the Client in
// a MvTime Packet.
func (s *Session) SetDuration(t time.Duration, j int) (*Job, error) {
	switch {
	case j == -1:
	case j < 0:
		s.jitter = 0
	case j > 100:
		s.jitter = 100
	default:
		s.jitter = uint8(j)
	}
	if t > 0 {
		s.sleep = t
	}
	// NOTE(dij): This may cause a de-sync issue when combined with a smaller
	//            initial timeout only on channels.
	//            (Just the bail below)
	if s.parent == nil {
		return nil, ErrNoTask
	}
	n := &com.Packet{ID: task.MvTime, Device: s.Device.ID}
	n.WriteUint8(s.jitter)
	n.WriteUint64(uint64(s.sleep))
	return s.Task(n)
}
func (s *Session) handleInfoResult(i uint16, t uint8, n *com.Packet) {
	// Handle server-side updates to info
	switch t {
	case task.MvTime:
		if err := n.ReadUint8(&s.jitter); err != nil {
			if cout.Enabled {
				s.log.Warning("[%s/ShC] Error reading MvTime Job %d jitter: %s!", s.ID, i, err.Error())
			}
			break
		}
		if err := n.ReadInt64((*int64)(&s.sleep)); err != nil {
			if cout.Enabled {
				s.log.Warning("[%s/ShC] Error reading MvTime Job %d sleep: %s!", s.ID, i, err.Error())
			}
		}
	case task.MvRefresh:
		if err := s.Device.UnmarshalStream(n); err != nil {
			if cout.Enabled {
				s.log.Warning("[%s/ShC] Error reading MvRefresh Job %d Machine info: %s!", s.ID, i, err.Error())
			}
		}
	default:
		return
	}
	n.Seek(0, 0)
}
