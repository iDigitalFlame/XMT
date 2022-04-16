//go:build !implant

package task

import (
	"io"
	"os"
)

// DLLFile will create a Tasklet that will instruct the client to local a
// DLL from the specified local (server - the one calling this function) file
// source. (THIS WILL MAKE A WRITE TO DISK!)
//
// To prevent writes to disk, use the 'cmd.DLLToASM' function on the server
// (or any non 'implant' tagged build) to build a shellcode DLL+loader using
// SRDi and launch as Assembly instead.
//
// The Filter attribute will attempt to set the target that runs the Process.
// If none are specified, the Process will be ran under the client process.
//
// C2 Details:
//  ID: WvInjectDLL
//
//  Input:
//      DLL struct {
//          string          // Path
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
//          []byte          // Raw DLL Data
//      }
//  Output:
//      uint64              // Handle
//      uint32              // PID
//      int32               // Exit Code
func DLLFile(s string) (*DLL, error) {
	b, err := os.ReadFile(s)
	if err != nil {
		return nil, err
	}
	return &DLL{Data: b}, nil
}

// DLLReader will create a Tasklet that will instruct the client to local a DLL
// from the specified reader source. (THIS WILL MAKE A WRITE TO DISK!)
//
// To prevent writes to disk, use the 'cmd.DLLToASM' function on the server
// (or any non 'implant' tagged build) to build a shellcode DLL+loader using
// SRDi and launch as Assembly instead.
//
// The Filter attribute will attempt to set the target that runs the Process.
// If none are specified, the Process will be ran under the client process.
//
// C2 Details:
//  ID: WvInjectDLL
//
//  Input:
//      DLL struct {
//          string          // Path
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
//          []byte          // Raw DLL Data
//      }
//  Output:
//      uint64              // Handle
//      uint32              // PID
//      int32               // Exit Code
func DLLReader(r io.Reader) (*DLL, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &DLL{Data: b}, nil
}
