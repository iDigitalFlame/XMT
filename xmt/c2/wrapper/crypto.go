package wrapper

import (
	"crypto/cipher"
	"errors"
	"fmt"
	"io"

	"github.com/iDigitalFlame/xmt/xmt/crypto"
	"github.com/iDigitalFlame/xmt/xmt/data"
)

const (
	cryptoWrapID  uint8 = 0xC0
	cryptoBlockID uint8 = 0xC2
)

var (
	// ErrInvalid is returned when the arguments provided
	// to any of the New* functions when the arguments are nil or
	// empty.
	ErrInvalid = errors.New("provided crypto arguments cannot be nil")

	// ErrInvalidRead is an error that is returned when the attempted writer
	// returned from a Read function is not valid as a c2.Writer. This error can be
	// returned during the last sanity checks.
	ErrInvalidRead = errors.New("received an invalid reader/writer value type")
)

type cipherBlock struct {
	iv    []byte
	block cipher.Block
}
type cipherWrapper struct {
	w crypto.Writer
	r crypto.Reader
}

func (c *cipherBlock) MarshalStream(w data.Writer) error {
	if err := w.WriteUint8(cryptoBlockID); err != nil {
		return err
	}
	if c.block != nil {
		b, ok := c.block.(data.Writable)
		if !ok {
			return fmt.Errorf("crypto block \"%T\" does not support the \"data.Writable\" interface", c.block)
		}
		if err := b.MarshalStream(w); err != nil {
			return err
		}
	} else {
		if err := w.WriteUint8(0); err != nil {
			return err
		}
	}
	return w.WriteBytes(c.iv)
}
func (c *cipherBlock) UnmarshalStream(r data.Reader) error {
	i, err := readShallow(r)
	if err != nil {
		return err
	}
	if i != nil {
		if b, ok := i.(cipher.Block); ok {
			c.block = b
		} else {
			return ErrInvalidRead
		}
	} else {
		return ErrInvalidRead
	}
	if c.iv, err = r.Bytes(); err != nil {
		return err
	}
	return nil
}
func (c *cipherWrapper) MarshalStream(w data.Writer) error {
	if err := w.WriteUint8(cryptoWrapID); err != nil {
		return err
	}
	if c.r != nil {
		i, ok := c.r.(data.Writable)
		if !ok {
			return fmt.Errorf("crypto reader \"%T\" does not support the \"data.Writable\" interface", c.r)
		}
		if err := i.MarshalStream(w); err != nil {
			return err
		}
	} else {
		if err := w.WriteUint8(0); err != nil {
			return err
		}
	}
	if c.w != nil {
		o, ok := c.w.(data.Writable)
		if !ok {
			return fmt.Errorf("crypto reader \"%T\" does not support the \"data.Writable\" interface", c.w)
		}
		if err := o.MarshalStream(w); err != nil {
			return err
		}
	} else {
		if err := w.WriteUint8(0); err != nil {
			return err
		}
	}
	return nil
}
func (c *cipherWrapper) UnmarshalStream(r data.Reader) error {
	i, err := readShallow(r)
	if err != nil {
		return err
	}
	if i != nil {
		if q, ok := i.(crypto.Reader); ok {
			c.r = q
		} else {
			return ErrInvalidRead
		}
	} else {
		return ErrInvalidRead
	}
	o, err := readShallow(r)
	if err != nil {
		return err
	}
	if o != nil {
		if q, ok := o.(crypto.Writer); ok {
			c.w = q
		} else {
			return ErrInvalidRead
		}
	} else {
		return ErrInvalidRead
	}
	return nil
}

// NewCryptoBlock returns a Wrapper based on a Block Cipher,
// such as AES.
func NewCryptoBlock(b cipher.Block, iv []byte) (Wrapper, error) {
	if b == nil || iv == nil {
		return nil, ErrInvalid
	}
	return &cipherBlock{iv: iv, block: b}, nil
}

// NewCrypto returns a Wrapper based on the crypto.Writer and crypto.Reader
// interfaces, such as XOR and CBK.
func NewCrypto(r crypto.Reader, w crypto.Writer) (Wrapper, error) {
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
