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

import (
	"time"

	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// WorkHours is a struct that can be used to indicate to a Session when it should
// and shouldn't operate. When a WorkHours struct is set, a Session will check it
// before operating and will wait until the next set timeslot (depending on WorkHours
// values).
//
// This struct is also compatible with the Setting interface and can be used
// directly in Pack or Build functions.
//
// Using this in a Pack or Build function when it is empty will clear an other
// previous and/or set WorkHours values.
type WorkHours struct {
	// Days is a bitmask of the days that this WorkHours struct allows us to work
	// on. The bit values are 0 (Sunday) to 7 (Saturday). Values 0, 255 and anything
	// over 126 are treated as all days selected.
	Days uint8
	// StartHour is the 0-23 value of the hour that this WorkHours struct allows
	// us to work on. If this value is greater than 23, an error will be returned
	// when attempting to use it.
	//
	// If this value and StartMin are zero, the WorkHours will start after midnight
	// of the next day (if Day or EndHour and EndMin are set).
	StartHour uint8
	// StartMin is the 0-59 value of the minute that this WorkHours struct allows
	// us to work on. If this value is greater than 59, an error will be returned
	// when attempting to use it.
	//
	// If this value and StartHour are zero, the WorkHours will start after midnight
	// of the next day (if Day or EndHour and EndMin are set).
	StartMin uint8
	// EndHour is the 0-23 value of the hour that this WorkHours struct stops us
	// from working. If this value is greater than 23, an error will be returned
	// when attempting to use it.
	//
	// If this value and EndMin are zero, the WorkHours will continue unchanged.
	EndHour uint8
	// EndMin is the 0-59 value of the minute that this WorkHours struct stops us
	// from working. If this value is greater than 59, an error will be returned
	// when attempting to use it.
	//
	// If this value and EndHour are zero, the WorkHours will continue unchanged.
	EndMin uint8
}

func (w WorkHours) id() cBit {
	return valWorkHours
}

// Empty returns true if this WorkHours struct is considered empty as nothing
// is set and all values are zero, false otherwise.
func (w WorkHours) Empty() bool {
	return w.StartHour == 0 && w.StartMin == 0 && w.EndHour == 0 && w.EndMin == 0 && (w.Days == 0 || w.Days > 126)
}
func (w WorkHours) args() []byte {
	return []byte{byte(valWorkHours), w.Days, w.StartHour, w.StartMin, w.EndHour, w.EndMin}
}

// Verify checks the values set in this WorkHours struct and returns any errors
// due to the number/time values being invalid.
func (w WorkHours) Verify() error {
	switch {
	case w.EndMin > 59:
		return xerr.Sub("invalid EndMin value", 0x73)
	case w.EndHour > 23:
		return xerr.Sub("invalid EndHour value", 0x72)
	case w.StartMin > 59:
		return xerr.Sub("invalid StartMin value", 0x71)
	case w.StartHour > 23:
		return xerr.Sub("invalid StartHour value", 0x70)
	}
	return nil
}

// Work returns the time that should be waitied for this WorkHours to be active.
// If zero, then this means that the WorkHours applies currently and work can be
// done.
//
// This function will not return no more than time to reach the next day.
func (w WorkHours) Work() time.Duration {
	if (w.Days == 0 || w.Days > 126) && w.StartHour == 0 && w.StartMin == 0 && w.EndHour == 0 && w.EndMin == 0 {
		return 0
	}
	n := time.Now()
	// Bit-shit to see if we have that day enabled.
	// Fastpath check if we need to check days at all.
	if w.Days > 0 && w.Days < 127 && (w.Days&(1<<uint(n.Weekday()))) == 0 {
		// Figure out how much time until the next day.
		y, m, d := n.Date()
		return time.Date(y, m, d+1, 0, 0, 0, 0, n.Location()).Sub(n)
	}
	// End == 0 is valid
	// - This can be used to lazy-mans end at midnight as the next check for day
	//   will tell it to wait (so next day).
	//
	// End == 0 && Days == 0 is also valid
	// - This means don't start UNTIL after Start for the given day.
	if w.StartHour == 0 && w.StartMin == 0 && w.EndHour == 0 && w.EndMin == 0 {
		return 0
	}
	var (
		y, m, d = n.Date()
		l       = n.Location()
		s       time.Time
	)
	if (w.StartHour == 0 && w.StartMin == 0) || w.StartHour > 23 || w.StartMin > 60 {
		// Set start to today at zero
		s = time.Date(y, m, d, 0, 0, 0, 0, l)
	} else {
		if s = time.Date(y, m, d, int(w.StartHour), int(w.StartMin), 0, 0, l); s.After(n) { // Wait until start time
			return s.Sub(n)
		}
	}
	if (w.EndHour == 0 && w.EndMin == 0) || w.EndHour > 23 || w.EndMin > 60 {
		return 0
	}
	e := time.Date(y, m, d, int(w.EndHour), int(w.EndMin), 0, 0, l)
	if e.Before(s) { // End is before start, bail.
		return 0
	}
	if n.After(e) { // Wait until start for next
		return s.AddDate(0, 0, 1).Sub(n)
	}
	return 0
}

// MarshalStream writes the data for this WorkHours struct to the supplied Writer.
func (w WorkHours) MarshalStream(s data.Writer) error {
	if err := s.WriteUint8(w.Days); err != nil {
		return err
	}
	if err := s.WriteUint8(w.StartHour); err != nil {
		return err
	}
	if err := s.WriteUint8(w.StartMin); err != nil {
		return err
	}
	if err := s.WriteUint8(w.EndHour); err != nil {
		return err
	}
	return s.WriteUint8(w.EndMin)
}

// UnmarshalStream transforms this struct from a binary format that is read from
func (w *WorkHours) UnmarshalStream(r data.Reader) error {
	if err := r.ReadUint8(&w.Days); err != nil {
		return err
	}
	if err := r.ReadUint8(&w.StartHour); err != nil {
		return err
	}
	if err := r.ReadUint8(&w.StartMin); err != nil {
		return err
	}
	if err := r.ReadUint8(&w.EndHour); err != nil {
		return err
	}
	return r.ReadUint8(&w.EndMin)
}
