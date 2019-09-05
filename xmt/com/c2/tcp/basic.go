package tcp

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/iDigitalFlame/xmt/xmt/com/c2"
)

type streamClient struct {
	c *streamConnector
}
type streamListener struct {
	listen  net.Listener
	network string
	timeout time.Duration
}
type streamConnector struct {
	tls     *tls.Config
	dial    *net.Dialer
	network string
}
type streamConnection struct {
	timeout time.Duration
	net.Conn
}

func (s *streamListener) Close() error {
	return s.listen.Close()
}
func (s *streamConnection) IP() string {
	return s.RemoteAddr().String()
}
func (s *streamListener) String() string {
	return fmt.Sprintf("Stream(%s) %s", strings.ToUpper(s.network), s.listen.Addr().String())
}
func (s *streamConnection) Read(b []byte) (int, error) {
	if s.timeout > 0 {
		s.Conn.SetReadDeadline(time.Now().Add(s.timeout))
	}
	return s.Conn.Read(b)
}
func (s *streamListener) Accept() (c2.Connection, error) {
	c, err := s.listen.Accept()
	if err != nil {
		return nil, err
	}
	return &streamConnection{
		Conn:    c,
		timeout: s.timeout,
	}, nil
}
func (s *streamConnector) Listen(a string) (c2.Listener, error) {
	var err error
	var c net.Listener
	if s.tls != nil {
		if (s.tls.Certificates == nil || len(s.tls.Certificates) == 0) || s.tls.GetCertificate == nil {
			return nil, ErrInvalidTLSConfig
		}
		c, err = tls.Listen(s.network, a, s.tls)
	} else {
		c, err = net.Listen(s.network, a)
	}
	if err != nil {
		return nil, err
	}
	return &streamListener{
		listen:  c,
		network: s.network,
		timeout: s.dial.Timeout,
	}, nil
}

// Connector creates a new simple stream based connector from
// the supplied network type and timeout.  Stream based connectors are
// only valid for TCP and UNIX sockets.  UDP/ICMP/IP will return an
// ErrInvalidNetwork error.
func Connector(n string, t time.Duration) (c2.Connector, error) {
	return ConnectorSecure(n, t, nil)
}
func (s *streamClient) Connect(a string) (c2.Connection, error) {
	return s.c.Connect(a)
}
func (s *streamConnector) Connect(a string) (c2.Connection, error) {
	var err error
	var c net.Conn
	if s.tls != nil {
		c, err = tls.DialWithDialer(s.dial, s.network, a, s.tls)
	} else {
		c, err = s.dial.Dial(s.network, a)
	}
	if err != nil {
		return nil, err
	}
	return &streamConnection{
		Conn:    c,
		timeout: s.dial.Timeout,
	}, nil
}

// ConnectorSecure creates a new simple stream based connector from
// the supplied network type and timeout.  Stream based connectors are
// only valid for TCP and UNIX sockets.  UDP/ICMP/IP will return an
// ErrInvalidNetwork error. This stream uses TLS with the provided config.
func ConnectorSecure(n string, t time.Duration, c *tls.Config) (c2.Connector, error) {
	switch n {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
	default:
		return nil, c2.ErrInvalidNetwork
	}
	return &streamConnector{
		tls: c,
		dial: &net.Dialer{
			Timeout:   t,
			KeepAlive: t,
			DualStack: true,
		},
		network: n,
	}, nil
}
