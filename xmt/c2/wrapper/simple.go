package wrapper

import (
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/iDigitalFlame/xmt/xmt/data"
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

	hexID    uint8 = 0xD1
	zlibID   uint8 = 0xD2
	gzipID   uint8 = 0xD3
	base64ID uint8 = 0xD4
)

type simple bool
type hexWrapper bool
type zlibWrapper int8
type gzipWrapper int8
type base64Wrapper bool
type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error {
	return nil
}

// NewZlib returns a Zlib compreession wrapper. This function will return and error
// if the commpression level is invalid.
func NewZlib(level int) (Wrapper, error) {
	if level < zlib.HuffmanOnly || level > zlib.BestCompression {
		return nil, fmt.Errorf("zlib: invalid compression level: %d", level)
	}
	return zlibWrapper(level), nil
}

// NewGzip returns a Gzip compreession wrapper. This function will return and error
// if the commpression level is invalid.
func NewGzip(level int) (Wrapper, error) {
	if level < gzip.HuffmanOnly || level > gzip.BestCompression {
		return nil, fmt.Errorf("gzip: invalid compression level: %d", level)
	}
	return gzipWrapper(level), nil
}
func (hexWrapper) MarshalStream(w data.Writer) error {
	return w.WriteUint8(hexID)
}
func (base64Wrapper) MarshalStream(w data.Writer) error {
	return w.WriteUint8(base64ID)
}
func (z zlibWrapper) MarshalStream(w data.Writer) error {
	if err := w.WriteUint8(zlibID); err != nil {
		return err
	}
	return w.WriteInt8(int8(z))
}
func (g gzipWrapper) MarshalStream(w data.Writer) error {
	if err := w.WriteUint8(gzipID); err != nil {
		return err
	}
	return w.WriteInt8(int8(g))
}
func (z *zlibWrapper) UnmarshalStream(r data.Reader) error {
	l, err := r.Int8()
	if err != nil {
		return err
	}
	*z = zlibWrapper(l)
	return nil
}
func (g *gzipWrapper) UnmarshalStream(r data.Reader) error {
	l, err := r.Int8()
	if err != nil {
		return err
	}
	*g = gzipWrapper(l)
	return nil
}
func (hexWrapper) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return &nopWriteCloser{hex.NewEncoder(w)}, nil
}
func (hexWrapper) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	return ioutil.NopCloser(hex.NewDecoder(r)), nil
}
func (gzipWrapper) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	return gzip.NewReader(r)
}
func (zlibWrapper) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	return zlib.NewReader(r)
}
func (z zlibWrapper) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return zlib.NewWriterLevel(w, int(z))
}
func (g gzipWrapper) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return gzip.NewWriterLevel(w, int(g))
}
func (base64Wrapper) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return base64.NewEncoder(base64.StdEncoding, w), nil
}
func (base64Wrapper) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	return ioutil.NopCloser(base64.NewDecoder(base64.StdEncoding, r)), nil
}
