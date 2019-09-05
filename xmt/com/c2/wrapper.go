package c2

import "io"

// Wrapper is an interface that allows for wrapping the
// binary streams into separate stream types. This allows for
// using encryption or compression.
type Wrapper interface {
	Wrap(io.WriteCloser) io.WriteCloser
	Unwrap(io.ReadCloser) io.ReadCloser
}
type rawWrapper struct{}

func (r *rawWrapper) Wrap(o io.WriteCloser) io.WriteCloser {
	return o
}
func (r *rawWrapper) Unwrap(i io.ReadCloser) io.ReadCloser {
	return i
}
