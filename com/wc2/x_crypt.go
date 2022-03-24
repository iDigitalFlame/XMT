//go:build crypt

package wc2

import (
	"net/http"

	"github.com/iDigitalFlame/xmt/util/crypt"
)

var userAgent = crypt.Get(44) // User-Agent

func (addr) Network() string {
	return crypt.Get(45) // wc2
}
func (complete) Error() string {
	return crypt.Get(43) // deadline exceeded
}
func modHeaders(h http.Header) {
	h.Set(crypt.Get(46), crypt.Get(47)) // Upgrade, websocket
	h.Set(crypt.Get(48), crypt.Get(46)) // Connection, Upgrade
}
