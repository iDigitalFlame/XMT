package transform

import (
	"encoding/base64"
	"io"
)

// Base64 is a transform that auto converts the data to and from Base64 encoding. This instance does not include
// any shifting.
const Base64 = b64(0)

type b64 byte

// Value is an interface that can modify the data BEFORE it is written or AFTER is read from a Connection.
// Transforms may be used to mask and unmask communications as benign protocols such as DNS, FTP or HTTP. This
// is just a compatibility interface to prevent import dependency cycles
type Value interface {
	Read(io.Writer, []byte) error
	Write(io.Writer, []byte) error
}

// Base64Shift returns a Base64 Transform that also shifts the bytes by the specified amount before writes
// and after reads. This is useful for evading detection by avoiding commonly flagged Base64 values.
func Base64Shift(n int) Value {
	return b64(n)
}
func (b b64) Read(w io.Writer, p []byte) error {
	var (
		i []byte
		c = base64.StdEncoding.DecodedLen(len(p))
	)
	if c < dnsSize {
		i = *bufs.Get().(*[]byte)
		defer bufs.Put(&i)
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
	var (
		c = base64.StdEncoding.EncodedLen(len(p))
		o []byte
	)
	if c < dnsSize {
		o = *bufs.Get().(*[]byte)
		defer bufs.Put(&o)
	} else {
		o = make([]byte, c)
	}
	base64.StdEncoding.Encode(o, p)
	_, err := w.Write(o[:c])
	return err
}
