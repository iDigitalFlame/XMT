//go:build !client
// +build !client

package xerr

// New creates a new string backed error struct and returns it. This error struct does not support Unwrapping.
// The resulting structs created will be comparable.
func New(s string) error {
	return strErr(s)
}

// Wrap creates a new error that wraps the specified error. If not nil, this function will append ": " + 'Error()'
// to the resulting string message.
func Wrap(s string, e error) error {
	if e != nil {
		return &err{s: s + ": " + e.Error(), e: e}
	}
	return &err{s: s}
}
