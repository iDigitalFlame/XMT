//go:build windows && cgo && (amd64 || arm64 || loong64 || mips64 || mips64le || ppc64 || ppc64le || riscv64)
// +build windows
// +build cgo
// +build amd64 arm64 loong64 mips64 mips64le ppc64 ppc64le riscv64

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
	BaseAddress       uintptr
	AllocationBase    uintptr
	AllocationProtect uint32
	Partition         uint16
	RegionSize        uint64
	State             uint32
	Protect           uint32
	Type              uint32
}
