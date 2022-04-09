//go:build implant

package c2

import (
	"io"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/device"
)

const maxEvents = 512

// Server is the manager for all C2 Listener and Sessions connection and states.
// This struct also manages all events and connection changes.
type Server struct {
	Oneshot func(*com.Packet)
}

// Listener is a struct that is passed back when a C2 Listener is added to the
// Server.
//
// The Listener struct allows for controlling the Listener and setting callback
// functions to be used when a client connects, registers or disconnects.
type Listener struct {
	s *Server
	m messager
}
type proxyState struct{}

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

// Remove removes and closes the Session and releases all it's associated
// resources.
//
// This does not close the Session on the client's end, use the Shutdown
// function to properly shutdown the client process.
func (*Listener) Remove(_ device.ID) {
}

// JSON returns the data of this Server as a JSON blob.
func (Server) JSON(_ io.Writer) error {
	return nil
}

// JSON returns the data of this Session as a JSON blob.
func (Session) JSON(_ io.Writer) error {
	return nil
}
func (*Listener) tryRemove(_ *Session) {
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
func (*Session) updateProxyInfo(_ []proxyData) {}
