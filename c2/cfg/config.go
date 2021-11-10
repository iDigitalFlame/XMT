package cfg

import (
	"io"
	"os"
	"time"

	"github.com/iDigitalFlame/xmt/c2"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

var (
	// ErrMultipleHints is an error returned by the 'Profile' function if more that one Connection Hint Setting is
	// attempted to be applied by the Config.
	ErrMultipleHints = xerr.New("attempted to add multiple hints")
	// ErrInvalidSetting is an error returned by the 'Profile' function if any of the specified Settings are invalid
	// or do contain valid information. The error returned will be a wrapped version of this error.
	ErrInvalidSetting = xerr.New("setting is invalid")
	// ErrMultipleTransforms is an error returned by the 'Profile' function if more that one Transform Setting is
	// attempted to be applied by the Config. Unlink Wrappers, Transforms cannot be stacked.
	ErrMultipleTransforms = xerr.New("attempted to add multiple transforms")
)

// Config is a raw binary representation of settings for a C2 Profile.
// This can be used to save/load Profiles from a file or network location.
type Config []byte
type profile struct {
	w      c2.Wrapper
	t      c2.Transform
	c      interface{}
	host   string
	sleep  time.Duration
	jitter uint8
}

// Pack will combine the supplied settings into a Config instance.
func Pack(s ...Setting) Config {
	return Bytes(s...)
}

// Bytes will combine the supplied settings into a byte slice that can be used
// as a Config or written to disk.
func Bytes(s ...Setting) []byte {
	if len(s) == 0 {
		return nil
	}
	var c []byte
	for i := range s {
		if s[i] == nil {
			continue
		}
		if a := s[i].args(); len(a) > 0 {
			c = append(c, a...)
			continue
		}
		c = append(c, byte(s[i].id()))
	}
	return c
}
func (p *profile) Host() string {
	return p.host
}
func (p *profile) Jitter() uint8 {
	return p.jitter
}

// Raw will parse the raw bytes and return a compiled Profile interface.
//
// Validation or setting errors will be returned if they occur.
func Raw(b []byte) (c2.Profile, error) {
	return Config(b).Build()
}
func (p *profile) Wrapper() c2.Wrapper {
	return p.w
}
func (p *profile) Sleep() time.Duration {
	return p.sleep
}

// File will attempt to read the file contents, parse the contents and return
// a compiled Profile interface.
//
// Validation or setting errors will be returned if they occur or if any
// file I/O errors occur.
func File(s string) (c2.Profile, error) {
	f, err := os.Open(s)
	if err != nil {
		return nil, err
	}
	p, err := Reader(f)
	f.Close()
	return p, err
}
func (p *profile) Listener() c2.Accepter {
	if c, ok := p.c.(c2.Accepter); ok {
		return c
	}
	return nil
}
func (p *profile) Connector() c2.Connector {
	if c, ok := p.c.(c2.Connector); ok {
		return c
	}
	return nil
}
func (p *profile) Transform() c2.Transform {
	return p.t
}

// Write will combine the supplied settings into a byte slice that will be
// written to the supplied writer. Any errors during writing will be returned.
func Write(w io.Writer, s ...Setting) error {
	return Pack(s...).Write(w)
}

// Reader will attempt to read the reader data, parse the raw data and return
// a compiled Profile interface.
//
// Validation or setting errors will be returned if they occur or if any
// I/O errors occur.
func Reader(r io.Reader) (c2.Profile, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return Config(b).Build()
}

// Build will combine the supplied settings and return a compiled Profile
// interface.
//
// Validation or setting errors will be returned if they occur.
func Build(s ...Setting) (c2.Profile, error) {
	return Pack(s...).Build()
}
