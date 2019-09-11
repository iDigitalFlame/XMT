package wrapper

import (
	"io"

	"github.com/iDigitalFlame/xmt/xmt/com/c2"
)

type Multi struct {
	list []c2.Wrapper
}

func NewMulti(w ...c2.Wrapper) *Multi {
	return &Multi{list: w}
}

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
