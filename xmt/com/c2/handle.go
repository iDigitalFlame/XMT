package c2

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/device"
)

var (
	// ErrInvalidPacketID is a error returned inside the client thread when the received packet
	// ID does not match the client ID and does not match any proxy client connected.
	ErrInvalidPacketID = errors.New("received a Packet ID that does not match our own ID")

	// ErrInvalidPacketCount is returned when attempting to read a packet marked
	// as multi or frag an the total count returned is zero.
	ErrInvalidPacketCount = errors.New("frag total is zero on a multi or frag packet")

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
	Mux        Mux
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
	controller *Server
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
					h.controller.events <- &callback{
						session:     h.sessions[v],
						sessionFunc: h.Disconnect,
					}
				}
				h.controller.Log.Trace("[%s] Removing session hash 0x%X.", h.name, v)
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
	h.controller.close <- h.name
}

// Close stops the operation of the Listener associated with
// this Handle and any clients that may be connected. Resources used
// with this Ticket and Listener will be freed up for reuse.
func (h *Handle) Close() error {
	defer func() { recover() }()
	h.cancel()
	err := h.listener.Close()
	close(h.close)
	return err
}
func returnBuffer(i, o *buffer) {
	i.Reset()
	o.Reset()
	buffers.Put(i)
	buffers.Put(o)
}

// String returns the Name of this Handle.
func (h *Handle) String() string {
	return h.name
}

// IsActive returns true if the Listener associated with this
// Handle is still able to send and receive Packets.
func (h *Handle) IsActive() bool {
	return h.ctx.Err() == nil
}

// Remove removes and closes the Session and releases all
// it's associated resources.  This does not close the
// Session on the client's end, use the Shutdown function on
// the Session to shutdown the client process.
func (h *Handle) Remove(i device.ID) {
	h.close <- i.Hash()
}

// Sessions returns an array of all the current Sessions connected
// to this Handle.
func (h *Handle) Sessions() []*Session {
	l := make([]*Session, 0, len(h.sessions))
	for _, v := range h.sessions {
		l = append(l, v)
	}
	return l
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
		h.controller.Log.Trace("[%s:%s] %s: Received an ignore packet.", h.name, p.Device, c.IP())
		return
	}
	if p.Flags&com.FlagOneshot != 0 {
		h.controller.Log.Trace("[%s:%s] %s: Received an oneshot packet.", h.name, p.Device, c.IP())
		process(h, nil, p)
		return
	}
	if p.Flags&com.FlagMulti == 0 || p.Flags&com.FlagMultiDevice == 0 {
		if s := h.client(c, p); s != nil {
			v, err := next(s.send, s.Device.ID)
			if err != nil {
				h.controller.Log.Warning("[%s:%s] %s: Received an error gathering packet data! (%s)", h.name, s.Device.ID, c.IP(), err.Error())
				return
			}
			h.controller.Log.Trace("[%s:%s] %s: Sending Packet \"%s\" to client.", h.name, s.Device.ID, c.IP(), v.String())
			if err := write(c, h.Wrapper, h.Transform, v); err != nil {
				h.controller.Log.Warning("[%s:%s] %s: Received an error writing data to client! (%s)", h.name, s.Device.ID, c.IP(), err.Error())
			}
		}
		return
	}
	n := p.Flags.FragTotal()
	if n == 0 {
		h.controller.Log.Warning("[%s:%s] %s: Received an invalid multi Packet!", h.name, p.Device, c.IP())
		return
	}
	var i, t uint16
	m := &com.Packet{ID: MsgMultiple}
	for ; i < n; i++ {
		v := &com.Packet{}
		if err := v.UnmarshalStream(p); err != nil {
			h.controller.Log.Warning("[%s:%s] %s: Received an error when attempting to read a Packet from \"%s\"! (%s)", h.name, p.Device, c.IP(), err.Error())
			return
		}
		if v.Flags&com.FlagOneshot != 0 {
			h.controller.Log.Trace("[%s:%s] %s: Received an oneshot packet.", h.name, v.Device, c.IP())
			process(h, nil, v)
			continue
		}
		s := h.client(c, v)
		if s == nil {
			continue
		}
		r, err := next(s.send, s.Device.ID)
		if err != nil {
			h.controller.Log.Warning("[%s:%s] %s: Received an error gathering packet data! (%s)", h.name, s.Device.ID, c.IP(), err.Error())
			return
		}
		if err := r.MarshalStream(m); err != nil {
			h.controller.Log.Warning("[%s:%s] %s: Received an error writing data to client buffer! (%s)", h.name, s.Device.ID, c.IP(), err.Error())
			return
		}
		t++
	}
	p.Close()
	m.Close()
	m.Flags.SetFragTotal(t)
	m.Flags = m.Flags | com.FlagMulti | com.FlagMultiDevice
	h.controller.Log.Trace("[%s:%s] %s: Sending Packet \"%s\" to client.", h.name, p.Device, c.IP(), m.String())
	if err := write(c, h.Wrapper, h.Transform, m); err != nil {
		h.controller.Log.Warning("[%s:%s] %s: Received an error writing data to client! (%s)", h.name, p.Device, c.IP(), err.Error())
		return
	}
}

