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
	if i > len(c) {
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
		if _ = c[n]; c[i+7] == 0 {
			return n
		}
		for x := int(c[i+7]); x > 0; x-- {
			n += int(c[n]) + int(c[n+1]) + 2
		}
		return n
	case valXOR, valHost:
		if i+2 >= len(c) {
			return -1
		}
		_ = c[i+2]
		return i + 3 + int(c[i+2]) | int(c[i+1])<<8
	case valAES:
		if i+2 >= len(c) {
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
		if i+3 >= len(c) {
			return -1
		}
		_ = c[i+3]
		return i + 4 + int(c[i+3]) | int(c[i+2])<<8
	case valTLSCert:
		if i+4 >= len(c) {
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
		if n = c.next(i); n == i || n > len(c) || n == -1 {
			n = len(c)
		}
		switch _ = c[n-1]; cBit(c[i]) {
		case Seperator:
			break loop
		case invalid:
			return -1, ErrInvalidSetting
		case valHost, valSleep, valJitter, valWeight:
			if len(c) < n {
				return -1, ErrInvalidSetting
			}
		case SelectorRoundRobin, SelectorLastValid, SelectorRandom, SelectorSemiRandom, SelectorSemiRoundRobin:
		case ConnectTCP, ConnectTLS, ConnectUDP, ConnectICMP, ConnectPipe, ConnectTLSNoVerify:
			fallthrough
		case valTLSx, valMuTLS, valTLSxCA, valTLSCert:
			if p {
				return -1, ErrMultipleConnections
			}
			p = true
		case valIP:
			if p {
				return -1, ErrMultipleConnections
			}
			if c[i+1] == 0 {
				return -1, ErrInvalidSetting
			}
			p = true
		case valWC2:
			if p {
				return -1, ErrMultipleConnections
			}
			if c[i+7] > 0 {
				var (
					q = (int(c[i+2]) | int(c[i+1])<<8) + i + 8 + (int(c[i+4]) | int(c[i+3])<<8)
					v = q + (int(c[i+6]) | int(c[i+5])<<8)
				)
				for j := 0; v < n && q < n && j < n; {
					q, j = v+2, int(c[v])+v+2
					if v = int(c[v+1]) + j; q == j {
						return -1, ErrInvalidSetting
					}
				}
			}
			p = true
		case WrapHex, WrapZlib, WrapGzip, WrapBase64, valXOR, valCBK:
		case valAES:
			var (
				v = int(c[i+1]) + i + 3
				z = int(c[i+2]) + v
			)
			switch v - (i + 3) {
			case 16, 24, 32:
			default:
				return -1, ErrInvalidSetting
			}
			if z-v != 16 {
				return -1, ErrInvalidSetting
			}
		case TransformB64, valDNS, valB64Shift:
			if t {
				return -1, ErrMultipleTransforms
			}
			t = true
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
		if n = c.next(i); n == i || n > len(c) || n == -1 {
			n = len(c)
		}
		switch _ = c[n-1]; cBit(c[i]) {
		case Seperator:
			break loop
		case invalid:
			return nil, -1, -1, ErrInvalidSetting
		case valHost:
			p.hosts = append(p.hosts, string(c[i+3:(int(c[i+2])|int(c[i+1])<<8)+i+3]))
		case valSleep:
			p.sleep = time.Duration(
				uint64(c[i+8]) | uint64(c[i+7])<<8 | uint64(c[i+6])<<16 | uint64(c[i+5])<<24 |
					uint64(c[i+4])<<32 | uint64(c[i+3])<<40 | uint64(c[i+2])<<48 | uint64(c[i+1])<<56,
			)
			if p.sleep <= 0 {
				p.sleep = c2.DefaultSleep
			}
		case valJitter:
			if p.jitter = int8(c[i+1]); p.jitter > 100 {
				p.jitter = 100
			}
		case valWeight:
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
			if c[i+1] == 0 {
				return nil, -1, -1, xerr.Wrap("ip", ErrInvalidSetting)
			}
			p.conn = com.NewIP(com.DefaultTimeout, c[i+1])
		case valWC2:
			// TODO(dij): Add an entry for a WC2 listener list of proxy
			//            redirects.
			//
			//            Going to handle this in Cirrus as there needs to be
			//            server content and I don't see the use of using a WC2
			//            connector internally as a proxy. *shrug*
			if p.conn != nil {
				return nil, -1, -1, xerr.Wrap("wc2", ErrMultipleConnections)
			}
			var (
				v, q = (int(c[i+2]) | int(c[i+1])<<8) + i + 8, i + 8
				t    wc2.Target
			)
			if v > q {
				t.URL = text.Matcher(c[q:v])
			}
			if v, q = (int(c[i+4])|int(c[i+3])<<8)+v, v; v > q {
				t.Host = text.Matcher(c[q:v])
			}
			if v, q = (int(c[i+6])|int(c[i+5])<<8)+v, v; v > q {
				t.Agent = text.Matcher(c[q:v])
			}
			if c[i+7] > 0 {
				t.Headers = make(map[string]wc2.Stringer, c[i+7])
				for j := 0; v < n && q < n && j < n; {
					q, j = v+2, int(c[v])+v+2
					if v = int(c[v+1]) + j; q == j {
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
			t, err := com.NewTLSConfig(false, uint16(c[i+1]), nil, nil, nil)
			if err != nil {
				return nil, -1, -1, xerr.Wrap("tls-ex", err)
			}
			p.conn = com.NewTLS(com.DefaultTimeout, t)
		case valMuTLS:
			if p.conn != nil {
				return nil, -1, -1, xerr.Wrap("mtls", ErrMultipleConnections)
			}
			var (
				a      = (int(c[i+3]) | int(c[i+2])<<8) + i + 8
				b      = (int(c[i+5]) | int(c[i+4])<<8) + a
				k      = (int(c[i+7]) | int(c[i+6])<<8) + b
				t, err = com.NewTLSConfig(true, uint16(c[i+1]), c[i+8:a], c[a:b], c[b:k])
			)
			if err != nil {
				return nil, -1, -1, xerr.Wrap("mtls", err)
			}
			p.conn = com.NewTLS(com.DefaultTimeout, t)
		case valTLSxCA:
			if p.conn != nil {
				return nil, -1, -1, xerr.Wrap("tls-ca", ErrMultipleConnections)
			}
			var (
				a      = (int(c[i+3]) | int(c[i+2])<<8) + i + 4
				t, err = com.NewTLSConfig(false, uint16(c[i+1]), c[i+4:a], nil, nil)
			)
			if err != nil {
				return nil, -1, -1, xerr.Wrap("tls-ca", err)
			}
			p.conn = com.NewTLS(com.DefaultTimeout, t)
		case valTLSCert:
			if p.conn != nil {
				return nil, -1, -1, xerr.Wrap("tls-cert", ErrInvalidSetting)
			}
			var (
				b      = (int(c[i+3]) | int(c[i+2])<<8) + i + 6
				k      = (int(c[i+5]) | int(c[i+4])<<8) + b
				t, err = com.NewTLSConfig(true, uint16(c[i+1]), nil, c[i+6:b], c[b:k])
			)
			if err != nil {
				return nil, -1, -1, xerr.Wrap("tls-cert", err)
			}
			p.conn = com.NewTLS(com.DefaultTimeout, t)
		case WrapHex:
			w = append(w, wrapper.Hex)
		case WrapZlib:
			w = append(w, wrapper.Hex)
		case WrapGzip:
			w = append(w, wrapper.Gzip)
		case WrapBase64:
			w = append(w, wrapper.Base64)
		case valXOR:
			x := crypto.XOR(c[i+3 : (int(c[i+2])|int(c[i+1])<<8)+i+3])
			w = append(w, wrapper.Stream(x, x))
		case valCBK:
			w = append(w, wrapper.NewCBK(c[i+2], c[i+3], c[i+4], c[i+5], c[i+1]))
		case valAES:
			var (
				v      = int(c[i+1]) + i + 3
				z      = int(c[i+2]) + v
				b, err = crypto.NewAes(c[i+3 : v])
			)
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
			_ = c[i+1]
			d := make(transform.DNSTransform, 0, c[i+1])
			for x, v, e := int(c[i+1]), i+2, i+2; x > 0 && v < n; x-- {
				if v += int(c[v]) + 1; e+1 > v || e+1 == v {
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
