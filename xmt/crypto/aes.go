package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"

	"github.com/iDigitalFlame/xmt/xmt/data"
)

const (
	// AesID is the numeric ID value for the
	// AES Cipher when read and written to/from a stream.
	AesID uint8 = 0xC4
	// DesID is the numeric ID value for the
	// DES Cipher when read and written to/from a stream.
	DesID uint8 = 0xC5
	// TrippleDesID is the numeric ID value for the
	// Tripple DES Cipher when read and written to/from a stream.
	TrippleDesID uint8 = 0xC6
)

// Block is the basic block Cipher wrapper.
// Implement this struct or use the Wrap function
// to enable reading and writing this cipher to
// streams.
type Block struct {
	ID uint8

	key   []byte
	block cipher.Block
}

// CipherAes is a wrapper for the AES Block
// Cipher. This allows for it to be read and written
// to a stream.
type CipherAes struct {
	*Block
}

// CipherDes is a wrapper for the DES Block
// Cipher. This allows for it to be read and written
// to a stream.
type CipherDes struct {
	*Block
}

// CipherTrippleDes is a wrapper for the Tripple DES Block
// Cipher. This allows for it to be read and written
// to a stream.
type CipherTrippleDes struct {
	*Block
}

// BlockSize returns the cipher's block size.
func (c *Block) BlockSize() int {
	return c.block.BlockSize()
}

// Encrypt encrypts the first block in src into dst.
// Dst and src must overlap entirely or not at all.
func (c *Block) Encrypt(d, s []byte) {
	c.block.Encrypt(d, s)
}

// Decrypt decrypts the first block in src into dst.
// Dst and src must overlap entirely or not at all.
func (c *Block) Decrypt(d, s []byte) {
	c.block.Decrypt(d, s)
}

// NewDes attempts to create a new DES block Cipher from the
// provided key data. Errors will be returned if the key length
// is invalid.
func NewDes(k []byte) (cipher.Block, error) {
	b, err := des.NewCipher(k)
	if err != nil {
		return nil, err
	}
	return &CipherDes{&Block{
		ID:    DesID,
		key:   k,
		block: b,
	}}, nil
}

// NewAes attempts to create a new AES block Cipher from the
// provided key data. Errors will be returned if the key length
// is invalid.
func NewAes(k []byte) (cipher.Block, error) {
	b, err := aes.NewCipher(k)
	if err != nil {
		return nil, err
	}
	return &CipherAes{&Block{
		ID:    AesID,
		key:   k,
		block: b,
	}}, nil
}

// MarshalStream allows this Cipher to be written to a stream.
func (c *Block) MarshalStream(w data.Writer) error {
	if err := w.WriteUint8(c.ID); err != nil {
		return err
	}
	return w.WriteBytes(c.key)
}

// NewTrippleDes attempts to create a new Tripple DES block Cipher from the
// provided key data. Errors will be returned if the key length
// is invalid.
func NewTrippleDes(k []byte) (cipher.Block, error) {
	b, err := des.NewTripleDESCipher(k)
	if err != nil {
		return nil, err
	}
	return &CipherTrippleDes{&Block{
		ID:    TrippleDesID,
		key:   k,
		block: b,
	}}, nil
}

// Wrap creates a writable Block cipher from the specified
// ID, key and cipher.Block. This wrapping allows the block to be
// written to a stream when the enclosing wrapper is also written.
// Be sure to register the ID using the wrapper.Register function
// in order to read this block cipher from a stream.
func Wrap(id uint8, k []byte, b cipher.Block) *Block {
	return &Block{
		ID:    id,
		key:   k,
		block: b,
	}
}

// UnmarshalStream allows this Cipher to be read from a stream.
func (c *Block) UnmarshalStream(r data.Reader) error {
	var err error
	if c.key, err = r.Bytes(); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream allows this Cipher to be read from a stream.
func (c *CipherAes) UnmarshalStream(r data.Reader) error {
	var err error
	if err := c.Block.UnmarshalStream(r); err != nil {
		return err
	}
	if c.block, err = aes.NewCipher(c.key); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream allows this Cipher to be read from a stream.
func (c *CipherDes) UnmarshalStream(r data.Reader) error {
	var err error
	if err := c.Block.UnmarshalStream(r); err != nil {
		return err
	}
	if c.block, err = des.NewCipher(c.key); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream allows this Cipher to be read from a stream.
func (c *CipherTrippleDes) UnmarshalStream(r data.Reader) error {
	var err error
	if err := c.Block.UnmarshalStream(r); err != nil {
		return err
	}
	if c.block, err = des.NewTripleDESCipher(c.key); err != nil {
		return err
	}
	return nil
}
