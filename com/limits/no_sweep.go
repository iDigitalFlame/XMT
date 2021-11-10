//go:build nosweep
// +build nosweep

package limits

import "time"

// MemorySweep enables the GC memory sweeper, which keeps the process memory clean to
// prevent any crashes while in DLL format or injected. This function only needs to be
// called once and will return immediately.
//
// Defaults to a time of one minute.
func MemorySweep() {}

// MemorySweepEx enables the GC memory sweeper, which keeps the process memory clean to
// prevent any crashes while in DLL format or injected. This function only needs to be
// called once and will return immediately.
//
// Allows for specification of the timespace between sweeps.
func MemorySweepEx(_ time.Duration) {}
