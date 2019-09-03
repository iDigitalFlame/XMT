package tcp

import (
	"crypto/tls"
	"net"
	"time"

	"golang.org/x/xerrors"
)

var (
	// Raw is the TCP Raw connector.  This connector uses raw TCP
	// connections without any encoding or Transforms.
	Raw = &streamConnector{
		dial: &net.Dialer{
			Timeout: RawDefaultTimeout,
		},
		network: "tcp",
	}

	// TLS is the TCP over TLS connector profile. This connector uses TCP
	// wrapped in TLS encryption using certificates.  This default
	// connector is only valid for clients that connect to servers with
	// properly signed and trusted root certificates. Otherwise connections
	// will fail.
	TLS = &tlsStreamConnector{
		dial: &net.Dialer{
			Timeout: RawDefaultTimeout,
		},
		config:  &tls.Config{},
		network: "tcp",
	}

	// RawDefaultTimeout is the default timeout used for the Raw TCP connector.
	// The default is 5 seconds.
	RawDefaultTimeout = time.Duration(5) * time.Second

	// ErrInvalidTLSConfig is returned when attempting to use the default TLS
	// Connector. This error is also returned when attemtping to use a TLS configuration
	// that does not have an valid server certificates.
	ErrInvalidTLSConfig = xerrors.New("tls configuration is missing certificates")
)
