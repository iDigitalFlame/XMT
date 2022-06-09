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

import "time"

const (
	// Seperator is an entry that can be used to create Groups in Config instances.
	//
	// It is recommended to use the 'AddGroup' functions instead, but this can
	// be used to create more advanced Groupings.
	Seperator = cBit(0xFA)

	invalid = cBit(0)

	valHost   = cBit(0xA0)
	valSleep  = cBit(0xA1)
	valJitter = cBit(0xA2)
	valWeight = cBit(0xA3)
)

type cBit byte
type cBytes []byte

// Setting is an interface represents a C2 Profile setting in binary form.
//
// This can be used inside to generate a C2 Profile from binary data or write
// a Profile to a binary stream or from a JSON payload.
type Setting interface {
	id() cBit
	args() []byte
}

func (c cBit) id() cBit {
	return c
}
func (cBit) args() []byte {
	return nil
}
func (c cBytes) id() cBit {
	if len(c) == 0 {
		return invalid
	}
	return cBit(c[0])
}

// Weight returns a Setting that will specify the Weight of the generated
// Profile. Weight is taken into account when multiple Profiles are included to
// make a multi-profile.
//
// This option MUST be included in the Group to take effect. Not including this
// will set the value to zero (0). Multiple values in a Group will take
// the last value.
func Weight(w uint) Setting {
	if w == 0 {
		return nil
	}
	return cBytes{byte(valWeight), byte(w)}
}

// Jitter returns a Setting that will specify the Jitter setting of the generated
// Profile. Only Jitter values from zero to one-hundred [0-100] are valid.
//
// Other values are ignored and replaced with the default.
func Jitter(n uint) Setting {
	return cBytes{byte(valJitter), byte(n)}
}

// Host will return a Setting that will specify a host setting to the profile.
// If empty, this value is ignored.
//
// This may be included multiple times to add multiple Host entries to be used
// in a single Group entry.
func Host(s string) Setting {
	if len(s) == 0 {
		return nil
	}
	n := len(s)
	if n > 0xFFFF {
		n = 0xFFFF
	}
	return append(cBytes{byte(valHost), byte(n >> 8), byte(n)}, s[:n]...)
}
func (c cBytes) args() []byte {
	return c
}

// Sleep returns a Setting that will specify the Sleep timeout setting of the
// generated Profile. Values of zero and below are ignored.
func Sleep(t time.Duration) Setting {
	if t <= 0 {
		return nil
	}
	return cBytes{
		byte(valSleep), byte(t >> 56), byte(t >> 48), byte(t >> 40), byte(t >> 32),
		byte(t >> 24), byte(t >> 16), byte(t >> 8), byte(t),
	}
}