// Session returns the session that matches the specified
// device ID.  This function will return nil if the session matching
// the ID is not found.
func (h *Handle) Session(i device.ID) *Session {
	if i == nil {
		return nil
	}
	if s, ok := h.sessions[i.Hash()]; ok {
		return s
	}
	return nil
}
func processFully(h *Handle, s *Session, p *com.Packet) {
	if h != nil && h.Receive != nil {
		h.controller.events <- &callback{
			packet:     p,
			session:    s,
			packetFunc: h.Receive,
		}
	}
	if s == nil {
		return
	}
	if s.Receive != nil {
		s.controller.events <- &callback{
			packet:     p,
			session:    s,
			packetFunc: s.Receive,
		}
	}
	if len(s.recv) == cap(s.recv) {
		<-s.recv
	}
	s.recv <- p
	if p.ID == MsgShutdown {
		s.send <- p
		switch {
		case p.Job == 1 && s.parent == nil:
			s.controller.Log.Debug("[%s] Server acknowledged shutdown, closing Session.", s.ID)
		case s.parent != nil:
			s.controller.Log.Debug("[%s] Client indicated shutdown, closing Session.", s.ID)
		default:
			s.controller.Log.Debug("[%s] Server indicated shutdown, closing Session.", s.ID)
		}
		s.Close()
		if s.parent != nil {
			s.parent.close <- s.Device.ID.Hash()
		}
		return
	}
	if s.Mux != nil {
		s.controller.events <- &callback{
			packet:     p,
			session:    s,
			packetFunc: s.Mux.Handle,
		}
	} else if s.parent.Mux != nil {
		s.controller.events <- &callback{
			packet:     p,
			session:    s,
			packetFunc: s.parent.Mux.Handle,
		}
	} else if s.controller.Mux != nil {
		s.controller.events <- &callback{
			packet:     p,
			session:    s,
			packetFunc: s.controller.Mux.Handle,
		}
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
			h.controller.events <- &callback{
				packet:     p,
				packetFunc: h.Oneshot,
			}
		}
		if h.Receive != nil {
			h.controller.events <- &callback{
				packet:     p,
				packetFunc: h.Receive,
			}
		}
		return nil
	}
	if s == nil {
		return nil
	}
	if (p.ID == MsgPing || p.ID == MsgHello || p.ID == MsgSleep) && p.Flags&com.FlagData == 0 {
		return nil
	}
	if p.Flags&com.FlagMultiDevice == 0 && s.Update != nil {
		s.controller.events <- &callback{
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
	case p.Flags&com.FlagFrag != 0 && p.Flags&com.FlagMulti == 0:
		if p.Flags.FragTotal() == 0 {
			return ErrInvalidPacketCount
		}
		g := p.Flags.FragGroup()
		c, ok := s.frags[g]
		if !ok {
			s.frags[g] = p
			return nil
		}
		if err := c.Combine(p); err != nil {
			return err
		}
		c.Flags.SetFragPosition(p.Flags.FragPosition() + 1)
		if c.Flags.FragPosition() >= c.Flags.FragTotal() {
			processFully(h, s, c)
			delete(s.frags, g)
		}
		p.ResetFull()
		return nil
	default:
		processFully(h, s, p)
	}
	return nil
}
func (h *Handle) client(c Connection, p *com.Packet) *Session {
	i := p.Device.Hash()
	h.controller.Log.Trace("[%s:%s] %s: Received a packet \"%s\".", h.name, p.Device, c.IP(), p.String())
	s, ok := h.sessions[i]
	if !ok {
		if p.ID != MsgHello {
			h.controller.Log.Warning("[%s:%s] %s: Received a non-hello packet from a unregistered client!", h.name, p.Device, c.IP())
			if err := write(c, h.Wrapper, h.Transform, &com.Packet{ID: MsgRegister}); err != nil {
				h.controller.Log.Warning("[%s:%s] %s: Received an error writing data to client! (%s)", h.name, p.Device, c.IP(), err.Error())
			}
			return nil
		}
		s = &Session{
			ID:         p.Device,
			send:       make(chan *com.Packet, h.size),
			recv:       make(chan *com.Packet, h.size),
			frags:      make(map[uint16]*com.Packet),
			parent:     h,
			Device:     &device.Machine{},
			wrapper:    h.Wrapper,
			Created:    time.Now(),
			transform:  h.Transform,
			controller: h.controller,
		}
		if h.Mux != nil {
			s.Mux = h.Mux
		}
		s.ctx, s.cancel = context.WithCancel(h.ctx)
		h.sessions[i] = s
		h.controller.Log.Debug("[%s:%s] %s: New client registered as \"%s\" hash 0x%X.", h.name, s.ID, c.IP(), s.ID, i)
	}
	s.server = c.IP()
	s.Last = time.Now()
	if p.ID == MsgHello {
		if err := s.Device.UnmarshalStream(p); err != nil {
			h.controller.Log.Warning("[%s:%s] %s: Received an error reading data from client! (%s)", h.name, s.ID, c.IP(), err.Error())
			return nil
		}
		h.controller.Log.Trace("[%s:%s] %s: Received client device info: (OS: %s, %s).", h.name, s.ID, c.IP(), s.Device.OS.String(), s.Device.Version)
		if p.Flags&com.FlagProxy == 0 {
			s.send <- &com.Packet{ID: MsgRegistered, Device: p.Device, Job: p.Job}
		}
		if h.Register != nil {
			h.controller.events <- &callback{
				session:     s,
				sessionFunc: h.Register,
			}
		}
		process(h, s, p)
		return s
	}
	if h.Connect != nil {
		h.controller.events <- &callback{
			session:     s,
			sessionFunc: h.Connect,
		}
	}
	if err := process(h, s, p); err != nil {
		h.controller.Log.Warning("[%s:%s] %s: Received an error processing packet data! (%s)", h.name, s.ID, c.IP(), err.Error())
		return nil
	}
	return s
}
func read(c io.Reader, w Wrapper, t Transform) (*com.Packet, error) {
	v := bufs.Get().([]byte)
	i, o := buffers.Get().(*buffer), buffers.Get().(*buffer)
	defer bufs.Put(v)
	defer returnBuffer(i, o)
	var n int
	var err error
	for {
		n, err = c.Read(v)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if _, err = i.Write(v[:n]); err != nil {
			return nil, err
		}
		if n < len(v) {
			break
		}
	}
	if n == 0 {
		return nil, io.EOF
	}
	i.Close()
	var b wrapBuffer
	if t != nil {
		if err := t.Read(i.buf[:len(i.buf)-i.wpos], o); err != nil {
			return nil, err
		}
		o.Close()
		if o.r, err = w.Unwrap(o); err != nil {
			return nil, err
		}
		b = wrapBuffer(*o)
	} else {
		if i.r, err = w.Unwrap(i); err != nil {
			return nil, err
		}
		b = wrapBuffer(*i)
	}
	if err := b.Close(); err != nil {
		return nil, err
	}

	p := &com.Packet{}
	if err := p.UnmarshalStream(&b); err != nil {
		return nil, err
	}
	i.r, o.r = nil, nil
	return p, nil
}
func write(c io.Writer, w Wrapper, t Transform, p *com.Packet) error {
	i, o := buffers.Get().(*buffer), buffers.Get().(*buffer)
	defer returnBuffer(i, o)
	var err error
	if i.w, err = w.Wrap(i); err != nil {
		return err
	}
	b := wrapBuffer(*i)
	if err := p.MarshalStream(&b); err != nil {
		return err
	}
	if err := b.Close(); err != nil {
		return err
	}
	i.Close()
	i.w, o.w = nil, nil
	if t != nil {
		if err := t.Write(i.buf[:len(i.buf)-i.wpos], o); err != nil {
			return err
		}
		o.Close()
		if _, err := c.Write(o.buf[:len(o.buf)-o.wpos]); err != nil {
			return err
		}
		return nil
	}
	if _, err := c.Write(i.buf[:len(i.buf)-i.wpos]); err != nil {
		return err
	}
	return nil
}
