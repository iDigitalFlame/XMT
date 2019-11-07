package transform

import (
	"encoding/base64"
	"io"

	"github.com/iDigitalFlame/xmt/xmt/data"
)

const (
	// Base64 is a transform that auto converts the data to and
	// from Base64 encoding.
	Base64 = b64(0)

	base64ID uint8 = 0xE1
)

type b64 byte

// Base64Shift returns a Base64 Transform that
// also shifts the bytes by the specified amount before
// writes and after reads. This is useful for evading detection
// by avoiding commonly flagged Base64 strings.
func Base64Shift(n int) Transform {
	return b64(n)
}
func (b b64) Read(w io.Writer, p []byte) error {
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
func (b b64) Write(w io.Writer, p []byte) error {
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
func (b b64) MarshalStream(w data.Writer) error {
	if err := w.WriteUint8(base64ID); err != nil {
		return err
	}
	return w.WriteUint8(uint8(b))
}
func (b *b64) UnmarshalStream(r data.Reader) error {
	v, err := r.Uint8()
	if err != nil {
		return err
	}
	*b = b64(v)
	return nil
}
