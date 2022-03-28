//go:build !implant

package task

import (
	"io"
	"os"
)

// AssemblyFile will create a Tasklet that will instruct the client to run
// shellcode from a file source on the local (server - the one calling this
// function) machine.
//
// This will attempt to read the file and will return an error if it fails.
//
// C2 Details:
//  ID: TvAssembly
//
//  Input:
//      Assembly struct {
//          bool            // Wait
//          int64           // Timeout
//          bool            // Filter Status
//          Filter struct { // Filter
//              uint32      // PID
//              bool        // Fallback
//              uint8       // Session
//              uint8       // Elevated
//              []string    // Exclude
//              []string    // Include
//          }
//          []byte          // Assembly Data
//      }
//  Output:
//      uint64              // Handle
//      uint32              // PID
//      int32               // Exit Code
//
// C2 Client Command:
//    asm <file>
//    assembly <file>
func AssemblyFile(s string) (*Assembly, error) {
	b, err := os.ReadFile(s)
	if err != nil {
		return nil, err
	}
	return &Assembly{Data: b}, nil
}

// AssemblyReader will create a Tasklet that will instruct the client to run
// shellcode from the contents of the supplied Reader.
//
// This will attempt to read from the Reader and will return an error if it
// fails.
//
// C2 Details:
//  ID: TvAssembly
//
//  Input:
//      Assembly struct {
//          bool            // Wait
//          int64           // Timeout
//          bool            // Filter Status
//          Filter struct { // Filter
//              uint32      // PID
//              bool        // Fallback
//              uint8       // Session
//              uint8       // Elevated
//              []string    // Exclude
//              []string    // Include
//          }
//          []byte          // Assembly Data
//      }
//  Output:
//      uint64              // Handle
//      uint32              // PID
//      int32               // Exit Code
//
// C2 Client Command:
//    asm <file>
//    assembly <file>
func AssemblyReader(r io.Reader) (*Assembly, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &Assembly{Data: b}, nil
}
