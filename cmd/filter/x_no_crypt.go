//go:build !crypt

package filter

import "encoding/json"

// MarshalJSON will attempt to convert the data in this Filter into the returned
// JSON byte array.
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
