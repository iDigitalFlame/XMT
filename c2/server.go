//go:build !implant

package c2

import (
	"context"
	"strings"
	"sync"

	"github.com/PurpleSec/logx"
	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Server is the manager for all C2 Listener and Sessions connection and states.
// This struct also manages all events and connection changes.
type Server struct {
	New      func(*Session)
	Oneshot  func(*com.Packet)
	Shutdown func(*Session)

	ch   chan struct{}
	log  *cout.Log
	ctx  context.Context
	new  chan *Listener
	lock sync.RWMutex
	init sync.Once

	delSession  chan uint32
	delListener chan string

	events   chan event
	cancel   context.CancelFunc
	active   map[string]*Listener
	sessions map[uint32]*Session
}

// Wait will block until the current Server is closed and shutdown.
func (s *Server) Wait() {
	if s.ch == nil {
		return
	}
	<-s.ch
}
func (s *Server) listen() {
	if bugtrack.Enabled {
		defer bugtrack.Recover("c2.Server.listen()")
	}
	if cout.Enabled {
		s.log.Info("Server-side event processing thread started!")
	}
	for s.ch = make(chan struct{}); ; {
		select {
		case <-s.ctx.Done():
			s.shutdown()
			return
		case l := <-s.new:
			s.active[l.name] = l
		case r := <-s.delListener:
			delete(s.active, r)
		case i := <-s.delSession:
			s.lock.Lock()
			if v, ok := s.sessions[i]; ok {
				if s.Shutdown != nil {
					s.queue(event{s: v, sf: s.Shutdown})
				}
				if delete(s.sessions, i); cout.Enabled {
					s.log.Info("[%s] Removed closed Session 0x%X.", v.parent.name, i)
				}
			}
			s.lock.Unlock()
		case e := <-s.events:
			e.process(s.log)
		}
	}
}
func (s *Server) shutdown() {
	s.cancel()
	for _, v := range s.sessions {
		v.Close()
	}
	for _, v := range s.active {
		v.Close()
	}
	for len(s.active) > 0 {
		delete(s.active, <-s.delListener)
	}
	if cout.Enabled {
		s.log.Debug("Stopping event processor.")
	}
	s.active = nil
	close(s.new)
	close(s.delListener)
	close(s.delSession)
	if close(s.events); s.ch != nil {
		close(s.ch)
	}
}

// Close stops the processing thread from this Server and releases all associated
// resources.
//
// This will signal the shutdown of all attached Listeners and Sessions.
func (s *Server) Close() error {
	if s.cancel(); s.ch != nil {
		<-s.ch
	}
	return nil
}
func (s *Server) queue(e event) {
	s.events <- e
}

// IsActive returns true if this Server is still able to Process events.
func (s *Server) IsActive() bool {
	select {
	case <-s.ch:
		return false
	case <-s.ctx.Done():
		return false
	default:
		return true
	}
}

// NewServer creates a new Server instance for managing C2 Listeners and Sessions.
//
// If the supplied Log is nil, the 'logx.NOP' log will be used.
func NewServer(l logx.Log) *Server {
	return NewServerContext(context.Background(), l)
}

// SetLog will set the internal logger used by the Server and any underlying
// Listeners, Sessions and Proxies.
//
// This function is a NOP if the logger is nil or logging is not enabled via the
// 'implant' build tag.
func (s *Server) SetLog(l logx.Log) {
	s.log.Set(l)
}

// Sessions returns an array of all the current Sessions connected to Listeners
// running on this Server instance.
func (s *Server) Sessions() []*Session {
	s.lock.RLock()
	l := make([]*Session, 0, len(s.sessions))
	for _, v := range s.sessions {
		l = append(l, v)
	}
	s.lock.RUnlock()
	return l
}

// Done returns a channel that's closed when this Server is closed.
//
// This can be used to monitor a Server's status using a select statement.
func (s *Server) Done() <-chan struct{} {
	return s.ch
}

// Listeners returns all the Listeners current active on this Server.
func (s *Server) Listeners() []*Listener {
	l := make([]*Listener, 0, len(s.active))
	for _, v := range s.active {
		l = append(l, v)
	}
	return l
}

// Listener returns the lister with the provided name if it exists, nil
// otherwise.
func (s *Server) Listener(n string) *Listener {
	if len(n) == 0 {
		return nil
	}
	return s.active[n]
}

// Session returns the Session that matches the specified Device ID.
//
// This function  will return nil if no matching Device ID is found.
func (s *Server) Session(i device.ID) *Session {
	if i.Empty() {
		return nil
	}
	s.lock.RLock()
	v := s.sessions[i.Hash()]
	s.lock.RUnlock()
	return v
}

// Remove removes and closes the Session and releases all it's associated
// resources from this server instance.
//
// If shutdown is false, this does not close the Session on the client's end and
// will just remove the entry, but can be re-added and if the client connects
// again.
//
// If shutdown is true, this will trigger a Shutdown packet to be sent to close
// down the client and will wait until the client acknowledges the shutdown
// request before removing.
func (s *Server) Remove(i device.ID, shutdown bool) {
	if !shutdown {
		if !s.IsActive() {
			return
		}
		s.delSession <- i.Hash()
		return
	}
	if !s.IsActive() {
		return
	}
	s.lock.RLock()
	v, ok := s.sessions[i.Hash()]
	if s.lock.RUnlock(); !ok {
		return
	}
	v.Close()
}

// NewServerContext creates a new Server instance for managing C2 Listeners and
// Sessions.
//
// If the supplied Log is nil, the 'logx.NOP' log will be used.
//
// This function will use the supplied Context as the base context for
// cancelation.
func NewServerContext(x context.Context, l logx.Log) *Server {
	s := &Server{
		log:         cout.New(l),
		new:         make(chan *Listener, 4),
		active:      make(map[string]*Listener),
		events:      make(chan event, maxEvents),
		sessions:    make(map[uint32]*Session),
		delSession:  make(chan uint32, 64),
		delListener: make(chan string, 16),
	}
	s.ctx, s.cancel = context.WithCancel(x)
	return s
}

// Listen adds the Listener under the name provided. A Listener struct to
// control and receive callback functions is added to assist in managing
// connections to this Listener.
func (s *Server) Listen(name, addr string, p Profile) (*Listener, error) {
	return s.ListenContext(s.ctx, name, addr, p)
}

// ListenContext adds the Listener under the name and address provided. A Listener
// struct to control and receive callback functions is added to assist in managing
// connections to this Listener.
//
// This function version allows for overriting the Context passed to the Session.
func (s *Server) ListenContext(x context.Context, name, addr string, p Profile) (*Listener, error) {
	if p == nil {
		return nil, ErrInvalidProfile
	}
	if len(name) == 0 {
		return nil, xerr.Sub("empty Listener name", 0x1B)
	}
	n := strings.ToLower(name)
	if _, ok := s.active[n]; ok {
		return nil, xerr.Sub("listener already exists", 0x1C)
	}
	h, w, t := p.Next()
	if len(addr) > 0 {
		h = addr
	}
	if len(h) == 0 {
		return nil, ErrNoHost
	}
	v, err := p.Listen(x, h)
	if err != nil {
		return nil, xerr.Wrap("unable to listen", err)
	} else if v == nil {
		return nil, xerr.Sub("unable to listen", 0x18)
	}
	l := &Listener{
		ch:         make(chan struct{}),
		name:       n,
		listener:   v,
		connection: connection{s: s, m: s, p: p, w: w, t: t, log: s.log},
	}
	if s.init.Do(func() { go s.listen() }); cout.Enabled {
		s.log.Info("[%s] Added Listener on %q!", n, h)
	}
	l.ctx, l.cancel = context.WithCancel(x)
	s.new <- l
	go l.listen()
	return l, nil
}
