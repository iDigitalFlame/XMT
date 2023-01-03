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
	"context"
	"crypto/tls"
	"net"
	"time"
)

type unixConnector tcpConnector

// NewUNIX creates a new simple UNIX socket based connector with the supplied
// timeout.
func NewUNIX(t time.Duration) Connector {
	return unixConnector(tcpConnector{Dialer: net.Dialer{Timeout: t, KeepAlive: t}})
}

// NewSecureUNIX creates a new simple TLS wrapped UNIX socket based connector
// with the supplied timeout.
func NewSecureUNIX(t time.Duration, c *tls.Config) Connector {
	if t < 0 {
		t = DefaultTimeout
	}
	return unixConnector(tcpConnector{tls: c, Dialer: net.Dialer{Timeout: t, KeepAlive: t}})
}
func (u unixConnector) Connect(x context.Context, s string) (net.Conn, error) {
	return newStreamConn(x, NameUnix, s, tcpConnector(u))
}
func (u unixConnector) Listen(x context.Context, s string) (net.Listener, error) {
	c, err := newStreamListener(x, NameUnix, s, tcpConnector(u))
	if err != nil {
		return nil, err
	}
	return &tcpListener{timeout: u.Timeout, Listener: c}, nil
}
