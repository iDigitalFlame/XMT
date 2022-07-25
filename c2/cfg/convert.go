// Copyright (C) 2020 - 2022 iDigitalFlame
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

package cfg

import (
	"sort"
	"time"

	"github.com/iDigitalFlame/xmt/c2"
	"github.com/iDigitalFlame/xmt/c2/transform"
	"github.com/iDigitalFlame/xmt/c2/wrapper"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/com/pipe"
	"github.com/iDigitalFlame/xmt/com/wc2"
	"github.com/iDigitalFlame/xmt/data/crypto"
	"github.com/iDigitalFlame/xmt/util/text"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

func init() {
	c2.ProfileParser = Raw
}
func (c Config) next(i int) int {
	if i > len(c) || i < 0 {
		return -1
	}
	switch cBit(c[i]) {
	case WrapHex, WrapZlib, WrapGzip, WrapBase64:
		fallthrough
	case SelectorLastValid, SelectorRoundRobin, SelectorRandom, SelectorSemiRandom, SelectorSemiRoundRobin:
		fallthrough
	case Seperator, ConnectTCP, ConnectTLS, ConnectUDP, ConnectICMP, ConnectPipe, ConnectTLSNoVerify, TransformB64:
		return i + 1
	case valIP, valB64Shift, valJitter, valWeight, valTLSx:
		return i + 2
	case valCBK:
		return i + 6
	case valSleep:
		return i + 9
	case valWC2:
		if i+7 >= len(c) {
			return -1
		}
		_ = c[i+7]
		n := i + 8 + (int(c[i+2]) | int(c[i+1])<<8) + (int(c[i+4]) | int(c[i+3])<<8) + (int(c[i+6]) | int(c[i+5])<<8)
		if n >= len(c) {
			return -1
		}
		if _ = c[n]; c[i+7] == 0 {
			return n
		}
		for x := int(c[i+7]); x > 0 && n < len(c) && n > 0; x-- {
			n += int(c[n]) + int(c[n+1]) + 2
		}
		return n
	case valXOR, valHost:
		if i+3 >= len(c) {
			return -1
		}
		_ = c[i+2]
		return i + 3 + int(c[i+2]) | int(c[i+1])<<8
	case valAES:
		if i+3 >= len(c) {
			return -1
		}
		_ = c[i+2]
		return i + 3 + int(c[i+1]) + int(c[i+2])
	case valMuTLS:
		if i+7 >= len(c) {
			return -1
		}
		_ = c[i+7]
		return i + 8 + (int(c[i+3]) | int(c[i+2])<<8) + (int(c[i+5]) | int(c[i+4])<<8) + (int(c[i+7]) | int(c[i+6])<<8)
	case valTLSxCA:
		if i+4 >= len(c) {
			return -1
		}
		_ = c[i+3]
		return i + 4 + int(c[i+3]) | int(c[i+2])<<8
	case valTLSCert:
		if i+6 >= len(c) {
			return -1
		}
		_ = c[i+4]
		return i + 6 + (int(c[i+3]) | int(c[i+2])<<8) + (int(c[i+5]) | int(c[i+4])<<8)
	case valDNS:
		if i+1 >= len(c) {
			return -1
		}
		_ = c[i+1]
		n := i + 2
		for x := int(c[i+1]); x > 0 && n < len(c); x-- {
			n += int(c[n]) + 1
		}
		return n
	}
	return -1
}

// Validate is similar to the 'Build' function but will instead only validate
// that the supplied Config will build into a Profile without returning an
// error. The error returned (if not nil) will be the same as the error returned
// during a Build call.
//
// This function will return an 'ErrInvalidSetting' if any value in this Config
// instance is invalid or 'ErrMultipleConnections' if more than one connection
// is contained in this Config.
//
// The similar error 'ErrMultipleTransforms' is similar to 'ErrMultipleConnections'
// but applies to Transforms, if more than one Transform is contained.
//
// Multiple 'c2.Wrapper' instances will be combined into a 'c2.MultiWrapper' in
// the order they are found.
//
// Other functions that may return errors on creation, like encryption wrappers
// for example, will stop the build process and will return that wrapped error.
func (c Config) Validate() error {
	if len(c) == 0 {
		return nil
	}
	var (
		n   int
		err error
	)
	for i := 0; i < len(c); i = n {
		if n, err = c.validate(i); err != nil {
			return err
		}
		if n-i == 1 && c[i] == byte(Seperator) {
			continue
		}
	}
	return nil
}

// Build will attempt to generate a 'c2.Profile' interface from this Config
// instance.
//
// This function will return an 'ErrInvalidSetting' if any value in this Config
// instance is invalid or 'ErrMultipleConnections' if more than one connection
// is contained in this Config.
//
// The similar error 'ErrMultipleTransforms' is similar to 'ErrMultipleConnections'
// but applies to Transforms, if more than one Transform is contained.
//
// Multiple 'c2.Wrapper' instances will be combined into a 'c2.MultiWrapper' in
// the order they are found.
//
// Other functions that may return errors on creation, like encryption wrappers
// for example, will stop the build process and will return that wrapped error.
func (c Config) Build() (c2.Profile, error) {
	if len(c) == 0 {
		return nil, nil
	}
	var (
		e   []*profile
		n   int
		v   *profile
		s   int8
		g   uint8
		err error
	)
	for i := 0; i < len(c); i = n {
		if v, n, s, err = c.build(i); err != nil {
			return nil, err
		}
		if v == nil || (n-i == 1 && c[i] == byte(Seperator)) {
			continue
		}
		if e = append(e, v); s >= 0 {
			g = uint8(s)
		}
		v = nil
	}
	if len(e) == 0 {
		return nil, nil
	}
	if len(e) == 1 {
		e[0].src = c
		return e[0], nil
	}
	r := &Group{sel: g, entries: e, src: c}
	sort.Sort(r)
	return r, nil
}
func (c Config) validate(x int) (int, error) {
	var (
		n    int
		p, t bool
	)
loop:
	for i := x; n >= 0 && n < len(c); i = n {
		if n = c.next(i); n == i || n > len(c) || n == -1 || n < i {
			n = len(c)
		}
		switch _ = c[n-1]; cBit(c[i]) {
		case Seperator:
			break loop
		case invalid:
			return -1, ErrInvalidSetting
		case valHost:
			if i+3 >= n {
				return -1, xerr.Wrap("host", ErrInvalidSetting)
			}
			if v := (int(c[i+2]) | int(c[i+1])<<8) + i; v > n || v < i {
				return -1, xerr.Wrap("host", ErrInvalidSetting)
			}
		case valSleep:
			if i+8 >= n {
				return -1, xerr.Wrap("sleep", ErrInvalidSetting)
			}
		case valJitter:
			if i+1 >= n {
				return -1, xerr.Wrap("jitter", ErrInvalidSetting)
			}
		case valWeight:
			if i+1 >= n {
				return -1, xerr.Wrap("weight", ErrInvalidSetting)
			}
		case SelectorRoundRobin, SelectorLastValid, SelectorRandom, SelectorSemiRandom, SelectorSemiRoundRobin:
		case ConnectTCP, ConnectTLS, ConnectUDP, ConnectICMP, ConnectPipe, ConnectTLSNoVerify:
			if p {
				return -1, ErrMultipleConnections
			}
			p = true
		case valIP:
			if p {
				return -1, xerr.Wrap("ip", ErrMultipleConnections)
			}
			if p = true; i+1 >= n {
				return -1, xerr.Wrap("ip", ErrInvalidSetting)
			}
			if c[i+1] == 0 {
				return -1, xerr.Wrap("ip", ErrInvalidSetting)
			}
		case valWC2:
			if p {
				return -1, xerr.Wrap("wc2", ErrMultipleConnections)
			}
			if p = true; i+7 >= n {
				return -1, xerr.Wrap("wc2", ErrInvalidSetting)
			}
			v, q := (int(c[i+2])|int(c[i+1])<<8)+i+8, i+8
			if v > n || q > n || q < i || v < i {
				return -1, xerr.Wrap("wc2", ErrInvalidSetting)
			}
			if v, q = (int(c[i+4])|int(c[i+3])<<8)+v, v; v > n || q > n || v < q || q < i || v < i {
				return -1, xerr.Wrap("wc2", ErrInvalidSetting)
			}
			if v, q = (int(c[i+6])|int(c[i+5])<<8)+v, v; v > n || q > n || v < q || q < i || v < i {
				return -1, xerr.Wrap("wc2", ErrInvalidSetting)
			}
			if c[i+7] > 0 {
				for j := 0; v < n && q < n && j < n; {
					q, j = v+2, int(c[v])+v+2
					if v = int(c[v+1]) + j; q == j || j > n || q > n || v > n || v < j || j < q || q < i || j < i || v < i {
						return -1, xerr.Wrap("wc2", ErrInvalidSetting)
					}
				}
			}
		case valTLSx:
			if p {
				return -1, xerr.Wrap("tls-ex", ErrMultipleConnections)
			}
			if p = true; i+1 >= n {
				return -1, xerr.Wrap("tls-ex", ErrInvalidSetting)
			}
		case valMuTLS:
			if p {
				return -1, xerr.Wrap("mtls", ErrMultipleConnections)
			}
			if p = true; i+7 >= n {
				return -1, xerr.Wrap("mtls", ErrInvalidSetting)
			}
			var (
				a = (int(c[i+3]) | int(c[i+2])<<8) + i + 8
				b = (int(c[i+5]) | int(c[i+4])<<8) + a
				k = (int(c[i+7]) | int(c[i+6])<<8) + b
			)
			if a > n || b > n || k > n || b < a || k < b || a < i || b < i || k < i {
				return -1, xerr.Wrap("mtls", ErrInvalidSetting)
			}
		case valTLSxCA:
			if p {
				return -1, xerr.Wrap("tls-ca", ErrMultipleConnections)
			}
			if p = true; i+4 >= n {
				return -1, xerr.Wrap("tls-ca", ErrInvalidSetting)
			}
			if a := (int(c[i+3]) | int(c[i+2])<<8) + i + 4; a > n || a < i {
				return -1, xerr.Wrap("tls-ca", ErrInvalidSetting)
			}
		case valTLSCert:
			if p {
				return -1, xerr.Wrap("tls-cert", ErrMultipleConnections)
			}
			if p = true; i+6 >= n {
				return -1, xerr.Wrap("tls-cert", ErrInvalidSetting)
			}
			var (
				b = (int(c[i+3]) | int(c[i+2])<<8) + i + 6
				k = (int(c[i+5]) | int(c[i+4])<<8) + b
			)
			if b > n || k > n || b < i || k < i || k < b {
				return -1, xerr.Wrap("tls-cert", ErrInvalidSetting)
			}
		case WrapHex, WrapZlib, WrapGzip, WrapBase64:
		case valXOR:
			if i+3 >= n {
				return -1, xerr.Wrap("xor", ErrInvalidSetting)
			}
			if k := (int(c[i+2]) | int(c[i+1])<<8) + i; k > n || k < i {
				return -1, xerr.Wrap("xor", ErrInvalidSetting)
			}
		case valCBK:
			if i+5 >= n {
				return -1, xerr.Wrap("cbk", ErrInvalidSetting)
			}
		case valAES:
			if i+3 >= n {
				return -1, xerr.Wrap("aes", ErrInvalidSetting)
			}
			var (
				v = int(c[i+1]) + i + 3
				z = int(c[i+2]) + v
			)
			if v == z || i+3 == v || z > n || v > n || z < i || v < i || z < v {
				return -1, xerr.Wrap("aes", ErrInvalidSetting)
			}
			switch v - (i + 3) {
			case 16, 24, 32:
			default:
				return -1, xerr.Wrap("aes", ErrInvalidSetting)
			}
			if z-v != 16 {
				return -1, xerr.Wrap("aes", ErrInvalidSetting)
			}
		case TransformB64:
			if t {
				return -1, xerr.Wrap("base64T", ErrMultipleTransforms)
			}
			t = true
		case valDNS:
			if t {
				return -1, xerr.Wrap("dns", ErrMultipleTransforms)
			}
			if t = true; i+1 >= n {
				return -1, xerr.Wrap("dns", ErrInvalidSetting)
			}
			for x, v, e := int(c[i+1]), i+2, i+2; x > 0 && v < n; x-- {
				if v += int(c[v]) + 1; e+1 > v || e+1 == v || v < e || v > n || e > n || x > n || e < i || v < i || x < i {
					return -1, xerr.Wrap("dns", ErrInvalidSetting)
				}
				e = v
			}
		case valB64Shift:
			if t {
				return -1, xerr.Wrap("b64S", ErrMultipleTransforms)
			}
			if t = true; i+1 >= n {
				return -1, xerr.Wrap("b64S", ErrInvalidSetting)
			}
		default:
			return -1, ErrInvalidSetting
		}
	}
	return n, nil
}
func (c Config) build(x int) (*profile, int, int8, error) {
	var (
		p profile
		w []c2.Wrapper
		n int
		z int8 = -1
	)
loop:
	for i := x; n >= 0 && n < len(c); i = n {
		if n = c.next(i); n == i || n > len(c) || n == -1 || n < i {
			n = len(c)
		}
		switch _ = c[n-1]; cBit(c[i]) {
		case Seperator:
			break loop
		case invalid:
			return nil, -1, -1, ErrInvalidSetting
		case valHost:
			if i+3 >= n {
				return nil, -1, -1, xerr.Wrap("host", ErrInvalidSetting)
			}
			v := (int(c[i+2]) | int(c[i+1])<<8) + i
			if v > n || v < i {
				return nil, -1, -1, xerr.Wrap("host", ErrInvalidSetting)
			}
			p.hosts = append(p.hosts, string(c[i+3:v+3]))
		case valSleep:
			if i+8 >= n {
				return nil, -1, -1, xerr.Wrap("sleep", ErrInvalidSetting)
			}
			p.sleep = time.Duration(
				uint64(c[i+8]) | uint64(c[i+7])<<8 | uint64(c[i+6])<<16 | uint64(c[i+5])<<24 |
					uint64(c[i+4])<<32 | uint64(c[i+3])<<40 | uint64(c[i+2])<<48 | uint64(c[i+1])<<56,
			)
			if p.sleep <= 0 {
				p.sleep = c2.DefaultSleep
			}
		case valJitter:
			if i+1 >= n {
				return nil, -1, -1, xerr.Wrap("jitter", ErrInvalidSetting)
			}
			if p.jitter = int8(c[i+1]); p.jitter > 100 {
				p.jitter = 100
			}
		case valWeight:
			if i+1 >= n {
				return nil, -1, -1, xerr.Wrap("weight", ErrInvalidSetting)
			}
			if p.weight = c[i+1]; p.weight > 100 {
				p.weight = 100
			}
		case SelectorRandom, SelectorRoundRobin, SelectorLastValid, SelectorSemiRandom, SelectorSemiRoundRobin:
			z = int8(c[i] - byte(SelectorLastValid))
		case ConnectTCP:
			if p.conn != nil {
				return nil, -1, -1, xerr.Wrap("tcp", ErrMultipleConnections)
			}
			p.conn = com.TCP
		case ConnectTLS:
			if p.conn != nil {
				return nil, -1, -1, xerr.Wrap("tls", ErrMultipleConnections)
			}
			p.conn = com.TLS
		case ConnectUDP:
			if p.conn != nil {
				return nil, -1, -1, xerr.Wrap("udp", ErrMultipleConnections)
			}
			p.conn = com.UDP
		case ConnectICMP:
			if p.conn != nil {
				return nil, -1, -1, xerr.Wrap("icmp", ErrMultipleConnections)
			}
			p.conn = com.ICMP
		case ConnectPipe:
			if p.conn != nil {
				return nil, -1, -1, xerr.Wrap("pipe", ErrMultipleConnections)
			}
			p.conn = pipe.Pipe
		case ConnectTLSNoVerify:
			if p.conn != nil {
				return nil, -1, -1, xerr.Wrap("tls-insecure", ErrMultipleConnections)
			}
			p.conn = com.TLSInsecure
		case valIP:
			if p.conn != nil {
				return nil, -1, -1, xerr.Wrap("ip", ErrMultipleConnections)
			}
			if i+1 >= n {
				return nil, -1, -1, xerr.Wrap("ip", ErrInvalidSetting)
			}
			if c[i+1] == 0 {
				return nil, -1, -1, xerr.Wrap("ip", ErrInvalidSetting)
			}
			p.conn = com.NewIP(com.DefaultTimeout, c[i+1])
		case valWC2:
			if p.conn != nil {
				return nil, -1, -1, xerr.Wrap("wc2", ErrMultipleConnections)
			}
			if i+7 >= n {
				return nil, -1, -1, xerr.Wrap("wc2", ErrInvalidSetting)
			}
			var (
				v, q = (int(c[i+2]) | int(c[i+1])<<8) + i + 8, i + 8
				t    wc2.Target
			)
			if v > n || q > n || q < i || v < i {
				return nil, -1, -1, xerr.Wrap("wc2", ErrInvalidSetting)
			}
			if v > q {
				t.URL = text.Matcher(c[q:v])
			}
			if v, q = (int(c[i+4])|int(c[i+3])<<8)+v, v; v > q {
				if v > n || q > n || v < q || q < i || v < i {
					return nil, -1, -1, xerr.Wrap("wc2", ErrInvalidSetting)
				}
				t.Host = text.Matcher(c[q:v])
			}
			if v, q = (int(c[i+6])|int(c[i+5])<<8)+v, v; v > q {
				if v > n || q > n || v < q || q < i || v < i {
					return nil, -1, -1, xerr.Wrap("wc2", ErrInvalidSetting)
				}
				t.Agent = text.Matcher(c[q:v])
			}
			if c[i+7] > 0 {
				t.Headers = make(map[string]wc2.Stringer, c[i+7])
				for j := 0; v < n && q < n && j < n; {
					q, j = v+2, int(c[v])+v+2
					if v = int(c[v+1]) + j; q == j || q > n || j > n || v > n || v < j || j < q || q < i || j < i || v < i {
						return nil, -1, -1, xerr.Wrap("wc2", ErrInvalidSetting)
					}
					t.Header(string(c[q:j]), text.Matcher(c[j:v]))
				}
			}
			p.conn = wc2.NewClient(com.DefaultTimeout, &t)
		case valTLSx:
			if p.conn != nil {
				return nil, -1, -1, xerr.Wrap("tls-ex", ErrMultipleConnections)
			}
			if i+1 >= n {
				return nil, -1, -1, xerr.Wrap("tls-ex", ErrInvalidSetting)
			}
			t, err := com.NewTLSConfig(false, uint16(c[i+1]), nil, nil, nil)
			if err != nil {
				return nil, -1, -1, xerr.Wrap("tls-ex", err)
			}
			p.conn = com.NewTLS(com.DefaultTimeout, t)
		case valMuTLS:
			if p.conn != nil {
				return nil, -1, -1, xerr.Wrap("mtls", ErrMultipleConnections)
			}
			if i+7 >= n {
				return nil, -1, -1, xerr.Wrap("mtls", ErrInvalidSetting)
			}
			var (
				a = (int(c[i+3]) | int(c[i+2])<<8) + i + 8
				b = (int(c[i+5]) | int(c[i+4])<<8) + a
				k = (int(c[i+7]) | int(c[i+6])<<8) + b
			)
			if a > n || b > n || k > n || b < a || k < b || a < i || b < i || k < i {
				return nil, -1, -1, xerr.Wrap("mtls", ErrInvalidSetting)
			}
			t, err := com.NewTLSConfig(true, uint16(c[i+1]), c[i+8:a], c[a:b], c[b:k])
			if err != nil {
				return nil, -1, -1, xerr.Wrap("mtls", err)
			}
			p.conn = com.NewTLS(com.DefaultTimeout, t)
		case valTLSxCA:
			if p.conn != nil {
				return nil, -1, -1, xerr.Wrap("tls-ca", ErrMultipleConnections)
			}
			if i+4 >= n {
				return nil, -1, -1, xerr.Wrap("tls-ca", ErrInvalidSetting)
			}
			a := (int(c[i+3]) | int(c[i+2])<<8) + i + 4
			if a > n || a < i {
				return nil, -1, -1, xerr.Wrap("tls-ca", ErrInvalidSetting)
			}
			t, err := com.NewTLSConfig(false, uint16(c[i+1]), c[i+4:a], nil, nil)
			if err != nil {
				return nil, -1, -1, xerr.Wrap("tls-ca", err)
			}
			p.conn = com.NewTLS(com.DefaultTimeout, t)
		case valTLSCert:
			if p.conn != nil {
				return nil, -1, -1, xerr.Wrap("tls-cert", ErrMultipleConnections)
			}
			if i+6 >= n {
				return nil, -1, -1, xerr.Wrap("tls-cert", ErrInvalidSetting)
			}
			var (
				b = (int(c[i+3]) | int(c[i+2])<<8) + i + 6
				k = (int(c[i+5]) | int(c[i+4])<<8) + b
			)
			if b > n || k > n || b < i || k < i || k < b {
				return nil, -1, -1, xerr.Wrap("tls-cert", ErrInvalidSetting)
			}
			t, err := com.NewTLSConfig(true, uint16(c[i+1]), nil, c[i+6:b], c[b:k])
			if err != nil {
				return nil, -1, -1, xerr.Wrap("tls-cert", err)
			}
			p.conn = com.NewTLS(com.DefaultTimeout, t)
		case WrapHex:
			w = append(w, wrapper.Hex)
		case WrapZlib:
			w = append(w, wrapper.Zlib)
		case WrapGzip:
			w = append(w, wrapper.Gzip)
		case WrapBase64:
			w = append(w, wrapper.Base64)
		case valXOR:
			if i+3 >= n {
				return nil, -1, -1, xerr.Wrap("xor", ErrInvalidSetting)
			}
			k := (int(c[i+2]) | int(c[i+1])<<8) + i
			if k > n || k < i {
				return nil, -1, -1, xerr.Wrap("xor", ErrInvalidSetting)
			}
			x := crypto.XOR(c[i+3 : k+3])
			w = append(w, wrapper.Stream(x, x))
		case valCBK:
			if i+5 >= n {
				return nil, -1, -1, xerr.Wrap("cbk", ErrInvalidSetting)
			}
			w = append(w, wrapper.NewCBK(c[i+2], c[i+3], c[i+4], c[i+5], c[i+1]))
		case valAES:
			if i+3 >= n {
				return nil, -1, -1, xerr.Wrap("aes", ErrInvalidSetting)
			}
			var (
				v = int(c[i+1]) + i + 3
				z = int(c[i+2]) + v
			)
			if v == z || i+3 == v || v > n || z > n || z < i || v < i || z < v {
				return nil, -1, -1, xerr.Wrap("aes", ErrInvalidSetting)
			}
			b, err := crypto.NewAes(c[i+3 : v])
			if err != nil {
				return nil, -1, -1, xerr.Wrap("aes", err)
			}
			u, err := wrapper.Block(b, c[v:z])
			if err != nil {
				return nil, -1, -1, xerr.Wrap("aes", err)
			}
			w = append(w, u)
		case TransformB64:
			if p.t != nil {
				return nil, -1, -1, xerr.Wrap("base64T", ErrMultipleTransforms)
			}
			p.t = transform.Base64
		case valDNS:
			if p.t != nil {
				return nil, -1, -1, xerr.Wrap("dns", ErrMultipleTransforms)
			}
			if i+1 >= n {
				return nil, -1, -1, xerr.Wrap("dns", ErrInvalidSetting)
			}
			_ = c[i+1]
			d := make(transform.DNSTransform, 0, c[i+1])
			for x, v, e := int(c[i+1]), i+2, i+2; x > 0 && v < n; x-- {
				if v += int(c[v]) + 1; e+1 > v || e+1 == v || v < e || v > n || e > n || x > n || e < i || v < i || x < i {
					return nil, -1, -1, xerr.Wrap("dns", ErrInvalidSetting)
				}
				d = append(d, string(c[e+1:v]))
				e = v
			}
			p.t = d
		case valB64Shift:
			if p.t != nil {
				return nil, -1, -1, xerr.Wrap("b64S", ErrMultipleTransforms)
			}
			if i+1 >= n {
				return nil, -1, -1, xerr.Wrap("b64S", ErrInvalidSetting)
			}
			p.t = transform.B64(c[i+1])
		default:
			return nil, -1, -1, ErrInvalidSetting
		}
	}
	if len(w) > 1 {
		p.w = c2.MultiWrapper(w)
	} else if len(w) == 1 {
		p.w = w[0]
	}
	return &p, n, z, nil
}
