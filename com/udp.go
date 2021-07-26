package com

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/iDigitalFlame/xmt/util/xerr"

	"github.com/iDigitalFlame/xmt/com/limits"
)

const udpLimit = 32768

type udpConn struct {
	_      [0]func()
	buf    chan byte
	addr   net.Addr
	ident  string
	parent *udpListener
}
type udpStream struct {
	_ [0]func()
	net.Conn
	timeout time.Duration
}
type udpListener struct {
	socket  net.PacketConn
	delete  chan string
	active  map[string]*udpConn
	buf     []byte
	timeout time.Duration
}
type udpConnector struct {
	_      [0]func()
	dialer *net.Dialer
}

func (u *udpConn) Close() error {
	if u.parent != nil {
		close(u.buf)
		u.parent.delete <- u.ident
		u.buf, u.parent = nil, nil
	}
	return nil
}
func (u *udpListener) Close() error {
	if u.delete == nil || u.socket == nil {
		return nil
	}
	for _, v := range u.active {
		if err := v.Close(); err != nil {
			return err
		}
	}
	close(u.delete)
	err := u.socket.Close()
	u.delete, u.socket = nil, nil
	return err
}
func (u udpListener) String() string {
	return "UDP[" + u.socket.LocalAddr().String() + "]"
}
func (u udpListener) Addr() net.Addr {
	return u.socket.LocalAddr()
}
func (u udpConn) LocalAddr() net.Addr {
	return u.addr
}
func (u udpConn) RemoteAddr() net.Addr {
	return u.addr
}

// NewUDP creates a new simple UDP based connector with the supplied timeout.
func NewUDP(t time.Duration) Connector {
	return &udpConnector{dialer: &net.Dialer{Timeout: t, KeepAlive: t, DualStack: true}}
}
func (u *udpConn) Read(b []byte) (int, error) {
	if len(u.buf) == 0 || u.parent == nil {
		return 0, io.EOF
	}
	var n int
	for ; len(u.buf) > 0 && n < len(b); n++ {
		b[n] = <-u.buf
	}
	return n, nil
}
func (udpConn) SetDeadline(_ time.Time) error {
	return nil
}
func (u *udpConn) Write(b []byte) (int, error) {
	if u.parent == nil {
		return 0, io.ErrUnexpectedEOF
	}
	var (
		n, c int
		err  error
	)
	for s, x := 0, udpLimit; n < len(b) && s < len(b); {
		if x > len(b) {
			x = len(b)
		}
		if c, err = u.parent.socket.WriteTo(b[s:x], u.addr); err != nil {
			break
		}
		s += c
		x += c
		n += c
	}
	return n, err
}
func (u *udpStream) Read(b []byte) (int, error) {
	if u.timeout > 0 {
		u.Conn.SetReadDeadline(time.Now().Add(u.timeout))
	}
	var (
		n, c int
		err  error
	)
	for s, x := 0, udpLimit; n < len(b) && s < len(b); {
		if x > len(b) {
			x = len(b)
		}
		//println("read ", x-s)
		if c, err = u.Conn.Read(b[s:x]); err != nil {
			break
		}
		n += c
		//println(c, err, x-s)
		if c < udpLimit {
			break
		}
		s += c
		x += c
	}
	//println(c, err, n)
	return n, err
}
func (u *udpStream) Write(b []byte) (int, error) {
	if u.timeout > 0 {
		u.Conn.SetWriteDeadline(time.Now().Add(u.timeout))
	}
	var (
		n, c int
		err  error
	)
	for s, x := 0, udpLimit; n < len(b) && s < len(b); {
		if x > len(b) {
			x = len(b)
		}
		if c, err = u.Conn.Write(b[s:x]); err != nil {
			break
		}
		s += c
		x += c
		n += c
	}
	return n, err
}

// Accept will block and listen for a connection to it's current listening port. This function
// wil return only when a connection is made or it is closed. The return error will most likely
// be nil unless the listener is closed. This function will return nil for both the connection and
// the error if the connection received was an existing tracked connection or did not complete.
func (u *udpListener) Accept() (net.Conn, error) {
	if u.socket == nil {
		return nil, io.ErrClosedPipe
	}
	if u.timeout > 0 {
		u.socket.SetDeadline(time.Now().Add(u.timeout))
	}
	// note: Apparently there's a bug in this method about not getting the full length of
	// the packet from this call?
	// Not that we'll need it but I think I should note this down just incase I need to bang my
	// head against the desk when the connector doesn't work.
	n, a, err := u.socket.ReadFrom(u.buf)
	for len(u.delete) > 0 {
		delete(u.active, <-u.delete)
	}
	if n == 0 && err != nil {
		return nil, err
	}
	//println("UDP: Read", n, a.String())
	if a == nil || n <= 1 {
		// Returning nil here as this happens due to a PacketCon hiccup in Golang.
		// Returning an error would trigger a closure of the socket, which we don't want.
		// Both returning nil means that we can continue listening.
		return nil, nil
	}
	var s string
	// TODO: Replace this with a compound integer.
	switch i := a.(type) {
	case *net.IPAddr:
		s = i.IP.String()
	case *net.UDPAddr:
		s = i.IP.String()
	default:
		return nil, xerr.New(`invalid type "` + a.Network() + `" supplied to UDP listener`)
	}
	c, ok := u.active[s]
	if !ok {
		c = &udpConn{buf: make(chan byte, limits.Buffer), addr: a, ident: s, parent: u}
		u.active[s] = c
	}
	//println("UDP: Push", n, "bytes")
	for i := 0; i < n; i++ {
		c.buf <- u.buf[i]
	}
	//println("UDP: Push", n, "bytes done, ret", !ok)
	if !ok {
		return c, nil
	}
	return nil, nil
}
func (udpConn) SetReadDeadline(_ time.Time) error {
	return nil
}
func (udpConn) SetWriteDeadline(_ time.Time) error {
	return nil
}
func (u udpConnector) Connect(s string) (net.Conn, error) {
	c, err := u.dialer.Dial(netUDP, s)
	if err != nil {
		return nil, err
	}
	return &udpStream{Conn: c, timeout: u.dialer.Timeout}, nil
}
func (u udpConnector) Listen(s string) (net.Listener, error) {
	c, err := ListenConfig.ListenPacket(context.Background(), netUDP, s)
	if err != nil {
		return nil, err
	}
	l := &udpListener{
		buf:     make([]byte, limits.Buffer),
		delete:  make(chan string, 32),
		socket:  c,
		active:  make(map[string]*udpConn),
		timeout: u.dialer.Timeout,
	}
	return l, nil
}
