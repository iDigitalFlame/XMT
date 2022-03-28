//go:build implant

package regedit

// String returns the string representation of the data held in the Data buffer.
// Invalid values of keys return an empty string.
func (Entry) String() string {
	return ""
}

// TypeName returns a string representation of the Type value, which represents
// the value data type.
func (Entry) TypeName() string {
	return ""
}
