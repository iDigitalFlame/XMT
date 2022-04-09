//go:build noproxy

package c2

import (
	"io"

	"github.com/iDigitalFlame/xmt/com"
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
func (*Session) checkProxyMarshal() bool {
	return true
}

// GetProxy returns the current Proxy (if enabled). This function take a name
// argument that is a string that specifies the Proxy name.
//
// By default, the name is ignored as multiproxy support is disabled.
//
// WHen proxy support is disabled, this always returns nil.
func (*Session) GetProxy(_ string) *Proxy {
	return nil
}
func (proxyBase) accept(_ *com.Packet) bool {
	return false
}

// Proxy establishes a new listening Proxy connection using the supplied Profile
// name and bind address that will send any received Packets "upstream" via the
// current Session.
//
// Packets destined for hosts connected to this proxy will be routed back and
// forth on this Session.
//
// This function will return an error if this is not a client Session or
// listening fails.
func (*Session) Proxy(_, _ string, _ Profile) (*Proxy, error) {
	return nil, xerr.Sub("proxy support disabled", 0x2)
}

func (*Session) writeProxyInfo(w io.Writer, d *[8]byte) error {
	(*d)[0] = 0
	return writeFull(w, 1, (*d)[0:1])
}
func readProxyInfo(r io.Reader, d *[8]byte) ([]proxyData, error) {
	if err := readFull(r, 1, (*d)[0:1]); err != nil {
		return nil, err
	}
	m := int((*d)[0])
	if m == 0 {
		return nil, nil
	}
	var (
		o   = make([]proxyData, m)
		err error
	)
	for i := 0; i < m && i < 0xFF; i++ {
		if err = readFull(r, 4, (*d)[0:4]); err != nil {
			return nil, err
		}
		n, s := make([]byte, uint16((*d)[1])|uint16((*d)[0])<<8), make([]byte, uint16((*d)[3])|uint16((*d)[2])<<8)
		if len(n) > 0 {
			if err = readFull(r, len(n), n); err != nil {
				return nil, err
			}
		}
		if len(s) > 0 {
			if err = readFull(r, len(s), s); err != nil {
				return nil, err
			}
		}
		if o[i].p, err = readSlice(r, d); err != nil {
			return nil, err
		}
		o[i].b, o[i].n = string(n), string(s)
	}
	return nil, nil
}
