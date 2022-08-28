//go:build !nosweep

// Copyright (C) 2020 - 2022 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package limits

import (
	"context"
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/iDigitalFlame/xmt/device"
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
		select {
		case <-v.C:
		case <-x.Done():
			break loop
		}
		if bugtrack.Enabled {
			bugtrack.Track("limits.sweep(): Starting GC and Free.")
		}
		runtime.GC()
		device.FreeOSMemory()
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
// Allows for specification of the time span between sweeps.
func MemorySweepEx(x context.Context, d time.Duration) {
	if d <= 0 || atomic.LoadUint32(&enabled) == 1 {
		return
	}
	debug.SetGCPercent(60)
	debug.SetTraceback("none")
	runtime.SetCPUProfileRate(0)
	runtime.SetBlockProfileRate(0)
	runtime.SetMutexProfileFraction(0)
	// NOTE(dij): Let's ignore this one for now.
	//            We'll set it in our own env.
	// runtime.GOMAXPROCS(runtime.NumCPU())
	atomic.StoreUint32(&enabled, 1)
	go sweep(x, d)
}
