//go:build js || plan9

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

package cmd

import (
	"context"
	"io"
	"os"
	"os/exec"
	"sync/atomic"

	"github.com/iDigitalFlame/xmt/cmd/filter"
)

type executable struct {
	e       *exec.Cmd
	r       *os.File
	closers []io.Closer
}

func (e *executable) close() {
	if len(e.closers) > 0 {
		for i := range e.closers {
			e.closers[i].Close()
		}
	}
	//if e.e.Process != nil {
	// NOTE(dij): This causes *nix systems to create a Zombie process
	//            (not what we want). Not sure if it matters enough to fix
	//            tho.
	// e.e.Process.Release()
	//}
}
func (executable) Resume() error {
	return nil
}
func (executable) Suspend() error {
	return nil
}
func (e *executable) Pid() uint32 {
	if e.e.ProcessState != nil {
		return uint32(e.e.ProcessState.Pid())
	}
	return uint32(e.e.Process.Pid)
}

// ResumeProcess will attempt to resume the process via its PID. This will
// attempt to resume the process using an OS-dependent syscall.
//
// This will not affect already running processes.
func ResumeProcess(_ uint32) error {
	return nil
}
func (executable) Handle() uintptr {
	return 0
}

// SuspendProcess will attempt to suspend the process via its PID. This will
// attempt to suspend the process using an OS-dependent syscall.
//
// This will not affect already suspended processes.
func SuspendProcess(_ uint32) error {
	return nil
}
func (e *executable) isStarted() bool {
	return e.e != nil && e.e.Process != nil
}
func (e *executable) isRunning() bool {
	return e.isStarted() && e.e.ProcessState == nil
}
func (executable) SetToken(_ uintptr) {
}
func (e *executable) wait(p *Process) {
	err := e.e.Wait()
	if _, ok := err.(*exec.ExitError); err != nil && !ok {
		p.stopWith(exitStopped, err)
		return
	}
	if err2 := p.ctx.Err(); err2 != nil {
		p.stopWith(exitStopped, err2)
		return
	}
	if atomic.StoreUint32(&p.cookie, 2); e.e.ProcessState != nil {
		p.exit = uint32(e.e.ProcessState.ExitCode())
	}
	if p.exit != 0 {
		p.stopWith(p.exit, &ExitError{Exit: p.exit})
		return
	}
	p.stopWith(p.exit, nil)
}
func (executable) SetFullscreen(_ bool) {
}
func (executable) SetWindowDisplay(_ int) {
}
func (executable) SetWindowTitle(_ string) {
}
func (executable) SetLogin(_, _, _ string) {}
func (executable) SetWindowSize(_, _ uint32) {
}
func (executable) SetUID(_ int32, _ *Process) {
}
func (executable) SetGID(_ int32, _ *Process) {
}
func (executable) SetWindowPosition(_, _ uint32) {
}
func (executable) SetChroot(_ string, _ *Process) {
}
func (executable) SetNoWindow(_ bool, _ *Process) {
}
func (executable) SetDetached(_ bool, _ *Process) {
}
func (executable) SetSuspended(_ bool, _ *Process) {
}
func (executable) SetNewConsole(_ bool, _ *Process) {
}
func (e *executable) kill(x uint32, p *Process) error {
	if p.exit = x; e.e == nil || e.e.Process == nil {
		return nil
	}
	return e.e.Process.Kill()
}
func (executable) SetParent(_ *filter.Filter, _ *Process) {
}
func (e *executable) start(x context.Context, p *Process, _ bool) error {
	if e.e != nil {
		return ErrAlreadyStarted
	}
	e.e = exec.CommandContext(x, p.Args[0])
	e.e.Args = p.Args
	e.e.Dir, e.e.Env = p.Dir, p.Env
	e.e.Stdin, e.e.Stdout, e.e.Stderr = p.Stdin, p.Stdout, p.Stderr
	if !p.split {
		z := os.Environ()
		if e.e.Env == nil {
			e.e.Env = make([]string, 0, len(z))
		}
		for n := range z {
			e.e.Env = append(e.e.Env, z[n])
		}
	}
	if e.r != nil {
		e.r.Close()
		e.r = nil
	}
	if err := e.e.Start(); err != nil {
		return err
	}
	go e.wait(p)
	return nil
}
