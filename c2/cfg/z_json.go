//go:build !nojson

package cfg

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/PurpleSec/escape"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

type config struct {
	Type string
	Args json.RawMessage
}
type mapper map[string]json.RawMessage

func (c cBit) String() string {
	switch c {
	case Seperator:
		return "|"
	case valHost:
		return "host"
	case valSleep:
		return "sleep"
	case valJitter:
		return "jitter"
	case valWeight:
		return "weight"
	case SelectorLastValid:
		return "select-last"
	case SelectorRoundRobin:
		return "select-round-robin"
	case SelectorRandom:
		return "select-random"
	case SelectorSemiRoundRobin:
		return "select-semi-round-robin"
	case SelectorSemiRandom:
		return "select-semi-random"
	case ConnectTCP:
		return "tcp"
	case ConnectTLS:
		return "tls"
	case ConnectUDP:
		return "udp"
	case ConnectICMP:
		return "icmp"
	case ConnectPipe:
		return "pipe"
	case ConnectTLSNoVerify:
		return "tls-insecure"
	case valIP:
		return "ip"
	case valWC2:
		return "wc2"
	case valTLSx:
		return "tls-ex"
	case valMuTLS:
		return "mtls"
	case valTLSxCA:
		return "tls-ca"
	case valTLSCert:
		return "tls-cert"
	case WrapHex:
		return "hex"
	case WrapZlib:
		return "zlib"
	case WrapGzip:
		return "gzip"
	case WrapBase64:
		return "base64"
	case valXOR:
		return "xor"
	case valCBK:
		return "cbk"
	case valAES:
		return "aes"
	case TransformB64:
		return "b64t"
	case valDNS:
		return "dns"
	case valB64Shift:
		return "b64s"
	}
	return "<invalid>"
}
func bitFromName(s string) cBit {
	switch strings.ToLower(s) {
	case "host":
		return valHost
	case "sleep":
		return valSleep
	case "jitter":
		return valJitter
	case "weight":
		return valWeight
	case "select-last":
		return SelectorLastValid
	case "select-round-robin":
		return SelectorRoundRobin
	case "select-random":
		return SelectorRandom
	case "select-semi-round-robin":
		return SelectorSemiRoundRobin
	case "select-semi-random":
		return SelectorSemiRandom
	case "tcp":
		return ConnectTCP
	case "tls":
		return ConnectTLS
	case "udp":
		return ConnectUDP
	case "icmp":
		return ConnectICMP
	case "pipe":
		return ConnectPipe
	case "tls-insecure":
		return ConnectTLSNoVerify
	case "ip":
		return valIP
	case "wc2":
		return valWC2
	case "tls-ex":
		return valTLSx
	case "mtls":
		return valMuTLS
	case "tls-ca":
		return valTLSxCA
	case "tls-cert":
		return valTLSCert
	case "hex":
		return WrapHex
	case "zlib":
		return WrapZlib
	case "gzip":
		return WrapGzip
	case "base64":
		return WrapBase64
	case "xor":
		return valXOR
	case "cbk":
		return valCBK
	case "aes":
		return valAES
	case "b64t":
		return TransformB64
	case "dns":
		return valDNS
	case "b64s":
		return valB64Shift
	}
	return invalid
}

// String returns a string representation of the data included in this Config
// instance. Each separate setting will be seperated by commas.
func (c Config) String() string {
	if len(c) == 0 || c[0] == 0 {
		return ""
	}
	var b util.Builder
	for i, x := 0, cBit(c[0]); i >= 0 && i < len(c); {
		b.WriteString(x.String())
		if i = c.next(i); i < 0 || i >= len(c) {
			break
		}
		if x == Seperator {
			x = cBit(c[i])
			continue
		}
		if x = cBit(c[i]); i >= 0 && i < len(c) && x != Seperator {
			b.WriteByte(',')
		}
	}
	return b.Output()
}

// JSON will combine the supplied settings into a JSON payload and returned in
// a byte slice. This will return any validation errors during conversion.
//
// Not valid when the 'nojson' tag is specified.
func JSON(s ...Setting) ([]byte, error) {
	return json.Marshal(Pack(s...))
}

