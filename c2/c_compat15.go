//go:build !go1.15
// +build !go1.15

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

import "time"

type sleeper struct {
	C chan time.Time
	t *time.Timer
	p time.Duration
	w chan struct{}
}

func (s *sleeper) Stop() {
	close(s.w)
}
func (s *sleeper) sleep() {
inner:
	for {
		select {
		case <-s.w:
			break inner
		case t := <-s.t.C:
			select {
			case s.C <- t:
			default:
			}
			s.t.Stop()
			s.t.Reset(s.p)
		}
	}
	s.t.Stop()
	close(s.C)
}
func (s *sleeper) Reset(d time.Duration) {
	for s.t.Stop(); len(s.t.C) > 0; {
		<-s.t.C
	}
	s.p = d
	s.t.Reset(d)
}
func newSleeper(d time.Duration) *sleeper {
	s := &sleeper{C: make(chan time.Time, 1), t: time.NewTimer(d), p: d, w: make(chan struct{})}
	go s.sleep()
	return s
}
