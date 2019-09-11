package c2

import (
	"context"
	"fmt"

	"github.com/iDigitalFlame/xmt/xmt/com"
)

// Proxy is a struct that controls a Proxied
// connection between a client and a server and allows
// for packets to be routed through a current established
// Session.
type Proxy struct {
	ctx       context.Context
	send      map[uint32]*proxyClient
	parent    *Session
	cancel    context.CancelFunc
	wrapper   Wrapper
	listener  Listener
	transform Transform
}
type proxyClient struct {
	ctx  context.Context
	hash uint32
	send chan *com.Packet
}

// Wait will block until the current Proxy
// is closed and shutdown.
func (p *Proxy) Wait() {
	<-p.ctx.Done()
}
func (p *Proxy) listen() {
	p.parent.controller.Log.Trace("[%s:Proxy] Starting listen \"%s\"...", p.parent.ID, p.listener.String())
	for p.ctx.Err() == nil {
		c, err := p.listener.Accept()
		if err != nil {
			p.parent.controller.Log.Error("[%s:Proxy] Received error during listener operation! (%s)", p.parent.ID, err.Error())
			if p.ctx.Err() != nil {
				break
			}
		}
		if c == nil {
			continue
		}
		p.parent.controller.Log.Trace("[%s:Proxy] Received a connection from \"%s\"...", p.parent.ID, c.IP())
		go p.session(c)
	}
	p.parent.controller.Log.Debug("[%s:Proxy] Stopping listen...", p.parent.ID)
	p.cancel()
	p.listener.Close()
	for k, v := range p.send {
		if p.parent.ctx.Err() == nil {
			p.parent.del <- k
		}
		close(v.send)
	}
}

// Close closes the proxy listener and terminates
// all current proxy connections.
func (p *Proxy) Close() error {
	p.cancel()
	return nil
}

// IsActive returns true if the Proxy is still
// able to send and receive Packets.
func (p *Proxy) IsActive() bool {
	return p.ctx.Err() == nil
}
func (p *Proxy) session(c Connection) {
	defer c.Close()
	d, err := read(c, p.wrapper, p.transform)
	if err != nil {
		p.parent.controller.Log.Warning("[%s:Proxy] Received an error when attempting to read a Packet from \"%s\"! (%s)", p.parent.ID, c.IP(), err.Error())
		return
	}
	if d == nil || d.IsEmpty() {
		p.parent.controller.Log.Warning("[%s:Proxy] Received an empty or invalid Packet from \"%s\"!", p.parent.ID, c.IP())
		return
	}
	if d.Flags&com.FlagIgnore != 0 {
		p.parent.controller.Log.Trace("[%s:Proxy] Received an ignore packet from \"%s\".", p.parent.ID, c.IP())
		return
	}
	if d.Flags&com.FlagMulti == 0 || d.Flags&com.FlagMultiDevice == 0 {
		p.client(c, d)
		return
	}
	n := d.Flags.FragTotal()
	if n == 0 {
		p.parent.controller.Log.Warning("[%s:Proxy] Received an invalid multi Packet from \"%s\"!", p.parent.ID, c.IP())
		return
	}
	v := &com.Packet{}
	for i := uint16(0); i < n && p.parent.ctx.Err() == nil; i++ {
		if err := v.UnmarshalStream(d); err != nil {
			p.parent.controller.Log.Warning("[%s:Proxy] Received an error when attempting to read a Packet from \"%s\"! (%s)", p.parent.ID, c.IP(), err.Error())
			return
		}
		p.client(c, v)
		v.Reset()
	}
	d.Close()
}
func (p *Proxy) client(c Connection, d *com.Packet) {
	if p.ctx.Err() != nil {
		return
	}
	i := d.Device.Hash()
	d.Flags |= com.FlagProxy
	p.parent.controller.Log.Trace("[%s:Proxy] Received a packet \"%s\" from \"%s\" (%s), session hash 0x%X.", p.parent.ID, d.String(), d.Device, c.IP(), i)
	if p.parent.ctx.Err() != nil {
		return
	}
	p.parent.send <- d
	x, ok := p.send[i]
	if !ok {
		x = &proxyClient{
			ctx:  p.ctx,
			hash: i,
			send: make(chan *com.Packet, cap(p.parent.send)),
		}
		p.send[i] = x
		p.parent.new <- x
		if d.ID == MsgHello {
			v := &com.Packet{ID: MsgRegistered, Device: d.Device, Job: d.Job}
			if err := write(c, p.wrapper, p.transform, v); err != nil {
				p.parent.controller.Log.Warning("[%s:Proxy] Received an error writing data to client \"%s\"! (%s)", p.parent.ID, c.IP(), err.Error())
			}
			return
		}
	}
	v, err := next(x.send, d.Device)
	if err != nil {
		p.parent.controller.Log.Warning("[%s:Proxy] Received an error gathering packet data for client \"%s\"! (%s)", p.parent.ID, c.IP(), err.Error())
		return
	}
	p.parent.controller.Log.Trace("[%s:Proxy] Sending Packet \"%s\" to client \"%s\".", p.parent.ID, v.String(), c.IP())
	if err := write(c, p.wrapper, p.transform, v); err != nil {
		p.parent.controller.Log.Warning("[%s:Proxy] Received an error writing data to client \"%s\"! (%s)", p.parent.ID, c.IP(), err.Error())
	}
}

// Proxy establishes a new listening Proxy connection using the supplied
// listener that will send any received Packets "upstream" via the current
// Session. Packets destined for hosts connected to this proxy will be routed
// back on forth on this Session.
func (s *Session) Proxy(b string, v Connector, p *Profile) (*Proxy, error) {
	if v == nil {
		return nil, ErrNoConnector
	}
	l, err := v.Listen(b)
	if err != nil {
		return nil, fmt.Errorf("unable to listen on \"%s\": %w", b, err)
	}
	if l == nil {
		return nil, fmt.Errorf("unable to listen on \"%s\"", b)
	}
	i := &Proxy{
		send:     make(map[uint32]*proxyClient),
		parent:   s,
		listener: l,
	}
	if p != nil {
		i.wrapper = p.Wrapper
		i.transform = p.Transform
	}
	if i.wrapper == nil {
		i.wrapper = DefaultWrapper
	}
	i.ctx, i.cancel = context.WithCancel(s.ctx)
	s.controller.Log.Debug("[%s] Added listener Proxy type \"%s\"...", s.ID, l.String())
	if s.proxies == nil {
		s.del = make(chan uint32, DefaultBufferSize)
		s.new = make(chan *proxyClient, DefaultBufferSize)
		s.proxies = make(map[uint32]*proxyClient)
	}
	go i.listen()
	return i, nil
}
