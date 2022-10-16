//go:build !implant

// Copyright (C) 2020 - 2022 iDigitalFlame
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

package result

import (
	"io"
	"io/fs"
	"time"

	"github.com/iDigitalFlame/xmt/c2"
	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/device/regedit"
)

// Window is a struct that represents a Windows Window. The handles are the same
// for the duration of the Window's existence.
//
// Copied here from the winapi package.
type Window struct {
	Name          string
	Flags         uint8
	Handle        uintptr
	X, Y          int32
	Width, Height int32
}
type fileInfo struct {
	mod  time.Time
	name string
	size int64
	mode fs.FileMode
}

func (fileInfo) Sys() any {
	return nil
}
func (f fileInfo) Size() int64 {
	return f.size
}
func (f fileInfo) IsDir() bool {
	return f.mode.IsDir()
}
func (f fileInfo) Name() string {
	return f.name
}

// IsMinimized returns true if the Window state was minimized at the time of
// discovery.
func (w Window) IsMinimized() bool {
	return w.Flags&0x2 != 0
}

// IsMaximized returns true if the Window state was maximized at the time of
// discovery.
func (w Window) IsMaximized() bool {
	return w.Flags&0x1 != 0
}
func (f fileInfo) Mode() fs.FileMode {
	return f.mode
}
func (f fileInfo) ModTime() time.Time {
	return f.mod
}

// Pwd will parse the RvResult Packet from a MvPwd task.
//
// The return result is the current directory the client is located in.
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func Pwd(n *com.Packet) (string, error) {
	if n == nil || n.Empty() || n.Flags&com.FlagError != 0 {
		return "", c2.ErrMalformedPacket
	}
	return n.StringVal()
}

// Spawn will parse the RvResult Packet from a MvSpawn task.
//
// The return result is the new PID of the resulting Spawn operation.
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func Spawn(n *com.Packet) (uint32, error) {
	if n == nil || n.Empty() || n.Flags&com.FlagError != 0 {
		return 0, c2.ErrMalformedPacket
	}
	return n.Uint32()
}

// CheckDLL will parse the RvResult Packet from a TvCheck task.
//
// The return result is true if the DLL provided is NOT hooked. A return value
// of false indicates that the DLL memory space differs from the on-disk value,
// which is an indicator of hooks.
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func CheckDLL(n *com.Packet) (bool, error) {
	if n == nil || n.Empty() || n.Flags&com.FlagError != 0 {
		return false, c2.ErrMalformedPacket
	}
	return n.Bool()
}

// Mounts will parse the RvResult Packet from a MvMounts task.
//
// The return result is a string list of all the exposed mount points on the
// client (drive letters on Windows).
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func Mounts(n *com.Packet) ([]string, error) {
	if n == nil || n.Empty() || n.Flags&com.FlagError != 0 {
		return nil, c2.ErrMalformedPacket
	}
	var s []string
	return s, data.ReadStringList(n, &s)
}

// IsDebugged will parse the RvResult Packet from a TvCheckDebug task.
//
// The return result is True if a debugger is detected, false otherwise.
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func IsDebugged(n *com.Packet) (bool, error) {
	if n == nil || n.Empty() || n.Flags&com.FlagError != 0 {
		return false, c2.ErrMalformedPacket
	}
	return n.Bool()
}

// Ls will parse the RvResult Packet from a MvList task.
//
// The return result is a slice of FileInfo interfaces that will return the
// data of the directory targeted.
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func Ls(n *com.Packet) ([]fs.FileInfo, error) {
	if n == nil || n.Empty() || n.Flags&com.FlagError != 0 {
		return nil, c2.ErrMalformedPacket
	}
	c, err := n.Uint32()
	if err != nil || c == 0 {
		return nil, err
	}
	e := make([]fs.FileInfo, c)
	for i := range e {
		var v fileInfo
		if err = n.ReadString(&v.name); err != nil {
			return nil, err
		}
		if err = n.ReadUint32((*uint32)(&v.mode)); err != nil {
			return nil, err
		}
		if err = n.ReadInt64(&v.size); err != nil {
			return nil, err
		}
		t, err := n.Int64()
		if err != nil {
			return nil, err
		}
		v.mod = time.Unix(t, 0)
		e[i] = v
	}
	return e, nil
}

