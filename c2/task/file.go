package task

import (
	"context"
	"io"
	"os"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
)

// Upload returns a Packet that will instruct a Client to upload the specified local file to the server.
func Upload(s string) *com.Packet {
	p := &com.Packet{ID: TvUpload}
	p.WriteString(s)
	return p
}

// Download returns a Packet that will instruct a Client to save the specified bytes to the local file location.
func Download(s string, b []byte) *com.Packet {
	p := &com.Packet{ID: TvDownload}
	p.WriteString(s)
	p.Write(b)
	return p
}

// DownloadFile returns a Packet that will instruct a Client to save the contents of the supplied local file to
// the remote file location. This will return an error if any errors occur during reading or opening the local file.
func DownloadFile(s, r string) (*com.Packet, error) {
	f, err := os.OpenFile(device.Expand(r), os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	p, err := DownloadReader(s, f)
	f.Close()
	return p, err
}

// DownloadReader returns a Packet that will instruct a Client to save the contents of the supplied reader to
// the remote file location. This will return an error if any errors occur during reading.
func DownloadReader(s string, r io.Reader) (*com.Packet, error) {
	p := &com.Packet{ID: TvDownload}
	p.WriteString(s)
	_, err := io.Copy(p, r)
	return p, err
}
func upload(x context.Context, p *com.Packet) (*com.Packet, error) {
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
func download(x context.Context, p *com.Packet) (*com.Packet, error) {
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
	w.WriteString(h)
	w.WriteInt64(n)
	return w, err
}
