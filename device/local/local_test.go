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

package local

import "testing"

func TestLocal(t *testing.T) {
	if Device.ID.Empty() {
		t.Fatalf("Local Device ID should not be empty!")
	}
	if len(Device.Hostname) == 0 {
		t.Fatalf("Local Device Hostname should not be empty!")
	}
	if len(Device.User) == 0 {
		t.Fatalf("Local Device User should not be empty!")
	}
	if len(Device.Version) == 0 {
		t.Fatalf("Local Device Version should not be empty!")
	}
	if Device.PID == 0 {
		t.Fatalf("Local Device PID should not be zero!")
	}
	if Device.PPID == 0 {
		t.Fatalf("Local Device PPID should not be zero!")
	}
	for i := range Device.Network {
		if Device.Network[i].Mac == 0 {
			t.Fatalf("Local Device Interface %s MAC address should not be empty!", Device.Network[i].Name)
		}
		for x := range Device.Network[i].Address {
			if Device.Network[i].Address[x].IsUnspecified() {
				t.Fatalf("Local Device Interface %s IP address %d should not be zero!", Device.Network[i].Name, x)
			}
		}
	}
}
