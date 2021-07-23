package task

import (
	"context"
	"io"
	"os"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
)

// File is a tasklet that is responsible for Uploads to Server (Pull) and Downloading to Client (Push).
const File = file(0)

type file uint8

func (file) Download(s string) *com.Packet {
	p := &com.Packet{ID: TvDownload}
	p.WriteString(s)
	return p
}
func (file) Upload(s string, b []byte) *com.Packet {
	p := &com.Packet{ID: TvUpload}
	p.WriteString(s)
	p.Write(b)
	return p
}
func (file) UploadFile(s, r string) (*com.Packet, error) {
	f, err := os.OpenFile(device.Expand(r), os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	p, err := File.UploadReader(s, f)
	f.Close()
	return p, err
}
func upload(x context.Context, p *com.Packet) (*com.Packet, error) {
	s, err := p.StringVal()
	if err != nil {
		return nil, err
	}
	var (
		h = device.Expand(s)
		f *os.File
	)
	if f, err = os.OpenFile(h, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		return nil, err
	}
	var (
		w = new(com.Packet)
		r = data.NewCtxReader(x, p)
	)
	n, err := io.Copy(f, r)
	r.Close()
	f.Close()
	p.Clear()
	w.WriteString(h)
	w.WriteInt64(n)
	return w, err
}
func (file) UploadReader(s string, r io.Reader) (*com.Packet, error) {
	p := &com.Packet{ID: TvUpload}
	p.WriteString(s)
	_, err := io.Copy(p, r)
	return p, err
}
func download(x context.Context, p *com.Packet) (*com.Packet, error) {
	var (
		s   string
		err error
	)
	if s, err = p.StringVal(); err != nil {
		return nil, err
	}
	var (
		h = device.Expand(s)
		f *os.File
		i os.FileInfo
	)
	if i, err = os.Stat(h); err != nil {
		return nil, err
	}
	var (
		d = i.IsDir()
		w = new(com.Packet)
	)
	w.WriteString(h)
	if w.WriteBool(d); d {
		w.WriteInt64(0)
		return w, nil
	}
	w.WriteInt64(i.Size())
	if f, err = os.OpenFile(h, os.O_RDONLY, 0); err != nil {
		return nil, err
	}
	r := data.NewCtxReader(x, f)
	_, err = io.Copy(w, r)
	r.Close()
	return w, err
}
