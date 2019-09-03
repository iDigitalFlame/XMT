package c2

import (
	"time"
)

// BasicProfile is an empty profile for use with
// testing or filling without having to define all the
// profile interface functions.  Implementing BasicProfile
// in a struct will automatically make it eligible to be a profile.
type BasicProfile struct{}

// CustomProfile is a profile struct that allows useage of a
// profile without creating any functions. Using the C_ values
// to modify the outputs of the profile functions.
type CustomProfile struct {
	CSize      int
	CSleep     time.Duration
	CJitter    int8
	CWrapper   Wrapper
	CTransform Transform
}

// Size returns the default size of the client and server
// buffers.  BasicProfile returns -1 and goes with the default value.
func (b *BasicProfile) Size() int {
	return -1
}

// Size returns the default size of the client and server
// buffers.
func (c *CustomProfile) Size() int {
	if c.CSize <= 0 {
		return -1
	}
	return c.CSize
}

// Jitter returns the default jitter percentage of the client connection
// BasicProfile returns -1 and goes with the default value.
func (b *BasicProfile) Jitter() int8 {
	return -1
}

// Jitter returns the default jitter percentage of the client connection
func (c *CustomProfile) Jitter() int8 {
	return c.CJitter
}

// Wrapper returns the default Wrapper for client and server data
// streams.  BasicProfile returns nil and goes with the default value.
func (b *BasicProfile) Wrapper() Wrapper {
	return nil
}

// Wrapper returns the default Wrapper for client and server data
// streams.
func (c *CustomProfile) Wrapper() Wrapper {
	return c.CWrapper
}

// Sleep returns the default sleep time between\ client connections
// BasicProfile returns -1 and goes with the default value.
func (b *BasicProfile) Sleep() time.Duration {
	return -1
}

// Transform returns the default Transform for client and server data
// streams.  BasicProfile returns nil and goes with the default value.
func (b *BasicProfile) Transform() Transform {
	return nil
}

// Sleep returns the default sleep time between\ client connections
func (c *CustomProfile) Sleep() time.Duration {
	if c.CSleep <= 0 {
		return -1
	}
	return c.CSleep
}

// Transform returns the default Transform for client and server data
// streams.
func (c *CustomProfile) Transform() Transform {
	return c.CTransform
}
