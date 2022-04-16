//go:build nosweep

package limits

import (
	"context"
	"time"
)

// MemorySweep enables the GC memory sweeper, which keeps the process memory
// clean to prevent any crashes while in DLL format or injected. This function
// only needs to be called once and will return immediately.
//
// The context is recommended to prevent any leaking Goroutines from being left
// behind.
//
// Defaults to a time of one minute.
func MemorySweep(_ context.Context) {}

// MemorySweepEx enables the GC memory sweeper, which keeps the process memory
// clean to prevent any crashes while in DLL format or injected. This function
// only needs to be called once and will return immediately.
//
// The context is recommended to prevent any leaking Goroutines from being left
// behind.
//
// Allows for specification of the timespace between sweeps.
func MemorySweepEx(_ context.Context, _ time.Duration) {}
