package crypto

import (
	"io"
	"sync"

	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
)

const bufMax = 2 << 14 // Should cover the default chunk.Write buffer size

var bufs = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 512, bufMax)
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
	n := copy(dst, src)
	x.Operate(dst[:n])
}

// Encrypt preforms the XOR operation on the specified byte array using the cipher as the key.
func (x XOR) Encrypt(dst, src []byte) {
	n := copy(dst, src)
	x.Operate(dst[:n])
}

// Read satisfies the crypto.Reader interface.
func (x XOR) Read(r io.Reader, b []byte) (int, error) {
	n, err := io.ReadFull(r, b)
	//        NOTE(dij): ErrUnexpectedEOF happens here on short (< buf size)
	//                   Reads, though is completely normal.
	x.Operate(b[:n])
	return n, err
}

// Write satisfies the crypto.Writer interface.
func (x XOR) Write(w io.Writer, b []byte) (int, error) {
	n := len(b)
	if n > bufMax {
		if bugtrack.Enabled {
			bugtrack.Track("crypto.XOR.Write(): Creating non-heap buffer, n=%d, bufMax=%d", n, bufMax)
		}
		o := make([]byte, n)
		copy(o, b)
		x.Operate(o)
		z, err := w.Write(o)
		// NOTE(dij): Make the GCs job easy
		o = nil
		return z, err
	}
	o := bufs.Get().(*[]byte)
	if len(*o) < n {
		if bugtrack.Enabled {
			bugtrack.Track("crypto.XOR.Write(): Increasing heap buffer size len(*o)=%d, n=%d", len(*o), n)
		}
		*o = append(*o, make([]byte, n-len(*o))...)
	}
	copy(*o, b)
	x.Operate((*o)[:n])
	z, err := w.Write((*o)[:n])
	bufs.Put(o)
	return z, err
}
