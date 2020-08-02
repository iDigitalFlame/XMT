package wrapper

import (
	"encoding/base64"
	"encoding/hex"
	"io"
	"io/ioutil"
)

const (
	// Hex is the Hex encoding Wrapper. This wraps the binary data as hex values.
	Hex = Simple(0x1)

	// Base64 is the Base64 Wrapper. This wraps the binary data as a Base64 byte string. This may be
	// combined with the Base64 transfrom.
	Base64 = Simple(0x2)
)

// Simple is an alias that allows for wrapping multiple types of simple mathematic-based Wrappers. This alias
// implements the 'c2.Wrapper' interface.
type Simple uint8
type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error {
	return nil
}

// Wrap satisfies the Wrapper interface.
func (s Simple) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	switch s {
	case Hex:
		return &nopCloser{hex.NewEncoder(w)}, nil
	case Base64:
		return base64.NewEncoder(base64.StdEncoding, w), nil
	}
	return nil, nil
}

// Unwrap satisfies the Wrapper interface.
func (s Simple) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	switch s {
	case Hex:
		return ioutil.NopCloser(hex.NewDecoder(r)), nil
	case Base64:
		return ioutil.NopCloser(base64.NewDecoder(base64.StdEncoding, r)), nil
	}
	return nil, nil
}
