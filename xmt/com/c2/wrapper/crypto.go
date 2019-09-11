package wrapper

import (
	"crypto/cipher"
	"errors"
	"io"

	"github.com/iDigitalFlame/xmt/xmt/com/c2"
	"github.com/iDigitalFlame/xmt/xmt/crypto"
)

var (
	// ErrInvalid is returned when the arguments provided
	// to any of the New* functions when the arguments are nil or
	// empty.
	ErrInvalid = errors.New("provider crypto arguments cannot be nil")
)

type cipherBlock struct {
	iv    []byte
	block cipher.Block
}
type cipherWrapper struct {
	w crypto.Writer
	r crypto.Reader
}

// NewCryptoBlock returns a Wrapper based on a Block Cipher,
// such as AES.
func NewCryptoBlock(b cipher.Block, iv []byte) (c2.Wrapper, error) {
	if b == nil || iv == nil {
		return nil, ErrInvalid
	}
	return &cipherBlock{iv: iv, block: b}, nil
}

// NewCrypto returns a Wrapper based on the crypto.Writer and crypto.Reader
// interfaces, such as XOR and CBK.
func NewCrypto(r crypto.Reader, w crypto.Writer) (c2.Wrapper, error) {
	if r == nil || w == nil {
		return nil, ErrInvalid
	}
	return &cipherWrapper{r: r, w: w}, nil
}
func (c *cipherBlock) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return crypto.BlockEncryptWriter(c.block, c.iv, w), nil
}
func (c *cipherBlock) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	return crypto.BlockDecryptReader(c.block, c.iv, r), nil
}
func (c *cipherWrapper) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return crypto.NewWriter(c.w, w), nil
}
func (c *cipherWrapper) Unwrap(r io.ReadCloser) (io.ReadCloser, error) {
	return crypto.NewReader(c.r, r), nil
}
