//go:build crypt
// +build crypt

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
	"net/http"

	"github.com/iDigitalFlame/xmt/util/crypt"
)

var userAgent = crypt.Get(27) // User-Agent

func (addr) Network() string {
	return crypt.Get(28) // wc2
}
func modHeaders(h http.Header) {
	h.Set(
		crypt.Get(29), // Upgrade
		crypt.Get(30), // websocket
	)
	h.Set(
		crypt.Get(31), // Connection
		crypt.Get(29), // Upgrade
	)
}
