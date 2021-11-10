package cfg

import (
	"io"
	"time"

	"github.com/iDigitalFlame/xmt/c2"
	"github.com/iDigitalFlame/xmt/c2/transform"
	"github.com/iDigitalFlame/xmt/c2/wrapper"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/com/pipe"
	"github.com/iDigitalFlame/xmt/com/wc2"
	"github.com/iDigitalFlame/xmt/data/crypto"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/text"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Len returns the length of this Config instance. This is the same as
// 'len(c)'.
func (c Config) Len() int {
	return len(c)
}

// Bytes returns the byte version of this Config. This is the same as casting
// the Config instance as '[]byte(c)'.
func (c Config) Bytes() []byte {
	return []byte(c)
}

// String returns a string representation of the data included in this
// Config instance. Each separate setting will be seperated by commas.
func (c Config) String() string {
	if len(c) == 0 || c[0] == 0 {
		return ""
	}
	var b util.Builder
	for i := 0; i >= 0 && i < len(c); i = c.next(i) {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(cBit(c[i]).String())
	}
	return b.Output()
}
func (c Config) next(i int) int {
	if i > len(c) {
		return -1
	}
	switch cBit(c[i]) {
	case WrapHex, WrapZlib, WrapGzip, WrapBase64:
		fallthrough
	case ConnectTCP, ConnectTLS, ConnectUDP, ConnectICMP, ConnectPipe, ConnectTLSNoVerify, TransformB64:
		return i + 1
	case valIP, valB64Shift, valJitter, valTLSx:
		return i + 2
	case valCBK:
		return i + 6
	case valSleep:
		return i + 9
	case valWC2:
		_ = c[8]
		n := i + 8 + (int(c[i+2]) | int(c[i+1])<<8) + (int(c[i+4]) | int(c[i+3])<<8) + (int(c[i+6]) | int(c[i+5])<<8)
		if _ = c[n]; c[i+7] == 0 {
			return n
		}
		for x := int(c[i+7]); x > 0; x-- {
			n += int(c[n]) + int(c[n+1]) + 2
		}
		return n
	case valXOR, valHost:
		_ = c[3]
		return i + 3 + int(c[i+2]) | int(c[i+1])<<8
	case valAES:
		_ = c[3]
		return i + 3 + int(c[i+1]) + int(c[i+2])
	case valMuTLS:
		_ = c[8]
		return i + 8 + (int(c[i+3]) | int(c[i+2])<<8) + (int(c[i+5]) | int(c[i+4])<<8) + (int(c[i+7]) | int(c[i+6])<<8)
	case valTLSxCA:
		_ = c[4]
		return i + 4 + int(c[i+3]) | int(c[i+2])<<8
	case valTLSCert:
		_ = c[6]
		return i + 6 + (int(c[i+3]) | int(c[i+2])<<8) + (int(c[i+5]) | int(c[i+4])<<8)
	}
	return -1
}

// Write will attempt to write the contents of this Config instance to the
// specified Writer.
//
// This function will return any errors that occurred during the write.
// This is a NOP if this Config is empty.
func (c Config) Write(w io.Writer) error {
	if len(c) == 0 {
		return nil
	}
	n, err := w.Write(c)
	if err == nil && n != len(c) {
		return io.ErrShortWrite
	}
	return err
}

// Build will attempt to generate a 'c2.Profile' interface from this Config
// instance.
//
// This function will return an 'ErrInvalidSetting' if any value in this Config
// instance is invalid or 'ErrMultipleHints' if more than one connection hint
// is contained in this Config.
//
// The similar error 'ErrMultipleTransforms' is similar to 'ErrMultipleHints'
// but applies to Transforms.
//
// Multiple 'cw.Wrapper' instances will be combined into a 'c2.MultiWrapper' in
// the order they are found.
//
// Other functions that may return errors on creation, like encryption wrappers
// for example, will stop the build process and will return that error.
func (c Config) Build() (c2.Profile, error) {
	var (
		p profile
		w []c2.Wrapper
	)
	for i, n := 0, 0; n >= 0 && n < len(c); i = n {
		if n = c.next(i); n == i || n > len(c) || n == -1 {
			n = len(c)
		}
		switch _ = c[n-1]; cBit(c[i]) {
		case invalid:
			return nil, ErrInvalidSetting
		case valHost:
			p.host = string(c[i+3 : (int(c[i+2])|int(c[i+1])<<8)+i+3])
		case valSleep:
			p.sleep = time.Duration(
				uint64(c[i+8]) | uint64(c[i+7])<<8 | uint64(c[i+6])<<16 | uint64(c[i+5])<<24 |
					uint64(c[i+4])<<32 | uint64(c[i+3])<<40 | uint64(c[i+2])<<48 | uint64(c[i+1])<<56,
			)
			if p.sleep <= 0 {
				p.sleep = c2.DefaultSleep
			}
		case valJitter:
			if p.jitter = c[i+1]; p.jitter > 100 {
				p.jitter = 100
			}
		case ConnectTCP:
			if p.c != nil {
				return nil, xerr.Wrap("tcp", ErrMultipleHints)
			}
			p.c = com.TCP
		case ConnectTLS:
			if p.c != nil {
				return nil, xerr.Wrap("tls", ErrMultipleHints)
			}
			p.c = com.TLS
		case ConnectUDP:
			if p.c != nil {
				return nil, xerr.Wrap("udp", ErrMultipleHints)
			}
			p.c = com.UDP
		case ConnectICMP:
			if p.c != nil {
				return nil, xerr.Wrap("icmp", ErrMultipleHints)
			}
			p.c = com.ICMP
		case ConnectPipe:
			if p.c != nil {
				return nil, xerr.Wrap("pipe", ErrMultipleHints)
			}
			p.c = pipe.Pipe
		case ConnectTLSNoVerify:
			if p.c != nil {
				return nil, xerr.Wrap("tls-insecure", ErrMultipleHints)
			}
			p.c = com.TLSInsecure
		case valIP:
			if p.c != nil {
				return nil, xerr.Wrap("ip", ErrMultipleHints)
			}
			if c[i+1] == 0 {
				return nil, xerr.Wrap("ip", ErrInvalidSetting)
			}
			p.c = com.NewIP(com.DefaultTimeout, c[i+1])
		case valWC2:
			if p.c != nil {
				return nil, xerr.Wrap("wc2", ErrMultipleHints)
			}
			var (
				v, z = (int(c[i+2]) | int(c[i+1])<<8) + i + 8, i + 8
				t    wc2.Target
			)
			if v > z {
				t.URL = text.Matcher(c[z:v])
			}
			if v, z = (int(c[i+4])|int(c[i+3])<<8)+v, v; v > z {
				t.Host = text.Matcher(c[z:v])
			}
			if v, z = (int(c[i+6])|int(c[i+5])<<8)+v, v; v > z {
				t.Agent = text.Matcher(c[z:v])
			}
			if c[i+7] == 0 {
				continue
			}
			t.Headers = make(map[string]wc2.Stringer, c[i+7])
			for j := 0; v < n && z < n && j < n; {
				z, j = v+2, int(c[v])+v+2
				if v = int(c[v+1]) + j; z == j {
					return nil, xerr.Wrap("wc2", ErrInvalidSetting)
				}
				t.Header(string(c[z:j]), text.Matcher(c[j:v]))
			}
			p.c = wc2.NewClient(com.DefaultTimeout, &t)
		case valTLSx:
			if p.c != nil {
				return nil, xerr.Wrap("tls-ex", ErrInvalidSetting)
			}
			t, err := com.NewTLSConfig(false, uint16(c[i+1]), nil, nil, nil)
			if err != nil {
				return nil, xerr.Wrap("tls-ex", err)
			}
			p.c = com.NewTLS(com.DefaultTimeout, t)
		case valMuTLS:
			if p.c != nil {
				return nil, xerr.Wrap("mtls", ErrInvalidSetting)
			}
			var (
				a      = (int(c[i+3]) | int(c[i+2])<<8) + i + 8
				b      = (int(c[i+5]) | int(c[i+4])<<8) + a
				k      = (int(c[i+7]) | int(c[i+6])<<8) + b
				t, err = com.NewTLSConfig(true, uint16(c[i+1]), c[i+8:a], c[a:b], c[b:k])
			)
			if err != nil {
				return nil, xerr.Wrap("mtls", err)
			}
			p.c = com.NewTLS(com.DefaultTimeout, t)
		case valTLSxCA:
			if p.c != nil {
				return nil, xerr.Wrap("tls-ca", ErrInvalidSetting)
			}
			var (
				a      = (int(c[i+3]) | int(c[i+2])<<8) + i + 4
				t, err = com.NewTLSConfig(false, uint16(c[i+1]), c[i+4:a], nil, nil)
			)
			if err != nil {
				return nil, xerr.Wrap("tls-ca", err)
			}
			p.c = com.NewTLS(com.DefaultTimeout, t)
		case valTLSCert:
			if p.c != nil {
				return nil, xerr.Wrap("tls-cert", ErrInvalidSetting)
			}
			var (
				b      = (int(c[i+3]) | int(c[i+2])<<8) + i + 6
				k      = (int(c[i+5]) | int(c[i+4])<<8) + b
				t, err = com.NewTLSConfig(true, uint16(c[i+1]), nil, c[i+6:b], c[b:k])
			)
			if err != nil {
				return nil, xerr.Wrap("tls-cert", err)
			}
			p.c = com.NewTLS(com.DefaultTimeout, t)
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
				return nil, xerr.Wrap("aes", err)
			}
			u, err := wrapper.Block(b, c[v:z])
			if err != nil {
				return nil, xerr.Wrap("aes", err)
			}
			w = append(w, u)
		case TransformB64:
			if p.t != nil {
				return nil, xerr.Wrap("base64T", ErrMultipleTransforms)
			}
			p.t = transform.Base64
		case valDNS:
			if p.t != nil {
				return nil, xerr.Wrap("dns", ErrMultipleTransforms)
			}
		case valB64Shift:
			if p.t != nil {
				return nil, xerr.Wrap("b64S", ErrMultipleTransforms)
			}
			p.t = transform.B64(c[i+1])
		}
	}
	if len(w) > 1 {
		p.w = c2.MultiWrapper(w)
	} else if len(w) == 1 {
		p.w = w[0]
	}
	return &p, nil
}
