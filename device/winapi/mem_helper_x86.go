//go:build windows && cgo && freemem && (386 || arm || mips || mipsle)
// +build windows
// +build cgo
// +build freemem
// +build 386 arm mips mipsle

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

package winapi

type memoryBasicInfo struct {
	// DO NOT REORDER
	BaseAddress       uint32
	AllocationBase    uint32
	AllocationProtect uint32
	RegionSize        uint32
	State             uint32
	Protect           uint32
	Type              uint32
}
