package c2

import (
	"io"
	"time"
)

// DefaultProfile is an simple profile for use with testing or filling without having to define all the
// profile properties.
var DefaultProfile = &Profile{Sleep: DefaultSleep, Jitter: uint(DefaultJitter)}

// Profile is a struct that represents a C2 profile. This is used for defining the specifics that will
// be used to listen by servers and for connections by clients.  Nil or empty values will be replaced with defaults.
type Profile struct {
	Wrapper   Wrapper
	Transform Transform
	hint      Setting

	Sleep  time.Duration
	Jitter uint
}

// MultiWrapper is an alias for an array of Wrappers. This will preform the wrapper/unwrapping operations in the
// order of the array. This is automatically created by a Config instance when multiple Wrappers are present.
type MultiWrapper []Wrapper

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
