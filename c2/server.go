package c2

import (
	"context"
	"io"
	"net"
	"strings"
	"time"

	"github.com/PurpleSec/escape"
	"github.com/PurpleSec/logx"
	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Default is the default Server instance. This can be used to directly use a Connector without having to
// setup a Server instance first. This instance will use the 'NOP' logger, as logging is not needed.
var Default = NewServerContext(context.Background(), logx.NOP)

// Server is the manager for all C2 Listener and Sessions connection and states. This struct also manages all
// events and connection changes.
type Server struct {
	New     func(*Session)
	Oneshot func(*com.Packet)

	ch     chan waker
	log    *cout.Log
	ctx    context.Context
	new    chan *Listener
	close  chan string
	events chan event
	cancel context.CancelFunc
	active map[string]*Listener
}

// Wait will block until the current Server is closed and shutdown.
func (s *Server) Wait() {
	<-s.ch
}
func (s *Server) listen() {
	if cout.Enabled {
		s.log.Info("Event processing thread started!")
	}
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
			e.process(s.log)
		}
	}
}
func (s *Server) shutdown() {
	s.cancel()
	for _, v := range s.active {
		v.Close()
	}
	for len(s.active) > 0 {
		delete(s.active, <-s.close)
	}
	if cout.Enabled {
		s.log.Debug("Stopping event processor.")
	}
	s.active = nil
	close(s.new)
	close(s.close)
	close(s.events)
	close(s.ch)
}

// Close stops the processing thread from this Server and releases all associated resources. This will
// signal the shutdown of all attached Listeners and Sessions.
func (s *Server) Close() error {
	s.cancel()
	<-s.ch
	return nil
}

// IsActive returns true if this Server is still able to Process events.
func (s *Server) IsActive() bool {
	return s.ctx.Err() == nil
}

// NewServer creates a new Server instance for managing C2 Listeners and Sessions. If the supplied Log is
// nil, the 'logx.NOP' log will be used.
func NewServer(l logx.Log) *Server {
	return NewServerContext(context.Background(), l)
}

// SetLog will set the internal logger used by the Server and any underlying Listeners, Sessions
// and Proxies. This function is a NOP if the logger is nil or logging is not enabled via the
// 'client' build tag.
func (s *Server) SetLog(l logx.Log) {
	s.log.Set(l)
}

// Connected returns an array of all the current Sessions connected to Listeners connected to this Server.
func (s *Server) Connected() []*Session {
	var l []*Session
	for _, v := range s.active {
		l = append(l, v.Connected()...)
	}
	return l
}

// Listeners returns all the Listeners current active on this Server.
func (s *Server) Listeners() []*Listener {
	l := make([]*Listener, 0, len(s.active))
	for _, v := range s.active {
		l = append(l, v)
	}
	return l
}

// JSON returns the data of this Server as a JSON blob.
func (s *Server) JSON(w io.Writer) error {
	if !cout.Enabled {
		return nil
	}
	if _, err := w.Write([]byte(`{"listeners": {`)); err != nil {
		return err
	}
	i := 0
	for k, v := range s.active {
		if i > 0 {
			if _, err := w.Write([]byte{0x2C}); err != nil {
				return err
			}
		}
		if _, err := w.Write([]byte(escape.JSON(k) + `:`)); err != nil {
			return err
		}
		if err := v.JSON(w); err != nil {
			return err
		}
		i++
	}
	_, err := w.Write([]byte(`}}`))
	return err
}

// Listener returns the lister with the provided name if it exists, nil otherwise.
func (s *Server) Listener(n string) *Listener {
	if len(n) == 0 {
		return nil
	}
	return s.active[n]
}

// GroupContext creates a Session using the supplied Group to connect to the listening server
// specified in the Group. This function will make 'g.Len()' attempts to connect before returning an error.
// This function uses the Default Server instance.
//
// This function version allows for overriting the context passed to the Session.
func GroupContext(g *Group) (*Session, error) {
	return Default.GroupContext(context.Background(), g)
}

// MarshalJSON fulfils the JSON Marshaler interface.
func (s *Server) MarshalJSON() ([]byte, error) {
	if !cout.Enabled {
		return nil, nil
	}
	b := buffers.Get().(*data.Chunk)
	s.JSON(b)
	d := b.Payload()
	returnBuffer(b)
	return d, nil
}

// Session returns the Session that matches the specified Device ID. This function will return nil if
// no matching Device ID is found.
func (s *Server) Session(i device.ID) *Session {
	if i.Empty() {
		return nil
	}
	for _, l := range s.active {
		if x := l.Session(i); x != nil {
			return x
		}
	}
	return nil
}

// Group creates a Session using the supplied Group to connect to the listening server
// specified in the Group. This function will make 'g.Len()' attempts to connect before returning an error.
func (s *Server) Group(g *Group) (*Session, error) {
	return s.GroupContext(s.ctx, g)
}

