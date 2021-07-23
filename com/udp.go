package com

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/iDigitalFlame/xmt/com/limits"
)

type udpConn struct {
	_      [0]func()
	buf    chan byte
	addr   net.Addr
	parent *udpListener
}
type udpStream struct {
	_ [0]func()
	net.Conn
	timeout time.Duration
}
type udpListener struct {
	socket  net.PacketConn
	delete  chan net.Addr
	active  map[net.Addr]*udpConn
	buf     []byte
	timeout time.Duration
}
type udpConnector struct {
	_      [0]func()
	dialer *net.Dialer
}

func (u *udpConn) Close() error {
	if u.buf == nil {
		close(u.buf)
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
	return u.parent.socket.WriteTo(b, u.addr)
}
func (u *udpStream) Read(b []byte) (int, error) {
	if u.timeout > 0 {
		u.Conn.SetReadDeadline(time.Now().Add(u.timeout))
	}
	n, err := u.Conn.Read(b)
	if err == nil && n < len(b) {
		err = io.EOF
	}
	return n, err
}
func (u *udpStream) Write(b []byte) (int, error) {
	if u.timeout > 0 {
		u.Conn.SetWriteDeadline(time.Now().Add(u.timeout))
	}
	return u.Conn.Write(b)
}

// Accept will block and listen for a connection to it's current listening port. This function
// wil return only when a connection is made or it is closed. The return error will most likely
// be nil unless the listener is closed. This function will return nil for both the connection and
// the error if the connection received was an existing tracked connection or did not complete.
func (u *udpListener) Accept() (net.Conn, error) {
	for len(u.delete) > 0 {
		delete(u.active, <-u.delete)
	}
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
	if err != nil {
		return nil, err
	}
	if a == nil || n <= 1 {
		// Returning nil here as this happens due to a PacketCon hiccup in Golang.
		// Returning an error would trigger a closure of the socket, which we don't want.
		// Both returning nil means that we can continue listening.
		return nil, nil
	}
	c, ok := u.active[a]
	if !ok {
		c = &udpConn{buf: make(chan byte, limits.Buffer), addr: a, parent: u}
		u.active[a] = c
	}
	for i := 0; i < n; i++ {
		c.buf <- u.buf[i]
	}
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
		delete:  make(chan net.Addr, 32),
		socket:  c,
		active:  make(map[net.Addr]*udpConn),
		timeout: u.dialer.Timeout,
	}
	return l, nil
}
