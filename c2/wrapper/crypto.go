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

// Block is a struct that contains an IV and Block-based Cipher that can be used to Wrap/Unwrap with the specified
// encryption algorithm.
type Block struct {
	v []byte
	cipher.Block
}

// Stream is a struct that contains a XMT Crypto Reader/Writer that can be used to Wrap/Unwrap using the specified
// streaming Reader and/or Writer types.
type Stream struct {
	_ [0]func()
	w crypto.Writer
	r crypto.Reader
}

// NewBlock returns a Wrapper based on a Block Cipher, such as AES.
func NewBlock(b cipher.Block, v []byte) (*Block, error) {
	if b == nil || len(v) == 0 {
		return nil, ErrInvalid
	}
	return &Block{v: v, Block: b}, nil
}

// Wrap satisfies the Wrapper interface.
func (b *Block) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return crypto.EncryptWriter(b.Block, b.v, w)
}

// Unwrap satisfies the Wrapper interface.
func (b *Block) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	return crypto.DecryptReader(b.Block, b.v, r)
}

// Wrap satisfies the Wrapper interface.
func (s *Stream) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return crypto.NewWriter(s.w, w), nil
}

// Unwrap satisfies the Wrapper interface.
func (s *Stream) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	return crypto.NewReader(s.r, r), nil
}

// NewCrypto returns a Wrapper based on the crypto.Writer and crypto.Reader interfaces, such as XOR and CBK.
func NewCrypto(r crypto.Reader, w crypto.Writer) (*Stream, error) {
	if r == nil || w == nil {
		return nil, ErrInvalid
	}
	return &Stream{r: r, w: w}, nil
}
