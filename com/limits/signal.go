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
	"os"
	"os/signal"

	// Importing runtime to "load in" the handler functions
	_ "runtime"
	"sync"
	"sync/atomic"
	"time"

	// Importing unsafe to use the linkname call
	_ "unsafe"
)

var (
	watchChan    chan struct{}
	watchStart   sync.Once
	watchStarted uint32
)

//go:linkname watchSignalLoop os/signal.watchSignalLoop
var watchSignalLoop func()

func watchSignals() {
	go func() {
		<-watchChan
		atomic.StoreUint32(&watchStarted, 2)
		signalSend(0)
	}()
	for {
		s := signalRecv()
		if s == 0 && atomic.LoadUint32(&watchStarted) == 2 {
			break
		}
		process(convertSignal(s))
	}
	close(watchChan)
}
func startSignals() {
	atomic.StoreUint32(&watchStarted, 1)
	watchChan = make(chan struct{})
	watchSignalLoop = watchSignals
	signalEnable(0)
}

//go:linkname signalRecv os/signal.signal_recv
func signalRecv() uint32

//go:linkname signalEnable os/signal.enableSignal
func signalEnable(uint32)

//go:linkname process os/signal.process
func process(sig os.Signal)

//go:linkname signalSend runtime.sigsend
func signalSend(uint32) bool

// StopNotify will stop the signal handling loop from running and will cause
// all signal handling to stop.
//
// This function will block until the Goroutine closes.
//
// This function has no effect if the loop is not started or stopped.
//
// The supplied chan can be nil but if non-nil will be passed to 'signal.Stop'
// for convince.
func StopNotify(c chan<- os.Signal) {
	if c != nil {
		signal.Stop(c)
	}
	if atomic.LoadUint32(&watchStarted) == 1 {
		watchChan <- struct{}{}
		// NOTE(dij): Add a small NOP here, so we don't pull the value out the
		//            channel that's made to signal the other thread. We technically
		//            could call GoYield(), but this might be easier.
		time.Sleep(time.Millisecond * 500)
		<-watchChan
	}
}

// Notify causes package signal to relay incoming signals to c.
// If no signals are provided, all incoming signals will be relayed to c.
// Otherwise, just the provided signals will.
//
// Package signal will not block sending to c: the caller must ensure
// that c has sufficient buffer space to keep up with the expected
// signal rate. For a channel used for notification of just one signal value,
// a buffer of size 1 is sufficient.
//
// It is allowed to call Notify multiple times with the same channel:
// each call expands the set of signals sent to that channel.
// The only way to remove signals from the set is to call Stop.
//
// It is allowed to call Notify multiple times with different channels
// and the same signals: each channel receives copies of incoming
// signals independently.
//
// This version will stop the signal handling loop once the 'StopNotify'
// function has been called.
func Notify(c chan<- os.Signal, s ...os.Signal) {
	watchStart.Do(startSignals)
	signal.Notify(c, s...)
}
