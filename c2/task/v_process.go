//go:build !implant
// +build !implant

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

package task

import (
	"io"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
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
//
//	ID: TvExecute
//
//	Input:
//	    Process struct {
//	        []string        // Args
//	        string          // Dir
//	        []string        // Environment
//	        uint32          // Flags
//	        bool            // Wait
//	        int64           // Timeout
//	        Filter struct { // Filter
//	            bool        // Filter Status
//	            uint32      // PID
//	            bool        // Fallback
//	            uint8       // Session
//	            uint8       // Elevated
//	            []string    // Exclude
//	            []string    // Include
//	        }
//	        []byte          // Stdin Data
//	    }
//	Output:
//	    uint32              // PID
//	    int32               // Exit Code
//	    []byte              // Output (Stdout and Stderr)
func Run(c string) Process {
	return Process{Args: cmd.Split(c), Wait: true}
}

// Packet will take the configured Process options and will return a Packet
// and any errors that may occur during building.
//
// This allows Process to fulfil the 'Tasklet' interface.
//
// C2 Details:
//
//	ID: TvAssembly
//
//	Input:
//	    Process struct {
//	        []string        // Args
//	        string          // Dir
//	        []string        // Environment
//	        uint32          // Flags
//	        bool            // Wait
//	        int64           // Timeout
//	        bool            // Hide
//	        string          // Username
//	        string          // Domain
//	        string          // Password
//	        Filter struct { // Filter
//	            bool        // Filter Status
//	            uint32      // PID
//	            bool        // Fallback
//	            uint8       // Session
//	            uint8       // Elevated
//	            []string    // Exclude
//	            []string    // Include
//	        }
//	        []byte          // Stdin Data
//	    }
//	Output:
//	    uint32              // PID
//	    int32               // Exit Code
//	    []byte              // Output (Stdout and Stderr)
func (p Process) Packet() (*com.Packet, error) {
	n := &com.Packet{ID: TvExecute}
	p.MarshalStream(n)
	return n, nil
}

// SetStdin wil attempt to read all the data from the supplied reader to fill
// the Stdin byte array for this Process struct.
//
// This function will return an error if any errors occurs during reading.
func (p *Process) SetStdin(r io.Reader) error {
	var err error
	p.Stdin, err = data.ReadAll(r)
	return err
}

// MarshalStream writes the data for this Process to the supplied Writer.
func (p Process) MarshalStream(w data.Writer) error {
	if err := data.WriteStringList(w, p.Args); err != nil {
		return err
	}
	if err := w.WriteString(p.Dir); err != nil {
		return err
	}
	if err := data.WriteStringList(w, p.Env); err != nil {
		return err
	}
	if err := w.WriteBool(p.Wait); err != nil {
		return err
	}
	if err := w.WriteUint32(p.Flags); err != nil {
		return err
	}
	if err := w.WriteInt64(int64(p.Timeout)); err != nil {
		return err
	}
	if err := w.WriteBool(p.Hide); err != nil {
		return err
	}
	if err := w.WriteString(p.User); err != nil {
		return err
	}
	if err := w.WriteString(p.Domain); err != nil {
		return err
	}
	if err := w.WriteString(p.Pass); err != nil {
		return err
	}
	if err := p.Filter.MarshalStream(w); err != nil {
		return err
	}
	return w.WriteBytes(p.Stdin)
}
