// Copyright (C) 2020 - 2023 iDigitalFlame
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

package cmd

import (
	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

var (
	// ErrNotSame is an error returned by the 'Start' or 'Run' functions when
	// attempting to start a ASM or DLL without a Filter target and not setting
	// the 'SelfProcess' attribute to 'true',
	//
	// This is used to prevent accidental self-injection
	ErrNotSame = xerr.Sub("cannot host new thread", 0x80)
	// ErrNotStarted is an error returned by multiple functions when attempting
	// to access a Runnable function that requires the Runnable to be started first.
	ErrNotStarted = xerr.Sub("process has not started", 0x3A)
	// ErrEmptyCommand is an error returned when attempting to start a Runnable
	// that has empty arguments.
	ErrEmptyCommand = xerr.Sub("process arguments are empty", 0x3B)
	// ErrStillRunning is returned when attempting to access the exit code on a
	// Runnable.
	ErrStillRunning = xerr.Sub("process still running", 0x3C)
	// ErrAlreadyStarted is an error returned by the 'Start' or 'Run' functions
	// when attempting to start a Runnable that has already been started via a
	// 'Start' or 'Run' function call.
	ErrAlreadyStarted = xerr.Sub("process already started", 0x3D)
)

// ExitError is a type of error that is returned by the Wait and Run functions
// when a function returns an error code other than zero.
type ExitError struct {
	Exit uint32
}

// Runnable is an interface that helps support the type of structs that can be
// used for execution, such as Assembly, DLL and Process, which share the same
// methods as this interface.
type Runnable interface {
	Run() error
	Pid() uint32
	Wait() error
	Stop() error
	Start() error
	Running() bool
	Release() error
	Handle() (uintptr, error)
	ExitCode() (int32, error)
	SetParent(*filter.Filter)
}

// Split will attempt to split the specified string based on the escape characters
// and spaces while attempting to preserve anything that is not a splitting space.
//
// This will automatically detect quotes and backslashes. The return result is a
// string array that can be used as args.
func Split(v string) []string {
	var (
		r []string
		s int
	)
	for i, m := 0, byte(0); i < len(v); i++ {
		switch {
		case v[i] == '\\' && i+1 < len(v) && (v[i+1] == '"' || v[i+1] == '\''):
			i++
		case v[i] == ' ' && m == 0 && s == i:
		case v[i] == ' ' && m == 0:
			r, s = append(r, v[s:i]), i+1
		case (v[i] == '"' || v[i] == '\'') && m == 0:
			s, m = i+1, v[i]
		case (v[i] == '"' || v[i] == '\'') && m == v[i]:
			r, m, s = append(r, v[s:i]), 0, i+1
		}
	}
	if s < len(v) {
		r = append(r, v[s:])
	}
	return r
}

// Error fulfills the error interface and returns a formatted string that
// specifies the Process Exit Code.
func (e ExitError) Error() string {
	return "exit " + util.Itoa(int64(e.Exit))
}