// NewServerContext creates a new Server instance for managing C2 Listeners and Sessions. If the supplied Log is
// nil, the 'logx.NOP' log will be used. This function will use the supplied Context as the base context for
// cancelation.
func NewServerContext(x context.Context, l logx.Log) *Server {
	s := &Server{
		ch:     make(chan waker, 1),
		log:    cout.New(l),
		new:    make(chan *Listener, 8),
		close:  make(chan string, 16),
		active: make(map[string]*Listener),
		events: make(chan event, maxEvents),
	}
	s.ctx, s.cancel = context.WithCancel(x)
	go s.listen()
	return s
}

// Connect creates a Session using the supplied Profile to connect to the listening server specified. A Session
// will be returned if the connection handshake succeeds.
//
// This function uses the Default Server instance.
func Connect(a string, c Connector, p Profile) (*Session, error) {
	return Default.ConnectContext(Default.ctx, a, c, p)
}

// Shoot sends the packet with the specified data to the server and does NOT register the device with the
// Server. This is used for spending specific data segments in single use connections.
//
// This function uses the Default Server instance.
func Shoot(a string, c Connector, p Profile, d *com.Packet) error {
	return Default.Shoot(a, c, p, d)
}

// Listen adds the Listener under the name provided. A Listener struct to control and receive callback functions
// is added to assist in manageing connections to this Listener.
//
// This function uses the Default Server instance.
func Listen(n, b string, c Accepter, p Profile) (*Listener, error) {
	return Default.ListenContext(Default.ctx, n, b, c, p)
}

// Connect creates a Session using the supplied Profile to connect to the listening server specified. A Session
// will be returned if the connection handshake succeeds.
func (s *Server) Connect(a string, c Connector, p Profile) (*Session, error) {
	return s.ConnectContext(s.ctx, a, c, p)
}

// GroupContext creates a Session using the supplied Group to connect to the listening server
// specified in the Group. This function will make 'g.Len()' attempts to connect before returning an error.
//
// This function version allows for overriting the context passed to the Session.
func (s *Server) GroupContext(x context.Context, g *Group) (*Session, error) {
	if g == nil {
		return nil, ErrNoConnector
	}
	var (
		l   = &Session{ID: device.UUID, Device: *device.Local.Machine}
		n   net.Conn
		err error
	)
	for c := 0; c < g.Len(); c++ {
		if n, err = g.connect(l); err == nil {
			break
		}
	}
	if err != nil {
		return nil, xerr.Wrap("unable to connect", err)
	}
	defer n.Close()
	v := &com.Packet{ID: SvHello, Device: l.ID, Job: uint16(util.FastRand())}
	l.Device.MarshalStream(v)
	if err = writePacket(n, l.w, l.t, v); err != nil {
		return nil, xerr.Wrap("unable to write Packet", err)
	}
	r, err := readPacket(n, l.w, l.t)
	if v.Clear(); err != nil {
		return nil, xerr.Wrap("unable to read Packet", err)
	}
	if r == nil || r.ID != SvComplete {
		return nil, xerr.New("server sent an invalid response")
	}
	if r.Clear(); cout.Enabled {
		s.log.Info("[%s] Client connected to %q!", l.ID, l.host)
	}
	if p := g.profile(); p != nil {
		l.sleep, l.jitter = p.Sleep(), p.Jitter()
	}
	if l.sleep == 0 {
		l.sleep = DefaultSleep
	}
	if l.tick = time.NewTicker(l.sleep); l.jitter > 100 {
		l.jitter = DefaultJitter
	}
	l.ctx, l.conn = x, g
	l.frags = make(map[uint16]*cluster)
	l.log, l.s, l.Mux = s.log, s, DefaultClientMux
	l.wake, l.ch = make(chan waker, 1), make(chan waker, 1)
	l.send, l.recv = make(chan *com.Packet, 256), make(chan *com.Packet, 256)
	go l.listen()
	return l, nil
}

// Shoot sends the packet with the specified data to the server and does NOT register the device with the
// Server. This is used for spending specific data segments in single use connections.
func (s *Server) Shoot(a string, c Connector, p Profile, d *com.Packet) error {
	if p != nil {
		if v, ok := p.(hinter); ok {
			if c == nil {
				c = v.Connector()
			}
			if len(a) == 0 {
				a = v.Host()
			}
		}
	}
	if c == nil {
		return ErrNoConnector
	}
	if len(a) == 0 {
		return ErrNoHost
	}
	var (
		w Wrapper
		t Transform
	)
	if p != nil {
		w, t = p.Wrapper(), p.Transform()
	}
	n, err := c.Connect(a)
	if err != nil {
		return xerr.Wrap("unable to connect to "+a, err)
	}
	if d == nil {
		d = &com.Packet{Device: device.UUID}
	}
	d.Flags |= com.FlagOneshot
	err = writePacket(n, w, t, d)
	if n.Close(); err != nil {
		return xerr.Wrap("unable to write packet", err)
	}
	return nil
}

