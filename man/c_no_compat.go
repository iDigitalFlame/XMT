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

package man

import (
	"context"
	"net"
	"net/http"
	"net/http/cookiejar"

	"github.com/iDigitalFlame/xmt/device"
)

func initDefaultClient() {
	j, _ := cookiejar.New(nil)
	client.v = &http.Client{
		Jar: j,
		Transport: &http.Transport{
			Proxy:                 device.Proxy,
			DialContext:           (&net.Dialer{Timeout: timeoutWeb, KeepAlive: timeoutWeb}).DialContext,
			MaxIdleConns:          64,
			IdleConnTimeout:       timeoutWeb * 2,
			DisableKeepAlives:     true,
			ForceAttemptHTTP2:     false,
			TLSHandshakeTimeout:   timeoutWeb,
			ExpectContinueTimeout: timeoutWeb,
			ResponseHeaderTimeout: timeoutWeb,
		},
	}
}
func newRequest(x context.Context) *http.Request {
	r, _ := http.NewRequestWithContext(x, http.MethodGet, "*", nil)
	return r
}
