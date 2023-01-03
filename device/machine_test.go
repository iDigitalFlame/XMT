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

package device

import (
	"testing"

	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device/arch"
)

// marshal a machine here

func TestMachine(t *testing.T) {
	m := &Machine{
		User:     "Test123",
		Version:  "Some Linux OS",
		Hostname: "Test Machine",
		PID:      123,
		PPID:     4567,
		Network: Network{
			device{
				Name: "my_device",
				Mac:  0xABCEEF012,
				Address: []Address{
					Address{hi: 0xFFFF, low: 0xABCDE},
				},
			},
		},
		Capabilities: 0xABCD,
		System:       uint8(Windows)<<4 | uint8(arch.ARM64),
		Elevated:     129,
	}
	m.ID[0] = 'A'
	m.ID[MachineIDSize] = 'A'
	m.ID.Seed([]byte("ABCD"))

	var c data.Chunk
	if err := m.MarshalStream(&c); err != nil {
		t.Fatalf("TestMachine(): MarshalStream returned an error: %s", err.Error())
	}

	c.Seek(0, 0)

	var g Machine
	if err := g.UnmarshalStream(&c); err != nil {
		t.Fatalf("TestMachine(): UnmarshalStream returned an error: %s", err.Error())
	}

	if g.ID != m.ID {
		t.Fatalf(`TestMachine(): Machine ID "%s" does not match "%s".`, g.ID.Full(), m.ID.Full())
	}
	if g.ID.Signature() != m.ID.Signature() {
		t.Fatalf(`TestMachine(): Machine ID Signature "%s" does not match "%s".`, g.ID.Signature(), m.ID.Signature())
	}
	if g.User != m.User {
		t.Fatalf(`TestMachine(): Machine User "%s" does not match "%s".`, g.User, m.User)
	}
	if g.Version != m.Version {
		t.Fatalf(`TestMachine(): Machine Version "%s" does not match "%s".`, g.Version, m.Version)
	}
	if g.Hostname != m.Hostname {
		t.Fatalf(`TestMachine(): Machine Hostname "%s" does not match "%s".`, g.Hostname, m.Hostname)
	}
	if g.PID != m.PID {
		t.Fatalf(`TestMachine(): Machine Hostname "%d" does not match "%d".`, g.PID, m.PID)
	}
	if g.PPID != m.PPID {
		t.Fatalf(`TestMachine(): Machine Hostname "%d" does not match "%d".`, g.PPID, m.PPID)
	}
	if g.System != m.System {
		t.Fatalf(`TestMachine(): Machine System "%d" does not match "%d".`, g.System, m.System)
	}
	if g.Network[0].Address[0].String() != m.Network[0].Address[0].String() {
		t.Fatalf(`TestMachine(): Machine Address[0] "%s" does not match "%s".`, g.Network[0].Address[0].String(), m.Network[0].Address[0].String())
	}
}
