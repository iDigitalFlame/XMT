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

// Package registry contains code to handle common Windows registry operations.
//
// Optimized copy from sys/windows/registry to work with Crypt.
//
package registry

import "github.com/iDigitalFlame/xmt/util/xerr"

// Registry value types.
const (
	TypeString       = 1
	TypeExpandString = 2
	TypeBinary       = 3
	TypeDword        = 4
	TypeStringList   = 7
	TypeQword        = 11
)

var (
	// ErrUnexpectedSize is returned when the key data size was unexpected.
	ErrUnexpectedSize = xerr.Sub("unexpected key size", 0x15)
	// ErrUnexpectedType is returned by Get*Value when the value's type was
	// unexpected.
	ErrUnexpectedType = xerr.Sub("unexpected key type", 0x16)
)
