package c2

import (
	"bytes"
	"context"
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
)

// Handle is a struct that is passed back when a Listener
// is added to the Controller.  The Handle allows for controlling
// the listener and setting callback functions to be used when a client
// connect, registers or disconnects.
type Handle struct {
	Wrapper    Wrapper
	Oneshot    func(*Session, *com.Packet)
	Receive    func(*Session, *com.Packet)
	Connect    func(*Session)
	Register   func(*Session)
	Transform  Transform
	Disconnect func(*Session)

	ctx        context.Context
	name       string
	size       int
	close      chan uint32
	cancel     context.CancelFunc
	sessions   map[uint32]*Session
	listener   Listener
	controller *controller
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

// Wait will block until the current Listener associated with
// this Handle is closed and shutdown.
func (h *Handle) Wait() {
	<-h.ctx.Done()
}
func (h *Handle) listen() {
	h.controller.Log.Trace("[%s] Starting listen \"%s\"...", h.name, h.listener.String())
	for h.ctx.Err() == nil {
		if len(h.close) > 0 {
			for x := 0; x < len(h.close); x++ {
				v := <-h.close
				if h.Disconnect != nil {
					h.controller.events <- &eventCallback{
						session:     h.sessions[v],
						sessionFunc: h.Disconnect,
					}
				}
				delete(h.sessions, v)
			}
		}
		c, err := h.listener.Accept()
		if err != nil {
			h.controller.Log.Error("[%s] Received error during listener operation! (%s)", h.name, err.Error())
			if h.ctx.Err() != nil {
				break
			}
		}
		if c == nil {
			continue
		}
		h.controller.Log.Trace("[%s] Received a connection from \"%s\"...", h.name, c.IP())
		go h.session(c)
	}
	h.controller.Log.Debug("[%s] Stopping listen...", h.name)
	h.Close()
	for _, v := range h.sessions {
		v.Close()
	}
}
func returnBuffer(b *buffer) {
	b.Reset()
	buffers.Put(b)
}

func (h *Handle) Remove(i device.ID) {
	h.close <- i.Hash()
}
func (h *Handle) Sessions() []*Session {
	l := make([]*Session, 0, len(h.sessions))
	for _, v := range h.sessions {
		l = append(l, v)
	}
	return l
}
func (h *Handle) Session(i device.ID) *Session {
	if i == nil {
		return nil
	}
	if s, ok := h.sessions[i.Hash()]; ok {
		return s
	}
	return nil
}

// Close stops the operation of the Listener associated with
// this Handle and any clients that may be connected. Resources used
// with this Ticket and Listener will be freed up for reuse.
func (h *Handle) Close() error {
	defer func() { recover() }()
	if h.ctx.Err() == nil {
	}
	h.cancel()
	err := h.listener.Close()
	close(h.close)
	return err
}

// IsActive returns true if the Listener associated with this
// Handle is still able to send and receive Packets.
func (h *Handle) IsActive() bool {
	return h.ctx.Err() == nil
}
func (h *Handle) session(c Connection) {
	defer c.Close()
	p, err := read(c, h.Wrapper, h.Transform)
	if err != nil {
		h.controller.Log.Warning("[%s] %s: Received an error when attempting to read a Packet! (%s)", h.name, c.IP(), err.Error())
		return
	}
	if p == nil || p.IsEmpty() {
		h.controller.Log.Warning("[%s] %s: Received an empty or invalid Packet!", h.name, c.IP())
		return
	}
	if p.Flags&com.FlagIgnore != 0 {
		h.controller.Log.Trace("[%s:%s] %s: Received an ignore packet.", h.name, p.Device.ID(), c.IP())
		return
	}
	if p.Flags&com.FlagOneshot != 0 {
		h.controller.Log.Trace("[%s:%s] %s: Received an oneshot packet.", h.name, p.Device.ID(), c.IP())
		process(h, nil, p)
		return
	}
	if p.Flags&com.FlagMulti == 0 || p.Flags&com.FlagMultiDevice == 0 {
		if s := h.client(c, p); s != nil {
			v, err := next(s.send, s.Device.ID)
			if err != nil {
				h.controller.Log.Warning("[%s:%s] %s: Received an error gathering packet data! (%s)", h.name, s.Device.ID.ID(), c.IP(), err.Error())
				return
			}
			h.controller.Log.Trace("[%s:%s] %s: Sending Packet \"%s\" to client.", h.name, s.Device.ID.ID(), c.IP(), v.String())
			if err := write(c, h.Wrapper, h.Transform, v); err != nil {
				h.controller.Log.Warning("[%s:%s] %s: Received an error writing data to client! (%s)", h.name, s.Device.ID.ID(), c.IP(), err.Error())
			}
		}
		return
	}
	n := p.Flags.FragTotal()
	if n == 0 {
		h.controller.Log.Warning("[%s:%s] %s: Received an invalid multi Packet!", h.name, p.Device.ID(), c.IP())
		return
	}
	var i, t uint16
	m := &com.Packet{ID: PacketMultiple}
	for ; i < n; i++ {
		v := &com.Packet{}
		if err := v.UnmarshalStream(p); err != nil {
			h.controller.Log.Warning("[%s:%s] %s: Received an error when attempting to read a Packet from \"%s\"! (%s)", h.name, p.Device.ID(), c.IP(), err.Error())
			return
		}
		if v.Flags&com.FlagOneshot != 0 {
			h.controller.Log.Trace("[%s:%s] %s: Received an oneshot packet.", h.name, v.Device.ID(), c.IP())
			process(h, nil, v)
			continue
		}
		s := h.client(c, v)
		if s == nil {
			continue
		}
		r, err := next(s.send, s.Device.ID)
		if err != nil {
			h.controller.Log.Warning("[%s:%s] %s: Received an error gathering packet data! (%s)", h.name, s.Device.ID.ID(), c.IP(), err.Error())
			return
		}
		if err := r.MarshalStream(m); err != nil {
			h.controller.Log.Warning("[%s:%s] %s: Received an error writing data to client buffer! (%s)", h.name, s.Device.ID.ID(), c.IP(), err.Error())
			return
		}
		t++
	}
	p.Close()
	m.Close()
	m.Flags.SetFragTotal(t)
	m.Flags = m.Flags | com.FlagMulti | com.FlagMultiDevice
	h.controller.Log.Trace("[%s:%s] %s: Sending Packet \"%s\" to client.", h.name, p.Device.ID(), c.IP(), m.String())
	if err := write(c, h.Wrapper, h.Transform, m); err != nil {
		h.controller.Log.Warning("[%s:%s] %s: Received an error writing data to client! (%s)", h.name, p.Device.ID(), c.IP(), err.Error())
		return
	}
}
func processFully(h *Handle, s *Session, p *com.Packet) {
	if h != nil && h.Receive != nil {
		h.controller.events <- &eventCallback{
			packet:     p,
			session:    s,
			packetFunc: h.Receive,
		}
	}
	if s == nil {
		return
	}
	if s.Receive != nil {
		s.controller.events <- &eventCallback{
			packet:     p,
			session:    s,
			packetFunc: s.Receive,
		}
	}
	if len(s.recv) == cap(s.recv) {
		<-s.recv
	}
	s.recv <- p
	if p.ID == PacketShutdown {
		s.controller.Log.Debug("[%s] Client indicated shutdown, closing Session.", s.ID.ID())
		s.Close()
	}
}
func process(h *Handle, s *Session, p *com.Packet) error {
	if h == nil && s == nil {
		return nil
	}
	if p == nil || p.IsEmpty() || p.Flags&com.FlagIgnore != 0 || p.Device == nil {
		return nil
	}
	if s != nil && !bytes.Equal(p.Device, s.Device.ID) && p.Flags&com.FlagMultiDevice == 0 {
		if s.proxies == nil {
			return ErrInvalidPacketID
		}
		if c, ok := s.proxies[p.Device.Hash()]; ok {
			c.send <- p
		}
		return nil
	}
	if h != nil && p.Flags&com.FlagOneshot != 0 {
		if h.Oneshot != nil {
			h.controller.events <- &eventCallback{
				packet:     p,
				packetFunc: h.Oneshot,
			}
		}
		if h.Receive != nil {
			h.controller.events <- &eventCallback{
				packet:     p,
				packetFunc: h.Receive,
			}
		}
		return nil
	}
	if s == nil {
		return nil
	}
	if (p.ID == PacketPing || p.ID == PacketHello || p.ID == PacketSleep) && p.Flags&com.FlagData == 0 {
		return nil
	}
	if p.Flags&com.FlagMultiDevice == 0 && s.Update != nil {
		s.controller.events <- &eventCallback{
			session:     s,
			sessionFunc: s.Update,
		}
	}
	switch {
	case p.Flags&com.FlagData != 0:
		v := &com.Packet{}
		if err := v.UnmarshalStream(p); err != nil {
			return err
		}
		process(h, s, v)
		return p.Close()
	case p.Flags&com.FlagFrag != 0:
		// Work on frag handeling...
	case p.Flags&com.FlagMulti != 0:
		n := p.Flags.FragTotal()
		if n == 0 {
			return ErrInvalidPacketCount
		}
		for i := uint16(0); i < n; i++ {
			v := &com.Packet{}
			if err := v.UnmarshalStream(p); err != nil {
				return err
			}
			process(h, s, v)
		}
		return p.Close()
	default:
		processFully(h, s, p)
	}
	return nil
}
func (h *Handle) client(c Connection, p *com.Packet) *Session {
	i := p.Device.Hash()
	h.controller.Log.Trace("[%s:%s] %s: Received a packet \"%s\".", h.name, p.Device.ID(), c.IP(), p.String())
	s, ok := h.sessions[i]
	if !ok {
		if p.ID != PacketHello {
			h.controller.Log.Warning("[%s:%s] %s: Received a non-hello packet from a unregistered client!", h.name, p.Device.ID(), c.IP())
			return nil
		}
		s = &Session{
			ID:         p.Device[device.MachineIDSize:],
			send:       make(chan *com.Packet, h.size),
			recv:       make(chan *com.Packet, h.size),
			parent:     h,
			Device:     &device.Machine{},
			wrapper:    h.Wrapper,
			Created:    time.Now(),
			transform:  h.Transform,
			controller: h.controller,
		}
		s.ctx, s.cancel = context.WithCancel(h.ctx)
		h.sessions[i] = s
		h.controller.Log.Debug("[%s:%s] %s: New client registered as \"%s\" hash \"%X\".", h.name, p.Device.ID(), c.IP(), s.ID, i)
	}
	s.server = c.IP()
	s.Last = time.Now()
	if p.ID == PacketHello {
		if err := s.Device.UnmarshalStream(p); err != nil {
			h.controller.Log.Warning("[%s:%s] %s: Received an error reading data from client! (%s)", h.name, p.Device.ID(), c.IP(), err.Error())
			return nil
		}
		h.controller.Log.Trace("[%s:%s] %s: Received client device info: (OS: %s, %s).", h.name, s.Device.ID.ID(), c.IP(), s.Device.OS.String(), s.Device.Version)
		if p.Flags&com.FlagProxy == 0 {
			if err := write(c, h.Wrapper, h.Transform, &com.Packet{ID: PacketRegistered, Device: p.Device}); err != nil {
				h.controller.Log.Warning("[%s:%s] %s: Received an error writing data to client! [%s]", h.name, s.Device.ID.ID(), c.IP(), err.Error())
				return nil
			}
		}
		if h.Register != nil {
			h.controller.events <- &eventCallback{
				session:     s,
				sessionFunc: h.Register,
			}
		}
		process(h, s, p)
		return s
	}
	if h.Connect != nil {
		h.controller.events <- &eventCallback{
			session:     s,
			sessionFunc: h.Connect,
		}
	}
	if err := process(h, s, p); err != nil {
		h.controller.Log.Warning("[%s:%s] %s: Received an error processing packet data! (%s)", h.name, s.Device.ID.ID(), c.IP(), err.Error())
		return nil
	}
	return s
}
func read(c io.Reader, w Wrapper, t Transform) (*com.Packet, error) {
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
	b.Close()
	b.r = w.Unwrap(b)
	p := &com.Packet{}
	q := wrapBuffer(*b)
	if err := p.UnmarshalStream(&q); err != nil {
		return nil, err
	}
	if err := q.Close(); err != nil {
		return nil, err
	}
	b.r = nil
	return p, nil
}
func write(c io.Writer, w Wrapper, t Transform, p *com.Packet) error {
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
