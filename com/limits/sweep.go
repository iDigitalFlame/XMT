//go:build !nosweep

package limits

import (
	"context"
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/iDigitalFlame/xmt/util/bugtrack"
)

var enabled uint32

// MemorySweep enables the GC memory sweeper, which keeps the process memory
// clean to prevent any crashes while in DLL format or injected. This function
// only needs to be called once and will return immediately.
//
// The context is recommended to prevent any leaking Goroutines from being left
// behind.
//
// Defaults to a time of one minute.
func MemorySweep(x context.Context) {
	MemorySweepEx(x, time.Minute)
}
func sweep(x context.Context, t time.Duration) {
	v := time.NewTicker(t)
loop:
	for {
		if bugtrack.Enabled {
			bugtrack.Track("limits.sweep(): Starting GC and Free.")
		}
		runtime.GC()
		debug.FreeOSMemory()
		select {
		case <-v.C:
		case <-x.Done():
			break loop
		}
	}
	if v.Stop(); bugtrack.Enabled {
		bugtrack.Track("limits.sweep(): Stopping GC and Free thread.")
	}
}

// MemorySweepEx enables the GC memory sweeper, which keeps the process memory
// clean to prevent any crashes while in DLL format or injected. This function
// only needs to be called once and will return immediately.
//
// The context is recommended to prevent any leaking Goroutines from being left
// behind.
//
// Allows for specification of the timespace between sweeps.
func MemorySweepEx(x context.Context, d time.Duration) {
	if d <= 0 || atomic.LoadUint32(&enabled) == 1 {
		return
	}
	debug.SetGCPercent(60)
	debug.SetTraceback("none")
	runtime.SetCPUProfileRate(0)
	runtime.SetBlockProfileRate(0)
	runtime.SetMutexProfileFraction(0)
	runtime.GOMAXPROCS(runtime.NumCPU())
	atomic.StoreUint32(&enabled, 1)
	go sweep(x, d)
}
