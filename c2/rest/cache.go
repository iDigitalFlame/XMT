package rest

import (
	"context"
	"sync"
	"time"

	"github.com/iDigitalFlame/xmt/c2"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/device"
)

type cache struct {
	osLock, devLock, jobLock sync.RWMutex
	dev                      map[string]*c2.Session
	job                      map[device.ID]map[uint16]*c2.Job
	os                       []packet
}
type packet struct {
	t time.Time
	p *com.Packet
}

func (c *cache) new(s *c2.Session) {
	if s == nil {
		return
	}
	c.devLock.Lock()
	c.dev[s.ID.String()] = s
	c.devLock.Unlock()
}
func (c *cache) complete(j *c2.Job) {
	if j.Status < c2.StatusCompleted {
		return
	}
	c.jobLock.Lock()
	if m, ok := c.job[j.Session.ID]; !ok {
		c.job[j.Session.ID] = map[uint16]*c2.Job{j.ID: j}
	} else {
		if _, ok := m[j.ID]; !ok {
			m[j.ID] = j
		}
	}
	c.jobLock.Unlock()
}
func (c *cache) catch(n *com.Packet) {
	if n == nil || n.ID == 0 || n.Flags&com.FlagOneshot == 0 {
		return
	}
	c.osLock.Lock()
	c.os = append(c.os, packet{p: n, t: time.Now()})
	c.osLock.Unlock()
}
func (c *cache) prune(x context.Context) {
	for t := time.NewTicker(time.Minute); ; {
		select {
		case n := <-t.C:
			if len(c.job) > 0 {
				c.jobLock.Lock()
				var r []device.ID
				for i, v := range c.job {
					for t, j := range v {
						if n.Sub(j.Start) < expire {
							continue
						}
						delete(v, t)
					}
					if len(v) == 0 {
						r = append(r, i)
					}
				}
				for i := range r {
					delete(c.job, r[i])
				}
				c.jobLock.Unlock()
			}
			// NOTE(dij): Removing this as it removes sessions that
			//            return a bad "Last" value (mostly due to hibernation)
			//            or time skipping.
			//if len(c.dev) > 0 {
			//	c.devLock.Lock()
			//	for i, s := range c.dev {
			//		if n.Sub(s.Last) < expire {
			//			continue
			//		}
			//		delete(c.dev, i)
			//	}
			//	c.devLock.Unlock()
			//}
		case <-x.Done():
			t.Stop()
			return
		}
	}
}
func (c *cache) device(i string) *c2.Session {
	if len(i) == 0 {
		return nil
	}
	if c.devLock.RLock(); len(c.dev) > 0 {
		if s, ok := c.dev[i]; ok {
			c.devLock.RUnlock()
			return s
		}
	}
	c.devLock.RUnlock()
	return nil
}
func (c *cache) track(i device.ID, j *c2.Job) {
	if j == nil {
		return
	}
	c.devLock.RLock()
	c.jobLock.Lock()
	if m, ok := c.job[i]; !ok {
		c.job[i] = map[uint16]*c2.Job{j.ID: j}
	} else {
		m[j.ID] = j
	}
	c.jobLock.Unlock()
	c.devLock.RUnlock()
	j.Update = c.complete
}
func (c *cache) remove(i string, s bool) bool {
	if len(i) == 0 {
		return false
	}
	c.devLock.Lock()
	c.jobLock.Lock()
	v, ok := c.dev[i]
	if ok {
		if delete(c.job, v.ID); s {
			v.Close()
		} else {
			v.Remove()
		}
		delete(c.dev, i)
	}
	c.devLock.Unlock()
	c.jobLock.Unlock()
	return ok
}
func (c *cache) retrive(i string, j uint16, d bool) *c2.Job {
	if len(i) == 0 || j == 0 {
		return nil
	}
	c.devLock.RLock()
	if c.jobLock.RLock(); len(c.dev) == 0 || len(c.job) == 0 {
		c.devLock.RUnlock()
		c.jobLock.RUnlock()
		return nil
	}
	s, ok := c.dev[i]
	c.devLock.RUnlock()
	if !ok {
		c.jobLock.RUnlock()
		return nil
	}
	l, ok := c.job[s.ID]
	if !ok || len(l) == 0 {
		c.jobLock.RUnlock()
		return nil
	}
	x, ok := l[j]
	if c.jobLock.RUnlock(); !ok {
		return nil
	}
	if !d {
		return x
	}
	c.jobLock.Lock()
	delete(l, j)
	c.jobLock.Unlock()
	return x
}
