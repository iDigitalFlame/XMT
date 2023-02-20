// Copyright (C) 2020 - 2023 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package com

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

// ErrInvalidTLSConfig is returned when attempting to use the default TLS Connector
// as a listener. This error is also returned when attempting to use a TLS
// configuration that does not have a valid server certificates.
var ErrInvalidTLSConfig = xerr.Sub("invalid or missing TLS certificates", 0x2D)

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
	var c tls.Config
	switch {
	case ver > 0 && ver < 0xFF:
		c.MinVersion = ver + tls.VersionTLS10
	case ver > tls.VersionTLS10:
		c.MinVersion = ver
	default:
		c.MinVersion = tls.VersionTLS12
	}
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
	return &c, nil
}
