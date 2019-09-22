package wrapper

import (
	"compress/gzip"
	"compress/zlib"
	"errors"
	"fmt"
	"io"

	"github.com/iDigitalFlame/xmt/xmt/crypto"
	"github.com/iDigitalFlame/xmt/xmt/crypto/cbk"
	"github.com/iDigitalFlame/xmt/xmt/crypto/xor"
	"github.com/iDigitalFlame/xmt/xmt/data"
)

var (
	// ErrInvalidWrapper is an error that is returned when the
	// attempted read received a nil value or a value that does not
	// implement the c2.Wrapper interface.
	ErrInvalidWrapper = errors.New("received an invalid wrapper")

	wrappers = map[uint8]func() interface{}{
		hexID:         func() interface{} { return Hex },
		base64ID:      func() interface{} { return Base64 },
		xor.XorID:     func() interface{} { return new(xor.Cipher) },
		cbk.CbkID:     func() interface{} { return new(cbk.Cipher) },
		crypto.AesID:  func() interface{} { return &crypto.CipherAes{Block: &crypto.Block{ID: crypto.AesID}} },
		crypto.DesID:  func() interface{} { return &crypto.CipherDes{Block: &crypto.Block{ID: crypto.DesID}} },
		cryptoWrapID:  func() interface{} { return new(cipherWrapper) },
		cryptoBlockID: func() interface{} { return new(cipherBlock) },

		zlibID: func() interface{} {
			z := zlibWrapper(zlib.DefaultCompression)
			return &z
		},
		gzipID: func() interface{} {
			g := zlibWrapper(gzip.DefaultCompression)
			return &g
		},
		crypto.TrippleDesID: func() interface{} {
			return &crypto.CipherTrippleDes{
				Block: &crypto.Block{ID: crypto.TrippleDesID},
			}
		},
	}
)

// Wrapper is an interface that allows for wrapping the
// binary streams into separate stream types. This allows for
// using encryption or compression.
type Wrapper interface {
	Wrap(io.WriteCloser) (io.WriteCloser, error)
	Unwrap(io.ReadCloser) (io.ReadCloser, error)
}

// Read attempts to read the first uint16 from the supplied
// Reader and create a new Wrapper with the data provided by
// the Reader.
func Read(r data.Reader) (Wrapper, error) {
	w, err := readShallow(r)
	if err != nil {
		return nil, err
	}
	if o, ok := w.(Wrapper); ok {
		return o, nil
	}
	return nil, ErrInvalidWrapper
}

// Write will attempt to write the supplied wrapper to the
// Writer. This will return an error is the current wrapper does
// not support writing. Wrappers are responsible for writing their
// ID values to be read.
func Write(i Wrapper, w data.Writer) error {
	if i == nil {
		if err := w.WriteUint8(0); err != nil {
			return err
		}
		return nil
	}
	o, ok := i.(data.Writable)
	if !ok {
		return fmt.Errorf("wrapper \"%T\" does not support the \"data.Writable\" interface", i)
	}
	if err := o.MarshalStream(w); err != nil {
		return fmt.Errorf("unable to marshal wrapper: %w", err)
	}
	return nil
}

// Register associates the uint16 number supplied with a function
// that will create a Wrapper that supports reading from a Reader.
// The Read function will take care of supplying the data once the
// new Wrapper is created.
func Register(i uint8, f func() interface{}) error {
	if _, ok := wrappers[i]; ok {
		return fmt.Errorf("wrapper ID %d is already registered", i)
	}
	wrappers[i] = f
	return nil
}
func readShallow(r data.Reader) (interface{}, error) {
	i, err := r.Uint8()
	if err != nil {
		return nil, fmt.Errorf("unable to read wrapper: %w", err)
	}
	f, ok := wrappers[i]
	if !ok {
		return nil, fmt.Errorf("no wrapper is registered with ID %d", i)
	}
	w := f()
	if w == nil {
		return nil, ErrInvalidWrapper
	}
	if z, ok := w.(data.Readable); ok {
		if err := z.UnmarshalStream(r); err != nil {
			return nil, fmt.Errorf("unable to unmarshal wrapper: %w", err)
		}
	}
	return w, nil
}
