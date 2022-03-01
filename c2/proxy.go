package c2

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Proxy is a struct that controls a Proxied connection between a client and a
// server and allows for packets to be routed through a current established
// Session.
type Proxy struct {
	listener net.Listener
	connection

	clients map[uint32]*proxyClient
	ch      chan struct{}
	close   chan uint32
	parent  *Session
	cancel  context.CancelFunc

	lock  sync.RWMutex
	state state
}
type proxyClient struct {
	send, chn chan *com.Packet
	wake      chan struct{}
	peek      *com.Packet
	ID        device.ID
	state     state
}

// Wait will block until the current Proxy is closed and shutdown.
func (p *Proxy) Wait() {
	<-p.ch
}
func (p *Proxy) prune() {
	for {
		select {
		case <-p.ch:
			return
		case <-p.ctx.Done():
			return
		case i := <-p.close:
			p.lock.RLock()
			if _, ok := p.clients[i]; ok {
				if delete(p.clients, i); cout.Enabled {
					p.log.Info("[%s/Proxy] Removed closed Session 0x%X.", p.parent.ID, i)
				}
			}
			p.lock.RUnlock()
		}
	}
}
func (p *Proxy) listen() {
	if cout.Enabled {
		p.log.Info("[%s/Proxy] Starting listen on %q..", p.parent.ID, p.listener)
	}
	go p.prune()
	for {
		select {
		case <-p.ctx.Done():
			p.state.Set(stateClosing)
		default:
		}
		if p.state.Closing() {
			break
		}
		c, err := p.listener.Accept()
		if err != nil {
			if p.state.Closing() {
				break
			}
			e, ok := err.(net.Error)
			if ok && e.Timeout() {
				continue
			}
			if cout.Enabled {
				p.parent.log.Error("[%s/P] Error during Listener accept: %s!", p.parent.ID, err)
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
			p.log.Trace("[%s/P] Received a connection from %q..", p.parent.ID, c.RemoteAddr())
		}
		go handle(p.log, c, p, c.RemoteAddr().String())
	}
	if cout.Enabled {
		p.parent.log.Debug("[%s/P] Stopping Listener..", p.parent.ID)
	}
	for _, v := range p.clients {
		v.Close()
	}
	if p.cancel(); !p.state.WakeClosed() {
		p.state.Set(stateWakeClose)
		close(p.close)
	}
	p.listener.Close()
	p.state.Set(stateClosed)
	p.parent.proxy = nil
	p.parent = nil
	close(p.ch)
}
func (p *Proxy) clientLock() {
	p.lock.RLock()
}

// Close stops the operation of the Proxy and any Sessions that may be connected.
//
// Resources used with this Proxy will be freed up for reuse.
func (p *Proxy) Close() error {
	if p.state.Closed() {
		return nil
	}
	p.state.Set(stateClosing)
	err := p.listener.Close()
	p.cancel()
	<-p.ch
	return err
}
func (c *proxyClient) Close() {
	if !c.state.SendClosed() {
		c.state.Set(stateSendClose)
		close(c.send)
	}
	if !c.state.WakeClosed() {
		c.state.Set(stateWakeClose)
		close(c.wake)
	}
	c.state.Set(stateClosed)
	c.state.Unset(stateChannelValue)
	c.state.Unset(stateChannelUpdated)
	c.state.Unset(stateChannel)
	c.peek = nil
}
func (p *Proxy) clientUnlock() {
	p.lock.RUnlock()
}

// IsActive returns true if the Proxy is still able to send and receive Packets.
func (p *Proxy) IsActive() bool {
	return !p.state.Closing()
}
func (p *Proxy) tags() []uint32 {
	if len(p.clients) == 0 {
		return nil
	}
	t := make([]uint32, 0, len(p.clients))
	p.lock.RLock()
	for i := range p.clients {
		if !p.clients[i].state.Tag() {
			continue
		}
		t = append(t, i)
	}
	p.lock.RUnlock()
	return t
}
func (c *proxyClient) chanWake() {
	if c.state.WakeClosed() || len(c.wake) >= cap(c.wake) {
		return
	}
	select {
	case c.wake <- wake:
	default:
	}
}

// Address returns the string representation of the address the Listener is
// bound to.
func (p *Proxy) Address() string {
	return p.listener.Addr().String()
}
func (p *Proxy) wrapper() Wrapper {
	return p.w
}
func (c *proxyClient) name() string {
	return c.ID.String()
}
func (proxyClient) accept(_ uint16) {}
func (c *proxyClient) chanWakeClear() {
	if c.state.WakeClosed() {
		return
	}
	for len(c.wake) > 0 {
		<-c.wake // Drain waker
	}
}
func (p *Proxy) clientClear(i uint32) {
	v, ok := p.clients[i]
	if !ok {
		return
	}
	v.chn = nil
	v.state.Unset(stateChannelProxy)
}
func (p *Proxy) transform() Transform {
	return p.t
}
func (c *proxyClient) chanStop() bool {
	return c.state.ChannelCanStop()
}
func (c *proxyClient) chanStart() bool {
	return c.state.ChannelCanStart()
}
func (c *proxyClient) update(_ string) {
	c.state.Set(stateSeen)
}

// Done returns a channel that's closed when this Proxy is closed.
//
// This can be used to monitor a Proxy's status using a select statement.
func (p *Proxy) Done() <-chan struct{} {
	return p.ch
}
func (proxyClient) frag(_, _, _ uint16) {}
func (c *proxyClient) chanRunning() bool {
	return c.state.Channel()
}
func (c *proxyClient) stateSet(v uint32) {
	c.state.Set(v)
}
func (c *proxyClient) stateUnset(v uint32) {
	c.state.Unset(v)
}
func (p *Proxy) accept(n *com.Packet) bool {
	if len(p.clients) == 0 {
		return false
	}
	p.lock.RLock()
	c, ok := p.clients[n.Device.Hash()]
	if p.lock.RUnlock(); !ok {
		return false
	}
	if isPacketNoP(n) {
		return true
	}
	c.queue(n)
	c.state.Set(stateReady)
	return true
}
func (c *proxyClient) queue(n *com.Packet) {
	if c.state.SendClosed() {
		return
	}
	if bugtrack.Enabled {
		if n.Device.Empty() {
			bugtrack.Track("c2.proxyClient.queue(): Calling queue with empty Device, n.ID=%d!", n.ID)
		}
	}
	if c.chn != nil {
		select {
		case c.chn <- n:
		default:
		}
		return
	}
	select {
	case c.send <- n:
	default:
	}
}
func (c *proxyClient) clientID() device.ID {
	return c.ID
}
func (proxyClient) deadlineRead() time.Time {
	return empty
}
func (proxyClient) deadlineWrite() time.Time {
	return empty
}
func (c *proxyClient) pick(i bool) *com.Packet {
	if c.peek != nil {
		n := c.peek
		c.peek = nil
		return n
	}
	if len(c.send) > 0 {
		return <-c.send
	}
	if i {
		return nil
	}
	if !c.state.Channel() {
		return &com.Packet{Device: c.ID}
	}
	select {
	case <-c.wake:
		return nil
	case n := <-c.send:
		return n
	}
}
func (c *proxyClient) next(i bool) *com.Packet {
	n := c.pick(i)
	if n == nil {
		c.state.Unset(stateReady)
		return nil
	}
	if len(c.send) == 0 && verifyPacket(n, c.ID) {
		if isPacketNoP(n) {
			c.state.Set(stateReady)
		} else {
			c.state.Unset(stateReady)
		}
		return n
	}
	if n, c.peek = nextPacket(c, c.send, n, c.ID, n.Tags); isPacketNoP(n) {
		c.state.Set(stateReady)
	} else {
		c.state.Unset(stateReady)
	}
	return n
}
func (c *proxyClient) sender() chan *com.Packet {
	return c.send
}

// Proxy establishes a new listening Proxy connection using the supplied Profile
// that will send any received Packets "upstream" via the current Session.
//
// Packets destined for hosts connected to this proxy will be routed back and
// forth on this Session.
//
// This function will return an error if this is not a client Session or
// listening fails.
func (s *Session) Proxy(p Profile) (*Proxy, error) {
	if s.parent != nil {
		return nil, xerr.Sub("must be a client session", 0x5)
	}
	// TODO(dij): Eventually this will be removed :P
	if s.proxy != nil {
		return nil, xerr.Sub("only a single Proxy per session can be active", 0x2)
	}
	if p == nil {
		return nil, ErrInvalidProfile
	}
	h, w, t := p.Next()
	if len(h) == 0 {
		return nil, ErrNoHost
	}
	l, err := p.Listen(s.ctx, h)
	if err != nil {
		return nil, xerr.Wrap("unable to listen", err)
	}
	if l == nil {
		return nil, xerr.Sub("unable to listen", 0x8)
	}
	v := &Proxy{
		ch:         make(chan struct{}, 1),
		close:      make(chan uint32, 8),
		parent:     s,
		clients:    make(map[uint32]*proxyClient),
		listener:   l,
		connection: connection{s: s.s, p: p, w: w, m: s.m, t: t, log: s.log},
	}
	if v.ctx, v.cancel = context.WithCancel(s.ctx); cout.Enabled {
		s.log.Info("[%s/P] Added Proxy Listener on %q!", s.ID, h)
	}
	s.proxy = v
	go v.listen()
	return v, nil
}
func (p *Proxy) clientGet(i uint32) (connHost, bool) {
	v, ok := p.clients[i]
	return v, ok
}
func (p *Proxy) clientSet(i uint32, c chan *com.Packet) {
	v, ok := p.clients[i]
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
func (p *Proxy) notify(h connHost, n *com.Packet) error {
	if isPacketNoP(n) {
		return nil
	}
	p.parent.queue(n)
	return nil
}
func (p *Proxy) talk(a string, n *com.Packet) (*conn, error) {
	if n.Device.Empty() || p.parent.state.Closing() {
		return nil, io.ErrShortBuffer
	}
	if n.Flags |= com.FlagProxy; cout.Enabled {
		p.log.Debug("[%s/P:%s] %s: Received a Packet %q..", p.parent.ID, n.Device, a, n)
	}
	p.lock.RLock()
	var (
		i     = n.Device.Hash()
		c, ok = p.clients[i]
	)
	if p.lock.RUnlock(); !ok {
		if n.ID != SvHello {
			if cout.Enabled {
				p.log.Warning("[%s/P:%s] %s: Received a non-hello Packet from a unregistered client!", p.parent.ID, n.Device, a)
			}
			var f com.Flag
			if n.Flags&com.FlagFrag != 0 {
				f.SetPosition(0)
				f.SetLen(n.Flags.Len())
				f.SetGroup(n.Flags.Group())
			}
			return &conn{next: &com.Packet{ID: SvRegister, Flags: f, Device: n.Device}}, nil
		}
		c = &proxyClient{
			ID:    n.Device,
			send:  make(chan *com.Packet, cap(p.parent.send)),
			wake:  make(chan struct{}, 1),
			state: state(stateReady),
		}
		p.lock.Lock()
		p.clients[i] = c
		if p.lock.Unlock(); cout.Enabled {
			p.log.Info("[%s/P:%s] %s: New client registered as %q hash 0x%X.", p.parent.ID, c.ID, a, c.ID, i)
		}
		p.parent.queue(n)
		c.queue(&com.Packet{ID: SvComplete, Device: n.Device, Job: n.Job})
	}
	if c.state.Set(stateSeen); n.ID == SvShutdown {
		select {
		case p.close <- i:
		default:
		}
		p.parent.queue(n)
		return &conn{next: &com.Packet{ID: SvShutdown, Device: n.Device, Job: n.Job}}, nil
	}
	v, err := p.resolve(c, a, n.Tags)
	if err != nil {
		return nil, err
	}
	if err = v.process(p.log, p, a, n, false); err != nil {
		return nil, err
	}
	return v, nil
}
func (p *Proxy) resolve(s *proxyClient, a string, t []uint32) (*conn, error) {
	if len(t) == 0 {
		return &conn{host: s}, nil
	}
	c := &conn{
		add:  make([]*com.Packet, 0, len(t)),
		subs: make(map[uint32]bool, len(t)),
		host: s,
	}
	return c, c.resolve(p.log, s, p, a, t, false)
}
func (p *Proxy) talkSub(a string, n *com.Packet, o bool) (connHost, uint32, *com.Packet, error) {
	if n.Device.Empty() || p.state.Closing() {
		return nil, 0, nil, io.ErrShortBuffer
	}
	if cout.Enabled {
		p.log.Trace("[%s/P:%s/M] %s: Received a Packet %q..", p.parent.ID, n.Device, a, n)
	}
	p.lock.RLock()
	var (
		i     = n.Device.Hash()
		c, ok = p.clients[i]
	)
	if p.lock.RUnlock(); !ok {
		if n.ID != SvHello {
			if cout.Enabled {
				p.log.Warning("[%s/P:%s/M] %s: Received a non-hello Packet from a unregistered client!", p.parent.ID, n.Device, a)
			}
			var f com.Flag
			if n.Flags&com.FlagFrag != 0 {
				f.SetPosition(0)
				f.SetLen(n.Flags.Len())
				f.SetGroup(n.Flags.Group())
			}
			return nil, 0, &com.Packet{ID: SvRegister, Flags: f, Device: n.Device}, nil
		}
		c = &proxyClient{
			ID:    n.Device,
			send:  make(chan *com.Packet, cap(p.parent.send)),
			wake:  make(chan struct{}, 1),
			state: state(stateReady),
		}
		p.lock.Lock()
		p.clients[i] = c
		if p.lock.Unlock(); cout.Enabled {
			p.log.Info("[%s/P:%s/M] %s: New client registered as %q hash 0x%X.", p.parent.ID, c.ID, a, c.ID, i)
		}
		c.queue(&com.Packet{ID: SvComplete, Device: n.Device, Job: n.Job})
	}
	switch c.state.Set(stateSeen); {
	case isPacketNoP(n):
		p.parent.queue(n)
		c.state.Set(stateReady)
	case n.ID == SvShutdown:
		select {
		case p.close <- i:
		default:
		}
		p.parent.queue(n)
		return nil, 0, &com.Packet{ID: SvShutdown, Device: n.Device, Job: n.Job}, nil
	}
	if o {
		return c, i, nil, nil
	}
	return c, i, c.next(true), nil
}
