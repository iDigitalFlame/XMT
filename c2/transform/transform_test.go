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

package transform

import (
	"bytes"
	"io"
	"testing"

	"github.com/iDigitalFlame/xmt/util"
)

type transform interface {
	Read([]byte, io.Writer) error
	Write([]byte, io.Writer) error
}

func TestTransformDNS(t *testing.T) {
	testWrapper(t, DNS)
}
func TestTransformBase64(t *testing.T) {
	testWrapper(t, Base64)
}
func TestTransformBase64Shift(t *testing.T) {
	testWrapper(t, B64Shift(int(util.FastRand())))
}
func testWrapper(t *testing.T, x transform) {
	var i, o bytes.Buffer
	if err := x.Write([]byte("hello world!"), &o); err != nil {
		t.Fatalf("Write failed with error: %s!", err.Error())
	}
	if err := x.Read(o.Bytes(), &i); err != nil {
		t.Fatalf("Write failed with error: %s!", err.Error())
	}
	v := make([]byte, 12)
	if _, err := i.Read(v); err != nil && err != io.EOF {
		t.Fatalf("Read failed with error: %s!", err.Error())
	}
	if string(v) != "hello world!" {
		t.Fatalf(`Result output "%s" did not match "hello world!"!`, v)
	}
}