// Netcat will parse the RvResult Packet from a TvNetcat task.
//
// The return result is a Reader with the resulting output data from the read
// request. If reading was not done, this will just return nil.
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil.
func Netcat(n *com.Packet) (io.Reader, error) {
	if n == nil || n.Flags&com.FlagError != 0 {
		return nil, c2.ErrMalformedPacket
	}
	if n.Empty() {
		return nil, nil
	}
	return n, nil
}

// Pull will parse the RvResult Packet from a TvPull task.
//
// The return result is the expended full file path on the host as a string, and
// the resulting count of bytes written to disk.
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func Pull(n *com.Packet) (string, uint64, error) {
	return Upload(n)
}

// WindowList will parse the RvResult Packet from a TvWindows task.
//
// The return result is a slice of 'Window' structs that will indicate
// the current Windows open on the target device.
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func WindowList(n *com.Packet) ([]Window, error) {
	if n == nil || n.Empty() || n.Flags&com.FlagError != 0 {
		return nil, c2.ErrMalformedPacket
	}
	c, err := n.Uint32()
	if err != nil {
		return nil, err
	}
	e := make([]Window, c)
	for i := range e {
		if err = e[i].UnmarshalStream(n); err != nil {
			return nil, err
		}
	}
	return e, nil
}

// ScreenShot will parse the RvResult Packet from a TvScreenShot task.
//
// The return result is a Reader with the resulting screenshot data encoded as
// a PNG image inside. (This can be directly written to disk as a PNG file).
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func ScreenShot(n *com.Packet) (io.Reader, error) {
	return ProcessDump(n)
}

// Script will parse the RvResult Packet from a MvScript task.
//
// The return result is a slice of the resulting Packet output. Some flags may
// have their error values set, so it is important to check beforehand.
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func Script(n *com.Packet) ([]*com.Packet, error) {
	if n == nil || n.Empty() || n.Flags&com.FlagError != 0 {
		return nil, c2.ErrMalformedPacket
	}
	var (
		d   []byte
		r   []*com.Packet
		e   bool
		err error
	)
	for err == nil {
		var v com.Packet
		if err = n.ReadUint8(&v.ID); err != nil {
			break
		}
		if err = n.ReadBool(&e); err != nil {
			break
		}
		if !e {
			var m string
			if err = n.ReadString(&m); err != nil {
				break
			}
			v.Flags = com.FlagError
			v.WriteString(m)
			r = append(r, &v)
			continue
		}
		if err = n.ReadBytes(&d); err != nil && err != io.EOF {
			break
		}
		v.Grow(len(d))
		v.Write(d)
		d, r, err = nil, append(r, &v), nil
	}
	if err == io.EOF {
		return r, nil
	}
	return r, err
}

// ProcessDump will parse the RvResult Packet from a TvProcDump task.
//
// The return result is a Reader with the resulting dump data inside.
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func ProcessDump(n *com.Packet) (io.Reader, error) {
	if n == nil || n.Empty() || n.Flags&com.FlagError != 0 {
		return nil, c2.ErrMalformedPacket
	}
	return n, nil
}

// Upload will parse the RvResult Packet from a TvUpload task.
//
// The return result is the expended full file path on the host as a string, and
// the resulting count of bytes written to disk.
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func Upload(n *com.Packet) (string, uint64, error) {
	if n == nil || n.Empty() || n.Flags&com.FlagError != 0 {
		return "", 0, c2.ErrMalformedPacket
	}
	var (
		s   string
		c   uint64
		err = n.ReadString(&s)
	)
	if err != nil {
		return "", 0, err
	}
	if err = n.ReadUint64(&c); err != nil {
		return "", 0, err
	}
	return s, c, nil
}

