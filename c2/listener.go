package c2

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/PurpleSec/escape"
	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
)

// Listener is a struct that is passed back when a C2 Listener is added to the Server. The Listener struct
// allows for controlling the Listener and setting callback functions to be used when a client connects,
// registers or disconnects.
type Listener struct {
	listener net.Listener
	connection

	New     func(*Session)
	Oneshot func(*com.Packet)

	ch       chan waker
	close    chan uint32
	sessions map[uint32]*Session

	cancel context.CancelFunc
	name   string
	sleep  time.Duration
	lock   sync.RWMutex
	state  state
}

// Wait will block until the current socket associated with this Listener is closed and shutdown.
func (l *Listener) Wait() {
	<-l.ch
}
func (l *Listener) prune() {
	for {
		select {
		case <-l.ch:
			return
		case <-l.ctx.Done():
			return
		case i := <-l.close:
			l.lock.Lock()
			if s, ok := l.sessions[i]; ok {
				if s.Shutdown != nil {
					l.s.events <- event{s: s, sf: s.Shutdown}
				}
				if delete(l.sessions, i); cout.Enabled {
					l.log.Debug("[%s] Removed closed Session 0x%X.", l.name, i)
				}
			}
			l.lock.Unlock()
		}
	}
}
func (l *Listener) listen() {
	if cout.Enabled {
		l.log.Info("[%s] Starting Listener %q...", l.name, l.listener)
	}
	go l.prune()
	for {
		select {
		case <-l.ctx.Done():
			l.state.Set(stateClosing)
		default:
		}
		if l.state.Closing() {
			break
		}
		c, err := l.listener.Accept()
		if err != nil {
			if l.state.Closing() {
				break
			}
			e, ok := err.(net.Error)
			if ok && e.Timeout() {
				continue
			}
			if cout.Enabled {
				l.log.Error("[%s] Error during Listener accept: %s!", l.name, err)
			}
			if ok && !e.Timeout() && !e.Temporary() {
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
	for _, v := range l.sessions {
		v.Close()
	}
	if l.cancel(); !l.state.WakeClosed() {
		l.state.Set(stateWakeClose)
		close(l.close)
	}
	l.listener.Close()
	l.s.close <- l.name
	l.state.Set(stateClosed)
	close(l.ch)
}
func (l *Listener) clientLock() {
	l.lock.RLock()
}

// Close stops the operation of the Listener and any Sessions that may be connected. Resources used with this
// Listener will be freed up for reuse. This function blocks until the listener socket is closed.
func (l *Listener) Close() error {
	if l.state.Closed() {
		return nil
	}
	l.state.Set(stateClosing)
	l.cancel()
	err := l.listener.Close()
	<-l.ch
	return err
}
func (l *Listener) clientUnlock() {
	l.lock.RUnlock()
}

// IsActive returns true if the Listener is still able to send and receive Packets.
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

// Address returns the string representation of the address the Listener is bound to.
func (l *Listener) Address() string {
	return l.listener.Addr().String()
}
func (l *Listener) wrapper() Wrapper {
	return l.w
}

// Remove removes and closes the Session and releases all it's associated resources. This does not close the
// Session on the client's end, use the Shutdown function to properly shutdown the client process.
func (l *Listener) Remove(i device.ID) {
	if l.state.WakeClosed() {
		return
	}
	l.close <- i.Hash()
}
func (l *Listener) clientClear(i uint32) {
	v, ok := l.sessions[i]
	if !ok {
		return
	}
	v.chn = nil
	v.state.Unset(stateChannelProxy)
}

// Shutdown triggers a remote Shutdown and closure of the Session associated with the Device ID. This will not
// immediately close a Session. The Session will be removed when the Client acknowledges the shutdown request.
func (l *Listener) Shutdown(i device.ID) {
	if l.state.WakeClosed() {
		return
	}
	l.lock.RLock()
	s, ok := l.sessions[i.Hash()]
	if l.lock.RUnlock(); !ok {
		return
	}
	s.Close()
}
func (l *Listener) transform() Transform {
	return l.t
}

// Connected returns an array of all the current Sessions connected to this Listener.
func (l *Listener) Connected() []*Session {
	if l.state.Closed() || len(l.sessions) == 0 {
		return nil
	}
	l.lock.RLock()
	d := make([]*Session, 0, len(l.sessions))
	for _, v := range l.sessions {
		d = append(d, v)
	}
	l.lock.RUnlock()
	return d
}

// JSON returns the data of this Listener as a JSON blob.
func (l *Listener) JSON(w io.Writer) error {
	if !cout.Enabled {
		return nil
	}
	if _, err := w.Write([]byte(`{"name":` + escape.JSON(l.name) + `,"address":` + escape.JSON(l.Address()))); err != nil {
		return err
	}
	if t, ok := l.listener.(stringer); ok {
		if _, err := w.Write([]byte(`,"type":` + escape.JSON(t.String()))); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte(`,"sessions":[`)); err != nil {
		return err
	}
	i := 0
	for _, v := range l.sessions {
		if i > 0 {
			if _, err := w.Write([]byte{0x2C}); err != nil {
				return err
			}
		}
		if err := v.JSON(w); err != nil {
			return err
		}
		i++
	}
	_, err := w.Write([]byte(`]}`))
	return err
}

// MarshalJSON fulfils the JSON Marshaler interface.
func (l *Listener) MarshalJSON() ([]byte, error) {
	if !cout.Enabled {
		return nil, nil
	}
	b := buffers.Get().(*data.Chunk)
	l.JSON(b)
	d := b.Payload()
	returnBuffer(b)
	return d, nil
}

// Session returns the Session that matches the specified Device ID. This function will return nil if
// no matching Device ID is found.
func (l *Listener) Session(i device.ID) *Session {
	if i.Empty() || l.state.Closed() {
		return nil
	}
	l.lock.RLock()
	s, ok := l.sessions[i.Hash()]
	if l.lock.RUnlock(); ok {
		return s
	}
	return nil
}
func (l *Listener) clientGet(i uint32) (connHost, bool) {
	s, ok := l.sessions[i]
	return s, ok
}
func (l *Listener) clientSet(i uint32, c chan *com.Packet) {
	v, ok := l.sessions[i]
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
	l.lock.RLock()
	var (
		i     = n.Device.Hash()
		s, ok = l.sessions[i]
	)
	if l.lock.RUnlock(); !ok {
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
			ch:         make(chan waker, 1),
			ID:         n.Device,
			jobs:       make(map[uint16]*Job),
			send:       make(chan *com.Packet, 256),
			wake:       make(chan waker, 1),
			frags:      make(map[uint16]*cluster),
			parent:     l,
			Created:    time.Now(),
			connection: connection{w: l.w, t: l.t, s: l.s, log: l.log, ctx: l.ctx},
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
		if cout.Enabled {
			l.log.Debug("[%s:%s] %s: Received client device info: (OS: %s, %s).", l.name, s.ID, a, s.Device.OS, s.Device.Version)
		}
		l.lock.Lock()
		l.sessions[i] = s
		if l.lock.Unlock(); cout.Enabled {
			l.log.Info("[%s:%s] %s: New client registered as %q (0x%X).", l.name, s.ID, a, s.ID, i)
		}
	}
	if s.sleep == 0 && ok {
		switch {
		case !s.Last.IsZero():
			s.sleep = time.Since(s.Last)
		case !s.Created.IsZero():
			s.sleep = time.Since(s.Created)
		}
	}
	if s.Last, s.host = time.Now(), a; !ok {
		if n.Flags&com.FlagProxy == 0 {
			s.queue(&com.Packet{ID: SvComplete, Device: n.Device, Job: n.Job})
		}
		switch {
		case l.New != nil:
			l.s.events <- event{s: s, sf: l.New}
		case l.s.New != nil:
			l.s.events <- event{s: s, sf: l.s.New}
		}
	}
	c, err := l.resolve(s, a, n.Tags)
	if err != nil {
		return nil, err
	}
	if err = c.process(l.log, l, a, n, false); err != nil {
		return nil, err
	}
	return c, nil
}
func (l *Listener) resolve(s *Session, a string, t []uint32) (*conn, error) {
	if len(t) == 0 {
		return &conn{host: s}, nil
	}
	c := &conn{
		add:  make([]*com.Packet, 0, len(t)),
		subs: make(map[uint32]bool, len(t)),
		host: s,
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
	l.lock.RLock()
	var (
		i     = n.Device.Hash()
		s, ok = l.sessions[i]
	)
	if l.lock.RUnlock(); !ok {
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
			ch:         make(chan waker, 1),
			ID:         n.Device,
			jobs:       make(map[uint16]*Job),
			send:       make(chan *com.Packet, 256),
			wake:       make(chan waker, 1),
			frags:      make(map[uint16]*cluster),
			parent:     l,
			Created:    time.Now(),
			connection: connection{w: l.w, t: l.t, s: l.s, log: l.log, ctx: l.ctx},
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
		if cout.Enabled {
			l.log.Debug("[%s:%s/M] %s: Received client device info: (OS: %s, %s).", l.name, s.ID, a, s.Device.OS, s.Device.Version)
		}
		l.lock.Lock()
		l.sessions[i] = s
		if l.lock.Unlock(); cout.Enabled {
			l.log.Info("[%s:%s/M] %s: New client registered as %q (0x%X).", l.name, s.ID, a, s.ID, i)
		}
	}
	if s.sleep == 0 && ok {
		switch {
		case !s.Last.IsZero():
			s.sleep = time.Since(s.Last)
		case !s.Created.IsZero():
			s.sleep = time.Since(s.Created)
		}
	}
	if s.Last, s.host = time.Now(), a; !ok {
		if n.Flags&com.FlagProxy == 0 {
			s.queue(&com.Packet{ID: SvComplete, Device: n.Device, Job: n.Job})
		}
		switch {
		case l.New != nil:
			l.s.events <- event{s: s, sf: l.New}
		case l.s.New != nil:
			l.s.events <- event{s: s, sf: l.s.New}
		}
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
	return s, i, s.next(true), nil
}
