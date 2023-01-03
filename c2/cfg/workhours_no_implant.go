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

package cfg

// Day bitmask constants
const (
	DaySunday uint8 = 1 << iota
	DayMonday
	DayTuesday
	DayWednesday
	DayThursday
	DayFriday
	DaySaturday
	DayEveryday = 0
)

// String returns the string version of the WorkHours.Day value. This can be
// parsed back into the numerical version.
func (w WorkHours) String() string {
	return dayNumToString(w.Days)
}
func dayNumToString(d uint8) string {
	if d == 0 || d > 126 {
		return "SMTWRFS"
	}
	var (
		b [7]byte
		n int
	)
	if d&DaySunday != 0 {
		b[n] = 'S'
		n++
	}
	if d&DayMonday != 0 {
		b[n] = 'M'
		n++
	}
	if d&DayTuesday != 0 {
		b[n] = 'T'
		n++
	}
	if d&DayWednesday != 0 {
		b[n] = 'W'
		n++
	}
	if d&DayThursday != 0 {
		b[n] = 'R'
		n++
	}
	if d&DayFriday != 0 {
		b[n] = 'F'
		n++
	}
	if d&DaySaturday != 0 {
		b[n] = 'S'
		n++
	}
	return string(b[:n])
}
