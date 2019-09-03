package tcp

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/iDigitalFlame/xmt/xmt/com/c2"
)

type streamListener struct {
	listen  net.Listener
	network string
	timeout time.Duration
}
type streamConnector struct {
	dial    *net.Dialer
	network string
}
type streamConnection struct {
	timeout time.Duration
	net.Conn
}
type tlsStreamConnector struct {
	dial    *net.Dialer
	config  *tls.Config
	network string
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
	return &streamConnection{Conn: c, timeout: s.timeout}, nil
}
func (s *streamConnector) Listen(a string) (c2.Listener, error) {
	c, err := net.Listen(s.network, a)
	if err != nil {
		return nil, nil
	}
	return &streamListener{network: s.network, listen: c, timeout: s.dial.Timeout}, nil
}

// Connector creates a new simple stream based connector from
// the supplied network type and timeout.  Stream based connectors are
// only valid for TCP and UNIX sockets.  UDP/ICMP/IP will return an
// ErrInvalidNetwork error.
func Connector(n string, t time.Duration) (c2.Connector, error) {
	switch n {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
	default:
		return nil, c2.ErrInvalidNetwork
	}
	return &streamConnector{dial: &net.Dialer{Timeout: t}, network: n}, nil
}
func (s *streamConnector) Connect(a string) (c2.Connection, error) {
	c, err := s.dial.Dial(s.network, a)
	if err != nil {
		return nil, err
	}
	return &streamConnection{Conn: c, timeout: s.dial.Timeout}, nil
}
func (t *tlsStreamConnector) Listen(a string) (c2.Listener, error) {
	if t.config == nil || (t.config.Certificates == nil || len(t.config.Certificates) == 0) || t.config.GetCertificate == nil {
		return nil, ErrInvalidTLSConfig
	}
	c, err := tls.Listen(t.network, a, t.config)
	if err != nil {
		return nil, nil
	}
	return &streamListener{network: t.network, listen: c, timeout: t.dial.Timeout}, nil
}
func (t *tlsStreamConnector) Connect(a string) (c2.Connection, error) {
	c, err := tls.DialWithDialer(t.dial, t.network, a, t.config)
	if err != nil {
		return nil, err
	}
	return &streamConnection{Conn: c, timeout: t.dial.Timeout}, nil
}

// SecureConnector creates a new simple stream based connector from
// the supplied network type and timeout.  Stream based connectors are
// only valid for TCP and UNIX sockets.  UDP/ICMP/IP will return an
// ErrInvalidNetwork error. This stream uses TLS with the provided config.
func SecureConnector(n string, t time.Duration, c *tls.Config) (c2.Connector, error) {
	return SecureServerEx(n, t, nil, c)
}

// SecureServer creates a new simple stream based connector from
// the supplied network type and timeout.  Stream based connectors are
// only valid for TCP and UNIX sockets.  UDP/ICMP/IP will return an
// ErrInvalidNetwork error. This stream uses TLS with the provided certificates.
func SecureServer(n string, t time.Duration, cert *tls.Certificate) (c2.Connector, error) {
	return SecureServerEx(n, t, cert, nil)
}

// SecureServerEx creates a new simple stream based connector from
// the supplied network type and timeout.  Stream based connectors are
// only valid for TCP and UNIX sockets.  UDP/ICMP/IP will return an
// ErrInvalidNetwork error. This stream uses TLS with the provided certificates and configuration.
func SecureServerEx(n string, t time.Duration, cert *tls.Certificate, c *tls.Config) (c2.Connector, error) {
	switch n {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
	default:
		return nil, c2.ErrInvalidNetwork
	}
	if c == nil {
		c = &tls.Config{}
	}
	if cert != nil {
		c.Certificates = []tls.Certificate{*cert}
	}
	return &tlsStreamConnector{dial: &net.Dialer{Timeout: t}, network: n, config: c}, nil
}
