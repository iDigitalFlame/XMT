package udp

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/com/limits"
)

var ipInfoSize = 20

type raw struct {
	*stream
}
type conn struct {
	buf    chan byte
	addr   net.Addr
	parent *listener
}
type stream struct {
	timeout time.Duration
	net.Conn
}
type provider struct {
	dial    *net.Dialer
	network string
}
type listener struct {
	buf     []byte
	del     chan net.Addr
	listen  net.PacketConn
	active  map[net.Addr]*conn
	network string
}

func (c *conn) Close() error {
	defer func() { recover() }()
	if c.parent != nil {
		c.parent = nil
		close(c.buf)
	}
	return nil
}
func (l *listener) Close() error {
	defer func() { recover() }()
	for _, v := range l.active {
		if err := v.Close(); err != nil {
			return err
		}
	}
	close(l.del)
	return l.listen.Close()
}
func (l listener) String() string {
	return fmt.Sprintf("Packet(%s) %s", strings.ToUpper(l.network), l.listen.LocalAddr().String())
}
func (l listener) Addr() net.Addr {
	return l.listen.LocalAddr()
}
func (c conn) LocalAddr() net.Addr {
	return c.addr
}
func (c conn) RemoteAddr() net.Addr {
	return c.addr
}
func (r *raw) Read(b []byte) (int, error) {
	if r.timeout > 0 {
		r.Conn.SetReadDeadline(time.Now().Add(r.timeout))
	}
	n, err := r.Conn.Read(b)
	if n > ipInfoSize {
		copy(b, b[ipInfoSize:])
		n -= ipInfoSize
	}
	return n, err
}
func (c *conn) Read(b []byte) (int, error) {
	var n int
	if len(c.buf) == 0 {
		return 0, io.EOF
	}
	for ; len(c.buf) > 0 && n < len(b); n++ {
		b[n] = <-c.buf
	}
	return n, nil
}
func (conn) SetDeadline(_ time.Time) error {
	return nil
}
func (c *conn) Write(b []byte) (int, error) {
	return c.parent.listen.WriteTo(b, c.addr)
}
func (s *stream) Read(b []byte) (int, error) {
	if s.timeout > 0 {
		s.Conn.SetReadDeadline(time.Now().Add(s.timeout))
	}
	return s.Conn.Read(b)
}
func (l *listener) Accept() (net.Conn, error) {
	for x := 0; x < len(l.del); x++ {
		delete(l.active, <-l.del)
	}
	n, a, err := l.listen.ReadFrom(l.buf)
	if err != nil {
		return nil, err
	}
	if a == nil || n <= 1 {
		return nil, nil
	}
	c, ok := l.active[a]
	if !ok {
		c = &conn{
			buf:    make(chan byte, limits.LargeLimit()),
			addr:   a,
			parent: l,
		}
		l.active[a] = c
	}
	for x := 0; x < n; x++ {
		c.buf <- l.buf[x]
	}
	if !ok {
		return c, nil
	}
	return nil, nil
}
func (conn) SetReadDeadline(_ time.Time) error {
	return nil
}
func (conn) SetWriteDeadline(_ time.Time) error {
	return nil
}
func (p *provider) Connect(s string) (net.Conn, error) {
	c, err := p.dial.Dial(p.network, s)
	if err != nil {
		return nil, err
	}
	if p.network[0] == 'i' && p.network[1] == 'p' {
		return &raw{&stream{
			Conn:    c,
			timeout: p.dial.Timeout,
		}}, nil
	}
	return &stream{
		Conn:    c,
		timeout: p.dial.Timeout,
	}, nil
}
func (p *provider) Listen(s string) (net.Listener, error) {
	l, err := net.ListenPacket(p.network, s)
	if err != nil {
		return nil, err
	}
	c := &listener{
		buf:     make([]byte, limits.LargeLimit()),
		del:     make(chan net.Addr, limits.SmallLimit()),
		listen:  l,
		active:  make(map[net.Addr]*conn),
		network: p.network,
	}
	return c, nil
}

// NewRaw creates a new packet based connector from the supplied
// network type and timeout. Packet based connectors are only valid for UDP,
// Datagram and IP sockets. TCP/Unix will return an ErrInvalidNetwork error.
func NewRaw(n string, t time.Duration) (com.Provider, error) {
	switch n {
	case "udp", "udp4", "udp6", "unixgram":
	default:
		if len(n) <= 3 || n[0] != 'i' || n[1] == 'p' || n[2] == ':' {
			return nil, com.ErrInvalidNetwork
		}
	}
	return &provider{
		dial: &net.Dialer{
			Timeout:   t,
			KeepAlive: t,
			DualStack: true,
		},
		network: n,
	}, nil
}
