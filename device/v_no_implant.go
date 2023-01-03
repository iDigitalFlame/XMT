//go:build !implant

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
	"net"

	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// State returns a string representation of the Login's Status type.
func (l Login) State() string {
	switch l.Status {
	case 0:
		return "Active"
	case 1:
		return "Connected"
	case 2:
		return "ConnectedQuery"
	case 3:
		return "Shadow"
	case 4:
		return "Disconnected"
	case 5:
		return "Idle"
	case 6:
		return "Listen"
	case 7:
		return "Reset"
	case 8:
		return "Down"
	case 9:
		return "Init"
	}
	return "Unknown"
}

// String returns a string representation of the OSType.
func (o OSType) String() string {
	switch o {
	case Windows:
		return "Windows"
	case Linux:
		return "Linux"
	case Unix:
		return "Unix/BSD"
	case Mac:
		return "MacOS"
	case IOS:
		return "iOS"
	case Android:
		return "Android"
	case Plan9:
		return "Plan9"
	case Unsupported:
		return "Unsupported"
	}
	return "Unknown"
}
func (d device) String() string {
	b := builders.Get().(*util.Builder)
	b.Grow(30)
	b.WriteString(d.Name + "(" + d.Mac.String() + "): [")
	for i := range d.Address {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(d.Address[i].String())
	}
	b.WriteByte(']')
	r := b.Output()
	builders.Put(b)
	return r
}

// String returns a simple string representation of the Machine instance.
func (m Machine) String() string {
	var e string
	if m.IsElevated() {
		e = "*"
	}
	return "[" + m.ID.String() + "] " + m.Hostname + " (" + m.Version + ") " + e + m.User
}

// MarshalJSON implements the json.Marshaler interface.
func (a Address) MarshalJSON() ([]byte, error) {
	return []byte(`"` + a.String() + `"`), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (a *Address) UnmarshalJSON(b []byte) error {
	if len(b) < 1 || b[len(b)-1] != '"' || b[0] != '"' {
		return xerr.Sub("invalid address value", 0x1E)
	}
	if i := net.ParseIP(string(b[1 : len(b)-2])); i != nil {
		a.Set(i)
	}
	return nil
}
