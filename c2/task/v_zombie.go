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
	"github.com/iDigitalFlame/xmt/data"
)

// SetStdin wil attempt to read all the data from the supplied reader to fill
// the Stdin byte array for this Process struct.
//
// This function will return an error if any errors occurs during reading.
func (z *Zombie) SetStdin(r io.Reader) error {
	var err error
	z.Stdin, err = data.ReadAll(r)
	return err
}

// ZombieAsm will create a Zombie Tasklet that can be used to run the supplied
// Assembly in a Zombie process that uses the specified command line arguments.
//
// The Filter attribute will attempt to set the target that runs the zombie
// Process. If none are specified, the Process will be ran under the client
// process.
//
// C2 Details:
//
//	ID: WvZombie
//
//	Input:
//	    Zombie struct {
//	        []byte          // Data
//	        []string        // Args
//	        string          // Dir
//	        []string        // Environment
//	        uint32          // Flags
//	        bool            // Wait
//	        int64           // Timeout
//	        bool            // Filter Status
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
func ZombieAsm(b []byte, args ...string) Zombie {
	return Zombie{Data: b, Args: args}
}

// ZombieAsmFile will create a Zombie Tasklet that can be used to run the
// supplied Assembly from the specified local (server) file source in a Zombie
// process that uses the specified command line arguments.
//
// The Filter attribute will attempt to set the target that runs the zombie
// Process. If none are specified, the Process will be ran under the client
// process.
//
// C2 Details:
//
//	ID: WvZombie
//
//	Input:
//	    Zombie struct {
//	        []byte          // Data
//	        []string        // Args
//	        string          // Dir
//	        []string        // Environment
//	        uint32          // Flags
//	        bool            // Wait
//	        int64           // Timeout
//	        bool            // Filter Status
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
func ZombieAsmFile(s string, args ...string) (Zombie, error) {
	b, err := data.ReadFile(s)
	if err != nil {
		return Zombie{}, err
	}
	return Zombie{Data: b, Args: args}, nil
}

// ZombieDLLFile will create a Zombie Tasklet that can be used to run the
// supplied DLL from the specified local (server) file source in a Zombie
// process that uses the specified command line arguments.
//
// The Filter attribute will attempt to set the target that runs the zombie
// Process. If none are specified, the Process will be ran under the client
// process.
//
// NOTE: This converts the DLL to Assembly.
//
// C2 Details:
//
//	ID: WvZombie
//
//	Input:
//	    Zombie struct {
//	        []byte          // Data
//	        []string        // Args
//	        string          // Dir
//	        []string        // Environment
//	        uint32          // Flags
//	        bool            // Wait
//	        int64           // Timeout
//	        bool            // Filter Status
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
func ZombieDLLFile(s string, args ...string) (Zombie, error) {
	b, err := data.ReadFile(s)
	if err != nil {
		return Zombie{}, err
	}
	return Zombie{Data: cmd.DLLToASM("", b), Args: args}, nil
}

// ZombieDLLReader will create a Zombie Tasklet that can be used to run the
// supplied DLL from the specified reader source in a Zombie process that uses
// the specified command line arguments.
//
// The Filter attribute will attempt to set the target that runs the zombie
// Process. If none are specified, the Process will be ran under the client
// process.
//
// NOTE: This converts the DLL to Assembly.
//
// C2 Details:
//
//	ID: WvZombie
//
//	Input:
//	    Zombie struct {
//	        []byte          // Data
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
func ZombieDLLReader(r io.Reader, args ...string) (Zombie, error) {
	b, err := data.ReadAll(r)
	if err != nil {
		return Zombie{}, err
	}
	return Zombie{Data: cmd.DLLToASM("", b), Args: args}, nil
}

// ZombieAsmReader will create a Zombie Tasklet that can be used to run the
// supplied Assembly from the specified reader source in a Zombie process that
// uses the specified command line arguments.
//
// The Filter attribute will attempt to set the target that runs the zombie
// Process. If none are specified, the Process will be ran under the client
// process.
//
// C2 Details:
//
//	ID: WvZombie
//
//	Input:
//	    Zombie struct {
//	        []byte          // Data
//	        []string        // Args
//	        string          // Dir
//	        []string        // Environment
//	        uint32          // Flags
//	        bool            // Wait
//	        int64           // Timeout
//	        bool            // Filter Status
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
func ZombieAsmReader(r io.Reader, args ...string) (Zombie, error) {
	b, err := data.ReadAll(r)
	if err != nil {
		return Zombie{}, err
	}
	return Zombie{Data: b, Args: args}, nil
}