// Listen adds the Listener under the name provided. A Listener struct to control and receive callback functions
// is added to assist in manageing connections to this Listener.
func (s *Server) Listen(n, b string, c Accepter, p Profile) (*Listener, error) {
	return s.ListenContext(s.ctx, n, b, c, p)
}

// ConnectContext creates a Session using the supplied Profile to connect to the listening server specified.
// This function uses the Default Server instance.
//
// This function version allows for overriting the context passed to the Session.
func ConnectContext(x context.Context, a string, c Connector, p Profile) (*Session, error) {
	return Default.ConnectContext(x, a, c, p)
}

// ListenContext adds the Listener under the name provided. A Listener struct to control and receive callback functions
// is added to assist in manageing connections to this Listener.
// This function uses the Default Server instance.
//
// This function version allows for overriting the context passed to the Session.
func ListenContext(x context.Context, n, b string, c Accepter, p Profile) (*Listener, error) {
	return Default.ListenContext(x, n, b, c, p)
}

// ConnectContext creates a Session using the supplied Profile to connect to the listening server specified.
//
// This function version allows for overriting the context passed to the Session.
func (s *Server) ConnectContext(x context.Context, a string, c Connector, p Profile) (*Session, error) {
	if p != nil {
		if v, ok := p.(hinter); ok {
			if c == nil {
				c = v.Connector()
			}
			if len(a) == 0 {
				a = v.Host()
			}
		}
	}
	if c == nil {
		return nil, ErrNoConnector
	}
	if len(a) == 0 {
		return nil, ErrNoHost
	}
	n, err := c.Connect(a)
	if err != nil {
		return nil, xerr.Wrap("unable to connect to "+a, err)
	}
	defer n.Close()
	var (
		l = &Session{ID: device.UUID, host: a, Device: *device.Local.Machine}
		v = &com.Packet{ID: SvHello, Device: l.ID, Job: uint16(util.FastRand())}
	)
	if p != nil {
		l.sleep, l.jitter = p.Sleep(), p.Jitter()
		l.w, l.t = p.Wrapper(), p.Transform()
	}
	if l.sleep == 0 {
		l.sleep = DefaultSleep
	}
	if l.tick = time.NewTicker(l.sleep); l.jitter > 100 {
		l.jitter = DefaultJitter
	}
	l.Device.MarshalStream(v)
	if err = writePacket(n, l.w, l.t, v); err != nil {
		return nil, xerr.Wrap("unable to write Packet", err)
	}
	r, err := readPacket(n, l.w, l.t)
	if v.Clear(); err != nil {
		return nil, xerr.Wrap("unable to read Packet", err)
	}
	if r == nil || r.ID != SvComplete {
		return nil, xerr.New("server sent an invalid response")
	}
	if r.Clear(); cout.Enabled {
		s.log.Info("[%s] Client connected to %q!", l.ID, a)
	}
	//l.ctx, l.socket = x, c
	l.frags = make(map[uint16]*cluster)
	l.ctx, l.conn = x, static{c: c, h: a}
	l.log, l.s, l.Mux = s.log, s, DefaultClientMux
	l.wake, l.ch = make(chan waker, 1), make(chan waker, 1)
	l.send, l.recv = make(chan *com.Packet, 256), make(chan *com.Packet, 256)
	go l.listen()
	return l, nil
}

// ListenContext adds the Listener under the name provided. A Listener struct to control and receive callback functions
// is added to assist in manageing connections to this Listener.
//
// This function version allows for overriting the context passed to the Session.
func (s *Server) ListenContext(x context.Context, n, b string, c Accepter, p Profile) (*Listener, error) {
	if c == nil && p != nil {
		if v, ok := p.(hinter); ok {
			c = v.Listener()
		}
	}
	if c == nil {
		return nil, ErrNoConnector
	}
	k := strings.ToLower(n)
	if _, ok := s.active[k]; ok {
		return nil, xerr.New("listener " + k + " already exists")
	}
	h, err := c.Listen(b)
	if err != nil {
		return nil, xerr.Wrap("unable to listen on "+b, err)
	}
	if h == nil {
		return nil, xerr.New("unable to listen on " + b + " (error unspecified)")
	}
	l := &Listener{
		ch:         make(chan waker, 1),
		name:       k,
		close:      make(chan uint32, 32),
		sessions:   make(map[uint32]*Session),
		listener:   h,
		connection: connection{s: s, log: s.log},
	}
	if p != nil {
		l.sleep, l.w, l.t = p.Sleep(), p.Wrapper(), p.Transform()
	}
	l.ctx, l.cancel = context.WithCancel(x)
	s.new <- l
	if cout.Enabled {
		s.log.Info("[%s] Added Listener on %q!", x, b)
	}
	go l.listen()
	return l, nil
}
