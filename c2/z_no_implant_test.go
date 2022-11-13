//go:build !implant

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

package c2

import (
	"encoding/json"
	"testing"

	"github.com/iDigitalFlame/xmt/device/local"
)

func TestMarshal(t *testing.T) {
	b, err := (&Job{}).MarshalJSON()
	if err != nil {
		t.Fatalf(`TestMarshal(): Job.MarshalJSON() returned an error: %s!`, err.Error())
	}
	if len(b) == 0 {
		t.Fatalf(`TestMarshal(): Job.MarshalJSON() returned an empty byte slice!`)
	}
	var v any
	if err = json.Unmarshal(b, &v); err != nil {
		t.Fatalf(`TestMarshal(): Unmarshal() Job returned an error: %s!`, err.Error())
	}
	s := NewServer(nil)
	b, err = s.MarshalJSON()
	if err != nil {
		t.Fatalf(`TestMarshal(): Server.MarshalJSON() returned an error: %s!`, err.Error())
	}
	if len(b) == 0 {
		t.Fatalf(`TestMarshal(): Server.MarshalJSON() returned an empty byte slice!`)
	}
	if err = json.Unmarshal(b, &v); err != nil {
		t.Fatalf(`TestMarshal(): Unmarshal() Server returned an error: %s!`, err.Error())
	}
	l := new(Listener)
	l.s = s
	b, err = l.MarshalJSON()
	if err != nil {
		t.Fatalf(`TestMarshal(): Listener.MarshalJSON() returned an error: %s!`, err.Error())
	}
	if len(b) == 0 {
		t.Fatalf(`TestMarshal(): Listener.MarshalJSON() returned an empty byte slice!`)
	}
	if err = json.Unmarshal(b, &v); err != nil {
		t.Fatalf(`TestMarshal(): Unmarshal() Listener returned an error: %s!`, err.Error())
	}
	b, err = (&Session{parent: l, Device: local.Device.Machine}).MarshalJSON()
	if err != nil {
		t.Fatalf(`TestMarshal(): Session.MarshalJSON() returned an error: %s!`, err.Error())
	}
	if len(b) == 0 {
		t.Fatalf(`TestMarshal(): Session.MarshalJSON() returned an empty byte slice!`)
	}
	if err = json.Unmarshal(b, &v); err != nil {
		t.Fatalf(`TestMarshal(): Unmarshal() Session returned an error: %s!`, err.Error())
	}
}
