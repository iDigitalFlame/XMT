//go:build !windows && !js && !plan9

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
	"syscall"

	"github.com/iDigitalFlame/xmt/cmd/filter"
)

const (
	flagUID    = 1 << 1
	flagGID    = 1 << 2
	flagChroot = 1 << 3
)

type executable struct {
	e        *exec.Cmd
	c        string
	r        *os.File
	closers  []io.Closer
	uid, gid uint32
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
func (e *executable) Pid() uint32 {
	if e.e.ProcessState != nil {
		return uint32(e.e.ProcessState.Pid())
	}
	return uint32(e.e.Process.Pid)
}

// ResumeProcess will attempt to resume the process via it's PID. This will
// attempt to resume the process using an OS-dependent syscall.
//
// This will not affect already running processes.
func ResumeProcess(p uint32) error {
	return syscall.Kill(int(p), syscall.SIGCONT)
}
func (executable) Handle() uintptr {
	return 0
}

// SuspendProcess will attempt to suspend the process via it's PID. This will
// attempt to suspend the process using an OS-dependent syscall.
//
// This will not affect already suspended processes.
func SuspendProcess(p uint32) error {
	return syscall.Kill(int(p), syscall.SIGSTOP)
}
func (e *executable) Resume() error {
	return e.e.Process.Signal(syscall.SIGCONT)
}
func (e *executable) Suspend() error {
	return e.e.Process.Signal(syscall.SIGSTOP)
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
func (e *executable) SetUID(u int32, p *Process) {
	if u < 0 {
		p.flags, e.uid = p.flags&^flagUID, 0
	} else {
		e.uid = uint32(u)
		p.flags |= flagUID
	}
}
func (e *executable) SetGID(g int32, p *Process) {
	if g < 0 {
		p.flags, e.gid = p.flags&^flagGID, 0
	} else {
		e.gid = uint32(g)
		p.flags |= flagGID
	}
}
func (executable) SetWindowPosition(_, _ uint32) {
}
func (executable) SetNoWindow(_ bool, _ *Process) {
}
func (executable) SetDetached(_ bool, _ *Process) {
}
func (executable) SetSuspended(_ bool, _ *Process) {
}
func (executable) SetNewConsole(_ bool, _ *Process) {
}
func (e *executable) SetChroot(s string, p *Process) {
	if len(s) == 0 {
		p.flags, e.c = p.flags&^flagChroot, ""
	} else {
		e.c = s
		p.flags |= flagChroot
	}
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
	if p.flags > 0 {
		e.e.SysProcAttr = &syscall.SysProcAttr{Chroot: e.c}
		switch {
		case p.flags&flagUID != 0 && p.flags&flagGID != 0:
			e.e.SysProcAttr.Credential = &syscall.Credential{Uid: e.uid, Gid: e.gid}
		case p.flags&flagUID != 0 && p.flags&flagGID == 0:
			e.e.SysProcAttr.Credential = &syscall.Credential{Uid: e.uid, Gid: uint32(os.Getgid())}
		case p.flags&flagUID == 0 && p.flags&flagGID != 0:
			e.e.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(os.Getuid()), Gid: e.gid}
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
