package crypto

import (
	"io"
	"sync"

	"github.com/iDigitalFlame/xmt/data"
)

var (
	bufs = &sync.Pool{
		New: func() interface{} {
			return make([]byte, smallBuf)
		},
	}

	smallBuf = 512
)

// XOR is an alias for a byte array that acts as the XOR
// key data buffer.
type XOR []byte

// Operate preforms the XOR operation on the specified byte
// array using the cipher as the key.
func (x XOR) Operate(b []byte) {
	if x == nil || len(x) == 0 {
		return
	}
	for i := 0; i < len(b); i++ {
		b[i] = b[i] ^ x[i%len(x)]
	}
}

// Decrypt preforms the XOR operation on the specified byte
// array using the cipher as the key.
func (x XOR) Decrypt(b []byte) {
	x.Operate(b)
}

// Encrypt preforms the XOR operation on the specified byte
// array using the cipher as the key.
func (x XOR) Encrypt(b []byte) {
	x.Operate(b)
}

// Flush satisfies the crypto.Writer interface.
func (XOR) Flush(w io.Writer) error {
	if f, ok := w.(data.Flusher); ok {
		return f.Flush()
	}
	return nil
}

// Read satisfies the crypto.Reader interface.
func (x XOR) Read(r io.Reader, b []byte) (int, error) {
	n, err := r.Read(b)
	x.Operate(b)
	return n, err
}

// Write satisfies the crypto.Writer interface.
func (x XOR) Write(w io.Writer, b []byte) (int, error) {
	n := len(b)
	var o []byte
	if n < smallBuf {
		o = bufs.Get().([]byte)
		defer bufs.Put(o)
	} else {
		o = make([]byte, n)
	}
	copy(o, b)
	x.Operate(o[:n])
	return w.Write(o[:n])
}
