//go:build windows && (arm64 || amd64) && go1.14 && !go1.15
// +build windows
// +build arm64 amd64
// +build go1.14
// +build !go1.15

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

// M constant offsets
// Golang 1.14 (x64)
const (
	ptrNext   = 288
	ptrThread = 792
)
