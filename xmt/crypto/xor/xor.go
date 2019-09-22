package xor

import (
	"io"
	"sync"

	"github.com/iDigitalFlame/xmt/xmt/data"
)

var (
	// XorID is the integer value used to represent
	// this Cipher when written to or read from a stream.
	XorID uint8 = 0xC1

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

// MarshalStream allows this Cipher to be written to a stream.
func (c Cipher) MarshalStream(w data.Writer) error {
	if err := w.WriteUint8(XorID); err != nil {
		return err
	}
	if err := w.WriteUint16(uint16(len(c))); err != nil {
		return err
	}
	if _, err := w.Write(c); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream allows this Cipher to be read from a stream.
func (c Cipher) UnmarshalStream(r data.Reader) error {
	l, err := r.Uint16()
	if err != nil {
		return err
	}
	c = make([]byte, l)
	n, err := r.Read(c)
	if err != nil && (err != io.EOF || l != uint16(n)) {
		return err
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
