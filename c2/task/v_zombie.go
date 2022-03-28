//go:build !implant

package task

import (
	"io"
	"os"
)

// SetStdin wil attempt to read all the data from the supplied reader to fill
// the Stdin byte array for this Process struct.
//
// This function will return an error if any errors occurs during reading.
func (z *Zombie) SetStdin(r io.Reader) error {
	var err error
	z.Stdin, err = io.ReadAll(r)
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
//  ID: WvZombie
//
//  Input:
//      Zombie struct {
//          string          // Path
//          []byte          // Data
//          bool            // IsDLL
//          []string        // Args
//          string          // Dir
//          []string        // Environment
//          uint32          // Flags
//          bool            // Wait
//          int64           // Timeout
//          bool            // Filter Status
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
//
// C2 Client Command:
//  zombie_asm <file> <args...>
//  zombie_dll <file> <args...>
//  zombie_dll_local <file> <args...>
func ZombieAsm(b []byte, args ...string) *Zombie {
	return &Zombie{Data: b, Args: args}
}

// ZombieDLL will create a Zombie Tasklet that can be used to run the supplied
// DLL in a Zombie process that uses the specified command line arguments.
//
// The Filter attribute will attempt to set the target that runs the zombie
// Process. If none are specified, the Process will be ran under the client
// process.
//
// C2 Details:
//  ID: WvZombie
//
//  Input:
//      Zombie struct {
//          string          // Path
//          []byte          // Data
//          bool            // IsDLL
//          []string        // Args
//          string          // Dir
//          []string        // Environment
//          uint32          // Flags
//          bool            // Wait
//          int64           // Timeout
//          bool            // Filter Status
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
//
// C2 Client Command:
//  zombie_asm <file> <args...>
//  zombie_dll <file> <args...>
//  zombie_dll_local <file> <args...>
func ZombieDLL(dll string, args ...string) *Zombie {
	return &Zombie{Path: dll, Args: args}
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
//  ID: WvZombie
//
//  Input:
//      Zombie struct {
//          string          // Path
//          []byte          // Data
//          bool            // IsDLL
//          []string        // Args
//          string          // Dir
//          []string        // Environment
//          uint32          // Flags
//          bool            // Wait
//          int64           // Timeout
//          bool            // Filter Status
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
//
// C2 Client Command:
//  zombie_asm <file> <args...>
//  zombie_dll <file> <args...>
//  zombie_dll_local <file> <args...>
func ZombieAsmFile(s string, args ...string) (*Zombie, error) {
	b, err := os.ReadFile(s)
	if err != nil {
		return nil, err
	}
	return &Zombie{Data: b, Args: args}, nil
}

// ZombieDLLFile will create a Zombie Tasklet that can be used to run the
// supplied DLL from the specified local (server) file source in a Zombie
// process that uses the specified command line arguments.
//
// The Filter attribute will attempt to set the target that runs the zombie
// Process. If none are specified, the Process will be ran under the client
// process.
//
// C2 Details:
//  ID: WvZombie
//
//  Input:
//      Zombie struct {
//          string          // Path
//          []byte          // Data
//          bool            // IsDLL
//          []string        // Args
//          string          // Dir
//          []string        // Environment
//          uint32          // Flags
//          bool            // Wait
//          int64           // Timeout
//          bool            // Filter Status
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
//
// C2 Client Command:
//  zombie_asm <file> <args...>
//  zombie_dll <file> <args...>
//  zombie_dll_local <file> <args...>
func ZombieDLLFile(s string, args ...string) (*Zombie, error) {
	b, err := os.ReadFile(s)
	if err != nil {
		return nil, err
	}
	return &Zombie{Data: b, IsDLL: true, Args: args}, nil
}

// ZombieDLLReader will create a Zombie Tasklet that can be used to run the
// supplied DLL from the specified reader source in a Zombie process that uses
// the specified command line arguments.
//
// The Filter attribute will attempt to set the target that runs the zombie
// Process. If none are specified, the Process will be ran under the client
// process.
//
// C2 Details:
//  ID: WvZombie
//
//  Input:
//      Zombie struct {
//          string          // Path
//          []byte          // Data
//          bool            // IsDLL
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
//
// C2 Client Command:
//  zombie_asm <file> <args...>
//  zombie_dll <file> <args...>
//  zombie_dll_local <file> <args...>
func ZombieDLLReader(r io.Reader, args ...string) (*Zombie, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &Zombie{Data: b, IsDLL: true, Args: args}, nil
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
//  ID: WvZombie
//
//  Input:
//      Zombie struct {
//          string          // Path
//          []byte          // Data
//          bool            // IsDLL
//          []string        // Args
//          string          // Dir
//          []string        // Environment
//          uint32          // Flags
//          bool            // Wait
//          int64           // Timeout
//          bool            // Filter Status
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
//
// C2 Client Command:
//  zombie_asm <file> <args...>
//  zombie_dll <file> <args...>
//  zombie_dll_local <file> <args...>
func ZombieAsmReader(r io.Reader, args ...string) (*Zombie, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &Zombie{Data: b, Args: args}, nil
}
