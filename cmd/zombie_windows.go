//go:build windows

package cmd

import (
	"context"
	"sync/atomic"
)

// Start will attempt to start the Zombie and will return an errors that occur
// while starting the Process.
//
// This function will return 'ErrEmptyCommand' if the 'Args' or 'Data'
// parameters are empty and 'ErrAlreadyStarted' if attempting to
// start a Zombie that already has been started previously.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func (z *Zombie) Start() error {
	if z.t.Running() {
		return ErrAlreadyStarted
	}
	if len(z.Args) == 0 || len(z.Data) == 0 {
		return ErrEmptyCommand
	}
	if z.ctx == nil {
		z.ctx = context.Background()
	}
	z.cancel = func() {}
	z.ch = make(chan struct{})
	atomic.StoreUint32(&z.cookie, 0)
	if err := z.x.start(z.ctx, &z.Process, true); err != nil {
		return z.stopWith(exitStopped, err)
	}
	/*if len(z.Data) > 0 {*/
	if err := z.t.Start(z.x.i.Process, z.Timeout, 0, z.Data); err != nil {
		return z.stopWith(exitStopped, z.t.stopWith(exitStopped, err))
	}
	// NOTE(dij): Removing ZombieDLL support for now as it does not work
	//            correctly. I would just convert DLL->ASM and run with that.
	/*} else {
		p, err := winapi.UTF16FromString(z.Path)
		if err != nil {
			return xerr.Wrap("could not convert path", err)
		}
		b := make([]byte, (len(p)*2)+1)
		for i := 0; i < len(b)-1; i += 2 {
			b[i], b[i+1] = byte(p[i/2]), byte(p[i/2]>>8)
		}
		// BUG(dij): This is broken and I'm not sure 100% why.
		//           The library gets /loaded/ but does not execute correctly.
		if err := z.t.Start(z.x.i.Process z.Timeout, winapi.LoadLibraryAddress(), b); err != nil {
			return z.stopWith(exitStopped, z.t.stopWith(exitStopped, err))
		}
	}*/
	go func() {
		z.t.callback = func() { z.stopWith(z.t.exit, z.t.err) }
		z.t.wait(z.x.i.ProcessID, z.x.i.ThreadID)
	}()
	return nil
}
