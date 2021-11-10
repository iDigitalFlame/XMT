package cfg

import (
	"time"
)

const (
	invalid = cBit(0)

	valHost   = cBit(0xA0)
	valSleep  = cBit(0xA1)
	valJitter = cBit(0xA2)
)

type cBit byte
type cBytes []byte

// Setting is an interface represents a C2 Profile setting in binary form. This can be used inside to generate
// a C2 Profile from binary data or write a Profile to a binary stream or from a JSON payload.
type Setting interface {
	id() cBit
	args() []byte
}

func (c cBit) id() cBit {
	return c
}
func (cBit) args() []byte {
	return nil
}
func (c cBytes) id() cBit {
	if len(c) == 0 {
		return invalid
	}
	return cBit(c[0])
}

// Jitter returns a Setting that will specify the Jitter setting of the generated Profile. Only Jitter values from
// zero to one-hundred [0-100] are valid.
//
// Other values are ignored and replaced with the default.
func Jitter(n uint) Setting {
	return cBytes{byte(valJitter), byte(n)}
}

// Host will return a Setting that will specify a host 'hint' to the profile, which can be used if the connecting
// address is empty.
//
// If empty, this value is ignored.
func Host(s string) Setting {
	n := len(s)
	if n > 0xFFFF {
		n = 0xFFFF
	}
	return append(cBytes{byte(valHost), byte(n >> 8), byte(n)}, s[:n]...)
}
func (c cBytes) args() []byte {
	return c
}

// Sleep returns a Setting that will specify the Sleep timeout setting of the generated Profile.
// Values of zero and below are ignored.
func Sleep(t time.Duration) Setting {
	return cBytes{
		byte(valSleep), byte(t >> 56), byte(t >> 48), byte(t >> 40), byte(t >> 32),
		byte(t >> 24), byte(t >> 16), byte(t >> 8), byte(t),
	}
}
