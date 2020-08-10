package com

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

// DefaultTimeout is the default timeout used for the default connectors. The default is 60 seconds (one minute).
const DefaultTimeout = time.Duration(60) * time.Second

const (
	netTCP  = "tcp"
	netUDP  = "udp"
	netUNIX = "unix"
)

// ListenConfig is the default listener config that is used to generate the Listeners. This can be used to specify the
// listen 'KeepALive' timeout.
var ListenConfig = net.ListenConfig{KeepAlive: DefaultTimeout}

var (
	// TCP is the TCP Raw connector. This connector uses raw TCP connections for communication.
	TCP = &TCPConnector{dialer: &net.Dialer{Timeout: DefaultTimeout, KeepAlive: DefaultTimeout, DualStack: true}}

	// UDP is the UDP Raw connector. This connector uses raw UDP connections for communication.
	UDP = NewUDP(DefaultTimeout)

	// ICMP is the ICMP Raw connector. This connector uses raw ICMP connections for communication.
	ICMP = NewIP(1, DefaultTimeout)

	// TLS is the TCP over TLS connector client. This client uses TCP wrapped in TLS encryption
	// using certificates. This client is only valid for clients that connect to servers with properly
	// signed and trusted certificates.
	TLS = &tcpClient{c: TCPConnector{tls: new(tls.Config), dialer: TCP.dialer}}
	// TLSNoCheck is the TCP over TLS connector profile. This client uses TCP wrapped in TLS encryption
	// using certificates. This instance DOES NOT check the server certificate for validity.
	TLSNoCheck = &tcpClient{c: TCPConnector{tls: &tls.Config{InsecureSkipVerify: true}, dialer: TCP.dialer}}
)

// ErrInvalidTLSConfig is returned when attempting to use the default TLS Connector as a listener. This error
// is also returned when attemtping to use a TLS configuration that does not have a valid server certificates.
var ErrInvalidTLSConfig = xerr.New("TLS configuration is missing certificates")

type deadline interface {
	SetDeadline(time.Time) error
}
