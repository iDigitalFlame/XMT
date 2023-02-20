//go:build !go1.13
// +build !go1.13

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

package wc2

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/iDigitalFlame/xmt/device"
)

type maskedConn struct {
	net.Conn
}

func (maskedConn) Close() error {
	return nil
}
func maskConn(c net.Conn) net.Conn {
	// NOTE(dij): Until go1.12, the HTTP library doesn't understand websockets
	//            so it tries to close ALL net.Conn's after each request completes.
	//            Let's make sure that doesn't happen so we can use it.
	return maskedConn{c}
}
func (t *transport) hook(x *http.Transport) {
	x.Dial = t.dial
	x.DialTLS = t.dialTLS
	x.DialContext = t.dialContext
}
func newRequest(x context.Context) *http.Request {
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	return r.WithContext(x) // added in go1.7
}
func newTransport(d time.Duration) *http.Transport {
	return &http.Transport{
		Proxy:                 device.Proxy,
		MaxIdleConns:          0,
		IdleConnTimeout:       0,
		DisableKeepAlives:     true,
		MaxIdleConnsPerHost:   0,
		TLSHandshakeTimeout:   d,
		ExpectContinueTimeout: d,
		ResponseHeaderTimeout: d,
	}
}
func baseContext(_ *listener, _ func(net.Listener) context.Context) {}
