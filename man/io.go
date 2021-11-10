package man

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"

	"github.com/iDigitalFlame/xmt/data/crypto"
	"github.com/iDigitalFlame/xmt/device/devtools"
	"github.com/iDigitalFlame/xmt/util"
)

var (
	// NOTE(dij): This is fucking annoying.. why? The error is ALWAYS nil!
	jar, _ = cookiejar.New(nil)

	client = &http.Client{
		Jar:     jar,
		Timeout: timeoutWeb,
		Transport: &http.Transport{
			Proxy:                 devtools.Proxy,
			DialContext:           (&net.Dialer{Timeout: timeoutWeb, KeepAlive: timeoutWeb, DualStack: true}).DialContext,
			MaxIdleConns:          64,
			IdleConnTimeout:       timeoutWeb * 4,
			ForceAttemptHTTP2:     true,
			TLSHandshakeTimeout:   timeoutWeb,
			ExpectContinueTimeout: timeoutWeb,
			ResponseHeaderTimeout: timeoutWeb,
		},
	}
)

// File will attempt to Marshal the Sentinel struct into the supplied file path.
// Any errors that occur during reading will be returned.
//
// This function will truncate and overrite any file that exists at 'p'.
func (s *Sentinel) File(p string) error {
	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	err = s.Write(f)
	f.Close()
	return err
}

// Write will Marshal and write the Sentinel data to the supplied Writer.
// Any errors that occur during writing will be returned.
func (s *Sentinel) Write(w io.Writer) error {
	var (
		v      = s.text()
		_, err = w.Write([]byte{byte(len(v) >> 8), byte(len(v))})
	)
	if err != nil {
		return err
	}
	for i := range v {
		if _, err := w.Write(v[i]); err != nil {
			return err
		}
		if _, err := w.Write([]byte{0x1E}); err != nil {
			return err
		}
	}
	return nil
}

// Reader will attempt to Marshal the Sentinel struct from the 'io.Reader'
// This function will also attempt to fill in the Filter and Linker parameters.
//
// Any errors that occur during reading will be returned.
func Reader(r io.Reader) (*Sentinel, error) {
	var (
		b      [64]byte
		n, err = r.Read(b[0:2])
	)
	if n = int(uint16(b[1]) | uint16(b[0])<<8); err != nil || n <= 0 {
		return nil, err
	}
	var (
		c int
		x util.Builder
		d = make([]string, n)
	)
	for x.Grow(len(b)); ; {
		if n, err = r.Read(b[:]); n == 0 && err != nil {
			break
		}
		for i := 0; i < n; i++ {
			if b[i] == 0x1E {
				d[c] = x.String()
				x.Reset()
				c++
				continue
			}
			x.WriteByte(b[i])
		}
	}
	switch {
	case err != io.EOF:
		return nil, err
	case len(d) == 0:
		return new(Sentinel), nil
	}
	var s Sentinel
	if len(d[0]) > 1 && d[0][0] == '{' && d[0][len(d[0])-1] == '}' {
		if err := json.Unmarshal([]byte(d[0]), &s.Filter); err != nil {
			return nil, err
		}
		d = d[1:]
	}
	if len(d) > 0 && len(d[0]) > 1 && d[0][0] == '*' {
		s.Linker = d[0][1:]
		d = d[1:]
	}
	s.Paths = d
	return &s, nil
}

// Crypt will attempt to Marshal the Sentinel struct from the supplied file path
// and 'cipher.Block'. If the Block is nil, this function will behave similar
// to 'File(p)'. This function will also attempt to fill in the Filter and
// Linker parameters.
//
// Any errors that occur during reading will be returned.
func Crypt(c cipher.Block, p string) (*Sentinel, error) {
	f, err := os.OpenFile(p, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	s, err := CryptReader(c, f)
	if f.Close(); err != nil {
		return nil, err
	}
	return s, nil
}

// Crypt will Marshal and write the Sentinel data to the supplied Writer and
// will encrypt it using the supplied 'cipher.Block' with a randomized IV value.
// If the Block is nil, this function is the same as 's.Write(w)'.
//
// Any errors that occur during writing will be returned.
func (s *Sentinel) Crypt(c cipher.Block, w io.Writer) error {
	if c == nil {
		return s.Write(w)
	}
	//  NOTE(dij): Replaced the pool object with an dynamically
	//   allocated slice. This /seems/ to be more efficient. (and readable!)
	var (
		b      = make([]byte, c.BlockSize())
		_, err = rand.Read(b)
	)
	if err != nil {
		return err
	}
	n, err := w.Write(b)
	if err != nil {
		return err
	}
	if n != c.BlockSize() {
		return io.ErrShortWrite
	}
	o, err := crypto.Encrypter(c, b, w)
	if err != nil {
		return err
	}
	err = s.Write(o)
	o.Close()
	return err
}

// CryptFile will attempt to Marshal the Sentinel struct into the supplied file
// path and 'cipher.Block'. If the Block is nil, this function will behave
// similar to 's.File(p)'. This function will also attempt to fill in the Filter
// and Linker parameters.
//
// Any errors that occur during reading will be returned.
// This function will truncate and overrite any file that exists at 'p'.
func (s *Sentinel) CryptFile(c cipher.Block, p string) error {
	if c == nil {
		return s.File(p)
	}
	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	err = s.Crypt(c, f)
	f.Close()
	return err
}

// CryptReader will attempt to Marshal the Sentinel struct from the 'io.Reader'
// and the supplied 'cipher.Block'. If the Block is nil, this function will behave
// similar to 'Reader(p)'. This function will also attempt to fill in the Filter
// and Linker parameters.
//
// Any errors that occur during reading will be returned.
func CryptReader(c cipher.Block, r io.Reader) (*Sentinel, error) {
	if c == nil {
		return Reader(r)
	}
	//  NOTE(dij): Replaced the pool object with an dynamically
	//   allocated slice. This /seems/ to be more efficient. (and readable!)
	var (
		b      = make([]byte, c.BlockSize())
		_, err = r.Read(b)
	)
	if err != nil {
		return nil, err
	}
	i, err := crypto.Decrypter(c, b, r)
	if err != nil {
		return nil, err
	}
	s, err := Reader(i)
	i.Close()
	return s, err
}
