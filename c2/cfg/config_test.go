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

package cfg

import (
	"bytes"
	"testing"
)

func TestLen(t *testing.T) {
	if n := Pack(ConnectTCP, WrapGzip).Len(); n != 2 {
		t.Fatalf(`TestLen(): Len should be "2" but got "%d"!`, n)
	}
}
func TestGroup(t *testing.T) {
	if v := (Config{}).Group(0); v != nil {
		t.Fatalf(`TestGroup(): Empty Group should be "nil" but got "%v"!`, v)
	}
	c := Pack(ConnectICMP)
	c.AddGroup(WrapCBK(1, 2, 3, 4), ConnectTCP, Host("test1"))
	c.AddGroup(WrapXOR([]byte("derp-master")))
	c.Add(ConnectPipe)
	if v := c.Group(-1); !bytes.Equal(v, c) {
		t.Fatalf(`TestGroup(): Group(-1) should return itself but got "%v"!`, v)
	}
	if n := c.Group(0).Len(); n != 1 {
		t.Fatalf(`TestGroup(): Group(0) len should be "1" but got "%d"!`, n)
	}
	if n := c.Group(1).Len(); n != 15 {
		t.Fatalf(`TestGroup(): Group(1) len should be "15" but got "%d"!`, n)
	}
	if n := c.Group(2).Len(); n != 15 {
		t.Fatalf(`TestGroup(): Group(2) len should be "15" but got "%d"!`, n)
	}
	if n := c.Group(3).Len(); n != 15 {
		t.Fatalf(`TestGroup(): Group(3) len should be "15" but got "%d"!`, n)
	}
}
func TestBuild(t *testing.T) {
	if _, err := Build(ConnectICMP, WrapBase64, Host("test123")); err != nil {
		t.Fatalf(`TestBuild(): Build should pass but got "%s"!`, err.Error())
	}
	if err := Pack(ConnectICMP, WrapBase64, Host("test123")).Validate(); err != nil {
		t.Fatalf(`TestBuild(): Validate should pass but got "%s"!`, err.Error())
	}
	if _, err := Build(ConnectICMP, ConnectPipe, WrapBase64, Host("test123")); err == nil {
		t.Fatalf("TestBuild(): Invalid Build should not pass!")
	}
}
func TestGroups(t *testing.T) {
	c := Pack(ConnectICMP)
	c.AddGroup(WrapBase64)
	c.AddGroup(WrapZlib, WrapHex)
	c.AddGroup()
	c.Add(ConnectPipe)
	c.Add()
	if n := c.Groups(); n != 3 {
		t.Fatalf(`TestGroups(): Group count should be "3" but got "%d"!`, n)
	}
}
func TestReadWrite(t *testing.T) {
	var b bytes.Buffer
	if err := Write(&b, ConnectICMP, WrapBase64, Host("test123")); err != nil {
		t.Fatalf(`TestReadWrite(): Write failed with error "%s"!`, err.Error())
	}
	if n := b.Len(); n != 12 {
		t.Fatalf(`TestReadWrite(): Written bytes count should be "12" but got "%d"!`, n)
	}
	if _, err := Reader(bytes.NewReader(b.Bytes())); err != nil {
		t.Fatalf(`TestReadWrite(): Reader should pass but got "%s"!`, err.Error())
	}
}
