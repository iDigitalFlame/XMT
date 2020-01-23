package com

import (
	"fmt"
	"net"
	"time"
)

// IPStream is a struct that represents a IP stream (direct) based network connection.
// This struct can be used to control and manage the current connection.
type IPStream struct {
	net.Conn
}

// IPListener is a struct that represents a IP based network connection listener.
// This struct can be used to accept and create new IP connections.
type IPListener struct {
	net.Listener
}

// IPConnector is a struct that represents a IP based network connection handler.
// This struct can be used to create new IP listeners.
type IPConnector struct {
	protocol byte
	UDPConnector
}

// String returns a string representation of this UNIXListener.
func (i IPListener) String() string {
	return fmt.Sprintf("IP[%s]", i.Addr().String())
}

// NewIP creates a new simple IP based connector with the supplied timeout and
// protocol number.
func NewIP(p byte, t time.Duration) *IPConnector {
	return &IPConnector{
		protocol: p,
		UDPConnector: UDPConnector{
			dialer: &net.Dialer{
				Timeout:   t,
				KeepAlive: t,
				DualStack: true,
			},
		},
	}
}

// Connect instructs the connector to create a connection to the supplied address. This function will
// return a connection handle if successful. Otherwise the returned error will be non-nil.
func (i IPConnector) Connect(s string) (net.Conn, error) {
	c, err := i.UDPConnector.Connect(s)
	if err != nil {
		return nil, err
	}
	return &IPStream{c}, nil
}

// Listen instructs the connector to create a listener on the supplied listeneing address. This function
// will return a handler to a listener and an error if there are any issues creating the listener.
func (i IPConnector) Listen(s string) (net.Listener, error) {
	l, err := i.UDPConnector.Listen(s)
	if err != nil {
		return nil, err
	}
	return &IPListener{l}, nil
}
