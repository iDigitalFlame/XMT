//go:build windows
// +build windows

package cmd

import (
	"context"
	"sync/atomic"

	"github.com/iDigitalFlame/xmt/util/xerr"
	"golang.org/x/sys/windows"
)

// Start will attempt to start the Zombie and will return an errors that occur
// while starting the Process.
//
// This function will return 'ErrEmptyCommand' if the 'Args', the 'Data' or
// the 'Path; parameters are empty and 'ErrAlreadyStarted' if attempting to
// start a Zombie that already has been started previously.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func (z *Zombie) Start() error {
	if z.t.Running() {
		return ErrAlreadyStarted
	}
	if len(z.Args) == 0 || (len(z.Path) == 0 && len(z.Data) == 0) {
		return ErrEmptyCommand
	}
	if z.ctx == nil {
		z.ctx = context.Background()
	}
	z.cancel = func() {}
	if atomic.StoreUint32(&z.cookie, 0); z.reader != nil {
		z.reader.Close()
		z.reader = nil
	}
	z.ch = make(chan finished)
	z.flags |= windows.CREATE_SUSPENDED
	if err := z.start(false); err != nil {
		return z.stopWith(exitStopped, err)
	}
	if len(z.Data) > 0 {
		if err := z.t.Start(z.Process.opts.info.Process, z.Timeout, 0, z.Data); err != nil {
			return z.stopWith(exitStopped, z.t.stopWith(exitStopped, err))
		}
	} else {
		var b []byte
		if loadLibFunc == "LoadLibraryW" {
			p, err := windows.UTF16FromString(z.Path)
			if err != nil {
				return xerr.Wrap("could not convert path", err)
			}
			b = make([]byte, len(p)*2)
			for i := 0; i < len(b); i += 2 {
				b[i], b[i+1] = byte(p[i/2]), byte(p[i/2]>>8)
			}
		} else {
			b = append([]byte(z.Path), 0)
		}
		if err := z.t.Start(z.opts.info.Process, z.Timeout, windows.Handle(funcLoadLibrary.Addr()), b); err != nil {
			return z.stopWith(exitStopped, z.t.stopWith(exitStopped, err))
		}
	}
	go func() {
		z.t.cb = func() {
			z.stopWith(z.t.exit, z.t.err)
		}
		z.t.wait()
	}()
	return nil
}
