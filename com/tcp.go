package com

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

type tcpConn struct {
	_ [0]func()
	net.Conn
	timeout time.Duration
}
type tcpClient struct {
	_ [0]func()
	c tcpConnector
}
type deadline interface {
	SetDeadline(time.Time) error
}
type tcpListener struct {
	_ [0]func()
	net.Listener
	timeout time.Duration
}
type tcpConnector struct {
	_      [0]func()
	tls    *tls.Config
	dialer *net.Dialer
}

func (t tcpListener) String() string {
	return "TCP[" + t.Addr().String() + "]"
}
func (t *tcpConn) CloseWrite() error {
	v, ok := t.Conn.(*net.TCPConn)
	if !ok {
		return nil
	}
	return v.CloseWrite()
}
func (t *tcpConn) Read(b []byte) (int, error) {
	if t.timeout > 0 {
		t.Conn.SetReadDeadline(time.Now().Add(t.timeout))
	}
	return t.Conn.Read(b)
}
func (t *tcpConn) Write(b []byte) (int, error) {
	if t.timeout > 0 {
		t.Conn.SetWriteDeadline(time.Now().Add(t.timeout))
	}
	return t.Conn.Write(b)
}
func (t tcpListener) Accept() (net.Conn, error) {
	if d, ok := t.Listener.(deadline); ok {
		d.SetDeadline(time.Now().Add(t.timeout))
	}
	c, err := t.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &tcpConn{timeout: t.timeout, Conn: c}, nil
}

// NewTCP creates a new simple TCP based connector with the supplied timeout.
func NewTCP(t time.Duration) (Connector, error) {
	return newConnector(netTCP, t, nil)
}
func (t tcpClient) Connect(s string) (net.Conn, error) {
	return t.c.Connect(s)
}
func (t tcpConnector) Connect(s string) (net.Conn, error) {
	c, err := newConn(netTCP, s, t)
	if err != nil {
		return nil, err
	}
	return &tcpConn{timeout: t.dialer.Timeout, Conn: c}, nil
}
func newConn(n, s string, t tcpConnector) (net.Conn, error) {
	if t.tls != nil {
		return tls.DialWithDialer(t.dialer, n, s, t.tls)
	}
	return t.dialer.Dial(n, s)
}
func (t tcpConnector) Listen(s string) (net.Listener, error) {
	c, err := newListener(netTCP, s, t)
	if err != nil {
		return nil, err
	}
	return &tcpListener{timeout: t.dialer.Timeout, Listener: c}, nil
}
func newListener(n, s string, t tcpConnector) (net.Listener, error) {
	if t.tls != nil && (len(t.tls.Certificates) == 0 || t.tls.GetCertificate == nil) {
		return nil, ErrInvalidTLSConfig
	}
	l, err := ListenConfig.Listen(context.Background(), n, s)
	if err != nil {
		return nil, err
	}
	if t.tls == nil {
		return l, nil
	}
	return tls.NewListener(l, t.tls), nil
}

// NewSecureTCP creates a new simple TLS wrapped TCP based connector with the supplied timeout.
func NewSecureTCP(t time.Duration, c *tls.Config) (Connector, error) {
	return newConnector(netTCP, t, c)
}
func newConnector(n string, t time.Duration, c *tls.Config) (*tcpConnector, error) {
	if t < 0 {
		return nil, xerr.New("invalid timeout value " + t.String())
	}
	switch n {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
	default:
		return nil, xerr.New("invalid network type " + n)
	}
	return &tcpConnector{tls: c, dialer: &net.Dialer{Timeout: t, KeepAlive: t, DualStack: true}}, nil
}
