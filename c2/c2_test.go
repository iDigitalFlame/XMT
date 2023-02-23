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
	"testing"
	"time"

	"github.com/PurpleSec/logx"
	"github.com/iDigitalFlame/xmt/c2/cfg"
	"github.com/iDigitalFlame/xmt/c2/task"
	"github.com/iDigitalFlame/xmt/com"
)

func TestC2(t *testing.T) {
	s := NewServer(logx.NOP)
	s.New = func(x *Session) {
		t.Logf("TestC2(): New Session %s connected!", x.ID)
		j, err := x.Task(task.Whoami())
		if err != nil {
			t.Fatalf("TestC2(): %s Job creation for failed: %s!", x.ID, err.Error())
		}
		j.Update = func(z *Job) {
			if z.Status == StatusError {
				t.Fatalf("TestC2(): %s Job ID %d returned an error: %s!", x.ID, z.ID, z.Error)
			}
			t.Logf("TestC2(): %s Job ID %d returned!", x.ID, j.ID)
		}
		t.Logf("TestC2(): %s Job ID %d created!", x.ID, j.ID)
	}

	c1, err := s.Listen("test1_tcp", "localhost:9091", cfg.Static{L: com.TCP})
	if err != nil {
		t.Fatalf("TestC2(): Listen test1_tcp failed with an error: %s!", err.Error())
	}
	c2, err := s.Listen("test2_tcp", "localhost:9092", cfg.Static{L: com.TCP})
	if err != nil {
		t.Fatalf("TestC2(): Listen test1_udp failed with an error: %s!", err.Error())
	}

	v1, err := Connect(logx.NOP, cfg.Static{H: "localhost:9091", C: com.TCP, S: time.Second * 2})
	if err != nil {
		t.Fatalf("TestC2(): Connect test1_tcp failed with an error: %s!", err.Error())
	}

	time.Sleep(time.Second * 3)

	for _, v := range s.Sessions() {
		v.Close()
	}

	v1.Wait()
	c1.Close()
	c2.Close()
}
