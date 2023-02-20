//go:build !go1.15
// +build !go1.15

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
)

func newStreamConn(x context.Context, n, s string, t *tcpConnector) (net.Conn, error) {
	if t.tls != nil {
		return tls.DialWithDialer(&t.Dialer, n, s, t.tls)
	}
	return t.DialContext(x, n, s)
}

// ConnectConfig will use the supplied TCP client to connect to the supplied address
// using the TLSConfig supplied.
func (t tcpClient) ConnectConfig(x context.Context, c *tls.Config, s string) (net.Conn, error) {
	return tls.DialWithDialer(&t.c.Dialer, NameTCP, s, c)
}
