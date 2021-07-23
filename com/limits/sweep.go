// +build !nosweep

package limits

import (
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"time"
)

var enabled uint32

// MemorySweep enables the GC memory sweeper, which keeps the process memory clean to
// prevent any crashes while in DLL format or injected. This function only needs to be
// called once and will return immediately.
//
// Defaults to a time of one minute.
func MemorySweep() {
	MemorySweepEx(time.Minute)
}
func sweep(t time.Duration) {
	for {
		runtime.GC()
		debug.FreeOSMemory()
		time.Sleep(t)
	}
}

// MemorySweepEx enables the GC memory sweeper, which keeps the process memory clean to
// prevent any crashes while in DLL format or injected. This function only needs to be
// called once and will return immediately.
//
// Allows for specification of the timespace between sweeps.
func MemorySweepEx(d time.Duration) {
	if d <= 0 || atomic.LoadUint32(&enabled) == 1 {
		return
	}
	atomic.StoreUint32(&enabled, 1)
	go sweep(d)
}
