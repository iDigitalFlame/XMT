//go:build !386 && !arm && !mips && !mipsle

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

package data

// MaxSlice is the max slice value used when creating slices to prevent OOM
// issues. XMT will refuse to  make a slice any larger than this and will return
// 'ErrToLarge'
const MaxSlice = 4_398_046_511_104 // (2 << 41)
