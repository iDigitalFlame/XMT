//go:build small

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

package limits

// Frag is the max size used to fragment packets into.
// Any packet over this byte size will be fragmented.
const Frag = 1_048_576

// Packets determines how many Packets may be processed by the Session thread before
// waiting another wait cycle. If this is set to anything less than one, only a
// single Packet will be processed at a time.
const Packets = 64
