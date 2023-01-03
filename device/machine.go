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
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device/arch"
)

// Machine is a struct that contains information about a specific device.
// This struct contains generic Operating System Information such as Version,
// Arch and network information.
type Machine struct {
	User     string
	Version  string
	Hostname string

	Network      Network
	PID, PPID    uint32
	Capabilities uint32

	ID       ID
	System   uint8
	Elevated uint8
}

// OS returns the Machine's OSType value.
// This value is gained by shifting the 'System' value by bits 4 to the right.
func (m Machine) OS() OSType {
	return OSType(m.System >> 4)
}

// IsElevated will return true if the elevated flag is set to true on this
// device's 'Elevated' flags.
func (m Machine) IsElevated() bool {
	return m.Elevated&1 != 0
}

// IsDomainJoined will return true if the domain joined flag is set to true
// on this device's 'Elevated' flags.
func (m Machine) IsDomainJoined() bool {
	return m.OS() == Windows && m.Elevated&128 != 0
}

// Arch returns the Machine's Architecture value.
// This value is gained by masking the OS bits of the 'System' value and returning
// the lower 4 bits.
func (m Machine) Arch() arch.Architecture {
	return arch.Architecture(m.System & 0xF)
}

// MarshalStream transforms this struct into a binary format and writes to the
// supplied data.Writer.
func (m Machine) MarshalStream(w data.Writer) error {
	if err := m.ID.MarshalStream(w); err != nil {
		return err
	}
	if err := w.WriteUint8(m.System); err != nil {
		return err
	}
	if err := w.WriteUint32(m.PID); err != nil {
		return err
	}
	if err := w.WriteUint32(m.PPID); err != nil {
		return err
	}
	if err := w.WriteString(m.User); err != nil {
		return err
	}
	if err := w.WriteString(m.Version); err != nil {
		return err
	}
	if err := w.WriteString(m.Hostname); err != nil {
		return err
	}
	if err := w.WriteUint8(m.Elevated); err != nil {
		return err
	}
	if err := w.WriteUint32(m.Capabilities); err != nil {
		return err
	}
	return m.Network.MarshalStream(w)
}

// UnmarshalStream transforms this struct from a binary format that is read from
// the supplied data.Reader.
func (m *Machine) UnmarshalStream(r data.Reader) error {
	if err := m.ID.UnmarshalStream(r); err != nil {
		return err
	}
	if err := r.ReadUint8(&m.System); err != nil {
		return err
	}
	if err := r.ReadUint32(&m.PID); err != nil {
		return err
	}
	if err := r.ReadUint32(&m.PPID); err != nil {
		return err
	}
	if err := r.ReadString(&m.User); err != nil {
		return err
	}
	if err := r.ReadString(&m.Version); err != nil {
		return err
	}
	if err := r.ReadString(&m.Hostname); err != nil {
		return err
	}
	if err := r.ReadUint8(&m.Elevated); err != nil {
		return err
	}
	if err := r.ReadUint32(&m.Capabilities); err != nil {
		return err
	}
	return m.Network.UnmarshalStream(r)
}
