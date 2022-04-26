package com

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

// ErrInvalidTLSConfig is returned when attempting to use the default TLS Connector
// as a listener. This error is also returned when attemtping to use a TLS
// configuration that does not have a valid server certificates.
var ErrInvalidTLSConfig = xerr.Sub("missing TLS certificates", 0x3C)

// NewTLSConfig generates a new 'tls.Config' struct from the provided TLS details.
// This can be used to generate mTLS or just simple CA-based TLS server/clients
// Connectors.
//
// The provided ca bytes (in PEM format) can be used to validate client certificates
// while the pem and key bytes (in PEM format) are used for the listening socket.
//
// The 'ver' integer represents the TLS-min version requirement. Setting it to zero
// will default to TLSv1. SSLv3 is NOT SUPPORTED!
//
// This function returns an error if the ca, pem and/or key are empty.
// The 'mu' bool will determine if mTLS should be enforced.
//
// mTLS insights sourced from: https://kofo.dev/how-to-mtls-in-golang
func NewTLSConfig(mu bool, ver uint16, ca, pem, key []byte) (*tls.Config, error) {
	c := &tls.Config{}
	if ver > 0 && ver < 0xFF {
		ver = tls.VersionTLS10 + ver
	}
	if ver > tls.VersionTLS10 {
		c.MinVersion = ver
	}
	// NOTE(dij): I kinda want to add a setting here if TLS min ver is unset
	//            that we set it to /at least/ TLSv1.2, but that might break
	//            stuff.
	if len(pem) > 0 && len(key) > 0 {
		x, err := tls.X509KeyPair(pem, key)
		if err != nil {
			return nil, err
		}
		c.Certificates = []tls.Certificate{x}
	}
	if len(ca) > 0 {
		if c.RootCAs = x509.NewCertPool(); !c.RootCAs.AppendCertsFromPEM(ca) {
			return nil, ErrInvalidTLSConfig
		}
		if mu {
			c.ClientCAs = c.RootCAs
			c.ClientAuth = tls.RequireAndVerifyClientCert
		}
	}
	return c, nil
}
