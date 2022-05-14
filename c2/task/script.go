package task

import (
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Script is a Tasklet type that allows for chaining the results of multiple
// Tasks in a single instance to be run as one.
//
// All script tasks will be run in the same thread and will execute in order
// until all tasks are complete.
//
// Each Script has two boolean options, 'Output' (default: true), which determines
// if the Script result should be returned and 'StopOnError' (default: false),
// which will determine the action taken if an error occurs in one of the Script
// tasks.
type Script struct {
	d data.Chunk
	c bool
}

// Clear will reset the Script and empty it's contents.
//
// This also removes the error and output settings. Use the 'Reset' function
// to keep those instead.
func (s *Script) Clear() {
	s.d.Clear()
}

// Reset will reset the Script and empty it's contents, but keep its error and
// output settings intact.
func (s *Script) Reset() {
	v, _ := s.d.Uint16()
	s.d.Clear()
	s.d.WriteUint16(v)
}

// Output controls the 'return output' setting for this Script.
//
// If set to True (the default), the results of all executed Tasks in this
// script will return their resulting output (if applicable and with no errors).
// Otherwise, False will disable output and all Task output will be ignored,
// unless errors occur.
func (s *Script) Output(e bool) {
	if s.d.Empty() {
		s.d.WriteUint8(0)
		s.d.WriteBool(e)
		return
	}
	s.d.WriteBoolPos(1, e)
}

// Channel (if true) will set this Script payload to enable Channeling mode
// (if supported) before running.
//
// NOTE: There is not a way to Scripts to disable channeling themselves.
func (s *Script) Channel(e bool) {
	s.c = e
}

// Payload returns the raw, underlying bytes in this Script.
// If this script is empty or initialized, the return will be empty.
func (s *Script) Payload() []byte {
	if s.d.Empty() {
		return nil
	}
	return s.d.Payload()
}

// Replace will clear the Script data and replace it with the supplied byte
// array.
//
// It is the callers responsibility to ensure that the first type bytes are
// correct values for error and output.
func (s *Script) Replace(b []byte) {
	s.d.Clear()
	s.d.Write(b)
}

// StopOnError controls the 'stop on error' setting for this Script.
//
// If set to True, the Script will STOP processing if one of the Tasks returns
// an error during runtime. Otherwise False (the default), will report the error
// in the chain and will keep going.
func (s *Script) StopOnError(e bool) {
	if s.d.Empty() {
		s.d.WriteBool(e)
		s.d.WriteUint8(1)
		return
	}
	s.d.WriteBoolPos(0, e)
}

// Add will add the supplied Task (in Packet form), to the Script. If this Script
// was not initalized, it will be initalized with the default options first.
//
// This function will return an error if the Packet supplied is invalid for
// Script usage.
//
// An invalid Script Packet is one of the following:
// - Any fragmented Packet
// - Any Packet with control (error/oneshot/proxy/multi/frag) Flags set
// - Any NoP Packet
// - Any Packet with a System ID
// - Any Script
func (s *Script) Add(n *com.Packet) error {
	if n == nil || n.ID == 0 || n.ID < MvRefresh || n.Flags > 0 || n.ID == MvScript {
		return xerr.Sub("invalid Packet", 0xF)
	}
	if s.d.Empty() {
		s.d.WriteUint16(1)
	}
	s.d.WriteUint8(n.ID)
	s.d.WriteBytes(n.Payload())
	return nil
}

// NewScript returns a new Script instance with the Settings for 'stop on error'
// and 'return output' set to the values specified.
//
// Non intalized Scripts can be used instead of calling this function directly.
func NewScript(errors, output bool) *Script {
	s := new(Script)
	s.d.WriteBool(errors)
	s.d.WriteBool(output)
	return s
}

// AddTasklet will add the supplied Tasklet result, to the Script. If this Script
// was not initalized, it will be initalized with the default options first.
//
// This function will return an error if the Packet supplied is invalid for
// Script usage or the Tasklet action returned an error or is invalid.
//
// An invalid Script Packet is one of the following:
// - Any fragmented Packet
// - Any Packet with control (error/oneshot/proxy/multi/frag) Flags set
// - Any NoP Packet
// - Any Packet with a System ID
// - Any Script
func (s *Script) AddTasklet(t Tasklet) error {
	if t == nil {
		return xerr.Sub("empty or nil Tasklet", 0x26)
	}
	n, err := t.Packet()
	if err != nil {
		return err
	}
	return s.Add(n)
}

// Packet will take the configured Script options/data and will return a Packet
// and any errors that may occur during building.
//
// This allows the Script struct to fulfil the 'Tasklet' interface.
//
// C2 Details:
//  ID: MvScript
//
//  Input:
//      bool      // Option 'output'
//      bool      // Option 'stop on error'
//      ...uint8  // Packet ID
//      ...[]byte // Packet Data
//  Output:
//      ...uint8  // Result Packet ID
//      ...bool   // Result is not error
//      ...[]byte // Result Data
func (s *Script) Packet() (*com.Packet, error) {
	n := &com.Packet{ID: MvScript} // Make a copy
	s.d.Seek(0, 0)
	n.Write(s.d.Payload())
	return n, nil
}

// Append will add the supplied Tasks (in Packet form), to the Script. If this
// Script was not initalized, it will be initalized with the default options first.
//
// This function is like 'Add' but takes a vardict of multiple Packets to be added
// in as single call.
//
// This function will return an error if any of the Packets supplied are invalid
// for Script usage.
//
// An invalid Script Packet is one of the following:
// - Any fragmented Packet
// - Any Packet with control (error/oneshot/proxy/multi/frag) Flags set
// - Any NoP Packet
// - Any Packet with a System ID
func (s *Script) Append(n ...*com.Packet) error {
	if len(n) == 0 {
		return nil
	}
	if len(n) == 1 {
		return s.Add(n[0])
	}
	for i := range n {
		if err := s.Add(n[i]); err != nil {
			return err
		}
	}
	return nil
}
