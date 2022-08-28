// Copyright (C) 2020 - 2022 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package task

import (
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const (
	flagChannel uint8 = 1 << iota
	flagNoReturnOutput
	flagStopOnError
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
	f uint8
}

// Clear will reset the Script and empty its contents.
//
// This does not remove the error and output settings.
func (s *Script) Clear() {
	s.d.Clear()
}

// Size returns the internal size of the backing Script buffer, similar to len(s).
func (s *Script) Size() int {
	return s.d.Size()
}

// Empty returns true if this Script's backing buffer is empty.
func (s *Script) Empty() bool {
	return s.d.Empty()
}

// Output controls the 'return output' setting for this Script.
//
// If set to True (the default), the results of all executed Tasks in this
// script will return their resulting output (if applicable and with no errors).
// Otherwise, False will disable output and all Task output will be ignored,
// unless errors occur.
func (s *Script) Output(e bool) {
	if e {
		s.f = s.f &^ flagNoReturnOutput
	} else {
		s.f |= flagNoReturnOutput
	}
}

// IsOutput returns true if the 'return output' setting is set to true.
func (s *Script) IsOutput() bool {
	return s.f&flagNoReturnOutput == 0
}

// Channel (if true) will set this Script payload to enable Channeling mode
// (if supported) before running.
//
// NOTE: There is not a way to Scripts to disable channeling themselves.
func (s *Script) Channel(e bool) {
	if e {
		s.f |= flagChannel
	} else {
		s.f = s.f &^ flagChannel
	}
}

// IsChannel returns true if the 'channel' setting is set to true.
func (s *Script) IsChannel() bool {
	return s.f&flagChannel != 0
}

// Payload returns the raw, underlying bytes in this Script.
// If this script is empty the return will be empty.
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
// an error during runtime, otherwise False (the default), will report the error
// in the chain and will keep going.
func (s *Script) StopOnError(e bool) {
	if e {
		s.f |= flagStopOnError
	} else {
		s.f = s.f &^ flagStopOnError
	}
}

// IsStopOnError returns true if the 'stop on error' setting is set to true.
func (s *Script) IsStopOnError() bool {
	return s.f&flagStopOnError != 0
}

// Truncate discards all but the first n unread bytes from the underlying buffer
// but continues to use the same allocated storage.
//
// This will return an error if n is negative or greater than the length of the
// buffer.
func (s *Script) Truncate(n int) error {
	return s.d.Truncate(n)
}

// Add will add the supplied Task (in Packet form), to the Script. If this Script
// was not initialized, it will be initialized with the default options first.
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
		return xerr.Sub("invalid Packet", 0x69)
	}
	s.d.WriteUint8(n.ID)
	s.d.WriteBytes(n.Payload())
	return nil
}

// NewScript returns a new Script instance with the Settings for 'stop on error'
// and 'return output' set to the values specified.
//
// Non initialized Scripts can be used instead of calling this function directly.
func NewScript(errors, output bool) *Script {
	s := new(Script)
	if errors {
		s.f |= flagStopOnError
	}
	if !output {
		s.f |= flagNoReturnOutput
	}
	return s
}

// AddTasklet will add the supplied Tasklet result, to the Script. If this Script
// was not initialized, it will be initialized with the default options first.
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
		return xerr.Sub("empty or nil Tasklet", 0x6A)
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
	if s.d.Empty() {
		return nil, xerr.Sub("script is empty", 0x6B)
	}
	n := &com.Packet{ID: MvScript}
	n.WriteUint8(s.f)
	s.d.Seek(0, 0)
	if n.Write(s.d.Payload()); s.IsChannel() {
		n.Flags |= com.FlagChannel
	}
	return n, nil
}

// Append will add the supplied Tasks (in Packet form), to the Script. If this
// Script was not initialized, it will be initialized with the default options first.
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
