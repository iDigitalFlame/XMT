//go:build crypt

package crypt

import "encoding/base64"

const cryptMax = 300

var (
	key     string
	values  [cryptMax]string
	payload string
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
	for s, e, n := 0, 0, 0; e < v && n < cryptMax; e++ {
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
func Get(i uint16) string {
	return values[i]
}
