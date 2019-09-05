package udp

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/iDigitalFlame/xmt/xmt/com/c2"
)

var (
	ipInfoSize = 20

	packetDeleteSize = 64
	packetBufferSize = 4096
)

type rawConnection struct {
	*streamConnection
}
type packetListener struct {
	buf     []byte
	del     chan net.Addr
	listen  net.PacketConn
	active  map[net.Addr]*packetConnection
	network string
}
type packetConnector struct {
	dial    *net.Dialer
	network string
}
type streamConnection struct {
	timeout time.Duration
	net.Conn
}
type packetConnection struct {
	buf    chan byte
	addr   net.Addr
	parent *packetListener
}

func (s *streamConnection) IP() string {
	return s.RemoteAddr().String()
}
func (p *packetListener) Close() error {
	defer func() { recover() }()
	for _, v := range p.active {
		if err := v.Close(); err != nil {
			return err
		}
	}
	close(p.del)
	return p.listen.Close()
}
func (p *packetConnection) IP() string {
	return p.addr.String()
}
func (p *packetConnection) Close() error {
	defer func() { recover() }()
	if p.parent != nil {
		p.parent = nil
		close(p.buf)
	}
	return nil
}
func (p *packetListener) String() string {
	return fmt.Sprintf("Packet(%s) %s", strings.ToUpper(p.network), p.listen.LocalAddr().String())
}
func (r *rawConnection) Read(b []byte) (int, error) {
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
func (s *streamConnection) Read(b []byte) (int, error) {
	if s.timeout > 0 {
		s.Conn.SetReadDeadline(time.Now().Add(s.timeout))
	}
	return s.Conn.Read(b)
}
func (p *packetConnection) Read(b []byte) (int, error) {
	var n int
	if len(p.buf) == 0 {
		return 0, io.EOF
	}
	for ; len(p.buf) > 0 && n < len(b); n++ {
		b[n] = <-p.buf
	}
	return n, nil
}
func (p *packetConnection) Write(b []byte) (int, error) {
	return p.parent.listen.WriteTo(b, p.addr)
}
func (p *packetListener) Accept() (c2.Connection, error) {
	for x := 0; x < len(p.del); x++ {
		delete(p.active, <-p.del)
	}
	n, a, err := p.listen.ReadFrom(p.buf)
	if err != nil {
		return nil, err
	}
	if a == nil || n <= 1 {
		return nil, nil
	}
	c, ok := p.active[a]
	if !ok {
		c = &packetConnection{
			buf:    make(chan byte, packetBufferSize),
			addr:   a,
			parent: p,
		}
		p.active[a] = c
	}
	for x := 0; x < n; x++ {
		c.buf <- p.buf[x]
	}
	if !ok {
		return c, nil
	}
	return nil, nil
}

// Connector creates a new packet based connector from the supplied
// network type and timeout. Packet based connectors are only valid for UDP,
// Datagram and IP sockets. TCP/Unix will return an ErrInvalidNetwork error.
func Connector(n string, t time.Duration) (c2.Connector, error) {
	switch n {
	case "udp", "udp4", "udp6", "unixgram":
	default:
		if len(n) <= 3 || n[0] != 'i' || n[1] == 'p' || n[2] == ':' {
			return nil, c2.ErrInvalidNetwork
		}
	}
	return &packetConnector{
		dial: &net.Dialer{
			Timeout:   t,
			KeepAlive: t,
			DualStack: true,
		},
		network: n,
	}, nil
}
func (p *packetConnector) Listen(a string) (c2.Listener, error) {
	l, err := net.ListenPacket(p.network, a)
	if err != nil {
		return nil, err
	}
	c := &packetListener{
		buf:     make([]byte, packetBufferSize),
		del:     make(chan net.Addr, packetDeleteSize),
		listen:  l,
		active:  make(map[net.Addr]*packetConnection),
		network: p.network,
	}
	return c, nil
}
func (p *packetConnector) Connect(a string) (c2.Connection, error) {
	c, err := p.dial.Dial(p.network, a)
	if err != nil {
		return nil, err
	}
	if p.network[0] == 'i' && p.network[1] == 'p' {
		return &rawConnection{
			&streamConnection{
				Conn:    c,
				timeout: p.dial.Timeout,
			},
		}, nil
	}
	return &streamConnection{
		Conn:    c,
		timeout: p.dial.Timeout,
	}, nil
}
