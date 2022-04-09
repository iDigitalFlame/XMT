package com

import (
	"context"
	"crypto/tls"
	"net"
	"time"
)

// DefaultTimeout is the default timeout used for the default connectors.
// The default is 15 seconds.
const DefaultTimeout = time.Second * 15 //30

// ListenConfig is the default listener config that is used to generate the
// Listeners. This can be used to specify the listen 'KeepAlive' timeout.
var ListenConfig = net.ListenConfig{KeepAlive: DefaultTimeout}

var (
	// TCP is the TCP Raw connector. This connector uses raw TCP connections for
	// communication.
	TCP = NewTCP(DefaultTimeout)

	// UDP is the UDP Raw connector. This connector uses raw UDP connections for
	// communication.
	UDP = NewUDP(DefaultTimeout)

	// ICMP is the ICMP Raw connector. This connector uses raw ICMP connections
	// for communication.
	//
	// TODO(dij): I think ICMP is bugged ATM, "NewIP(<anything greater than 1>, DefaultTimeout)" works, weird.
	ICMP = NewIP(DefaultTimeout, 1)

	// TLS is the TCP over TLS connector client. This client uses TCP wrapped in
	// TLS encryption using certificates.
	//
	// This client is only valid for clients that connect to servers with properly
	// signed and trusted certificates.
	TLS = &tcpClient{c: tcpConnector{tls: new(tls.Config), Dialer: TCP.(*tcpConnector).Dialer}}

	// TLSInsecure is the TCP over TLS connector profile. This client uses TCP
	// wrapped in TLS encryption using certificates.
	//
	// This instance DOES NOT check the server certificate for validity.
	TLSInsecure = &tcpClient{c: tcpConnector{tls: &tls.Config{InsecureSkipVerify: true}, Dialer: TCP.(*tcpConnector).Dialer}}
)

type deadliner interface {
	SetDeadline(time.Time) error
}

// Connector is an interface that represents an object that can create and
// establish connections on various protocols.
type Connector interface {
	Connect(context.Context, string) (net.Conn, error)
	Listen(context.Context, string) (net.Listener, error)
}

// DialTCP is a quick utility function that can be used to quickly create a TCP
// connection to the provided address.
//
// This function uses the 'com.TCP' var.
func DialTCP(x context.Context, s string) (net.Conn, error) {
	return TCP.Connect(x, s)
}

// ListenTCP is a quick utility function that can be used to quickly create a
// TCP listener using the 'TCP' Acceptor.
func ListenTCP(x context.Context, s string) (net.Listener, error) {
	return TCP.Listen(x, s)
}

// DialTLS is a quick utility function that can be used to quickly create a TLS
// connection to the provided address.
//
// This function uses the 'com.TLS' var if the provided tls config is nil.
func DialTLS(x context.Context, s string, c *tls.Config) (net.Conn, error) {
	if c == nil {
		return TLS.Connect(x, s)
	}
	return newStreamConn(x, NameTCP, s, tcpConnector{tls: c, Dialer: TCP.(*tcpConnector).Dialer})
}

// SetListenerDeadline attempts to set a deadline on the 'Accept; function of a
// Listener if applicable. This function will return any errors if they occur
// and always returns 'nil' if the Listener does not support deadlines.
func SetListenerDeadline(l net.Listener, t time.Time) error {
	if d, ok := l.(deadliner); ok {
		return d.SetDeadline(t)
	}
	return nil
}

// ListenTLS is a quick utility function that can be used to quickly create a TLS
// listener using the provided TLS config.
func ListenTLS(x context.Context, s string, c *tls.Config) (net.Listener, error) {
	return newStreamListener(x, NameTCP, s, tcpConnector{tls: c, Dialer: TCP.(*tcpConnector).Dialer})
}
