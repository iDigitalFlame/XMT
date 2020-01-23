package com

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

// UNIXListener is a struct that represents a UNIX socket based connection listener.
// This struct can be used to accept and create new UNIX socket connections.
type UNIXListener struct {
	TCPListener
}

// UNIXConnector is a struct that represents a UNIX socket based connection handler.
// This struct can be used to create new UNIX socket listeners.
type UNIXConnector struct {
	TCPConnector
}

// String returns a string representation of this UNIXListener.
func (u UNIXListener) String() string {
	return fmt.Sprintf("UNIX[%s]", u.Addr().String())
}

// NewUNIX creates a new simple UNIX socket based connector with the supplied timeout.
func NewUNIX(t time.Duration) (*UNIXConnector, error) {
	n, err := newConnector(netUNIX, t, nil)
	if err != nil {
		return nil, err
	}
	return &UNIXConnector{TCPConnector: *n}, nil
}

// Connect instructs the connector to create a connection to the supplied address. This function will
// return a connection handle if successful. Otherwise the returned error will be non-nil.
func (u UNIXConnector) Connect(s string) (net.Conn, error) {
	return newConn(netUNIX, s, u.TCPConnector)
}

// Listen instructs the connector to create a listener on the supplied listeneing address. This function
// will return a handler to a listener and an error if there are any issues creating the listener.
func (u UNIXConnector) Listen(s string) (net.Listener, error) {
	c, err := newListener(netUNIX, s, u.TCPConnector)
	if err != nil {
		return nil, err
	}
	return &TCPListener{
		timeout:  u.TCPConnector.dialer.Timeout,
		Listener: c,
	}, nil
}

// NewSecureUNIX creates a new simple TLS wrapped UNIX socket based connector with the supplied timeout.
func NewSecureUNIX(t time.Duration, c *tls.Config) (*UNIXConnector, error) {
	n, err := newConnector(netUNIX, t, c)
	if err != nil {
		return nil, err
	}
	return &UNIXConnector{TCPConnector: *n}, nil
}
