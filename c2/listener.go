package c2

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"time"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/device"
)

var (
	// ErrInvalidPacketID is a error returned inside the client thread when the received packet
	// ID does not match the client ID and does not match any proxy client connected.
	ErrInvalidPacketID = errors.New("received a Packet ID that does not match our own ID")
	// ErrInvalidPacketCount is returned when attempting to read a packet marked
	// as multi or frag an the total count returned is zero.
	ErrInvalidPacketCount = errors.New("frag total is zero on a multi or frag packet")
)

// Listener is a struct that is passed back when a C2 Listener is added to the Server. The Listener struct
// allows for controlling the Listener and setting callback functions to be used when a client connects,
// registers or disconnects.
type Listener struct {
	New     func(*Session)
	Connect func(*Session)
	Oneshot func(*com.Packet)
	Receive func(*Session, *com.Packet)

	ch       chan waker
	size     uint
	name     string
	done     uint32
	close    chan uint32
	listener net.Listener
	sessions map[uint32]*Session
	connection
}

// Wait will block until the current socket associated with this Listener is closed and shutdown.
func (l *Listener) Wait() {
	<-l.ch
}
func (l *Listener) listen() {
	l.log.Debug("[%s] Starting Listener %q...", l.name, l.listener)
	for atomic.LoadUint32(&l.done) == flagOpen {
		for len(l.close) > 0 {
			var (
				i     = <-l.close
				s, ok = l.sessions[i]
			)
			if !ok {
				continue
			}
			if s.Shutdown != nil {
				l.s.events <- event{s: s, sFunc: s.Shutdown}
			}
			delete(l.sessions, i)
			l.log.Debug("[%s] Removed closed Session 0x%X.", l.name, i)
		}
		c, err := l.listener.Accept()
		if err != nil {
			if l.done > flagOpen {
				break
			}
			e, ok := err.(net.Error)
			if ok && e.Timeout() {
				continue
			}
			l.log.Error("[%s] Error occurred during Listener accept: %s!", l.name, err.Error())
			if ok && !e.Timeout() && !e.Temporary() {
				break
			}
			continue
		}
		if c == nil {
			continue
		}
		l.log.Trace("[%s] Received a connection from %q...", l.name, c.RemoteAddr().String())
		go l.handle(c)
	}
	l.log.Debug("[%s] Stopping Listener.", l.name)
	for _, v := range l.sessions {
		v.Close()
	}
	l.cancel()
	close(l.close)
	l.listener.Close()
	atomic.StoreUint32(&l.done, flagFinished)
	l.s.close <- l.name
	close(l.ch)
}

// Close stops the operation of the Listener and any Sessions that may be connected. Resources used with this
// Listener will be freed up for reuse. This function blocks until the listener socket is closed.
func (l *Listener) Close() error {
	atomic.StoreUint32(&l.done, flagClose)
	err := l.listener.Close()
	l.Wait()
	return err
}

// String returns the Name of this Listener.
func (l Listener) String() string {
	return l.name
}

// IsActive returns true if the Listener is still able to send and receive Packets.
func (l Listener) IsActive() bool {
	return l.done == flagOpen
}
func (l *Listener) handle(c net.Conn) {
	if !l.handlePacket(c, false) {
		c.Close()
		return
	}
	l.log.Debug("[%s] %s: Triggered Channel mode, holding open Channel!", l.name, c.RemoteAddr().String())
	for atomic.LoadUint32(&l.done) == flagOpen {
		if l.handlePacket(c, true) {
			break
		}
	}
	l.log.Debug("[%s] %s: Closing Channel..", l.name, c.RemoteAddr().String())
	c.Close()
}

// Remove removes and closes the Session and releases all it's associated resources. This does not close the
// Session on the client's end, use the Shutdown function to properly shutdown the client process.
func (l *Listener) Remove(i device.ID) {
	l.close <- i.Hash()
}

// Shutdown triggers a remote Shutdown and closure of the Session associated with the Device ID. This will not
// immediately close a Session. The Session will be removed when the Client acknowledges the shutdown request.
func (l *Listener) Shutdown(i device.ID) {
	s, ok := l.sessions[i.Hash()]
	if !ok {
		return
	}
	s.Close()
}

