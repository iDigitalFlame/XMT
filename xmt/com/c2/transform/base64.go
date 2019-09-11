package transform

import (
	"encoding/base64"
	"io"

	"github.com/iDigitalFlame/xmt/xmt/com/c2"
)

const (
	// Base64 is a transform that auto converts the data to and
	// from Base64 encoding.
	Base64 = b64(0)
)

type b64 byte

// Base64Shift returns a Base64 Transform that
// also shifts the bytes by the specified amount before
// writes and after reads. This is useful for evading detection
// by avoiding commonly flagged Base64 strings.
func Base64Shift(n int) c2.Transform {
	return b64(n)
}
func (b b64) Read(p []byte, w io.Writer) error {
	c := base64.StdEncoding.DecodedLen(len(p))
	var i []byte
	if c < bufSize {
		i = bufs.Get().([]byte)
		defer bufs.Put(i)
	} else {
		i = make([]byte, c)
	}
	n, err := base64.StdEncoding.Decode(i, p)
	if err != nil {
		return err
	}
	if b != 0 {
		for x := 0; x < n && x < len(i); x++ {
			i[x] -= byte(b)
		}
	}
	_, err = w.Write(i[:n])
	return err
}
func (b b64) Write(p []byte, w io.Writer) error {
	if b != 0 {
		for i := range p {
			p[i] += byte(b)
		}
	}
	c := base64.StdEncoding.EncodedLen(len(p))
	var o []byte
	if c < bufSize {
		o = bufs.Get().([]byte)
		defer bufs.Put(o)
	} else {
		o = make([]byte, c)
	}
	base64.StdEncoding.Encode(o, p)
	_, err := w.Write(o[:c])
	return err
}
