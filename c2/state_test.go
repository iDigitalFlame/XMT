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

import "testing"

func TestState(t *testing.T) {
	var s state
	if s.Seen() {
		t.Fatalf("TestState(): Seen should return false!")
	}
	if s.Set(stateSeen); !s.Seen() {
		t.Fatalf("TestState(): Seen should return true!")
	}
	if s.SetLast(0xBEEF); s.Last() != 0xBEEF {
		t.Fatalf(`TestState(): Last should return "0xBEEF" not "0x%X"!`, s.Last())
	}
	if s.Moving() {
		t.Fatalf("TestState(): Moving should return false!")
	}
	if s.Set(stateMoving); !s.Moving() {
		t.Fatalf("TestState(): Moving should return true!")
	}
	if s.CanRecv() {
		t.Fatalf("TestState(): Moving CanRecv return false!")
	}
	if s.Set(stateCanRecv); !s.CanRecv() {
		t.Fatalf("TestState(): CanRecv should return true!")
	}
	if s.Set(stateClosed); s.CanRecv() {
		t.Fatalf("TestState(): CanRecv should return false!")
	}
	if !s.Closing() || !s.Shutdown() || !s.SendClosed() || !s.WakeClosed() {
		t.Fatalf("TestState(): Closing should return true!")
	}
	if s.ChannelProxy() {
		t.Fatalf("TestState(): ChannelProxy should return false!")
	}
	if s.Set(stateChannelProxy); !s.ChannelProxy() {
		t.Fatalf("TestState(): ChannelProxy should return true!")
	}
	if s.ChannelUpdated() {
		t.Fatalf("TestState(): ChannelUpdated should return false!")
	}
	if s.Set(stateChannelUpdated); !s.ChannelUpdated() {
		t.Fatalf("TestState(): ChannelUpdated should return true!")
	}
	if !s.SetChannel(true) {
		t.Fatalf("TestState(): SetChannel(true) should return false!")
	}
	if !s.SetChannel(false) {
		t.Fatalf("TestState(): SetChannel(true) should return false!")
	}
	if s.Unset(stateChannelValue); s.SetChannel(false) {
		t.Fatalf("TestState(): SetChannel(true) should return true!")
	}
}
