// Copyright (C) 2020 - 2023 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

// Package crypto contains helper functions and interfaces that can be used to
// easily read and write different types of encrypted data.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"io"

	"github.com/iDigitalFlame/xmt/data/crypto/subtle"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

type reader struct {
	_ [0]func()
	r io.Reader
	c *CBK
}
type writer struct {
	_ [0]func()
	w io.Writer
	c *CBK
}
type nopCloser struct {
	io.Reader
}
type flusher interface {
	Flush() error
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
func (nopCloser) Close() error {
	return nil
}

// UnwrapString is used to un-encode a string written in a XOR byte array "encrypted"
// by the specified key.
//
// This function returns the string value of the result but also modifies the
// input array, which can be used to re-use the resulting string.
func UnwrapString(key, data []byte) string {
	if len(key) == 0 || len(data) == 0 {
		return ""
	}
	subtle.XorOp(data, key)
	return string(data)
}

// NewAes attempts to create a new AES block cipher from the provided key data.
// Errors will be returned if the key length is invalid.
func NewAes(k []byte) (cipher.Block, error) {
	return aes.NewCipher(k)
}
func (r *reader) Read(b []byte) (int, error) {
	return r.c.Read(r.r, b)
}
func (w *writer) Write(b []byte) (int, error) {
	return w.c.Write(w.w, b)
}

// NewXORReader creates an io.WriteCloser type from the specified XOR cipher and
// Reader.
//
// This creates a Block cipher with a auto-generated IV based on the key value.
// To control the IV value, use the 'NewCBKReader' function instead.
func NewXORReader(x XOR, r io.Reader) io.Reader {
	v := make([]byte, len(x))
	for i := range x {
		v[i] = (x[i] + byte(i)) ^ 2
	}
	return nopCloser{&cipher.StreamReader{R: r, S: cipher.NewCTR(x, v)}}
}

// NewCBKReader creates an io.ReadCloser type from the specified CBK cipher and
// Reader.
func NewCBKReader(c CBK, r io.Reader) io.Reader {
	return &reader{c: &c, r: r}
}

// NewXORWriter creates an io.WriteCloser type from the specified XOR cipher and
// Writer.
//
// This creates a Block cipher with a auto-generated IV based on the key value.
// To control the IV value, use the 'NewBlockWriter' function instead.
func NewXORWriter(x XOR, w io.Writer) io.WriteCloser {
	v := make([]byte, len(x))
	for i := range x {
		v[i] = (x[i] + byte(i)) ^ 2
	}
	// return &cipher.StreamWriter{W: w, S: cipher.NewCFBEncrypter(x, v)}
	return &cipher.StreamWriter{W: w, S: cipher.NewCTR(x, v)}
}

// NewCBKWriter creates an io.WriteCloser type from the specified CBK cipher and
// Writer.
func NewCBKWriter(c CBK, w io.Writer) io.WriteCloser {
	return &writer{c: &c, w: w}
}

// NewBlockReader creates a data.Reader type from the specified block cipher,
// IV and Reader.
//
// This is used to Decrypt data. This function returns an error if the blocksize
// of the Block does not equal the length of the supplied IV.
//
// This uses CFB mode.
func NewBlockReader(b cipher.Block, iv []byte, r io.Reader) (io.ReadCloser, error) {
	if len(iv) != b.BlockSize() {
		return nil, xerr.Sub("block size must equal IV size", 0x29)
	}
	return nopCloser{&cipher.StreamReader{R: r, S: cipher.NewCTR(b, iv)}}, nil
}

// NewBlockWriter creates a data.Reader type from the specified block cipher,
// IV and Writer.
//
// This is used to Encrypt data. This function returns an error if the blocksize
// of the Block does not equal the length of the supplied IV.
//
// This uses CFB mode.
func NewBlockWriter(b cipher.Block, iv []byte, w io.Writer) (io.WriteCloser, error) {
	if len(iv) != b.BlockSize() {
		return nil, xerr.Sub("block size must equal IV size", 0x29)
	}
	return &cipher.StreamWriter{W: w, S: cipher.NewCTR(b, iv)}, nil
}
