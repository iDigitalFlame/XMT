//go:build !js
// +build !js

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

package cmd

type processList []ProcessInfo

func (p processList) Len() int {
	return len(p)
}
func (p processList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p processList) Less(i, j int) bool {
	if p[i].PPID == p[j].PPID {
		return p[i].PID < p[j].PID
	}
	return p[i].PPID < p[j].PPID
}
