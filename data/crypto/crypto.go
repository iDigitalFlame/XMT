package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"

	"github.com/iDigitalFlame/xmt/data"
)

type reader struct {
	_ [0]func()
	r io.Reader
	c Reader
}
type writer struct {
	_ [0]func()
	w io.Writer
	c Writer
}

// Reader is an interface that supports reading bytes from a Reader through the specified Cipher.
type Reader interface {
	Read(io.Reader, []byte) (int, error)
}

// Writer is an interface that supports writing bytes to a Writer through the specified Cipher.
type Writer interface {
	Flush(io.Writer) error
	Write(io.Writer, []byte) (int, error)
}

func (w *writer) Flush() error {
	if err := w.c.Flush(w.w); err != nil {
		return err
	}
	if f, ok := w.w.(data.Flusher); ok {
		return f.Flush()
	}
	return nil
}
func (w *writer) Close() error {
	if err := w.Flush(); err != nil {
		return err
	}
	if c, ok := w.w.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// NewAes attempts to create a new AES block Cipher from the provided key data. Errors will be returned
// if the key length is invalid.
func NewAes(k []byte) (cipher.Block, error) {
	return aes.NewCipher(k)
}
func (r *reader) Read(b []byte) (int, error) {
	return r.c.Read(r.r, b)
}
func (w *writer) Write(b []byte) (int, error) {
	return w.c.Write(w.w, b)
}

// NewReader creates a data.Reader type from the specified Cipher and Reader.
func NewReader(c Reader, r io.Reader) data.Reader {
	if c == nil {
		return data.NewReader(r)
	}
	return data.NewReader(&reader{c: c, r: r})
}

// NewWriter creates a data.Writer type from the specified Cipher and Writer.
func NewWriter(c Writer, w io.Writer) data.Writer {
	if c == nil {
		return data.NewWriter(w)
	}
	return data.NewWriter(&writer{c: c, w: w})
}

// DecryptReader creates a data.Reader type from the specified block Cipher, IV and Reader.
// This is used to Decrypt data. This function returns an error if the blocksize of the Block does not equal
// the length of the supplied IV.
func DecryptReader(b cipher.Block, iv []byte, r io.Reader) (data.Reader, error) {
	if len(iv) != b.BlockSize() {
		return nil, fmt.Errorf("blocksize (%d) must equal IV size (%d)", b.BlockSize(), len(iv))
	}
	return data.NewReader(&cipher.StreamReader{
		R: r,
		S: cipher.NewCFBDecrypter(b, iv),
	}), nil
}

// DecryptWriter creates a data.Writer type from the specified block Cipher, IV and Writer.
// This is used to Decrypt data. This function returns an error if the blocksize of the Block does not equal
// the length of the supplied IV.
func DecryptWriter(b cipher.Block, iv []byte, w io.Writer) (data.Writer, error) {
	if len(iv) != b.BlockSize() {
		return nil, fmt.Errorf("blocksize (%d) must equal IV size (%d)", b.BlockSize(), len(iv))
	}
	return data.NewWriter(&cipher.StreamWriter{
		W: w,
		S: cipher.NewCFBDecrypter(b, iv),
	}), nil
}

// EncryptReader creates a data.Reader type from the specified block Cipher, IV and Reader.
// This is used to Encrypt data. This function returns an error if the blocksize of the Block does not equal
// the length of the supplied IV.
func EncryptReader(b cipher.Block, iv []byte, r io.Reader) (data.Reader, error) {
	if len(iv) != b.BlockSize() {
		return nil, fmt.Errorf("blocksize (%d) must equal IV size (%d)", b.BlockSize(), len(iv))
	}
	return data.NewReader(&cipher.StreamReader{
		R: r,
		S: cipher.NewCFBEncrypter(b, iv),
	}), nil
}

// EncryptWriter creates a data.Reader type from the specified block Cipher, IV and Writer.
// This is used to Encrypt data. This function returns an error if the blocksize of the Block does not equal
// the length of the supplied IV.
func EncryptWriter(b cipher.Block, iv []byte, w io.Writer) (data.Writer, error) {
	if len(iv) != b.BlockSize() {
		return nil, fmt.Errorf("blocksize (%d) must equal IV size (%d)", b.BlockSize(), len(iv))
	}
	return data.NewWriter(&cipher.StreamWriter{
		W: w,
		S: cipher.NewCFBEncrypter(b, iv),
	}), nil
}
