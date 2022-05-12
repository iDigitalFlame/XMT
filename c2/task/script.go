package task

import (
	"github.com/iDigitalFlame/xmt/com"
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
	n *com.Packet
}

// Clear will reset the Script and empty it's contents.
func (s *Script) Clear() {
	s.n.Clear()
	s.n = nil
}

// Output controls the 'return output' setting for this Script.
//
// If set to True (the default), the results of all executed Tasks in this
// script will return their resulting output (if applicable and with no errors).
// Otherwise, False will disable output and all Task output will be ignored,
// unless errors occur.
func (s *Script) Output(e bool) {
	if s.n == nil {
		s.n = &com.Packet{ID: MvScript}
		s.n.WriteUint8(0)
		s.n.WriteBool(e)
		return
	}
	s.n.WriteBoolPos(1, e)
}

// Channel (if true) will set this Script payload to enable Channeling mode
// (if supported) before running.
//
// NOTE: There is not a way to Scripts to disable channeling themselves.
func (s *Script) Channel(e bool) {
	if s.n == nil {
		s.n = &com.Packet{ID: MvScript}
		s.n.WriteUint16(1)
	}
	if e {
		s.n.Flags.Set(com.FlagChannel)
	} else {
		s.n.Flags.Unset(com.FlagChannel)
	}
}

// StopOnError controls the 'stop on error' setting for this Script.
//
// If set to True, the Script will STOP processing if one of the Tasks returns
// an error during runtime. Otherwise False (the default), will report the error
// in the chain and will keep going.
func (s *Script) StopOnError(e bool) {
	if s.n == nil {
		s.n = &com.Packet{ID: MvScript}
		s.n.WriteBool(e)
		s.n.WriteUint8(1)
		return
	}
	s.n.WriteBoolPos(0, e)
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
func (s *Script) Add(n *com.Packet) error {
	if n == nil || n.ID == 0 || n.ID < MvRefresh || n.Flags > 0 {
		return xerr.Sub("invalid Packet", 0xF)
	}
	if s.n == nil {
		s.n = &com.Packet{ID: MvScript}
		s.n.WriteUint16(1)
	}
	s.n.WriteUint8(n.ID)
	s.n.WriteBytes(n.Payload())
	return nil
}

// NewScript returns a new Script instance with the Settings for 'stop on error'
// and 'return output' set to the values specified.
//
// Non intalized Scripts can be used instead of calling this function directly.
func NewScript(errors, output bool) *Script {
	s := &Script{n: &com.Packet{ID: MvScript}}
	s.n.WriteBool(errors)
	s.n.WriteBool(output)
	return s
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
	return s.n, nil
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
