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

package com

import (
	"bytes"
	"testing"

	"github.com/iDigitalFlame/xmt/device/local"
)

func TestPacket(t *testing.T) {
	b := Packet{
		ID:     120,
		Job:    29000,
		Device: local.Device.ID,
	}
	b.WriteInt32(0xFF)
	b.WriteFloat32(1.45)
	b.WriteInt8(120)
	b.WriteString("derp123")
	b.WriteUint64(0xFF00FF00FF123)
	b.WriteUint16Pos(0, 0xFF)
	var w bytes.Buffer
	if err := b.Marshal(&w); err != nil {
		t.Fatalf("TestPacket(): Marshal failed with error: %s!", err.Error())
	}
	var r Packet
	if err := r.Unmarshal(bytes.NewReader(w.Bytes())); err != nil {
		t.Fatalf("TestPacket(): Unmarshal failed with error: %s!", err.Error())
	}
	v, err := r.Int32()
	if err != nil {
		t.Fatalf("TestPacket(): Int32 failed with error: %s!", err.Error())
	}
	if v != 0xFF00FF {
		t.Fatalf(`TestPacket(): Int32 result "0x%X" does not match the expected value "0xFF00FF"!`, v)
	}
	f, err := r.Float32()
	if err != nil {
		t.Fatalf("TestPacket(): Float32 failed with error: %s!", err.Error())
	}
	if f != 1.45 {
		t.Fatalf(`TestPacket(): Float32 result "%.2f" does not match the expected value "1.45"!`, f)
	}
	n, err := r.Int8()
	if err != nil {
		t.Fatalf("TestPacket(): Int8 failed with error: %s!", err.Error())
	}
	if n != 120 {
		t.Fatalf(`TestPacket(): Int8 result "%d" does not match the expected value "120"!`, n)
	}
	s, err := r.StringVal()
	if err != nil {
		t.Fatalf("TestPacket(): StringVal failed with error: %s!", err.Error())
	}
	if s != "derp123" {
		t.Fatalf(`TestPacket(): StringVal result "%s" does not match the expected value 120!`, s)
	}
	u, err := r.Uint64()
	if err != nil {
		t.Fatalf("TestPacket(): Uint64 failed with error: %s!", err.Error())
	}
	if u != 0xFF00FF00FF123 {
		t.Fatalf(`TestPacket(): Uint64 result "0x%X" does not match the expected value "0xFF00FF00FF123"!`, u)
	}
}