// UnmarshalStream transforms this struct from a binary format that is read from
// the supplied data.Reader.
func (w *Window) UnmarshalStream(r data.Reader) error {
	v, err := r.Uint64()
	if err != nil {
		return err
	}
	w.Handle = uintptr(v)
	if err = r.ReadString(&w.Name); err != nil {
		return err
	}
	if err = r.ReadUint8(&w.Flags); err != nil {
		return err
	}
	if err = r.ReadInt32(&w.X); err != nil {
		return err
	}
	if err = r.ReadInt32(&w.Y); err != nil {
		return err
	}
	if err = r.ReadInt32(&w.Width); err != nil {
		return err
	}
	if err = r.ReadInt32(&w.Height); err != nil {
		return err
	}
	return nil
}

// UserLogins will parse the RvResult Packet from a TvLogins task.
//
// The return result is a slice of 'device.Login' structs that will indicate
// the current active Sessions (Logins) on the target device.
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func UserLogins(n *com.Packet) ([]device.Login, error) {
	if n == nil || n.Empty() || n.Flags&com.FlagError != 0 {
		return nil, c2.ErrMalformedPacket
	}
	c, err := n.Uint16()
	if err != nil {
		return nil, err
	}
	e := make([]device.Login, c)
	for i := range e {
		if err = e[i].UnmarshalStream(n); err != nil {
			return nil, err
		}
	}
	return e, nil
}

// DLL will parse the RvResult Packet from a TvDLL task.
//
// The return result is a handle to the memory location of the DLL (as an
// uintptr), the resulting PID of the DLL "host" and the exit code of the
// primary thread (if wait was specified, otherwise this is zero).
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func DLL(n *com.Packet) (uintptr, uint32, int32, error) {
	return Assembly(n)
}

// ProcessList will parse the RvResult Packet from a TvProcList task.
//
// The return result is a slice of 'cmd.ProcessInfo' structs that will indicate
// the current processes running on the target device.
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func ProcessList(n *com.Packet) ([]cmd.ProcessInfo, error) {
	if n == nil || n.Empty() || n.Flags&com.FlagError != 0 {
		return nil, c2.ErrMalformedPacket
	}
	c, err := n.Uint32()
	if err != nil {
		return nil, err
	}
	e := make([]cmd.ProcessInfo, c)
	for i := range e {
		if err = e[i].UnmarshalStream(n); err != nil {
			return nil, err
		}
	}
	return e, nil
}

// SystemIO will parse the RvResult Packet from a TvSystemIO task.
//
// The return result is dependent on the resulting operation. If the result is
// from a 'Move' or 'Copy' operation, this will return the resulting path and
// new file size.
//
// The boolean value will return true if the result was a valid command that
// returns no output, such as a Touch, Delete or Kill operation.
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func SystemIO(n *com.Packet) (string, uint64, bool, error) {
	if n == nil || n.Empty() || n.Flags&com.FlagError != 0 {
		return "", 0, false, c2.ErrMalformedPacket
	}
	o, err := n.Uint8()
	if err != nil {
		return "", 0, false, c2.ErrMalformedPacket
	}
	if o != 2 && o != 3 {
		return "", 0, true, nil
	}
	var (
		i uint64
		v string
	)
	if err = n.ReadString(&v); err != nil {
		return "", 0, false, err
	}
	if err = n.ReadUint64(&i); err != nil {
		return "", 0, false, err
	}
	return v, i, true, nil
}

// Registry will parse the RvResult Packet from a TvRegistry task.
//
// The return result is dependent on the resulting operation. If the result is
// from a 'RegLs' or 'RegGet' operation, this will return the resulting entries
// found (only one entry if this was a Get operation).
//
// The boolean value will return true if the result was a valid registry command
// that returns no output, such as a Set operation.
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func Registry(n *com.Packet) ([]regedit.Entry, bool, error) {
	if n == nil || n.Empty() || n.Flags&com.FlagError != 0 {
		return nil, false, c2.ErrMalformedPacket
	}
	o, err := n.Uint8()
	if err != nil {
		return nil, false, c2.ErrMalformedPacket
	}
	if o > 1 {
		return nil, o < 13, nil
	}
	var c uint32
	if o == 0 {
		if err = n.ReadUint32(&c); err != nil {
			return nil, false, err
		}
		if c == 0 {
			return nil, false, nil
		}
	} else {
		c = 1
	}
	r := make([]regedit.Entry, c)
	for i := range r {
		if err = r[i].UnmarshalStream(n); err != nil {
			return nil, false, err
		}
	}
	return r, true, nil
}

