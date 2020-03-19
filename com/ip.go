package com

import (
	"fmt"
	"net"
	"time"

	"github.com/iDigitalFlame/xmt/com/limits"
)

// IPStream is a struct that represents a IP stream (direct) based network connection. This struct can
// be used to control and manage the current connection.
type IPStream struct {
	timeout time.Duration
	net.Conn
}

// IPListener is a struct that represents a IP based network connection listener. This struct can
// be used to accept and create new IP connections.
type IPListener struct {
	proto byte
	net.Listener
}

// IPConnector is a struct that represents a IP based network connection handler. This struct can be
// used to create new IP listeners.
type IPConnector struct {
	proto  byte
	dialer *net.Dialer
}

// String returns a string representation of this UNIXListener.
func (i IPListener) String() string {
	return fmt.Sprintf("IP:%d[%s]", i.proto, i.Addr().String())
}

// Read will attempt to read len(b) bytes from the current connection and fill the supplied buffer.
// The return values will be the amount of bytes read and any errors that occurred.
func (i *IPStream) Read(b []byte) (int, error) {
	if i.timeout > 0 {
		i.Conn.SetReadDeadline(time.Now().Add(i.timeout))
	}
	n, err := i.Conn.Read(b)
	if n > 20 {
		copy(b, b[20:])
		n -= 20
	}
	return n, err
}

// NewIP creates a new simple IP based connector with the supplied timeout and protocol number.
func NewIP(p byte, t time.Duration) *IPConnector {
	return &IPConnector{proto: p, dialer: &net.Dialer{Timeout: t, KeepAlive: t, DualStack: true}}
}

// Connect instructs the connector to create a connection to the supplied address. This function will
// return a connection handle if successful. Otherwise the returned error will be non-nil.
func (i IPConnector) Connect(s string) (net.Conn, error) {
	c, err := i.dialer.Dial(fmt.Sprintf(netIP, i.proto), s)
	if err != nil {
		return nil, err
	}
	return &IPStream{timeout: i.dialer.Timeout, Conn: c}, nil
}

// Listen instructs the connector to create a listener on the supplied listeneing address. This function
// will return a handler to a listener and an error if there are any issues creating the listener.
func (i IPConnector) Listen(s string) (net.Listener, error) {
	c, err := net.ListenPacket(fmt.Sprintf(netIP, i.proto), s)
	if err != nil {
		return nil, err
	}
	l := &IPListener{
		proto: i.proto,
		Listener: &UDPListener{
			buf:    make([]byte, limits.LargeLimit()),
			delete: make(chan net.Addr, limits.SmallLimit()),
			socket: c,
			active: make(map[net.Addr]*UDPConn),
		},
	}
	return l, nil
}
