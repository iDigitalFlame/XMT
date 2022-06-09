//go:build !bugs

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

package bugtrack

// Enabled is the stats of the bugtrack package.
//
// This is true if bug tracking is enabled.
const Enabled = false

// Recover is a "guard" function to be used to gracefully shutdown a program
// when a panic is detected.
//
// Can be en enabled by using:
//    if bugtrack.Enabled {
//        defer bugtrack.Recover("thread-name")
//    }
//
// The specified name will be entered into the bugtrack log and a stack trace
// will be generated before gracefully execution to the program.
func Recover(_ string) {}

// Track is a simple logging function that takes the same arguments as a
// 'fmt.Sprintf' function. This can be used to track bugs or output values.
//
// Not recommended to be used in production environments.
//
// The "-tags bugs" option is required in order for this function to be used.
func Track(_ string, _ ...any) {}
