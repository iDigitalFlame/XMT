package wrapper

import (
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/iDigitalFlame/xmt/xmt/com/c2"
)

const (
	// Hex is the Hex encoding Wrapper. This wraps the binary
	// data as hex values.
	Hex = hexWrapper(false)

	// Zlib is the default Zlib Wrapper. This wrapper uses the
	// default compression level. Use the 'NewZlib' function to
	// create a wrapper with a different level.
	Zlib = zlibWrapper(zlib.DefaultCompression)

	// Gzip is the default Gzip Wrapper. This wrapper uses the
	// default compression level. Use the 'NewGzip' function to
	// create a wrapper with a different level.
	Gzip = gzipWrapper(zlib.DefaultCompression)

	// Base64 is the Base64 Wrapper. This wraps the binary
	// data as a Base64 byte array. This may be combined with the Base64
	// transfrom.
	Base64 = base64Wrapper(false)
)

type hexWrapper bool
type zlibWrapper int8
type gzipWrapper int8
type base64Wrapper bool
type nopWriteCloser struct {
	io.Writer
}

func (n *nopWriteCloser) Close() error {
	return nil
}

// NewZlib returns a Zlib compreession wrapper. This function will return and error
// if the commpression level is invalid.
func NewZlib(level int) (c2.Wrapper, error) {
	if level < zlib.HuffmanOnly || level > zlib.BestCompression {
		return nil, fmt.Errorf("zlib: invalid compression level: %d", level)
	}
	return zlibWrapper(level), nil
}

// NewGzip returns a Gzip compreession wrapper. This function will return and error
// if the commpression level is invalid.
func NewGzip(level int) (c2.Wrapper, error) {
	if level < gzip.HuffmanOnly || level > gzip.BestCompression {
		return nil, fmt.Errorf("gzip: invalid compression level: %d", level)
	}
	return gzipWrapper(level), nil
}
func (h hexWrapper) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return &nopWriteCloser{hex.NewEncoder(w)}, nil
}
func (h hexWrapper) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	return ioutil.NopCloser(hex.NewDecoder(r)), nil
}
func (z zlibWrapper) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return zlib.NewWriterLevel(w, int(z))
}
func (z zlibWrapper) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	return zlib.NewReader(r)
}
func (g gzipWrapper) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return gzip.NewWriterLevel(w, int(g))
}
func (g gzipWrapper) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	return gzip.NewReader(r)
}
func (b base64Wrapper) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return base64.NewEncoder(base64.StdEncoding, w), nil
}
func (b base64Wrapper) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	return ioutil.NopCloser(base64.NewDecoder(base64.StdEncoding, r)), nil
}
