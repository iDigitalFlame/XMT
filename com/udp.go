package com

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/iDigitalFlame/xmt/com/limits"
)

// UDPConn is a struct that represents a UDP based network connection. This struct can
// be used to control and manage the current connection.
type UDPConn struct {
	buf    chan byte
	addr   net.Addr
	parent *UDPListener
}

// UDPStream is a struct that represents a UDP stream (direct) based network connection. This struct
// can be used to control and manage the current connection.
type UDPStream struct {
	timeout time.Duration
	net.Conn
}

// UDPListener is a struct that represents a UDP based network connection listener. This struct can
// be used to accept and create new UDP connections.
type UDPListener struct {
	buf     []byte
	delete  chan net.Addr
	socket  net.PacketConn
	active  map[net.Addr]*UDPConn
	timeout time.Duration
}

// UDPConnector is a struct that represents a UDP based network connection handler. This struct
// can be used to create new UDP listeners.
type UDPConnector struct {
	dialer *net.Dialer
}

// Close closes this connetion and frees any related resources.
func (u *UDPConn) Close() error {
	if u.buf == nil {
		close(u.buf)
		u.buf, u.parent = nil, nil
	}
	return nil
}

// Close closes this listener. Any blocked Accept operations will be unblocked and return errors.
func (u *UDPListener) Close() error {
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

// String returns a string representation of this UDPListener.
func (u UDPListener) String() string {
	return fmt.Sprintf("UDP[%s]", u.socket.LocalAddr().String())
}

// Addr returns the listener's current bound network address.
func (u UDPListener) Addr() net.Addr {
	return u.socket.LocalAddr()
}

// LocalAddr returns the connected remote network address.
func (u UDPConn) LocalAddr() net.Addr {
	return u.addr
}

// RemoteAddr returns the connected remote network address.
func (u UDPConn) RemoteAddr() net.Addr {
	return u.addr
}

// NewUDP creates a new simple UDP based connector with the supplied timeout.
func NewUDP(t time.Duration) *UDPConnector {
	return &UDPConnector{dialer: &net.Dialer{Timeout: t, KeepAlive: t, DualStack: true}}
}

// Read will attempt to read len(b) bytes from the current connection and fill the supplied buffer.
// The return values will be the amount of bytes read and any errors that occurred.
func (u *UDPConn) Read(b []byte) (int, error) {
	if len(u.buf) == 0 || u.parent == nil {
		return 0, io.EOF
	}
	var n int
	for ; len(u.buf) > 0 && n < len(b); n++ {
		b[n] = <-u.buf
	}
	return n, nil
}

// SetDeadline sets the read and write deadlines associated with the connection. This function
// does nothing for this type of connection.
func (UDPConn) SetDeadline(_ time.Time) error {
	return nil
}

// Write will attempt to write len(b) bytes to the current connection from the supplied buffer.
// The return values will be the amount of bytes wrote and any errors that occurred.
func (u *UDPConn) Write(b []byte) (int, error) {
	if u.parent == nil {
		return 0, io.ErrUnexpectedEOF
	}
	return u.parent.socket.WriteTo(b, u.addr)
}

// Read will attempt to read len(b) bytes from the current connection and fill the supplied buffer.
// The return values will be the amount of bytes read and any errors that occurred.
func (u *UDPStream) Read(b []byte) (int, error) {
	if u.timeout > 0 {
		u.Conn.SetReadDeadline(time.Now().Add(u.timeout))
	}
	return u.Conn.Read(b)
}

// Write will attempt to write len(b) bytes to the current connection from the supplied buffer.
// The return values will be the amount of bytes wrote and any errors that occurred.
func (u *UDPStream) Write(b []byte) (int, error) {
	if u.timeout > 0 {
		u.Conn.SetWriteDeadline(time.Now().Add(u.timeout))
	}
	return u.Conn.Write(b)
}

// Accept will block and listen for a connection to it's current listening port. This function
// wil return only when a connection is made or it is closed. The return error will most likely
// be nil unless the listener is closed. This function will return nil for both the connection and
// the error if the connection received was an existing tracked connection or did not complete.
func (u *UDPListener) Accept() (net.Conn, error) {
	for len(u.delete) > 0 {
		delete(u.active, <-u.delete)
	}
	if u.socket == nil {
		return nil, io.ErrClosedPipe
	}
	if u.timeout > 0 {
		u.socket.SetDeadline(time.Now().Add(u.timeout))
	}
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
		c = &UDPConn{
			buf:    make(chan byte, limits.LargeLimit()),
			addr:   a,
			parent: u,
		}
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

// SetReadDeadline sets the deadline for future Read calls and any currently-blocked Read call. This
// function does nothing for this type of connection.
func (UDPConn) SetReadDeadline(_ time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls and any currently-blocked Write call. This
// function does nothing for this type of connection.
func (UDPConn) SetWriteDeadline(_ time.Time) error {
	return nil
}

// Connect instructs the connector to create a connection to the supplied address. This function will
// return a connection handle if successful. Otherwise the returned error will be non-nil.
func (u UDPConnector) Connect(s string) (net.Conn, error) {
	c, err := u.dialer.Dial(netUDP, s)
	if err != nil {
		return nil, err
	}
	return &UDPStream{Conn: c, timeout: u.dialer.Timeout}, nil
}

// Listen instructs the connector to create a listener on the supplied listeneing address. This function
// will return a handler to a listener and an error if there are any issues creating the listener.
func (u UDPConnector) Listen(s string) (net.Listener, error) {
	c, err := ListenConfig.ListenPacket(context.Background(), netUDP, s)
	if err != nil {
		return nil, err
	}
	l := &UDPListener{
		buf:     make([]byte, limits.LargeLimit()),
		delete:  make(chan net.Addr, limits.SmallLimit()),
		socket:  c,
		active:  make(map[net.Addr]*UDPConn),
		timeout: u.dialer.Timeout,
	}
	return l, nil
}
