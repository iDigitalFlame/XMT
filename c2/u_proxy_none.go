//go:build noproxy

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
	"github.com/iDigitalFlame/xmt/c2/cfg"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Proxy is a struct that controls a Proxied connection between a client and a
// server and allows for packets to be routed through a current established
// Session.
type Proxy struct{}
type proxyBase struct{}

// Close stops the operation of the Proxy and any Sessions that may be connected.
//
// Resources used with this Proxy will be freed up for reuse.
func (Proxy) Close() error {
	return nil
}
func (proxyBase) Close() error {
	return nil
}
func (proxyBase) subsRegister() {}
func (proxyBase) tags() []uint32 {
	return nil
}
func (proxyBase) IsActive() bool {
	return false
}

// Proxy returns the current Proxy (if enabled). This function take a name
// argument that is a string that specifies the Proxy name.
//
// By default, the name is ignored as multiproxy support is disabled.
//
// When proxy support is disabled, this always returns nil.
func (*Session) Proxy(_ string) *Proxy {
	return nil
}
func (*Session) checkProxyMarshal() bool {
	return true
}
func (proxyBase) accept(_ *com.Packet) bool {
	return false
}

// Replace allows for rebinding this Proxy to another address or using another
// Profile without closing the Proxy.
//
// The listening socket will be closed and the Proxy will be paused and
// cannot accept any more connections before being reopened.
//
// If the replacement fails, the Proxy will be closed.
func (Proxy) Replace(_ string, _ cfg.Profile) error {
	return xerr.Sub("proxy support disabled", 0x53)
}
func (*Session) writeProxyData(_ bool, w data.Writer) error {
	return w.WriteUint8(0)
}
func readProxyData(f bool, r data.Reader) ([]proxyData, error) {
	n, err := r.Uint8()
	if err != nil {
		return nil, err
	}
	for i := uint8(0); i < n; i++ {
		if _, err = r.Bytes(); err != nil {
			return nil, err
		}
		if _, err = r.Bytes(); err != nil {
			return nil, err
		}
		if !f {
			continue
		}
		if _, err = r.Bytes(); err != nil {
			return nil, err
		}
	}
	return nil, nil
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
func (*Session) NewProxy(_, _ string, _ cfg.Profile) (*Proxy, error) {
	return nil, xerr.Sub("proxy support disabled", 0x53)
}
