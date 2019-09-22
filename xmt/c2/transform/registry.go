package transform

import (
	"errors"
	"fmt"
	"io"

	"github.com/iDigitalFlame/xmt/xmt/data"
)

var (
	// ErrInvalidTransform is an error that is returned when the
	// attempted read received a nil value or a value that does not
	// implement the c2.Transform interface.
	ErrInvalidTransform = errors.New("received an invalid transform")

	transforms = map[uint8]func() interface{}{
		dnsID:    func() interface{} { return new(DNSClient) },
		base64ID: func() interface{} { return new(b64) },
	}
)

// Transform is an interface that can modify the data BEFORE
// it is written or AFTER is read from a Connection.
// Transforms may be used to mask and unmask communications
// as benign protocols such as DNS, FTP or HTTP.
type Transform interface {
	Read(io.Writer, []byte) error
	Write(io.Writer, []byte) error
}

// Read attempts to read the first uint16 from the supplied
// Reader and create a new Transform with the data provided by
// the Reader.
func Read(r data.Reader) (Transform, error) {
	w, err := readShallow(r)
	if err != nil {
		return nil, err
	}
	if o, ok := w.(Transform); ok {
		return o, nil
	}
	return nil, ErrInvalidTransform
}

// Write will attempt to write the supplied transform to the
// Writer. This will return an error is the current wrapper does
// not support writing. Transforms are responsible for writing their
// ID values to be read.
func Write(i Transform, w data.Writer) error {
	if i == nil {
		if err := w.WriteUint8(0); err != nil {
			return err
		}
		return nil
	}
	o, ok := i.(data.Writable)
	if !ok {
		return fmt.Errorf("transform \"%T\" does not support the \"data.Writable\" interface", i)
	}
	if err := o.MarshalStream(w); err != nil {
		return fmt.Errorf("unable to marshal transform: %w", err)
	}
	return nil
}

// Register associates the uint16 number supplied with a function
// that will create a Transform that supports reading from a Reader.
// The Read function will take care of supplying the data once the
// new Transform is created.
func Register(i uint8, f func() interface{}) error {
	if _, ok := transforms[i]; ok {
		return fmt.Errorf("transform ID %d is already registered", i)
	}
	transforms[i] = f
	return nil
}
func readShallow(r data.Reader) (interface{}, error) {
	i, err := r.Uint8()
	if err != nil {
		return nil, fmt.Errorf("unable to read transform: %w", err)
	}
	f, ok := transforms[i]
	if !ok {
		return nil, fmt.Errorf("no transform is registered with ID %d", i)
	}
	w := f()
	if w == nil {
		return nil, ErrInvalidTransform
	}
	if z, ok := w.(data.Readable); ok {
		if err := z.UnmarshalStream(r); err != nil {
			return nil, fmt.Errorf("unable to unmarshal transform: %w", err)
		}
	}
	return w, nil
}
