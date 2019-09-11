package xor

import (
	"io"
	"sync"

	"github.com/iDigitalFlame/xmt/xmt/data"
)

var (
	bufs = &sync.Pool{
		New: func() interface{} {
			return make([]byte, bufSize)
		},
	}

	bufSize = 512
)

// Cipher is an alias for a byte array that acts as the XOR
// key data.
type Cipher []byte

// XOR preforms the XOR operation on the specified byte
// array using the cipher as the key.
func (c Cipher) XOR(b []byte) {
	if c == nil || len(c) == 0 {
		return
	}
	for i := 0; i < len(b); i++ {
		b[i] = b[i] ^ c[i%len(c)]
	}
}

// Decrypt preforms the XOR operation on the specified byte
// array using the cipher as the key.
func (c Cipher) Decrypt(b []byte) {
	c.XOR(b)
}

// Encrypt preforms the XOR operation on the specified byte
// array using the cipher as the key.
func (c Cipher) Encrypt(b []byte) {
	c.XOR(b)
}

// Flush satisfies the crypto.Writer interface.
func (c Cipher) Flush(w io.Writer) error {
	if f, ok := w.(data.Flusher); ok {
		return f.Flush()
	}
	return nil
}

// Read satisfies the crypto.Reader interface.
func (c Cipher) Read(r io.Reader, b []byte) (int, error) {
	n, err := r.Read(b)
	c.XOR(b)
	return n, err
}

// Write satisfies the crypto.Writer interface.
func (c Cipher) Write(w io.Writer, b []byte) (int, error) {
	n := len(b)
	var o []byte
	if n < bufSize {
		o = bufs.Get().([]byte)
		defer bufs.Put(o)
	} else {
		o = make([]byte, n)
	}
	copy(o, b)
	c.XOR(o[:n])
	return w.Write(o[:n])
}
