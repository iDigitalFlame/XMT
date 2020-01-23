package util

// String is a wrapper for strings to support
// the fmt.Stringer interface.
type String string

// String returns the string value of itself.
func (s String) String() string {
	return string(s)
}
