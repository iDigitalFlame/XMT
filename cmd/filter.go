package cmd

import (
	"encoding/json"

	"github.com/iDigitalFlame/xmt/data"
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
	// AnyParent will attempt to locate a parent process that may be elevated
	// based on the current process permissions.
	//
	// This one will fallback to non-elevated if all checks fail.
	AnyParent = (&Filter{Fallback: true}).SetElevated(true)
	// RandomParent is a Filter that can be used by default to select ANY random
	// process on the target device to be used as the parent process without
	// creating a new Filter struct.
	RandomParent = &Filter{Fallback: false}
)

// Filter is a struct that can be used to set the Parent process for many types of
// 'Runnable' compatable interfaces.
//
// Each option can be set directly or chained using the function calls which all return
// the struct for chain usage.
//
// This struct can be serialized into JSON or written using a Stream Marshaler.
type Filter struct {
	// Exclude and Include determine the processes that can be included or omitted during
	// process listing. 'Exclude' always takes precedence over 'Include'.
	//
	// Ether one being nil or empty means no processes are included/excluded.
	// All matches are case-insensitive.
	Exclude []string `json:"exclude,omitempty"`
	Include []string `json:"include,omitempty"`
	// PID will attempt to select the PID to be used for the parent.
	// If set to zero, it will be ignored. Values less than 5 are not valid!
	PID uint32 `json:"pid,omitempty"`
	// Fallback specifies if the opts routine should try again with less constaints
	// than the previous attempt. All attempts will still respect the 'Exclude' and
	// 'Ignore' directives.
	Fallback bool `json:"fallback,omitempty"`
	// Session can be set to 'True' or 'False' to attempt to target processes that
	// are either in or not in a DWM session environment (ie: in a user deskop [True]
	// or a service context [False]). This value is ignored if set to 'Empty'.
	Session boolean `json:"session,omitempty"`
	// Elevated can be set 'True' or 'False' to attempt to target processes that are
	// in a High/System or Lower integrity context. 'True' will attempt to select
	// elevated processes, while 'False' will select lower integrity or non-elevated
	// processes. If set to 'Empty' or omitted, this will be set based on the current
	// process's integrity level (ie: 'True' if device.Elevated == true else 'False').
	Elevated boolean `json:"elevated,omitempty"`
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

// MarshalJSON will attempt to convert the data in this Filter into the
// returned JSON byte array.
func (f Filter) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{"fallback": f.Fallback}
	if f.PID != 0 {
		m["pid"] = f.PID
	}
	if f.Session > Empty {
		m["session"] = f.Session
	}
	if f.Elevated > Empty {
		m["elevated"] = f.Elevated
	}
	if len(f.Exclude) > 0 {
		m["exclude"] = f.Elevated
	}
	if len(f.Include) > 0 {
		m["include"] = f.Include
	}
	return json.Marshal(m)
}
func (b boolean) MarshalJSON() ([]byte, error) {
	switch b {
	case True:
		return []byte(`"true"`), nil
	case False:
		return []byte(`"false"`), nil
	default:
	}
	return []byte(`""`), nil
}

// UnmarshalJSON will attempt to parse the supplied JSON into this Filter.
func (f *Filter) UnmarshalJSON(b []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	if len(m) == 0 {
		return nil
	}
	if v, ok := m["pid"]; ok {
		if err := json.Unmarshal(v, &f.PID); err != nil {
			return err
		}
	}
	if v, ok := m["session"]; ok {
		if err := json.Unmarshal(v, &f.Session); err != nil {
			return err
		}
	}
	if v, ok := m["elevated"]; ok {
		if err := json.Unmarshal(v, &f.Elevated); err != nil {
			return err
		}
	}
	if v, ok := m["exclude"]; ok {
		if err := json.Unmarshal(v, &f.Exclude); err != nil {
			return err
		}
	}
	if v, ok := m["include"]; ok {
		if err := json.Unmarshal(v, &f.Include); err != nil {
			return err
		}
	}
	if v, ok := m["fallback"]; ok {
		if err := json.Unmarshal(v, &f.Fallback); err != nil {
			return err
		}
	}
	return nil
}
func (b *boolean) UnmarshalJSON(d []byte) error {
	if len(d) == 0 {
		*b = Empty
		return nil
	}
	if d[0] == '"' && len(d) >= 1 {
		switch d[1] {
		case '1', 'T', 't':
			*b = True
			return nil
		case '0', 'F', 'f':
			*b = False
			return nil
		}
		*b = Empty
		return nil
	}
	switch d[0] {
	case '1', 'T', 't':
		*b = True
		return nil
	case '0', 'F', 'f':
		*b = False
		return nil
	}
	*b = Empty
	return nil
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
func (f Filter) MarshalStream(w data.Writer) error {
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
	if err := data.WriteStringList(w, f.Include); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream will attempt to read the Filter data from the supplied Reader
// and return any errors that may occur.
func (f *Filter) UnmarshalStream(r data.Reader) error {
	if err := r.ReadUint32(&f.PID); err != nil {
		return err
	}
	if err := r.ReadBool(&f.Fallback); err != nil {
		return err
	}
	if err := r.ReadUint8((*uint8)((&f.Session))); err != nil {
		return err
	}
	if err := r.ReadUint8((*uint8)((&f.Elevated))); err != nil {
		return err
	}
	if err := data.ReadStringList(r, &f.Exclude); err != nil {
		return err
	}
	if err := data.ReadStringList(r, &f.Include); err != nil {
		return err
	}
	return nil
}
