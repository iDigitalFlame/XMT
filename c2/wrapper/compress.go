package wrapper

import (
	"compress/gzip"
	"compress/zlib"
	"io"
	"strconv"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

const (
	// Zlib is the default Zlib Wrapper. This wrapper uses the default compression level. Use the 'NewZlib'
	// function to create a wrapper with a different level.
	Zlib = ZlibWrap(zlib.DefaultCompression)

	// Gzip is the default Gzip Wrapper. This wrapper uses the default compression level. Use the 'NewGzip'
	// function to create a wrapper with a different level.
	Gzip = GzipWrap(zlib.DefaultCompression)
)

// ZlibWrap is a alias for a Zlib compression level that implements the 'c2.Wrapper' interface.
type ZlibWrap int8

// GzipWrap is a alias for a Gzip compression level that implements the 'c2.Wrapper' interface.
type GzipWrap int8

// NewZlib returns a Zlib compreession wrapper. This function will return and error if the commpression level
// is invalid.
func NewZlib(level int) (ZlibWrap, error) {
	if level < zlib.HuffmanOnly || level > zlib.BestCompression {
		return 0, xerr.New("invalid compression level " + strconv.Itoa(level))
	}
	return ZlibWrap(level), nil
}

// NewGzip returns a Gzip compreession wrapper. This function will return and error if the commpression level
// is invalid.
func NewGzip(level int) (GzipWrap, error) {
	if level < gzip.HuffmanOnly || level > gzip.BestCompression {
		return 0, xerr.New("invalid compression level " + strconv.Itoa(level))
	}
	return GzipWrap(level), nil
}

// Unwrap satisfies the Wrapper interface.
func (GzipWrap) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	return gzip.NewReader(r)
}

// Unwrap satisfies the Wrapper interface.
func (ZlibWrap) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	return zlib.NewReader(r)
}

// Wrap satisfies the Wrapper interface.
func (z ZlibWrap) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return zlib.NewWriterLevel(w, int(z))
}

// Wrap satisfies the Wrapper interface.
func (g GzipWrap) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return gzip.NewWriterLevel(w, int(g))
}
