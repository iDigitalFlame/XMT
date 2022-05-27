//go:build !bugs

package bugtrack

// Enabled is the stats of the bugtrack package.
//
// This is true if bug tracking is enabled.
const Enabled = false

// Recover is a "guard" function to be used to gracefully shutdown a program
// when a panic is detected.
//
// Can be en enabled by using:
//    if bugtrack.Enabled {
//        defer bugtrack.Recover("thread-name")
//    }
//
// The specified name will be entered into the bugtrack log and a stack trace
// will be generated before gracefully execution to the program.
func Recover(_ string) {}

// Track is a simple logging function that takes the same arguments as a
// 'fmt.Sprintf' function. This can be used to track bugs or output values.
//
// Not recommended to be used in production environments.
//
// The "-tags bugs" option is required in order for this function to be used.
func Track(_ string, _ ...any) {}
