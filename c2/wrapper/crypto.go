package wrapper

import (
	"crypto/cipher"
	"errors"
	"io"

	"github.com/iDigitalFlame/xmt/data/crypto"
)

// ErrInvalid is returned when the arguments provided to any of the New* functions when the
// arguments are nil or empty.
var ErrInvalid = errors.New("provided crypto arguments cannot be nil")

type block struct {
	v []byte
	cipher.Block
}
type writer struct {
	_ [0]func()
	w crypto.Writer
	r crypto.Reader
}

// NewBlock returns a Wrapper based on a Block Cipher, such as AES.
func NewBlock(b cipher.Block, v []byte) (Value, error) {
	if b == nil || len(v) == 0 {
		return nil, ErrInvalid
	}
	return &block{v: v, Block: b}, nil
}
func (b *block) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return crypto.EncryptWriter(b.Block, b.v, w)
}
func (b *block) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	return crypto.DecryptReader(b.Block, b.v, r)
}
func (c *writer) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return crypto.NewWriter(c.w, w), nil
}
func (c *writer) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	return crypto.NewReader(c.r, r), nil
}

// NewCrypto returns a Wrapper based on the crypto.Writer and crypto.Reader interfaces, such as XOR and CBK.
func NewCrypto(r crypto.Reader, w crypto.Writer) (Value, error) {
	if r == nil || w == nil {
		return nil, ErrInvalid
	}
	return &writer{r: r, w: w}, nil
}