// MarshalJSON will attempt to convert the raw binary data in this Config
// instance into a JSON formart.
//
// The only error that may occur is 'ErrInvalidSetting' if an invalid
// setting or data value is encountered during conversion.
func (c Config) MarshalJSON() ([]byte, error) {
	var (
		b util.Builder
		x cBit
	)
	b.WriteByte('[')
	for i, n, z := 0, 0, false; n >= 0 && n < len(c); i = n {
		if i == 0 || x == Seperator {
			b.WriteByte('[')
		}
		if x = cBit(c[i]); x == invalid {
			return nil, ErrInvalidSetting
		}
		if n = c.next(i); n == i || n > len(c) || n == -1 {
			n = len(c)
		}
		if x == Seperator {
			if n == len(c) {
				break
			}
			b.WriteString("],")
			z = true
			continue
		}
		if i > 0 && !z {
			b.WriteByte(',')
		}
		b.WriteString(`{"type":"` + x.String() + `"`)
		switch z = false; x {
		case WrapHex, WrapZlib, WrapGzip, WrapBase64:
			fallthrough
		case SelectorLastValid, SelectorRoundRobin, SelectorRandom, SelectorSemiRandom, SelectorSemiRoundRobin:
			fallthrough
		case ConnectTCP, ConnectTLS, ConnectUDP, ConnectICMP, ConnectPipe, ConnectTLSNoVerify:
			fallthrough
		case TransformB64:
			goto end
		}
		b.WriteString(`,"args":`)
		switch _ = c[n-1]; x {
		case valHost:
			b.WriteString(escape.JSON(string(c[i+3 : (int(c[i+2])|int(c[i+1])<<8)+i+3])))
		case valSleep:
			b.WriteString(escape.JSON(time.Duration(
				uint64(c[i+8]) | uint64(c[i+7])<<8 | uint64(c[i+6])<<16 | uint64(c[i+5])<<24 |
					uint64(c[i+4])<<32 | uint64(c[i+3])<<40 | uint64(c[i+2])<<48 | uint64(c[i+1])<<56,
			).String()))
		case valJitter, valWeight, valIP, valTLSx, valB64Shift:
			b.WriteString(strconv.FormatUint(uint64(c[i+1]), 10))
		case valWC2:
			var (
				v, z = (int(c[i+2]) | int(c[i+1])<<8) + i + 8, i + 8
				w    = v > 0
			)
			if v > z {
				b.WriteString(`{"url":` + escape.JSON(string(c[z:v])))
			}
			if v, z = (int(c[i+4])|int(c[i+3])<<8)+v, v; v > z {
				if !w {
					w = true
					b.WriteString(`{"host":`)
				} else {
					b.WriteString(`,"host":`)
				}
				b.WriteString(escape.JSON(string(c[z:v])))
			}
			if v, z = (int(c[i+6])|int(c[i+5])<<8)+v, v; v > z {
				if !w {
					w = true
					b.WriteString(`{"agent":`)
				} else {
					b.WriteString(`,"agent":`)
				}
				b.WriteString(escape.JSON(string(c[z:v])))
			}
			if c[i+7] == 0 {
				b.WriteByte('}')
				goto end
			}
			if !w {
				b.WriteString(`{"headers":{`)
			} else {
				b.WriteString(`,"headers":{`)
			}
			for j := 0; v < n && z < n && j < n; {
				if j > 0 {
					b.WriteByte(',')
				}
				z, j = v+2, int(c[v])+v+2
				if v = int(c[v+1]) + j; z == j {
					return nil, xerr.Wrap("wc2", ErrInvalidSetting)
				}
				b.WriteString(escape.JSON(string(c[z:j])))
				b.WriteByte(':')
				b.WriteString(escape.JSON(string(c[j:v])))
			}
			b.WriteString("}}")
		case valMuTLS:
			var (
				a = (int(c[i+3]) | int(c[i+2])<<8) + i + 8
				p = (int(c[i+5]) | int(c[i+4])<<8) + a
				k = (int(c[i+7]) | int(c[i+6])<<8) + p
			)
			b.WriteString(`{"version":` + strconv.FormatUint(uint64(c[i+1]), 10))
			b.WriteString(`,"ca":"`)
			e := base64.NewEncoder(base64.StdEncoding, &b)
			e.Write(c[i+8 : a])
			e.Close()
			b.WriteString(`","pem":"`)
			e = base64.NewEncoder(base64.StdEncoding, &b)
			e.Write(c[a:p])
			e.Close()
			b.WriteString(`","key":"`)
			e = base64.NewEncoder(base64.StdEncoding, &b)
			e.Write(c[p:k])
			e.Close()
			b.WriteString(`"}`)
		case valTLSxCA:
			a := (int(c[i+3]) | int(c[i+2])<<8) + i + 4
			b.WriteString(`{"version":` + strconv.FormatUint(uint64(c[i+1]), 10))
			b.WriteString(`,"ca":"`)
			e := base64.NewEncoder(base64.StdEncoding, &b)
			e.Write(c[i+4 : a])
			e.Close()
			b.WriteString(`"}`)
		case valTLSCert:
			var (
				p = (int(c[i+3]) | int(c[i+2])<<8) + i + 6
				k = (int(c[i+5]) | int(c[i+4])<<8) + p
			)
			b.WriteString(`{"version":` + strconv.FormatUint(uint64(c[i+1]), 10))
			b.WriteString(`,"pem":"`)
			e := base64.NewEncoder(base64.StdEncoding, &b)
			e.Write(c[i+6 : p])
			e.Close()
			b.WriteString(`","key":"`)
			e = base64.NewEncoder(base64.StdEncoding, &b)
			e.Write(c[p:k])
			e.Close()
			b.WriteString(`"}`)
		case valXOR:
			b.WriteByte('"')
			e := base64.NewEncoder(base64.StdEncoding, &b)
			e.Write(c[i+3 : (int(c[i+2])|int(c[i+1])<<8)+i+3])
			e.Close()
			b.WriteByte('"')
		case valCBK:
			b.WriteString(`{"size":`)
			b.WriteString(strconv.FormatUint(uint64(c[i+1]), 10))
			b.WriteString(`,"A":`)
			b.WriteString(strconv.FormatUint(uint64(c[i+2]), 10))
			b.WriteString(`,"B":`)
			b.WriteString(strconv.FormatUint(uint64(c[i+3]), 10))
			b.WriteString(`,"C":`)
			b.WriteString(strconv.FormatUint(uint64(c[i+4]), 10))
			b.WriteString(`,"D":`)
			b.WriteString(strconv.FormatUint(uint64(c[i+5]), 10))
			b.WriteByte('}')
		case valAES:
			var (
				v = int(c[i+1]) + i + 3
				z = int(c[i+2]) + v
			)
			if v == z || i+3 == v {
				return nil, xerr.Wrap("aes", ErrInvalidSetting)
			}
			b.WriteString(`{"key":"`)
			e := base64.NewEncoder(base64.StdEncoding, &b)
			e.Write(c[i+3 : v])
			e.Close()
			b.WriteString(`","iv":"`)
			e = base64.NewEncoder(base64.StdEncoding, &b)
			e.Write(c[v:z])
			e.Close()
			b.WriteString(`"}`)
		case valDNS:
			_ = c[i+1]
			b.WriteByte('[')
			for x, v, e := int(c[i+1]), i+2, i+2; x > 0 && v < n; x-- {
				if v += int(c[v]) + 1; e+1 > v || e+1 == v {
					return nil, xerr.Wrap("dns", ErrInvalidSetting)
				}
				if x != int(c[i+1]) {
					b.WriteByte(',')
				}
				b.WriteString(escape.JSON(string(c[e+1 : v])))
				e = v
			}
			b.WriteByte(']')
		}
	end:
		b.WriteByte('}')
	}
	b.WriteString("]]")
	return []byte(b.Output()), nil
}

