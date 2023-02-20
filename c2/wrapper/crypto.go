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

package wrapper

import (
	"crypto/cipher"
	"io"

	"github.com/iDigitalFlame/xmt/data/crypto"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// CBK is the crypto.CBK implementation of a Wrapper. This instance will create
// a new CBK instance with these supplied values and size.
type CBK [5]byte

// XOR is the crypto.XOR implementation of a Wrapper. This instance will create
// a XOR key stream with IV based on the supplied values.
type XOR struct {
	_  [0]func()
	k  crypto.XOR
	iv []byte
}

// Block is the cipher.Block implementation of a Wrapper. This instance will create
// a Block based Wrapper with ciphers such as AES.
type Block struct {
	_ [0]func()
	b cipher.Block
	v []byte
}
type nopCloser struct {
	io.Reader
}

// NewXOR is a function that is an alias for 'Stream(crypto.XOR(k), crypto.XOR(k))'
//
// This wil return a Stream-backed XOR Wrapper.
func NewXOR(k []byte) XOR {
	x := XOR{k: k, iv: make([]byte, len(k))}
	for i := range k {
		x.iv[i] = (k[i] + byte(i)) ^ 2
	}
	return x
}
func (nopCloser) Close() error {
	return nil
}

// NewCBK creates a special type of Wrapper for CBK-based encryptors.
//
// NOTE: This function will prevent CBK from using its index based block
// functions, not sure if there's a way to work around this.
func NewCBK(a, b, c, d, size byte) CBK {
	var e CBK
	e[0], e[1], e[2], e[3], e[4] = a, b, c, d, size
	return e
}

// Unwrap fulfils the Wrapper interface.
func (c CBK) Unwrap(r io.Reader) (io.Reader, error) {
	e, err := crypto.NewCBKSource(c[0], c[1], c[2], c[3], c[4])
	if err != nil {
		return nil, err
	}
	return crypto.NewCBKReader(e, r), nil
}

// Unwrap fulfils the Wrapper interface.
func (x XOR) Unwrap(r io.Reader) (io.Reader, error) {
	return nopCloser{&cipher.StreamReader{R: r, S: cipher.NewCFBDecrypter(x.k, x.iv)}}, nil
}

// Unwrap fulfils the Wrapper interface.
func (b Block) Unwrap(r io.Reader) (io.Reader, error) {
	return &cipher.StreamReader{R: r, S: cipher.NewCFBDecrypter(b.b, b.v)}, nil
}

// NewBlock returns a Wrapper based on a Block Cipher, such as AES.
func NewBlock(b cipher.Block, v []byte) (Block, error) {
	if b == nil || len(v) == 0 {
		return Block{}, xerr.Sub("arguments cannot be nil or empty", 0x6E)
	}
	if len(v) != b.BlockSize() {
		return Block{}, xerr.Sub("block size must equal IV size", 0x29)
	}
	return Block{v: v, b: b}, nil
}

// Wrap fulfils the Wrapper interface.
func (c CBK) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	e, err := crypto.NewCBKSource(c[0], c[1], c[2], c[3], c[4])
	if err != nil {
		return nil, err
	}
	return crypto.NewCBKWriter(e, w), nil
}

// Wrap fulfils the Wrapper interface.
func (x XOR) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return &cipher.StreamWriter{W: w, S: cipher.NewCFBEncrypter(x.k, x.iv)}, nil
}

// Wrap fulfils the Wrapper interface.
func (b Block) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return &cipher.StreamWriter{W: w, S: cipher.NewCFBEncrypter(b.b, b.v)}, nil
}
