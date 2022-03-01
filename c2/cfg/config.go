// Package cfg is used to generate Binary versions of C2 Profiles and can be
// used to create automatic Profile 'Groups' with multiple communication and
// encoding types to be used by a Single session.
package cfg

import (
	"io"
	"os"

	"github.com/iDigitalFlame/xmt/c2"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

var (
	// ErrInvalidSetting is an error returned by the 'Profile' function if any
	// of the specified Settings are invalid or do contain valid information.
	//
	// The error returned will be a wrapped version of this error.
	ErrInvalidSetting = xerr.Sub("setting is invalid", 0xD)
	// ErrMultipleTransforms is an error returned by the 'Profile' function if
	// more that one Transform Setting is attempted to be applied in the Config
	// Group.
	//
	// Unlike Wrappers, Transforms cannot be stacked.
	ErrMultipleTransforms = xerr.Sub("cannot add multiple transforms", 0x17)
	// ErrMultipleConnections is an error returned by the 'Profile' function if more
	// that one Connection Hint Setting is attempted to be applied in the Config
	// Group.
	ErrMultipleConnections = xerr.Sub("cannot add multiple connections", 0x17)
)

// Config is a raw binary representation of settings for a C2 Profile. This can
// be used to save/load Profiles from a file or network location.
type Config []byte

// Len returns the length of this Config instance. This is the same as 'len(c)'.
func (c Config) Len() int {
	return len(c)
}

// Groups returns the number of Groups included in this Config. This determines
// how many Profiles are contained in this Config and will be generated when
// built.
//
// Returns zero on an empty Config.
func (c Config) Groups() int {
	if len(c) == 0 {
		return 0
	}
	var n int
	for i := 0; i >= 0 && i < len(c); i = c.next(i) {
		if cBit(c[i]) == Seperator && i > 0 {
			n++
		}
	}
	return n + 1
}

// Bytes returns the byte version of this Config. This is the same as casting
// the Config instance as '[]byte(c)'.
func (c Config) Bytes() []byte {
	return []byte(c)
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

// Add will append the raw data of the supplied Settings to this Config instance.
func (c *Config) Add(s ...Setting) {
	if len(s) == 0 {
		return
	}
	*c = append(*c, Bytes(s...)...)
}

// Group will attempt to extract the Config Group out of this Config based on
// it's position. Attempts to modify this Config slice will NOT modify the
// resulting parent Config. Modifying the parent Config after extracting a Group
// may invalidate this Group.
//
// This can be used in combination with 'Groups' to iterate over the Groups
// in this Config.
//
// If supplied '-1', this Config returns itself.
func (c Config) Group(p int) Config {
	if len(c) == 0 {
		return nil
	}
	if p == -1 {
		return c
	}
	var l, s int
	for e := 0; e >= 0 && e < len(c); e = c.next(e) {
		if x := cBit(c[e]); x == Seperator {
			if e == 0 {
				continue
			}
			if p <= 0 && l == 0 {
				return c[0:e]
			}
			if p == l {
				return c[s:e]
			}
			s, l = e+1, l+1
		}
	}
	if l > 0 && s > 0 {
		return c[s:]
	}
	if p <= 0 && l == 0 {
		return c
	}
	return nil
}

// Raw will parse the raw bytes and return a compiled Profile interface.
//
// Validation or setting errors will be returned if they occur.
func Raw(b []byte) (c2.Profile, error) {
	return Config(b).Build()
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

// AddGroup will append the supplied Settings to this Config. This will append
// the raw data Setting data to this Config with a seperator, indicating a new
// Profile.
func (c *Config) AddGroup(s ...Setting) {
	if len(s) == 0 {
		return
	}
	if len(*c) > 0 {
		*c = append(*c, byte(Seperator))
	}
	*c = append(*c, Bytes(s...)...)
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

// Write will combine the supplied settings into a byte slice that will be
// written to the supplied writer. Any errors during writing will be returned.
func Write(w io.Writer, s ...Setting) error {
	return Pack(s...).Write(w)
}

// Reader will attempt to read the reader data, parse the raw data and return a
// compiled Profile interface.
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
