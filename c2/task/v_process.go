//go:build !implant

package task

import (
	"io"

	"github.com/iDigitalFlame/xmt/cmd"
)

// Run will create a Tasklet that will instruct the client to run a command.
// This command will parsed using the 'cmd.Split' function.
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
