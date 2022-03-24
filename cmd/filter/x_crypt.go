//go:build crypt

package filter

import (
	"encoding/json"

	"github.com/iDigitalFlame/xmt/util/crypt"
)

// MarshalJSON will attempt to convert the data in this Filter into the returned
// JSON byte array.
func (f Filter) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{crypt.Get(58): f.Fallback} // fallback
	if f.PID != 0 {
		m[crypt.Get(59)] = f.PID // pid
	}
	if f.Session > Empty {
		m[crypt.Get(60)] = f.Session // session
	}
	if f.Elevated > Empty {
		m[crypt.Get(61)] = f.Elevated // elevated
	}
	if len(f.Exclude) > 0 {
		m[crypt.Get(62)] = f.Elevated // exclude
	}
	if len(f.Include) > 0 {
		m[crypt.Get(63)] = f.Include // include
	}
	return json.Marshal(m)
}
func (b boolean) MarshalJSON() ([]byte, error) {
	switch b {
	case True:
		return []byte(crypt.Get(64)), nil // "true"
	case False:
		return []byte(crypt.Get(65)), nil // "false"
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
	if v, ok := m[crypt.Get(59)]; ok { // pid
		if err := json.Unmarshal(v, &f.PID); err != nil {
			return err
		}
	}
	if v, ok := m[crypt.Get(60)]; ok { // session
		if err := json.Unmarshal(v, &f.Session); err != nil {
			return err
		}
	}
	if v, ok := m[crypt.Get(61)]; ok { // elevated
		if err := json.Unmarshal(v, &f.Elevated); err != nil {
			return err
		}
	}
	if v, ok := m[crypt.Get(62)]; ok { // exclude
		if err := json.Unmarshal(v, &f.Exclude); err != nil {
			return err
		}
	}
	if v, ok := m[crypt.Get(63)]; ok { // include
		if err := json.Unmarshal(v, &f.Include); err != nil {
			return err
		}
	}
	if v, ok := m[crypt.Get(58)]; ok { // fallback
		if err := json.Unmarshal(v, &f.Fallback); err != nil {
			return err
		}
	}
	return nil
}
