package wrapper

import (
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
)

const (
	// Zlib is the default Zlib Wrapper. This wrapper uses the default compression level. Use the 'NewZlib'
	// function to create a wrapper with a different level.
	Zlib = zlibWrap(zlib.DefaultCompression)

	// Gzip is the default Gzip Wrapper. This wrapper uses the default compression level. Use the 'NewGzip'
	// function to create a wrapper with a different level.
	Gzip = gzipWrap(zlib.DefaultCompression)
)

type zlibWrap int8
type gzipWrap int8

// NewZlib returns a Zlib compreession wrapper. This function will return and error if the commpression level
// is invalid.
func NewZlib(level int) (Value, error) {
	if level < zlib.HuffmanOnly || level > zlib.BestCompression {
		return nil, fmt.Errorf("invalid compression level: %d", level)
	}
	return zlibWrap(level), nil
}

// NewGzip returns a Gzip compreession wrapper. This function will return and error if the commpression level
// is invalid.
func NewGzip(level int) (Value, error) {
	if level < gzip.HuffmanOnly || level > gzip.BestCompression {
		return nil, fmt.Errorf("invalid compression level: %d", level)
	}
	return gzipWrap(level), nil
}
func (gzipWrap) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	return gzip.NewReader(r)
}
func (zlibWrap) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	return zlib.NewReader(r)
}
func (z zlibWrap) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return zlib.NewWriterLevel(w, int(z))
}
func (g gzipWrap) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return gzip.NewWriterLevel(w, int(g))
}
