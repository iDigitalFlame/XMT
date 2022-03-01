//go:build !crypt
// +build !crypt

package wc2

import "net/http"

const userAgent = "User-Agent"

func (addr) Network() string {
	return "wc2"
}
func (complete) Error() string {
	return "deadline exceeded"
}
func modHeaders(h http.Header) {
	h.Set("Upgrade", "websocket")
	h.Set("Connection", "Upgrade")
}
