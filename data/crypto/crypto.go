package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"io"

	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util/xerr"
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
type flusher interface {
	Flush() error
}

// Reader is an interface that supports reading bytes from a Reader through the wrapped Cipher.
type Reader interface {
	Read(io.Reader, []byte) (int, error)
}

// Writer is an interface that supports writing bytes to a Writer through the wrapped Cipher.
type Writer interface {
	Flush(io.Writer) error
	Write(io.Writer, []byte) (int, error)
}

func (w *writer) Flush() error {
	if err := w.c.Flush(w.w); err != nil {
		return err
	}
	if f, ok := w.w.(flusher); ok {
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

// NewReader creates a io.ReadCloser type from the specified Cipher and Reader.
func NewReader(c Reader, r io.Reader) io.Reader {
	if c == nil {
		return r
	}
	return &reader{c: c, r: r}
}

// NewWriter creates a io.WriteCloser type from the specified Cipher and Writer.
func NewWriter(c Writer, w io.Writer) io.WriteCloser {
	if c == nil {
		return data.WriteCloser(w)
	}
	return &writer{c: c, w: w}
}

// Decrypter creates a data.Reader type from the specified block Cipher, IV and Reader.
//
// This is used to Decrypt data. This function returns an error if the blocksize of the Block does not equal
// the length of the supplied IV.
func Decrypter(b cipher.Block, iv []byte, r io.Reader) (io.ReadCloser, error) {
	if len(iv) != b.BlockSize() {
		return nil, xerr.Sub("block size must equal IV size", 0x10)
	}
	return io.NopCloser(&cipher.StreamReader{R: r, S: cipher.NewCFBDecrypter(b, iv)}), nil
}

// Encrypter creates a data.Reader type from the specified block Cipher, IV and Writer.
//
// This is used to Encrypt data. This function returns an error if the blocksize of the Block does not equal
// the length of the supplied IV.
func Encrypter(b cipher.Block, iv []byte, w io.Writer) (io.WriteCloser, error) {
	if len(iv) != b.BlockSize() {
		return nil, xerr.Sub("block size must equal IV size", 0x10)
	}
	return &cipher.StreamWriter{W: w, S: cipher.NewCFBEncrypter(b, iv)}, nil
}