// Connected returns an array of all the current Sessions connected to this Listener.
func (l *Listener) Connected() []*Session {
	d := make([]*Session, 0, len(l.sessions))
	for _, v := range l.sessions {
		d = append(d, v)
	}
	return d
}

// Context returns the current Listener's context. This function can be useful for canceling running
// processes when this Listener closes.
func (l *Listener) Context() context.Context {
	return l.ctx
}

// Session returns the Session that matches the specified Device ID. This function will return nil if
// no matching Device ID is found.
func (l *Listener) Session(i device.ID) *Session {
	if len(i) == 0 {
		return nil
	}
	if s, ok := l.sessions[i.Hash()]; ok {
		return s
	}
	return nil
}
func (l *Listener) handlePacket(c net.Conn, o bool) bool {
	p, err := readPacket(c, l.w, l.t)
	if err != nil {
		l.log.Warning("[%s] %s: Error occurred during Packet read: %s!", l.name, c.RemoteAddr().String(), err.Error())
		return o
	}
	if p.Flags&com.FlagOneshot != 0 {
		l.log.Trace("[%s] %s: Received an Oneshot Packet.", l.name, c.RemoteAddr().String())
		notify(l, nil, p)
		return false
	}
	l.log.Trace("[%s:%s] Received Packet %q.", l.name, c.RemoteAddr().String(), p)
	z := l.resolveTags(c.RemoteAddr().String(), p.Device, o, p.Tags)
	if p.Flags&com.FlagMultiDevice == 0 && p.Flags&com.FlagProxy == 0 {
		if s := l.client(c, p, o); s != nil {
			n, err := s.next(false)
			if err != nil {
				l.log.Warning("[%s:%s] %s: Received an error retriving Packet data: %s!", l.name, s.Device.ID, s.host, err.Error())
				return p.Flags&com.FlagChannel != 0
			}
			if len(z) > 0 {
				l.log.Trace("[%s:%s] %s: Resolved Tags added %d Packets!", l.name, s.Device.ID, s.host, len(z))
				u := &com.Packet{ID: MsgMultiple, Flags: com.FlagMulti | com.FlagMultiDevice}
				n.MarshalStream(u)
				for i := 0; i < len(z); i++ {
					z[i].MarshalStream(u)
				}
				u.Flags.SetLen(uint16(len(z) + 1))
				u.Close()
				n = u
			}
			l.log.Trace("[%s:%s] %s: Sending Packet %q to client...", l.name, s.Device.ID, s.host, n.String())
			if err = writePacket(c, s.w, s.t, n); err != nil {
				l.log.Warning("[%s:%s] %s: Received an error writing data to client: %s!", l.name, s.Device.ID, s.host, err.Error())
				return o
			}
		}
		return p.Flags&com.FlagChannel != 0
	}
	x := p.Flags.Len()
	if x == 0 {
		l.log.Warning("[%s:%s] %s: Received an invalid multi Packet!", l.name, p.Device, c.RemoteAddr().String())
		return p.Flags&com.FlagChannel != 0
	}
	var (
		i, t uint16
		n    *com.Packet
		m    = &com.Packet{ID: MsgMultiple, Flags: com.FlagMulti | com.FlagMultiDevice}
	)
	for ; i < x; i++ {
		n = new(com.Packet)
		if err := n.UnmarshalStream(p); err != nil {
			l.log.Warning("[%s:%s] %s: Received an error when attempting to read a Packet: %s!", l.name, p.Device, c.RemoteAddr().String(), err.Error())
			return p.Flags&com.FlagChannel != 0
		}
		if n.Flags&com.FlagOneshot != 0 {
			l.log.Trace("[%s:%s] %s: Received an Oneshot Packet.", l.name, n.Device, c.RemoteAddr().String())
			notify(l, nil, n)
			continue
		}
		s := l.client(c, n, o)
		if s == nil {
			continue
		}
		if r, err := s.next(false); err != nil {
			l.log.Warning("[%s:%s] %s: Received an error retriving Packet data: %s!", l.name, s.Device.ID, s.host, err.Error())
		} else {
			r.MarshalStream(m)
		}
		n = nil
		t++
	}
	if len(z) > 0 {
		l.log.Trace("[%s:%s] %s: Resolved Tags added %d Packets!", l.name, p.Device, c.RemoteAddr().String(), len(z))
		for i := 0; i < len(z); i++ {
			z[i].MarshalStream(m)
		}
		t += uint16(len(z))
	}
	m.Flags.SetLen(t)
	m.Close()
	l.log.Trace("[%s:%s] %s: Sending Packet %q to client...", l.name, p.Device, c.RemoteAddr().String(), m.String())
	if err := writePacket(c, l.w, l.t, m); err != nil {
		l.log.Warning("[%s:%s] %s: Received an error writing data to client: %s!", l.name, p.Device, c.RemoteAddr().String(), err.Error())
	}
	return p.Flags&com.FlagChannel != 0
}
func (l *Listener) client(c net.Conn, p *com.Packet, o bool) *Session {
	l.log.Trace("[%s:%s] %s: Received a Packet %q...", l.name, p.Device, c.RemoteAddr().String(), p.String())
	var (
		i     = p.Device.Hash()
		s, ok = l.sessions[i]
	)
	if !ok {
		if p.ID != MsgHello {
			l.log.Warning("[%s:%s] %s: Received a non-hello Packet from a unregistered client!", l.name, p.Device, c.RemoteAddr().String())
			if err := writePacket(c, l.w, l.t, &com.Packet{ID: MsgRegister}); err != nil {
				l.log.Warning("[%s:%s] %s: Received an error writing data to client: %s!", l.name, p.Device, c.RemoteAddr().String(), err.Error())
			}
			return nil
		}
		s = &Session{
			ch:      make(chan waker, 1),
			ID:      p.Device,
			send:    make(chan *com.Packet, l.size),
			recv:    make(chan *com.Packet, l.size),
			frags:   make(map[uint16]*cluster),
			parent:  l,
			Created: time.Now(),
			connection: connection{
				w:   l.w,
				t:   l.t,
				s:   l.s,
				log: l.log,
				Mux: l.Mux,
			},
		}
		s.ctx, s.cancel = context.WithCancel(l.ctx)
		l.sessions[i] = s
		l.log.Debug("[%s:%s] %s: New client registered as %q hash 0x%X.", l.name, s.ID, c.RemoteAddr().String(), s.ID, i)
	}
	s.Last = time.Now()
	s.host = c.RemoteAddr().String()
	if p.ID == MsgHello {
		if err := s.Device.UnmarshalStream(p); err != nil {
			l.log.Warning("[%s:%s] %s: Received an error reading data from client: %s!", l.name, s.ID, s.host, err.Error())
			return nil
		}
		l.log.Trace("[%s:%s] %s: Received client device info: (OS: %s, %s).", l.name, s.ID, s.host, s.Device.OS.String(), s.Device.Version)
		if p.Flags&com.FlagProxy == 0 {
			s.send <- &com.Packet{ID: MsgRegistered, Device: p.Device, Job: p.Job}
		}
		if l.New != nil {
			l.s.events <- event{s: s, sFunc: l.New}
		}
		if err := notify(l, s, p); err != nil {
			l.log.Warning("[%s:%s] %s: Received an error processing Packet data: %s!", l.name, s.ID, c.RemoteAddr().String(), err.Error())
		}
		return s
	}
	if l.Connect != nil && !o {
		l.s.events <- event{s: s, sFunc: l.Connect}
	}
	if err := notify(l, s, p); err != nil {
		l.log.Warning("[%s:%s] %s: Received an error processing Packet data: %s!", l.name, s.ID, c.RemoteAddr().String(), err.Error())
		return nil
	}
	return s
}
func (l *Listener) resolveTags(a string, i device.ID, o bool, t []uint32) []*com.Packet {
	var p []*com.Packet
	for x := 0; x < len(t); x++ {
		l.log.Trace("[%s:%s] %s: Received a Tag 0x%X...", l.name, i, a, t[x])
		s, ok := l.sessions[t[x]]
		if !ok {
			l.log.Warning("[%s:%s] %s: Received an invalid Tag 0x%X!", l.name, i, a, t[x])
			continue
		}
		s.host, s.Last = a, time.Now()
		if l.Connect != nil && !o {
			l.s.events <- event{s: s, sFunc: l.Connect}
		}
		n, err := s.next(true)
		if err != nil {
			l.log.Warning("[%s:%s] %s: Received an error retriving Packet data: %s!", l.name, i, a, err.Error())
			continue
		}
		if n == nil {
			continue
		}
		p = append(p, n)
	}
	return p
}
