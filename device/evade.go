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

package device

const (
	// EvadeWinPatchTrace is an evasion flag that instructs the client to patch
	// ETW tracing functions.
	EvadeWinPatchTrace uint8 = 1 << iota
	// EvadeWinPatchAmsi is an evasion flag that instructs the client to patch
	// Amsi detection functions.
	EvadeWinPatchAmsi
	// EvadeWinHideThreads is an evasion flag that instructs the client to hide
	// all of it's current threads from debuggers.
	EvadeWinHideThreads
	// EvadeAll does exactly what it says, enables ALL Evasion functions.
	EvadeAll uint8 = 0xFF
)
