package com

import (
	"context"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/iDigitalFlame/xmt/com/limits"
)

type ipStream struct {
	net.Conn
	timeout time.Duration
}
type ipListener struct {
	net.Listener
	proto byte
}
type ipConnector struct {
	dialer *net.Dialer
	proto  byte
}

func (i ipListener) String() string {
	return "IP:" + strconv.Itoa(int(i.proto)) + "[" + i.Addr().String() + "]"
}

// NewIP creates a new simple IP based connector with the supplied timeout and protocol number.
func NewIP(p byte, t time.Duration) Connector {
	return &ipConnector{proto: p, dialer: &net.Dialer{Timeout: t, KeepAlive: t, DualStack: true}}
}
func (i *ipStream) Read(b []byte) (int, error) {
	if i.timeout > 0 {
		i.Conn.SetReadDeadline(time.Now().Add(i.timeout))
	}
	n, err := i.Conn.Read(b)
	if n > 20 {
		copy(b, b[20:])
		n -= 20
	}
	if err == nil && n < len(b)-20 {
		err = io.EOF
	}
	return n, err
}
func (i ipConnector) Connect(s string) (net.Conn, error) {
	c, err := i.dialer.Dial("ip:"+strconv.Itoa(int(i.proto)), s)
	if err != nil {
		return nil, err
	}
	return &ipStream{timeout: i.dialer.Timeout, Conn: c}, nil
}
func (i ipConnector) Listen(s string) (net.Listener, error) {
	c, err := ListenConfig.ListenPacket(context.Background(), "ip:"+strconv.Itoa(int(i.proto)), s)
	if err != nil {
		return nil, err
	}
	l := &ipListener{
		proto: i.proto,
		Listener: &udpListener{
			buf:     make([]byte, limits.Buffer),
			delete:  make(chan net.Addr, 32),
			socket:  c,
			active:  make(map[net.Addr]*udpConn),
			timeout: i.dialer.Timeout,
		},
	}
	return l, nil
}
