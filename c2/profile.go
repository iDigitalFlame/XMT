package c2

import (
	"time"

	"github.com/iDigitalFlame/xmt/c2/transform"
	"github.com/iDigitalFlame/xmt/c2/wrapper"
	"github.com/iDigitalFlame/xmt/com/limits"
)

// DefaultProfile is an simple profile for use with testing or filling without having to define all the
// profile properties.
var DefaultProfile = &Profile{
	Size:   limits.MediumLimit(),
	Sleep:  DefaultSleep,
	Jitter: DefaultJitter,
}

// Profile is a struct that represents a C2 profile. This is used for defining the specifics that will
// be used to listen by servers and for connections by clients.  Nil or empty values will be replaced with defaults.
type Profile struct {
	Size      int
	Sleep     time.Duration
	Jitter    int8
	Wrapper   wrapper.Wrapper
	Transform transform.Transform
}
