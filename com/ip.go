package com

import (
	"context"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/iDigitalFlame/xmt/util/bugtrack"
)

type ipStream struct {
	udpStream
}
type ipListener struct {
	net.Listener
	proto byte
}
type ipConnector struct {
	net.Dialer
	proto byte
}

func (i *ipListener) String() string {
	return "IP:" + strconv.Itoa(int(i.proto)) + "/" + i.Addr().String()
}

// NewIP creates a new simple IP based connector with the supplied timeout and protocol number.
func NewIP(t time.Duration, p byte) Connector {
	return &ipConnector{proto: p, Dialer: net.Dialer{Timeout: t, KeepAlive: t, DualStack: true}}
}
func (i *ipStream) Read(b []byte) (int, error) {
	n, err := i.udpStream.Read(b)
	if n > 20 {
		if bugtrack.Enabled {
			bugtrack.Track("com.ipStream.Read(): Cutting off IP header n=%d, after n=%d", n, n-20)
		}
		copy(b, b[20:])
		n -= 20
	}
	if err == nil && n < len(b)-20 {
		err = io.EOF
	}
	return n, err
}
func (i *ipConnector) Connect(s string) (net.Conn, error) {
	c, err := i.Dial("ip:"+strconv.Itoa(int(i.proto)), s)
	if err != nil {
		return nil, err
	}
	return &ipStream{udpStream{Conn: c}}, nil
}
func (i *ipConnector) Listen(s string) (net.Listener, error) {
	c, err := ListenConfig.ListenPacket(context.Background(), "ip:"+strconv.Itoa(int(i.proto)), s)
	if err != nil {
		return nil, err
	}
	l := &udpListener{
		new:  make(chan *udpConn, 16),
		del:  make(chan udpAddr, 16),
		cons: make(map[udpAddr]*udpConn),
		sock: c,
	}
	l.ctx, l.cancel = context.WithCancel(context.Background())
	go l.purge()
	go l.listen()
	return &ipListener{proto: i.proto, Listener: l}, nil
}
