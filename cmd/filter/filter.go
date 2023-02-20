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

// Package filter is a separate container for the 'Filter' struct that can be used
// to target a specific process or one that matches an attribute set.
package filter

import (
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const (
	// True is the 'true' bool value.
	True = boolean(2)
	// False is the 'false' bool value.
	False = boolean(1)
	// Empty represents the absence of a value.
	Empty = boolean(0)
)

var (
	// ErrNoProcessFound is returned by the SetParent* functions on Windows
	// devices when a specified parent could not be found.
	ErrNoProcessFound = xerr.Sub("could not find a suitable process", 0x3E)

	// Any will attempt to locate a parent process that may be elevated
	// based on the current process permissions.
	//
	// This one will fall back to non-elevated if all checks fail.
	Any = (&Filter{Fallback: true}).SetElevated(true)
	// Random is a Filter that can be used by default to select ANY random
	// process on the target device to be used as the parent process without
	// creating a new Filter struct.
	Random = &Filter{Fallback: false}
)

// Filter is a struct that can be used to set the Parent process for many types
// of 'Runnable' compatible interfaces.
//
// Each option can be set directly or chained using the function calls which all
// return the struct for chain usage.
//
// This struct can be serialized into JSON or written using a Stream Marshaler.
type Filter struct {
	// Exclude and Include determine the processes that can be included or omitted
	// during process listing. 'Exclude' always takes precedence over 'Include'.
	//
	// Either one being nil or empty means no processes are included/excluded.
	// All matches are case-insensitive.
	Exclude []string
	Include []string
	// PID will attempt to select the PID to be used for the parent.
	// If set to zero, it will be ignored. Values less than 5 are not valid!
	PID uint32
	// Fallback specifies if the opts routine should try again with less constraints
	// than the previous attempt. All attempts will still respect the 'Exclude'
	// and 'Ignore' directives.
	Fallback bool
	// Session can be set to 'True' or 'False' to attempt to target processes that
	// are either in or not in a DWM session environment (ie: in a user desktop
	// [True] or a service context [False]). This value is ignored if set to 'Empty'.
	Session boolean
	// Elevated can be set 'True' or 'False' to attempt to target processes that
	// are in a High/System or Lower integrity context. 'True' will attempt to
	// select elevated processes, while 'False' will select lower integrity or
	// non-elevated processes. If set to 'Empty' or omitted, this will choose
	// any process, regardless of integrity level.
	Elevated boolean
}
type boolean uint8
type filter func(uint32, bool, string, uintptr) bool

// F is a shortcut for 'new(Filter)'
func F() *Filter {
	return new(Filter)
}

// B is a shortcut for '&Filter{Fallback: f}'
func B(f bool) *Filter {
	return &Filter{Fallback: f}
}

// I is a shortcut for '&Filter{Include: s}'
func I(s ...string) *Filter {
	return &Filter{Include: s}
}

// E is a shortcut for '&Filter{Exclude: s}'
func E(s ...string) *Filter {
	return &Filter{Exclude: s}
}

// Empty will return true if this Filter is nil or unset.
func (f *Filter) Empty() bool {
	return f == nil || f.isEmpty()
}
func (f Filter) isEmpty() bool {
	return f.PID == 0 && f.Session == Empty && f.Elevated == Empty && len(f.Exclude) == 0 && len(f.Include) == 0
}

// Clear clears the Filter settings, except for 'Fallback' and returns itself.
func (f *Filter) Clear() *Filter {
	f.PID, f.Session, f.Elevated, f.Exclude, f.Include = 0, Empty, Empty, nil, nil
	return f
}

// SetPID sets the target PID and returns the Filter struct.
func (f *Filter) SetPID(p uint32) *Filter {
	f.PID = p
	return f
}

// SetSession sets the Session setting to 'True' or 'False' and returns itself.
func (f *Filter) SetSession(s bool) *Filter {
	if s {
		f.Session = True
	} else {
		f.Session = False
	}
	return f
}

// SetElevated sets the Elevated setting to 'True' or 'False' and returns itself.
func (f *Filter) SetElevated(e bool) *Filter {
	if e {
		f.Elevated = True
	} else {
		f.Elevated = False
	}
	return f
}

// SetFallback sets the Fallback setting and returns itself.
func (f *Filter) SetFallback(i bool) *Filter {
	f.Fallback = i
	return f
}

// SetInclude sets the Inclusion list and returns itself.
func (f *Filter) SetInclude(n ...string) *Filter {
	f.Include = n
	return f
}

// SetExclude sets the Exclusion list and returns itself.
func (f *Filter) SetExclude(n ...string) *Filter {
	f.Exclude = n
	return f
}

// MarshalStream will attempt to write the Filter data to the supplied Writer
// and return any errors that may occur.
func (f *Filter) MarshalStream(w data.Writer) error {
	if f == nil || f.isEmpty() {
		return w.WriteBool(false)
	}
	if err := w.WriteBool(true); err != nil {
		return err
	}
	if err := w.WriteUint32(f.PID); err != nil {
		return err
	}
	if err := w.WriteBool(f.Fallback); err != nil {
		return err
	}
	if err := w.WriteUint8(uint8(f.Session)); err != nil {
		return err
	}
	if err := w.WriteUint8(uint8(f.Elevated)); err != nil {
		return err
	}
	if err := data.WriteStringList(w, f.Exclude); err != nil {
		return err
	}
	return data.WriteStringList(w, f.Include)
}

// UnmarshalStream will attempt to read the Filter data from the supplied Reader
// and return any errors that may occur.
func (f *Filter) UnmarshalStream(r data.Reader) error {
	v, err := r.Bool()
	if err != nil {
		return err
	}
	if !v {
		return nil
	}
	if f == nil {
		f = new(Filter)
	}
	return f.unmarshalStream(r)
}
func (f *Filter) unmarshalStream(r data.Reader) error {
	if err := r.ReadUint32(&f.PID); err != nil {
		return err
	}
	if err := r.ReadBool(&f.Fallback); err != nil {
		return err
	}
	if err := r.ReadUint8((*uint8)(&f.Session)); err != nil {
		return err
	}
	if err := r.ReadUint8((*uint8)(&f.Elevated)); err != nil {
		return err
	}
	if err := data.ReadStringList(r, &f.Exclude); err != nil {
		return err
	}
	return data.ReadStringList(r, &f.Include)
}

// UnmarshalStream will attempt to read the Filter data from the supplied Reader
// and return any errors that may occur.
//
// This function takes a pointer of the Filter pointer so it can fill in any
// nil or empty Filters with data for a new Filter struct.
func UnmarshalStream(r data.Reader, f **Filter) error {
	v, err := r.Bool()
	if err != nil {
		return err
	}
	if !v {
		return nil
	}
	if *f == nil {
		*f = new(Filter)
	}
	return (*f).unmarshalStream(r)
}
