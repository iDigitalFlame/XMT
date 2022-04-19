// Package wrapper is a simple container package for c2 Wrapper types.
package wrapper

import (
	"encoding/base64"
	"encoding/hex"
	"io"

	"github.com/iDigitalFlame/xmt/data"
)

const (
	// Hex is the Hex encoding Wrapper. This wraps the binary data as hex values.
	Hex = simple(0x1)
	// Base64 is the Base64 Wrapper. This wraps the binary data as a Base64 byte string. This may be
	// combined with the Base64 transfrom.
	Base64 = simple(0x2)
)

type simple uint8

func (s simple) Unwrap(r io.Reader) (io.Reader, error) {
	switch s {
	case Hex:
		return hex.NewDecoder(r), nil
	case Base64:
		return base64.NewDecoder(base64.StdEncoding, r), nil
	}
	return r, nil
}
func (s simple) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	switch s {
	case Hex:
		return data.WriteCloser(hex.NewEncoder(w)), nil
	case Base64:
		return base64.NewEncoder(base64.StdEncoding, w), nil
	}
	return w, nil
}
