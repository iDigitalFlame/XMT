//go:build nokeyset
// +build nokeyset

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

package data

// Crypt will perform an "encryption" operation on the underlying Chunk buffer.
// No bytes are added or removed and this will not change the Chunk's size.
//
// If the Chunk is empty, 'nokeyset' was specified on build or the Key is nil,
// this is a NOP.
func (*Chunk) KeyCrypt(_ KeyPair) {}
