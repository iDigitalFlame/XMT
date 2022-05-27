//go:build crypt

package wc2

import (
	"net/http"

	"github.com/iDigitalFlame/xmt/util/crypt"
)

var userAgent = crypt.Get(4) // User-Agent

func (addr) Network() string {
	return crypt.Get(39) // wc2
}
func modHeaders(h http.Header) {
	h.Set(
		crypt.Get(40), // Upgrade
		crypt.Get(41), // websocket
	)
	h.Set(
		crypt.Get(42), // Connection
		crypt.Get(40), // Upgrade
	)
}
