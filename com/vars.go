package com

import (
	"crypto/tls"
	"errors"
	"net"
	"time"
)

const (
	// DefaultTimeout is the default timeout used for the Raw connectors.
	// The default is 5 seconds.
	DefaultTimeout = time.Duration(5) * time.Second

	netIP   = "ip:"
	netTCP  = "tcp"
	netUDP  = "udp"
	netUNIX = "unix"
)

var (
	// TCP is the TCP Raw connector. This connector uses raw TCP connections for communication.
	TCP = &TCPConnector{dialer: &net.Dialer{Timeout: DefaultTimeout, KeepAlive: DefaultTimeout, DualStack: true}}

	// UDP is the UDP Raw connector. This connector uses raw UDP connections for communication.
	UDP = NewUDP(DefaultTimeout)

	// TLS is the TCP over TLS connector client. This client uses TCP wrapped in TLS encryption
	// using certificates. This client is only valid for clients that connect to servers with properly
	// signed and trusted certificates.
	TLS = &tcpClient{c: TCPConnector{tls: &tls.Config{}, dialer: TCP.dialer}}

	// TLSNoCheck is the TCP over TLS connector profile. This client uses TCP wrapped in TLS encryption
	// using certificates. This instance DOES NOT check the server certificate for validity.
	TLSNoCheck = &tcpClient{c: TCPConnector{tls: &tls.Config{InsecureSkipVerify: true}, dialer: TCP.dialer}}
)

var (
	// ErrInvalidTimeout is an error returned on client or connector creation when a negative timeout
	// is supplied.
	ErrInvalidTimeout = errors.New("invalid timeout value")
	// ErrInvalidNetwork is an error returned from the New* functions when an improper network is used
	// that is not compatible with the New function return type.
	ErrInvalidNetwork = errors.New("invalid network type")
	// ErrInvalidTLSConfig is returned when attempting to use the default TLS Connector as a listener.
	// This error is also returned when attemtping to use a TLS configuration that does not have a valid
	// server certificates.
	ErrInvalidTLSConfig = errors.New("tls configuration is missing certificates")
)
