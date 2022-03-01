//go:build implant
// +build implant

package c2

import "io"

const maxEvents = 512

func (Session) name() string {
	return ""
}
func (Proxy) prefix() string {
	return ""
}
func (status) String() string {
	return ""
}

// String returns the details of this Session as a string.
func (Session) String() string {
	return ""
}

// JSON returns the data of this Job as a JSON blob.
func (Job) JSON(_ io.Writer) error {
	return nil
}

// JSON returns the data of this Server as a JSON blob.
func (Server) JSON(_ io.Writer) error {
	return nil
}

// JSON returns the data of this Session as a JSON blob.
func (Session) JSON(_ io.Writer) error {
	return nil
}

// JSON returns the data of this Listener as a JSON blob.
func (Listener) JSON(_ io.Writer) error {
	return nil
}

// MarshalJSON fulfils the JSON Marshaler interface.
func (Job) MarshalJSON() ([]byte, error) {
	return nil, nil
}

// MarshalJSON fulfils the JSON Marshaler interface.
func (Server) MarshalJSON() ([]byte, error) {
	return nil, nil
}

// MarshalJSON fulfils the JSON Marshaler interface.
func (Session) MarshalJSON() ([]byte, error) {
	return nil, nil
}

// MarshalJSON fulfils the JSON Marshaler interface.
func (Listener) MarshalJSON() ([]byte, error) {
	return nil, nil
}
