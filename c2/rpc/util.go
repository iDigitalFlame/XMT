package rpc

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PurpleSec/escape"
	"github.com/PurpleSec/routex"
	"github.com/iDigitalFlame/xmt/c2"
	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

func stringList(i interface{}) []string {
	v, ok := i.([]interface{})
	if !ok {
		return nil
	}
	r := make([]string, 0, len(v))
	for _, e := range v {
		s, ok := e.(string)
		if !ok {
			continue
		}
		r = append(r, s)
	}
	return r
}
func parseJitter(s string) (int, error) {
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
func parseSleep(s string) (time.Duration, error) {
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
func errors(c int, e string, w http.ResponseWriter, r *routex.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(c)
	w.Write([]byte(`{"error": ` + escape.JSON(e) + `, "code": ` + strconv.Itoa(c) + `}`))
}
func (r *Server) lookup(v string, x *routex.Request, c routex.Content) (*c2.Session, *cmd.Filter, string, bool, error) {
	n, err := x.Values.String("session")
	if err != nil {
		return nil, nil, "", false, err
	}
	e, err := c.String(v)
	if err != nil {
		return nil, nil, "", false, err
	}
	if len(e) == 0 {
		return nil, nil, "", false, xerr.Wrap(v, routex.ErrEmptyValue)
	}
	s := r.session(n)
	if s == nil {
		return nil, nil, "", true, routex.ErrEmptyValue
	}
	var f *cmd.Filter
	if d, err := c.Object("filter"); err == nil {
		f = new(cmd.Filter)
		if v := d.Raw("exclude"); v != nil {
			f.Exclude = stringList(v)
		}
		if v := d.Raw("include"); v != nil {
			f.Include = stringList(v)
		}
		if v, err := d.Uint64("pid"); err == nil && v > 0 {
			f.PID = uint32(v)
		}
		if v, err := d.Bool("session"); err == nil {
			f.SetSession(v)
		}
		if v, err := d.Bool("fallback"); err == nil {
			f.SetFallback(v)
		}
		if v, err := d.Bool("elevated"); err == nil {
			f.SetElevated(v)
		}
	}
	return s, f, e, false, nil
}
