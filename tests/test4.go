package main

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"github.com/iDigitalFlame/logx/logx"

	"github.com/iDigitalFlame/xmt/c2"
)

func main() {
	logx.Global.SetLevel(logx.LTrace)

	var (
		s = c2.NewServer(logx.Global)
		c = c2.Config{c2.Sleep(time.Second * 5), c2.Jitter(30), c2.ConnectTCP}
	)

	p, err := c.Profile()
	if err != nil {
		panic(err)
	}

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "c":
			client(s, p)
		case "p":
			proxy(s, p)
		case "q":
			proxyClient(s, p)
		default:
			panic(fmt.Sprintf("Not a valid operation %q!", os.Args[1]))
		}
	} else {
		server(s, p)
	}
}

func proxy(s *c2.Server, p *c2.Profile) {
	c, err := s.Connect("127.0.0.1:8080", nil, p)
	if err != nil {
		panic(err)
	}
	fmt.Printf("New Session [%s]\n", c)
	i, err := c.Proxy("127.0.0.1:9090", nil, p)
	if err != nil {
		panic(err)
	}
	fmt.Printf("New Proxy [%v]\n", i)
	c.Wait()
}
func client(s *c2.Server, p *c2.Profile) {
	c, err := s.Connect("127.0.0.1:8080", nil, p)
	if err != nil {
		panic(err)
	}
	fmt.Printf("New Session [%s]\n", c)
	c.Wait()
}
func server(s *c2.Server, p *c2.Profile) {
	http.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	http.Handle("/debug/pprof/block", pprof.Handler("block"))
	http.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	http.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	go http.ListenAndServe("127.0.0.1:9091", nil)

	l, err := s.Listen("test1", "127.0.0.1:8080", nil, p)
	if err != nil {
		panic(err)
	}
	fmt.Printf("New Listener [%s]\n", l)
	l.Wait()
}
func proxyClient(s *c2.Server, p *c2.Profile) {
	c, err := s.Connect("127.0.0.1:9090", nil, p)
	if err != nil {
		panic(err)
	}
	c.Wait()
}