// Assembly will parse the RvResult Packet from a TvAssembly task.
//
// The return result is a handle to the memory location of the Assembly code (as
// an uintptr), the resulting PID of the Assembly "host" and the exit code of the
// primary thread (if wait was specified, otherwise this is zero).
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func Assembly(n *com.Packet) (uintptr, uint32, int32, error) {
	if n == nil || n.Empty() || n.Flags&com.FlagError != 0 {
		return 0, 0, 0, c2.ErrMalformedPacket
	}
	var (
		h   uint64
		p   uint32
		x   int32
		err = n.ReadUint64(&h)
	)
	if err != nil {
		return 0, 0, 0, err
	}
	if err = n.ReadUint32(&p); err != nil {
		return 0, 0, 0, err
	}
	if err = n.ReadInt32(&x); err != nil {
		return 0, 0, 0, err
	}
	return uintptr(h), p, x, nil
}

// Zombie will parse the RvResult Packet from a TvZombie task.
//
// The return result is the spawned PID of the new process and the resulting
// exit code and Stdout/Stderr data (if wait was specified, otherwise this the
// return code is zero and the reader will be empty).
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func Zombie(n *com.Packet) (uint32, int32, io.Reader, error) {
	return Process(n)
}

// Process will parse the RvResult Packet from a TvExecute task.
//
// The return result is the spawned PID of the new process and the resulting
// exit code and Stdout/Stderr data (if wait was specified, otherwise this the
// return code is zero and the reader will be empty).
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func Process(n *com.Packet) (uint32, int32, io.Reader, error) {
	if n == nil || n.Empty() || n.Flags&com.FlagError != 0 {
		return 0, 0, nil, c2.ErrMalformedPacket
	}
	var (
		p   uint32
		x   int32
		err = n.ReadUint32(&p)
	)
	if err != nil {
		return 0, 0, nil, err
	}
	if err = n.ReadInt32(&x); err != nil {
		return 0, 0, nil, err
	}
	return p, x, n, nil
}

// PullExec will parse the RvResult Packet from a TvPullExecute task.
//
// The return result is the spawned PID of the new process and the resulting
// exit code and Stdout/Stderr data (if wait was specified, otherwise this the
// return code is zero and the reader will be empty).
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func PullExec(n *com.Packet) (uint32, int32, io.Reader, error) {
	return Process(n)
}

// UserProcessList will parse the RvResult Packet from a TvLoginsProc task.
//
// The return result is a slice of 'cmd.ProcessInfo' structs that will indicate
// the current processes running on the target device.
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func UserProcessList(n *com.Packet) ([]cmd.ProcessInfo, error) {
	return ProcessList(n)
}

// Download will parse the RvResult Packet from a TvDownload task.
//
// The return result is the expended full file path on the host as a string,
// a boolean representing if the path requested is a directory (true if the path
// is a directory, false otherwise), the size of the data in bytes (zero if the
// target is a directory) and a reader with the resulting file data (empty if
// the target is a directory).
//
// This function returns an error if any reading errors occur, the Packet is not
// in the expected format or the Packet is nil or empty.
func Download(n *com.Packet) (string, bool, uint64, io.Reader, error) {
	if n == nil || n.Empty() || n.Flags&com.FlagError != 0 {
		return "", false, 0, nil, c2.ErrMalformedPacket
	}
	var (
		s   string
		d   bool
		c   uint64
		err = n.ReadString(&s)
	)
	if err != nil {
		return "", false, 0, nil, err
	}
	if err = n.ReadBool(&d); err != nil {
		return "", false, 0, nil, err
	}
	if err = n.ReadUint64(&c); err != nil {
		return "", false, 0, nil, err
	}
	return s, d, c, n, nil
}
