//go:build go1.13
// +build go1.13

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

func maskConn(c net.Conn) net.Conn {
	return c
}
func (t *transport) hook(x *http.Transport) {
	x.Dial = t.dial
	x.DialTLS = t.dialTLS
	x.DialContext = t.dialContext
	x.DialTLSContext = t.dialTLSContext
}
func newRequest(x context.Context) *http.Request {
	r, _ := http.NewRequestWithContext(x, http.MethodGet, "", nil)
	return r
}
func newTransport(d time.Duration) *http.Transport {
	return &http.Transport{
		Proxy:                 device.Proxy,
		MaxIdleConns:          0,
		ReadBufferSize:        1, // Bugfix
		WriteBufferSize:       1, // Bugfix
		IdleConnTimeout:       0,
		MaxConnsPerHost:       0,
		ForceAttemptHTTP2:     false,
		DisableKeepAlives:     true,
		TLSHandshakeTimeout:   d,
		MaxIdleConnsPerHost:   0,
		ExpectContinueTimeout: d,
		ResponseHeaderTimeout: d,
	}
}
func baseContext(l *listener, f func(net.Listener) context.Context) {
	l.BaseContext = f
}
