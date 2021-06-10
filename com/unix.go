package com

import (
	"crypto/tls"
	"net"
	"time"
)

type unixConnector struct {
	tcpConnector
}

// NewUNIX creates a new simple UNIX socket based connector with the supplied timeout.
func NewUNIX(t time.Duration) (Connector, error) {
	n, err := newConnector(netUNIX, t, nil)
	if err != nil {
		return nil, err
	}
	return &unixConnector{tcpConnector: *n}, nil
}
func (u unixConnector) Connect(s string) (net.Conn, error) {
	return newConn(netUNIX, s, u.tcpConnector)
}
func (u unixConnector) Listen(s string) (net.Listener, error) {
	c, err := newListener(netUNIX, s, u.tcpConnector)
	if err != nil {
		return nil, err
	}
	return &tcpListener{timeout: u.tcpConnector.dialer.Timeout, Listener: c}, nil
}

// NewSecureUNIX creates a new simple TLS wrapped UNIX socket based connector with the supplied timeout.
func NewSecureUNIX(t time.Duration, c *tls.Config) (Connector, error) {
	n, err := newConnector(netUNIX, t, c)
	if err != nil {
		return nil, err
	}
	return &unixConnector{tcpConnector: *n}, nil
}
