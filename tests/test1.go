package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PurpleSec/logx"
	"github.com/iDigitalFlame/xmt/c2"
	"github.com/iDigitalFlame/xmt/c2/task"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/nakabonne/gosivy/agent"
)

func main() {

	if err := agent.Listen(agent.Options{}); err != nil {
		log.Fatal(err)
	}
	defer agent.Close()

	logx.Global.SetLevel(logx.Debug)

	var (
		s = c2.NewServer(logx.Global)
		c = c2.Config{
			c2.Sleep(time.Second * 5),
			c2.Jitter(10),
			c2.ConnectTCP,
			c2.WrapBase64,
			c2.WrapZlib,
		}
	)

	p, err := c.Profile()
	if err != nil {
		panic(err)
	}

	e := make(chan os.Signal, 1)
	signal.Notify(e, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	go func() {
		fmt.Printf("GOT SIGNAL %s\n", <-e)
		s.Close()
	}()

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

	time.Sleep(5 * time.Second)
	if err := c.Write(&com.Packet{ID: 0xDF}); err != nil {
		panic(err)
	}

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

	l.New = func(v *c2.Session) {
		var (
			pf = &task.Process{
				Dir:     "/tmp",
				Args:    []string{"ping", "-c", "2", "google.com"},
				Wait:    true,
				Timeout: time.Second * 10,
			}
			df = &com.Packet{ID: task.TvExecute}
		)
		pf.MarshalStream(df)

		j, err := v.Schedule(df)
		if err != nil {
			panic(err)
		}

		fmt.Printf("%s JOB ID %d submitted!\n", v.ID, j.ID)

		j.Update = func(ii *c2.Job) {
			fmt.Printf("JOB ID %d done!\n", ii.ID)
			fmt.Printf("%s:\n[%s]\n", ii.Result.String(), ii.Result.Payload())
		}
	}

	l.Wait()
}
func proxyClient(s *c2.Server, p *c2.Profile) {
	c, err := s.Connect("127.0.0.1:9090", nil, p)
	if err != nil {
		panic(err)
	}

	c.Wait()
}
