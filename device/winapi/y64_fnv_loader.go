//go:build windows && (altload || crypt) && (arm64 || amd64)

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

package winapi

type imageOptionalHeader struct {
	Magic               uint16
	_, _                uint8
	SizeOfCode          uint32
	_, _                uint32
	AddressOfEntryPoint uint32
	BaseOfCode          uint32
	ImageBase           uint64
	_, _                uint32
	_, _, _, _, _, _    uint16
	_                   uint32
	SizeOfImage         uint32
	SizeOfHeaders       uint32
	_                   uint32
	Subsystem           uint16
	DllCharacteristics  uint16
	_, _, _, _          uint64
	LoaderFlags         uint32
	_                   uint32
	Directory           [16]imageDataDirectory
}
