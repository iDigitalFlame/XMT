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

package com

import (
	"testing"

	"github.com/iDigitalFlame/xmt/data"
)

func TestFlag(t *testing.T) {
	var f Flag
	if n := f.Group(); n != 0 {
		t.Fatalf(`TestFlag(): Group returned "%d" but should be "0"!`, n)
	}
	if f.SetGroup(0xBEEF); f.Group() != 0xBEEF {
		t.Fatalf(`TestFlag(): Group returned "0x%X" but should be "0xBEEF"!`, f.Group())
	}
	if n := f.Position(); n != 0 {
		t.Fatalf(`TestFlag(): Position returned "%d" but should be "0"!`, n)
	}
	if f.SetPosition(0xDEFF); f.Position() != 0xDEFF {
		t.Fatalf(`TestFlag(): Position returned "0x%X" but should be "0xDEFF"!`, f.Position())
	}
	if n := f.Len(); n != 0 {
		t.Fatalf(`TestFlag(): Len returned "%d" but should be "0"!`, n)
	}
	if f.SetLen(0xABCD); f.Len() != 0xABCD {
		t.Fatalf(`TestFlag(): Len returned "0x%X" but should be "0xABCD"!`, f.Len())
	}
	if f&FlagFrag == 0 {
		t.Fatalf("TestFlag(): Frag flag should be set!")
	}
	if f.Clear(); f != 0 {
		t.Fatalf("TestFlag(): Flags should be empty!")
	}
	if f.Set(FlagCrypt); f != 256 {
		t.Fatalf(`TestFlag(): Flag value is "%d" but should be "256"!`, f)
	}
	if f.Unset(FlagError); f != 256 {
		t.Fatalf(`TestFlag(): Flag value is "%d" but should be "256"!`, f)
	}
	var c data.Chunk
	if err := f.MarshalStream(&c); err != nil {
		t.Fatalf("TestFlag(): MarshalStream returned an error: %s!", err.Error())
	}
	c.Seek(0, 0)
	var g Flag
	if err := g.UnmarshalStream(&c); err != nil {
		t.Fatalf("TestFlag(): UnmarshalStream returned an error: %s!", err.Error())
	}
	if f != g {
		t.Fatalf(`TestFlag(): Flag value is "%d" but should be "%d"!`, g, f)
	}
}
