package tcp

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/iDigitalFlame/xmt/xmt/com"
)

type conn struct {
	timeout time.Duration
	net.Conn
}
type client struct {
	p *provider
}
type listener struct {
	network string
	timeout time.Duration
	net.Listener
}
type provider struct {
	tls     *tls.Config
	dial    *net.Dialer
	network string
}

func (l listener) String() string {
	return fmt.Sprintf("Stream(%s) %s", strings.ToUpper(l.network), l.Addr().String())
}
func (c *conn) Read(b []byte) (int, error) {
	if c.timeout > 0 {
		c.Conn.SetReadDeadline(time.Now().Add(c.timeout))
	}
	return c.Conn.Read(b)
}
func (l *listener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &conn{
		Conn:    c,
		timeout: l.timeout,
	}, nil
}
func (c *client) Connect(s string) (net.Conn, error) {
	return c.p.Connect(s)
}
func (p *provider) Connect(s string) (net.Conn, error) {
	var err error
	var c net.Conn
	if p.tls != nil {
		c, err = tls.DialWithDialer(p.dial, p.network, s, p.tls)
	} else {
		c, err = p.dial.Dial(p.network, s)
	}
	if err != nil {
		return nil, err
	}
	return &conn{
		Conn:    c,
		timeout: p.dial.Timeout,
	}, nil
}
func (p *provider) Listen(s string) (net.Listener, error) {
	var err error
	var c net.Listener
	if p.tls != nil {
		if (p.tls.Certificates == nil || len(p.tls.Certificates) == 0) || p.tls.GetCertificate == nil {
			return nil, ErrInvalidTLSConfig
		}
		c, err = tls.Listen(p.network, s, p.tls)
	} else {
		c, err = net.Listen(p.network, s)
	}
	if err != nil {
		return nil, err
	}
	return &listener{
		network:  p.network,
		timeout:  p.dial.Timeout,
		Listener: c,
	}, nil
}

// NewRaw creates a new simple stream based connector from
// the supplied network type and timeout.  Stream based connectors are
// only valid for TCP and UNIX sockets.  UDP/ICMP/IP will return an
// ErrInvalidNetwork error.
func NewRaw(n string, t time.Duration) (com.Connector, error) {
	return NewRawTLS(n, t, nil)
}

// NewRawTLS creates a new simple stream based Provider from
// the supplied network type and timeout.  Stream based connectors are
// only valid for TCP and UNIX sockets.  UDP/ICMP/IP will return an
// ErrInvalidNetwork error. This stream uses TLS with the provided config.
func NewRawTLS(n string, t time.Duration, c *tls.Config) (com.Provider, error) {
	switch n {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
	default:
		return nil, com.ErrInvalidNetwork
	}
	return &provider{
		tls: c,
		dial: &net.Dialer{
			Timeout:   t,
			KeepAlive: t,
			DualStack: true,
		},
		network: n,
	}, nil
}
