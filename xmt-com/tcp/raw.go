package tcp

import (
	"crypto/tls"
	"errors"
	"net"
	"time"
)

const (
	network = "tcp"
)

var (
	// Raw is the TCP Raw provider. This provider uses raw TCP
	// connections without any encoding or Transforms.
	Raw = &provider{
		dial: &net.Dialer{
			Timeout:   RawDefaultTimeout,
			KeepAlive: RawDefaultTimeout,
			DualStack: true,
		},
		network: network,
	}

	// TLS is the TCP over TLS connector profile. This connector uses TCP
	// wrapped in TLS encryption using certificates.  This default
	// connector is only valid for clients that connect to servers with
	// properly signed and trusted root certificates.
	TLS = &client{
		p: &provider{
			tls:     &tls.Config{},
			dial:    Raw.dial,
			network: network,
		},
	}

	// TLSNoCertCheck is the TCP over TLS connector profile. This connector uses TCP
	// wrapped in TLS encryption using certificates.  This default
	// connector is only valid for clients that connect to servers with
	// properly signed and trusted root certificates.
	// This instance DOES NOT check the server certificate.
	TLSNoCertCheck = &client{
		p: &provider{
			tls: &tls.Config{
				InsecureSkipVerify: true,
			},
			dial:    Raw.dial,
			network: network,
		},
	}

	// RawDefaultTimeout is the default timeout used for the Raw TCP connector.
	// The default is 5 seconds.
	RawDefaultTimeout = time.Duration(5) * time.Second

	// ErrInvalidTLSConfig is returned when attempting to use the default TLS
	// Connector. This error is also returned when attemtping to use a TLS configuration
	// that does not have an valid server certificates.
	ErrInvalidTLSConfig = errors.New("tls configuration is missing certificates")
)
