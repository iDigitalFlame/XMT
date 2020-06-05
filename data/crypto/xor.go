package crypto

import (
	"io"
	"sync"

	"github.com/iDigitalFlame/xmt/data"
)

const smallBuf = 512

var bufs = sync.Pool{
	New: func() interface{} {
		b := make([]byte, smallBuf)
		return &b
	},
}

// XOR is an alias for a byte array that acts as the XOR
// key data buffer.
type XOR []byte

// BlockSize returns the cipher's block size.
func (x XOR) BlockSize() int {
	return len(x)
}

// Operate preforms the XOR operation on the specified byte
// array using the cipher as the key.
func (x XOR) Operate(b []byte) {
	if len(x) == 0 {
		return
	}
	for i := 0; i < len(b); i++ {
		b[i] = b[i] ^ x[i%len(x)]
	}
}

// Flush satisfies the crypto.Writer interface.
func (XOR) Flush(w io.Writer) error {
	if f, ok := w.(data.Flusher); ok {
		return f.Flush()
	}
	return nil
}

// Decrypt preforms the XOR operation on the specified byte array using the cipher as the key.
func (x XOR) Decrypt(dst, src []byte) {
	copy(dst, src)
	x.Operate(dst)
}

// Encrypt preforms the XOR operation on the specified byte array using the cipher as the key.
func (x XOR) Encrypt(dst, src []byte) {
	copy(dst, src)
	x.Operate(dst)
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
		o = *bufs.Get().(*[]byte)
		defer bufs.Put(&o)
	} else {
		o = make([]byte, n)
	}
	copy(o, b)
	x.Operate(o[:n])
	return w.Write(o[:n])
}
