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

type tcpClient struct {
	c tcpConnector
}
type tcpListener struct {
	_ [0]func()
	net.Listener
	timeout time.Duration
}
type tcpConnector struct {
	tls *tls.Config
	net.Dialer
}

// NewTCP creates a new simple TCP based connector with the supplied timeout.
func NewTCP(t time.Duration) Connector {
	if t < 0 {
		t = DefaultTimeout
	}
	return &tcpConnector{Dialer: net.Dialer{Timeout: t, KeepAlive: t}}
}
func (t tcpListener) Accept() (net.Conn, error) {
	if d, ok := t.Listener.(deadliner); ok {
		d.SetDeadline(time.Now().Add(t.timeout))
	}
	c, err := t.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// NewTLS creates a new simple TLS wrapped TCP based connector with the supplied
// timeout.
func NewTLS(t time.Duration, c *tls.Config) Connector {
	if t < 0 {
		t = DefaultTimeout
	}
	return &tcpConnector{tls: c, Dialer: net.Dialer{Timeout: t, KeepAlive: t}}
}
func (t tcpClient) Connect(x context.Context, s string) (net.Conn, error) {
	return t.c.Connect(x, s)
}
func (t tcpConnector) Connect(x context.Context, s string) (net.Conn, error) {
	c, err := newStreamConn(x, NameTCP, s, t)
	if err != nil {
		return nil, err
	}
	return c, nil
}
func (t tcpConnector) Listen(x context.Context, s string) (net.Listener, error) {
	c, err := newStreamListener(x, NameTCP, s, t)
	if err != nil {
		return nil, err
	}
	return &tcpListener{timeout: t.Timeout, Listener: c}, nil
}
func newStreamConn(x context.Context, n, s string, t tcpConnector) (net.Conn, error) {
	if t.tls != nil {
		return (&tls.Dialer{Config: t.tls, NetDialer: &t.Dialer}).DialContext(x, n, s)
	}
	return t.DialContext(x, n, s)
}
func newStreamListener(x context.Context, n, s string, t tcpConnector) (net.Listener, error) {
	if t.tls != nil && (len(t.tls.Certificates) == 0 || t.tls.GetCertificate == nil) {
		return nil, ErrInvalidTLSConfig
	}
	l, err := ListenConfig.Listen(x, n, s)
	if err != nil {
		return nil, err
	}
	if t.tls == nil {
		return l, nil
	}
	return tls.NewListener(l, t.tls), nil
}