// UnmarshalJSON will attempt to convert the JSON data provided into this Config
// instance.
//
// Errors during parsing or formatting will be returned along with the
// 'ErrInvalidSetting' error if parsed data contains invalid values.
func (c *Config) UnmarshalJSON(b []byte) error {
	var h []json.RawMessage
	if err := json.Unmarshal(b, &h); err != nil {
		return err
	}
	if len(h) == 0 {
		return nil
	}
	r := make([]Setting, 0, len(h)*4)
	for k := range h {
		var m []config
		if err := json.Unmarshal(h[k], &m); err != nil {
			return err
		}
		if len(m) == 0 {
			continue
		}
		for i := range m {
			switch x := bitFromName(m[i].Type); x {
			case invalid:
				return ErrInvalidSetting
			case WrapHex, WrapZlib, WrapGzip, WrapBase64:
				fallthrough
			case SelectorLastValid, SelectorRoundRobin, SelectorRandom, SelectorSemiRandom, SelectorSemiRoundRobin:
				fallthrough
			case ConnectTCP, ConnectTLS, ConnectUDP, ConnectICMP, ConnectPipe, ConnectTLSNoVerify:
				fallthrough
			case TransformB64:
				r = append(r, x)
			case valHost:
				var s string
				if err := json.Unmarshal(m[i].Args, &s); err != nil {
					return xerr.Wrap("host", err)
				}
				r = append(r, Host(s))
			case valSleep:
				var s string
				if err := json.Unmarshal(m[i].Args, &s); err != nil {
					return xerr.Wrap("sleep", err)
				}
				d, err := time.ParseDuration(s)
				if err != nil {
					return xerr.Wrap("sleep", err)
				}
				r = append(r, Sleep(d))
			case valJitter:
				var v uint8
				if err := json.Unmarshal(m[i].Args, &v); err != nil {
					return xerr.Wrap("jitter", err)
				}
				r = append(r, cBytes{byte(valJitter), v})
			case valWeight:
				var v uint8
				if err := json.Unmarshal(m[i].Args, &v); err != nil {
					return xerr.Wrap("weight", err)
				}
				r = append(r, cBytes{byte(valWeight), v})
			case valIP:
				var v uint8
				if err := json.Unmarshal(m[i].Args, &v); err != nil {
					return xerr.Wrap("ip", err)
				}
				r = append(r, cBytes{byte(valIP), v})
			case valWC2:
				var z mapper
				if err := json.Unmarshal(m[i].Args, &z); err != nil {
					return xerr.Wrap("wc2", err)
				}
				var (
					u, h, a string
					j       map[string]string
				)
				if err := z.Unmarshal("url", false, &u); err != nil {
					return err
				}
				if err := z.Unmarshal("host", false, &h); err != nil {
					return err
				}
				if err := z.Unmarshal("agent", false, &a); err != nil {
					return err
				}
				if d, ok := z["headers"]; ok {
					if err := json.Unmarshal(d, &j); err != nil {
						return xerr.Wrap("wc2", err)
					}
					for v := range j {
						if len(v) == 0 {
							return xerr.Wrap("wc2", ErrInvalidSetting)
						}
					}
				}
				r = append(r, ConnectWC2(u, h, a, j))
			case valTLSx:
				var v uint8
				if err := json.Unmarshal(m[i].Args, &v); err != nil {
					return xerr.Wrap("tls-ex", err)
				}
				r = append(r, cBytes{byte(valTLSx), v})
			case valMuTLS:
				var z mapper
				if err := json.Unmarshal(m[i].Args, &z); err != nil {
					return xerr.Wrap("mtls", err)
				}
				var (
					a, p, k []byte
					v       uint16
				)
				if err := z.Unmarshal("ca", false, &a); err != nil {
					return xerr.Wrap("mtls", err)
				}
				if err := z.Unmarshal("pem", true, &p); err != nil {
					return xerr.Wrap("mtls", err)
				}
				if err := z.Unmarshal("key", true, &k); err != nil {
					return xerr.Wrap("mtls", err)
				}
				if d, ok := z["version"]; ok {
					if err := json.Unmarshal(d, &v); err != nil {
						return xerr.Wrap("mtls", err)
					}
				}
				r = append(r, ConnectMuTLS(v, a, p, k))
			case valTLSxCA:
				var z mapper
				if err := json.Unmarshal(m[i].Args, &z); err != nil {
					return xerr.Wrap("tls-ca", err)
				}
				var (
					a []byte
					v uint16
				)
				if err := z.Unmarshal("ca", false, &a); err != nil {
					return xerr.Wrap("tls-ca", err)
				}
				if d, ok := z["version"]; ok {
					if err := json.Unmarshal(d, &v); err != nil {
						return xerr.Wrap("tls-ca", err)
					}
				}
				r = append(r, ConnectTLSExCA(v, a))
			case valTLSCert:
				var z mapper
				if err := json.Unmarshal(m[i].Args, &z); err != nil {
					return xerr.Wrap("tls-cert", err)
				}
				var (
					p, k []byte
					v    uint16
				)
				if err := z.Unmarshal("pem", true, &p); err != nil {
					return xerr.Wrap("tls-cert", err)
				}
				if err := z.Unmarshal("key", true, &k); err != nil {
					return xerr.Wrap("tls-cert", err)
				}
				if d, ok := z["version"]; ok {
					if err := json.Unmarshal(d, &v); err != nil {
						return xerr.Wrap("tls-cert", err)
					}
				}
				r = append(r, ConnectTLSCerts(v, p, k))
			case valXOR:
				var (
					v      = make([]byte, base64.StdEncoding.DecodedLen(len(m[i].Args)-2))
					n, err = base64.StdEncoding.Decode(v, m[i].Args[1:len(m[i].Args)-1])
				)
				if err != nil {
					return xerr.Wrap("xor", err)
				}
				r = append(r, WrapXOR(v[:n]))
			case valCBK:
				var z mapper
				if err := json.Unmarshal(m[i].Args, &z); err != nil {
					return xerr.Wrap("cbk", err)
				}
				var e, t, y, u, s uint8 = 0, 0, 0, 0, 128
				if d, ok := z["size"]; ok {
					if err := json.Unmarshal(d, &s); err != nil {
						return xerr.Wrap("cbk", err)
					}
				}
				if err := z.Unmarshal("A", true, &e); err != nil {
					return xerr.Wrap("cbk", err)
				}
				if err := z.Unmarshal("B", true, &t); err != nil {
					return xerr.Wrap("cbk", err)
				}
				if err := z.Unmarshal("C", true, &y); err != nil {
					return xerr.Wrap("cbk", err)
				}
				if err := z.Unmarshal("D", true, &u); err != nil {
					return xerr.Wrap("cbk", err)
				}
				r = append(r, cBytes{byte(valCBK), s, e, t, y, u})
			case valAES:
				var z mapper
				if err := json.Unmarshal(m[i].Args, &z); err != nil {
					return xerr.Wrap("aes", err)
				}
				var k, v []byte
				if err := z.Unmarshal("iv", true, &v); err != nil {
					return xerr.Wrap("aes", err)
				}
				if err := z.Unmarshal("key", true, &k); err != nil {
					return xerr.Wrap("aes", err)
				}
				r = append(r, WrapAES(k, v))
			case valDNS:
				var d []string
				if err := json.Unmarshal(m[i].Args, &d); err != nil {
					return xerr.Wrap("dns", err)
				}
				r = append(r, TransformDNS(d...))
			case valB64Shift:
				var v uint8
				if err := json.Unmarshal(m[i].Args, &v); err != nil {
					return xerr.Wrap("b64S", err)
				}
				r = append(r, cBytes{byte(valB64Shift), v})
			}
		}
		if k+1 < len(h) {
			r = append(r, Seperator)
		}
	}
	*c = Bytes(r...)
	return nil
}
func (c *config) UnmarshalJSON(b []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	v, ok := m["type"]
	if !ok {
		return xerr.Sub(`missing "type" string`, 0xD)
	}
	if err := json.Unmarshal(v, &c.Type); err != nil {
		return err
	}
	if v, ok = m["args"]; ok {
		if err := json.Unmarshal(v, &c.Args); err != nil {
			return err
		}
	}
	return nil
}
func (m mapper) Unmarshal(s string, r bool, v interface{}) error {
	d, ok := m[s]
	if !ok {
		if !r {
			return nil
		}
		if xerr.Concat {
			return xerr.Sub(`"`+s+`" not found`, 0xD)
		}
		return xerr.Sub("key not found", 0xD)
	}
	return json.Unmarshal(d, v)
}
