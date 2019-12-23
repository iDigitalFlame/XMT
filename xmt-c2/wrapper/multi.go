package wrapper

import (
	"fmt"
	"io"

	data "github.com/iDigitalFlame/xmt/xmt-data"
)

const (
	multiID uint8 = 0xD5
)

// Multi is a struct that contains multiple Wrappers
// that can be used in a sequence to add multiple wrapping
// layers.
type Multi struct {
	list []Wrapper
}

// NewMulti creates a new Multi struct based on the passed
// Wrapper vardict. This function will IGNORE all Multi Wrapper
// structs.
func NewMulti(w ...Wrapper) *Multi {
	m := &Multi{list: make([]Wrapper, 0, len(w))}
	for x := range w {
		if _, ok := w[x].(*Multi); !ok {
			m.list = append(m.list, w[x])
		}
	}
	return m
}

// MarshalStream writes this Multi Wrapper's data to the supplied
// Writer.
func (m *Multi) MarshalStream(w data.Writer) error {
	if err := w.WriteUint8(multiID); err != nil {
		return err
	}
	if err := w.WriteUint8(uint8(len(m.list))); err != nil {
		return err
	}
	var ok bool
	var o data.Writeable
	for x := range m.list {
		if o, ok = m.list[x].(data.Writeable); !ok {
			return fmt.Errorf("wrapper \"%T\" does not support the \"data.Writeable\" interface", m.list[x])
		}
		if err := o.MarshalStream(w); err != nil {
			return err
		}
	}
	return nil
}

// UnmarshalStream attempts to read this Multi Wrapper's data from the
// supplied Reader.
func (m *Multi) UnmarshalStream(r data.Reader) error {
	n, err := r.Uint8()
	if err != nil {
		return err
	}
	m.list = make([]Wrapper, n)
	for x := range m.list {
		w, err := Read(r)
		if err != nil {
			return nil
		}
		m.list[x] = w
	}
	return nil
}

// Wrap satisfies the Wrapper interface.
func (m *Multi) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	o := w
	var err error
	for x := len(m.list) - 1; x > 0; x-- {
		if o, err = m.list[x].Wrap(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}

// Unwrap satisfies the Wrapper interface.
func (m *Multi) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	o := r
	var err error
	for x := len(m.list) - 1; x > 0; x-- {
		if o, err = m.list[x].Unwrap(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}
