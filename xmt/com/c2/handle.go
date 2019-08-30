package c2

import (
	"context"
	"hash"
	"hash/fnv"
	"io"
	"sync"
	"time"

	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/device"
)

var (
	bufs = &sync.Pool{
		New: func() interface{} {
			return make([]byte, DefaultBufferSize)
		},
	}
	buffers = &sync.Pool{
		New: func() interface{} {
			return &buffer{
				buf:  make([]byte, 0, DefaultBufferSize),
				rbuf: make([]byte, 8),
			}
		},
	}
	hashers = &sync.Pool{
		New: func() interface{} {
			return fnv.New32()
		},
	}

	pingPacket     = &com.Packet{ID: PacketPing}
	sleepPacket    = &com.Packet{ID: PacketSleep}
	registerPacket = &com.Packet{ID: PacketRegistered}
)

// Handle is a struct that is passed back when a Listener
// is added to the Controller.  The Handle allows for controlling
// the listener and setting callback functions to be used when a client
// connect, registers or disconnects.
type Handle struct {
	Size       int
	Wrapper    Wrapper
	Receive    func(*Session, *com.Packet)
	Connect    func(*Session)
	Register   func(*Session)
	Sessions   map[uint32]*Session
	Transport  Transport
	Disconnect func(*Session)

	ctx      context.Context
	hasher   hash.Hash32
	cancel   context.CancelFunc
	listener Listener
}
type buffer struct {
	w    io.WriteCloser
	r    io.ReadCloser
	buf  []byte
	rbuf []byte
	rpos int
	wpos int
}
type wrapBuffer buffer

func returnBuffer(b *buffer) {
	b.Reset()
	buffers.Put(b)
}

// Close stops the operation of the Listener associated with
// this Handle and any clients that may be connected. Resources used
// with this Ticket and Listener will be freed up for reuse.
func (h *Handle) Close() error {
	h.cancel()
	return h.listener.Close()
}
func (h *Handle) listen(c *controller) {
	c.Log.Trace("[%s] Starting listen...", h.listener.String())
	for h.ctx.Err() == nil {
		n, err := h.listener.Accept()
		if err != nil {
			c.Log.Error("[%s] Received error during listener operation! (%s)", h.listener.String(), err.Error())
			break
		}
		c.Log.Trace("[%s] Received a connected from \"%s\"...", h.listener.String(), n.IP())
		go h.client(c, n)
	}
	c.Log.Debug("[%s] Stopping listen...", h.listener.String())
	h.Close()
}
func (h *Handle) client(a *controller, c Connection) {
	defer c.Close()
	p, err := read(c, h.Wrapper, h.Transport)
	if err != nil || p == nil || p.Empty() {
		a.Log.Warning("[%s] Received an error when attempting to read a Packet from \"%s\"! (%s)", h.listener.String(), c.IP(), err.Error())
		return
	}
	x := hashers.Get().(hash.Hash32)
	x.Reset()
	x.Write(p.Device)
	i := x.Sum32()
	hashers.Put(x)
	a.Log.Trace("[%s] Received a packet \"%s\" from \"%s\" (\"%s\") hash \"%X\".", h.listener.String(), p.String(), p.Device.ID(), c.IP(), i)
	s, ok := h.Sessions[i]
	if !ok && p.ID != PacketHello {
		a.Log.Warning("[%s] Received a non-hello packet from non-registered client \"%s\"!", h.listener.String(), c.IP())
		return
	}
	if !ok {
		s = &Session{
			ID:        p.Device[device.MachineIDSize:],
			Host:      &device.Machine{},
			send:      make(chan *com.Packet, h.Size),
			recv:      make(chan *com.Packet, h.Size),
			parent:    h,
			wrapper:   h.Wrapper,
			Created:   time.Now(),
			transport: h.Transport,
		}
		s.ctx, s.cancel = context.WithCancel(h.ctx)
		h.Sessions[i] = s
		a.Log.Debug("[%s] New client \"%s\" (\"%s\") registered as \"%s\"!", h.listener.String(), p.Device.ID(), c.IP(), s.ID)
	}
	s.server = c.IP()
	s.Last = time.Now()
	if p.ID == PacketHello {
		if err := s.Host.UnmarshalStream(p); err != nil {
			a.Log.Warning("[%s] Received an error reading data from client \"%s\"! (%s)", h.listener.String(), c.IP(), err.Error())
			return
		}
		a.Log.Trace("[%s] Received client \"%s\" (\"%s\") device info! OS: %s, %s", h.listener.String(), s.Host.ID.ID(), c.IP(), s.Host.OS.String(), s.Host.Version)
		if err := write(c, h.Wrapper, h.Transport, registerPacket); err != nil {
			a.Log.Warning("[%s] Received an error writing data to client \"%s\"! (%s)", h.listener.String(), c.IP(), err.Error())
			return
		}
		if h.Register != nil {
			h.Register(s)
		}
		return
	}
	if s.Receive != nil {
		s.Receive(p)
	}
	if h.Receive != nil {
		h.Receive(s, p)
	}
	if len(s.recv) == cap(s.recv) {
		<-s.recv
	}
	s.recv <- p
	v := sleepPacket
	if len(s.send) > 0 {
		v = <-s.send
	}
	a.Log.Trace("[%s] Sending Packet \"%s\" to client \"%s\".", h.listener.String(), v.String(), c.IP())
	if err := write(c, h.Wrapper, h.Transport, v); err != nil {
		a.Log.Warning("[%s] Received an error writing data to client \"%s\"! (%s)", h.listener.String(), c.IP(), err.Error())
		return
	}
}
func read(c io.Reader, w Wrapper, t Transport) (*com.Packet, error) {
	b := buffers.Get().(*buffer)
	defer returnBuffer(b)
	var n int
	var err error
	v := bufs.Get().([]byte)
	defer bufs.Put(v)
	b.Reset()
	for {
		n, err = c.Read(v)
		if err != nil {
			return nil, err
		}
		if _, err = b.Write(v[:n]); err != nil {
			return nil, err
		}
		if n < len(v) {
			break
		}
	}
	if n == 0 {
		return nil, io.EOF
	}
	b.buf, err = t.Read(b.buf[:len(b.buf)-b.wpos])
	if err != nil {
		return nil, err
	}
	b.r = w.Unwrap(b)
	p := &com.Packet{}
	q := wrapBuffer(*b)
	b.Close()
	if err := p.UnmarshalStream(&q); err != nil {
		return nil, err
	}
	if err := q.Close(); err != nil {
		return nil, err
	}
	b.r = nil
	return p, nil
}
func write(c io.Writer, w Wrapper, t Transport, p *com.Packet) error {
	b := buffers.Get().(*buffer)
	defer returnBuffer(b)
	b.w = w.Wrap(b)
	q := wrapBuffer(*b)
	if err := p.MarshalStream(&q); err != nil {
		return err
	}
	if err := q.Close(); err != nil {
		return err
	}
	o, err := t.Write(b.buf[:len(b.buf)-b.wpos])
	if err != nil {
		return err
	}
	if _, err := c.Write(o); err != nil {
		return err
	}
	b.w = nil
	return nil
}
