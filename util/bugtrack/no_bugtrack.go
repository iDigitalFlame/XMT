//go:build !bugs
// +build !bugs

package bugtrack

// Enabled is the stats of the bugtrack package.
// This is true if bug tracking is enabled.
const Enabled = false

// Track is a simple logging function that takes the same arguments as a 'fmt.Sprintf'
// function. This can be used to track bugs or output values.
//
// Not recommended to be used in production environments.
// The "-tags bugs" option is required in order for this function to be used.
func Track(_ string, _ ...interface{}) {}
