package transform

import (
	"encoding/base64"
	"io"
	"sync"

	"github.com/iDigitalFlame/xmt/util/bugtrack"
)

// Base64 is a transform that auto converts the data to and from Base64
// encoding. This instance does not include any shifting.
const Base64 = B64(0)

const bufMax = 2 << 14

var bufs = sync.Pool{
	New: func() any {
		b := make([]byte, 512, bufMax)
		return &b
	},
}

// B64 is the underlying type for the Base64 Transform. This Transform encodes
// data into a Base64 string before the final write to the output.
type B64 byte

// B64Shift returns a Base64 Transform that also shifts the bytes by the
// specified amount before writes and after reads. This is useful for evading
// detection by avoiding commonly flagged Base64 values.
func B64Shift(n int) B64 {
	return B64(n)
}

// Read satisfies the Transform interface requirements.
func (b B64) Read(p []byte, w io.Writer) error {
	n := base64.StdEncoding.DecodedLen(len(p))
	if n > bufMax {
		if bugtrack.Enabled {
			bugtrack.Track("transform.B64.Read(): Creating non-heap buffer, n=%d, bufMax=%d", n, bufMax)
		}
		var (
			o   = make([]byte, n)
			err = decodeShift(w, byte(b), p, &o)
		)
		o = nil
		return err
	}
	o := bufs.Get().(*[]byte)
	if len(*o) < n {
		if bugtrack.Enabled {
			bugtrack.Track("transform.B64.Read(): Increasing heap buffer size len(*o)=%d, n=%d", len(*o), n)
		}
		*o = append(*o, make([]byte, n-len(*o))...)
	}
	err := decodeShift(w, byte(b), p, o)
	bufs.Put(o)
	return err
}

// Write satisfies the Transform interface requirements.
func (b B64) Write(p []byte, w io.Writer) error {
	if b != 0 {
		for i := range p {
			p[i] += byte(b)
		}
	}
	var (
		e      = base64.NewEncoder(base64.StdEncoding, w)
		c, err = e.Write(p)
	)
	if e.Close(); c != len(p) {
		return io.ErrShortWrite
	}
	e = nil
	return err
}
func decodeShift(w io.Writer, b byte, p []byte, o *[]byte) error {
	n, err := base64.StdEncoding.Decode(*o, p)
	if err != nil {
		return err
	}
	if b != 0 {
		for x := 0; x < n; x++ {
			(*o)[x] -= b
		}
	}
	c, err := w.Write((*o)[:n])
	if err != nil {
		return err
	}
	if c != n {
		return io.ErrShortWrite
	}
	return nil
}
