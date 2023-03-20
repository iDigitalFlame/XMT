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

package cfg

import (
	"time"

	"github.com/iDigitalFlame/xmt/data"
)

// Separator is an entry that can be used to create Groups in Config instances.
//
// It is recommended to use the 'AddGroup' functions instead, but this can
// be used to create more advanced Groupings.
const Separator = cBit(0xFA)

const (
	invalid = cBit(0)

	valHost      = cBit(0xA0)
	valSleep     = cBit(0xA1)
	valJitter    = cBit(0xA2)
	valWeight    = cBit(0xA3)
	valKeyPin    = cBit(0xA6)
	valKillDate  = cBit(0xA4)
	valWorkHours = cBit(0xA5)
)

type cBit byte
type cBytes []byte

// Setting is an interface that represents a C2 Profile setting in binary form.
//
// This can be used to generate a C2 Profile from binary data or write a Profile
// to a binary stream or JSON payload (if enabled).
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
	return &cBytes{byte(valWeight), byte(w)}
}

// Jitter returns a Setting that will specify the Jitter setting of the generated
// Profile. Only Jitter values from zero to one-hundred [0-100] are valid.
//
// Other values are ignored and replaced with the default.
func Jitter(n uint) Setting {
	return &cBytes{byte(valJitter), byte(n)}
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
	c := append(cBytes{byte(valHost), byte(n >> 8), byte(n)}, s[:n]...)
	return &c
}
func (c cBytes) args() []byte {
	return c
}

// KillDate returns a Setting that will specify the KillDate setting of the
// generated Profile. Zero values will clear the set value.
func KillDate(t time.Time) Setting {
	if t.IsZero() {
		return &cBytes{byte(valKillDate), 0, 0, 0, 0, 0, 0, 0, 0}
	}
	v := t.Unix()
	return &cBytes{
		byte(valKillDate), byte(v >> 56), byte(v >> 48), byte(v >> 40), byte(v >> 32),
		byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v),
	}
}

// Sleep returns a Setting that will specify the Sleep timeout setting of the
// generated Profile. Values of zero and below are ignored.
func Sleep(t time.Duration) Setting {
	if t <= 0 {
		return nil
	}
	return &cBytes{
		byte(valSleep), byte(t >> 56), byte(t >> 48), byte(t >> 40), byte(t >> 32),
		byte(t >> 24), byte(t >> 16), byte(t >> 8), byte(t),
	}
}

// KeyPin returns a Setting that indicates to the client if the Server's PublicKey
// should be trusted. This Setting can be added multiple times to add multiple
// PublicKeys.
//
// This function takes a trusted PublicKey and hashes it to be matched by the
// client.
func KeyPin(k data.PublicKey) Setting {
	if k.Empty() {
		return nil
	}
	h := k.Hash()
	return &cBytes{byte(valKeyPin), byte(h >> 24), byte(h >> 16), byte(h >> 8), byte(h)}
}
