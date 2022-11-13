// Copyright (C) 2020 - 2022 iDigitalFlame
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
	"bytes"
	"crypto/aes"
	"io"
	"testing"

	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util"
)

type wrapper interface {
	Unwrap(io.Reader) (io.Reader, error)
	Wrap(io.WriteCloser) (io.WriteCloser, error)
}

func TestWrapHex(t *testing.T) {
	testWrapper(t, Hex)
}
func TestWrapCBK(t *testing.T) {
	c := NewCBK(uint8(util.FastRand()), uint8(util.FastRand()), uint8(util.FastRand()), uint8(util.FastRand()), 128)
	testWrapper(t, c)
}
func TestWrapAES(t *testing.T) {
	b := make([]byte, 32)
	util.Rand.Read(b)
	c, err := aes.NewCipher(b)
	if err != nil {
		t.Fatalf("TestWrapAES(): NewCipher failed with error: %s!", err.Error())
	}
	i := make([]byte, 16)
	util.Rand.Read(i)
	x, err := NewBlock(c, i)
	if err != nil {
		t.Fatalf("TestWrapAES(): Block failed with error: %s!", err.Error())
	}
	testWrapper(t, x)
}
func TestWrapXOR(t *testing.T) {
	b := make([]byte, 64)
	util.Rand.Read(b)
	x := NewXOR(b)
	testWrapper(t, x)
}
func TestWrapGzip(t *testing.T) {
	testWrapper(t, Gzip)
}
func TestWrapZlib(t *testing.T) {
	testWrapper(t, Zlib)
}
func TestWrapBase64(t *testing.T) {
	testWrapper(t, Base64)
}
func testWrapper(t *testing.T, x wrapper) {
	var b bytes.Buffer
	w, err := x.Wrap(data.WriteCloser(&b))
	if err != nil {
		t.Fatalf("TestWrapper(): Wrap failed with error: %s!", err.Error())
	}
	if _, err = w.Write([]byte("hello world!")); err != nil {
		t.Fatalf("TestWrapper(): Write failed with error: %s!", err.Error())
	}
	w.Close()
	r, err := x.Unwrap(bytes.NewReader(b.Bytes()))
	if err != nil {
		t.Fatalf("Unwrap failed with error: %s!", err.Error())
	}
	o := make([]byte, 12)
	if _, err = r.Read(o); err != nil && err != io.EOF {
		t.Fatalf("TestWrapper(): Read failed with error: %s!", err.Error())
	}
	if string(o) != "hello world!" {
		t.Fatalf(`TestWrapper(): Result output "%s" did not match "hello world!"!`, o)
	}
}
