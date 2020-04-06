package wrapper

import (
	"encoding/base64"
	"encoding/hex"
	"io"
	"io/ioutil"
)

const (
	// Hex is the Hex encoding Wrapper. This wraps the binary data as hex values.
	Hex = simpleWrapper(0x1)

	// Base64 is the Base64 Wrapper. This wraps the binary data as a Base64 byte string. This may be
	// combined with the Base64 transfrom.
	Base64 = simpleWrapper(0x2)
)

// Value is an interface that wraps the binary streams into separate stream types. This allows for using
// encryption or compression (or both!). This is just a compatibility interface to prevent import dependency cycles.
type Value interface {
	Wrap(io.WriteCloser) (io.WriteCloser, error)
	Unwrap(io.ReadCloser) (io.ReadCloser, error)
}
type nopCloser struct {
	io.Writer
}
type simpleWrapper uint8

func (nopCloser) Close() error {
	return nil
}
func (s simpleWrapper) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	switch s {
	case Hex:
		return &nopCloser{hex.NewEncoder(w)}, nil
	case Base64:
		return base64.NewEncoder(base64.StdEncoding, w), nil
	}
	return nil, nil
}
func (s simpleWrapper) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	switch s {
	case Hex:
		return ioutil.NopCloser(hex.NewDecoder(r)), nil
	case Base64:
		return ioutil.NopCloser(base64.NewDecoder(base64.StdEncoding, r)), nil
	}
	return nil, nil
}
