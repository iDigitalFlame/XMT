package com

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

// TCPConn is a struct that represents a TCP based network connection. This struct can be used
// to control and manage the current connection.
type TCPConn struct {
	_       [0]func()
	timeout time.Duration
	net.Conn
}
type tcpClient struct {
	_ [0]func()
	c TCPConnector
}

// TCPListener is a struct that represents a TCP based network connection listener. This struct can be
// used to accept and create new TCP connections.
type TCPListener struct {
	_       [0]func()
	timeout time.Duration
	net.Listener
}

// TCPConnector is a struct that represents a TCP based network connection handler. This struct can
// be used to create new TCP listeners.
type TCPConnector struct {
	_      [0]func()
	tls    *tls.Config
	dialer *net.Dialer
}

// String returns a string representation of this TCPListener.
func (t TCPListener) String() string {
	return fmt.Sprintf("TCP[%s]", t.Addr().String())
}

// Read will attempt to read len(b) bytes from the current connection and fill the supplied buffer.
// The return values will be the amount of bytes read and any errors that occurred.
func (t *TCPConn) Read(b []byte) (int, error) {
	if t.timeout > 0 {
		t.Conn.SetReadDeadline(time.Now().Add(t.timeout))
	}
	return t.Conn.Read(b)
}

// Write will attempt to write len(b) bytes to the current connection from the supplied buffer.
// The return values will be the amount of bytes wrote and any errors that occurred.
func (t *TCPConn) Write(b []byte) (int, error) {
	if t.timeout > 0 {
		t.Conn.SetWriteDeadline(time.Now().Add(t.timeout))
	}
	return t.Conn.Write(b)
}

// Accept will block and listen for a connection to it's current listening port. This function will return only
// when a connection is made or it is closed. The return error will most likely be nil unless the listener is closed.
func (t TCPListener) Accept() (net.Conn, error) {
	if d, ok := t.Listener.(deadline); ok {
		d.SetDeadline(time.Now().Add(t.timeout))
	}
	c, err := t.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &TCPConn{timeout: t.timeout, Conn: c}, nil
}

// NewTCP creates a new simple TCP based connector with the supplied timeout.
func NewTCP(t time.Duration) (*TCPConnector, error) {
	return newConnector(netTCP, t, nil)
}
func (t tcpClient) Connect(s string) (net.Conn, error) {
	return t.c.Connect(s)
}

// Connect instructs the connector to create a connection to the supplied address. This function will
// return a connection handle if successful. Otherwise the returned error will be non-nil.
func (t TCPConnector) Connect(s string) (net.Conn, error) {
	c, err := newConn(netTCP, s, t)
	if err != nil {
		return nil, err
	}
	return &TCPConn{timeout: t.dialer.Timeout, Conn: c}, nil
}
func newConn(n, s string, t TCPConnector) (net.Conn, error) {
	if t.tls != nil {
		return tls.DialWithDialer(t.dialer, n, s, t.tls)
	}
	return t.dialer.Dial(n, s)
}

// Listen instructs the connector to create a listener on the supplied listeneing address. This function
// will return a handler to a listener and an error if there are any issues creating the listener.
func (t TCPConnector) Listen(s string) (net.Listener, error) {
	c, err := newListener(netTCP, s, t)
	if err != nil {
		return nil, err
	}
	return &TCPListener{timeout: t.dialer.Timeout, Listener: c}, nil
}
func newListener(n, s string, t TCPConnector) (net.Listener, error) {
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
func NewSecureTCP(t time.Duration, c *tls.Config) (*TCPConnector, error) {
	return newConnector(netTCP, t, c)
}
func newConnector(n string, t time.Duration, c *tls.Config) (*TCPConnector, error) {
	if t < 0 {
		return nil, fmt.Errorf("%d: %w", t, ErrInvalidTimeout)
	}
	switch n {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
	default:
		return nil, fmt.Errorf("%s: %w", n, ErrInvalidNetwork)
	}
	return &TCPConnector{tls: c, dialer: &net.Dialer{Timeout: t, KeepAlive: t, DualStack: true}}, nil
}
