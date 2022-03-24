//go:build crypt

package crypt

import (
	"encoding/base64"
)

var (
	key     string
	values  [256]string
	payload string
)

// Named Network Constants
var (
	IP   = Get(2)   // ip
	TCP  = Get(3)   // tcp
	UDP  = Get(4)   // udp
	Unix = Get(5)   // unix
	Pipe = Get(6)   // pipe
	HTTP = Get(116) // http
)

func init() {
	if len(payload) == 0 || len(key) == 0 {
		return
	}
	var (
		b      = make([]byte, base64.URLEncoding.DecodedLen(len(payload)))
		v, err = base64.URLEncoding.Decode(b, []byte(payload))
	)
	if err != nil || len(b) == 0 || v == 0 {
		return
	}
	var (
		k = make([]byte, base64.URLEncoding.DecodedLen(len(key)))
		c int
	)
	if c, err = base64.URLEncoding.Decode(k, []byte(key)); err != nil || len(k) == 0 || c == 0 {
		return
	}
	for i := 0; i < v; i++ {
		b[i] = b[i] ^ k[i%c]
	}
	for s, e, n := 0, 0, 0; e < v && n < 256; e++ {
		if b[e] != 0 {
			continue
		}
		if e-s > 0 {
			values[n] = string(b[s:e])
		}
		s, n = e+1, n+1
	}
	key, payload = "", ""
}

// Get returns the crypt value at the provided string index.
func Get(i uint8) string {
	return values[i]
}
