package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/iDigitalFlame/logx/logx"

	"github.com/iDigitalFlame/xmt/com"

	"github.com/iDigitalFlame/xmt/c2"

	"github.com/iDigitalFlame/xmt/cmd"
)

func main() {
	c2c := c2.Config{
		c2.WrapXOR([]byte("derp123")),
		c2.WrapBase64,
		c2.WrapHex,
		c2.TransformDNS("google.com", "bing.com", "lol.com", "duckduckgo.com"),
		c2.Sleep(time.Second * 10),
		c2.Jitter(30),
	}

	fmt.Printf("derp [%s]\n[%v]\n", c2c, c2c)

	p, err := c2c.Profile()
	if err != nil {
		panic(err)
	}

	fmt.Printf("pro [%+v]\n", p)

	b := new(bytes.Buffer)

	if err := c2c.Write(b); err != nil {
		panic(err)
	}

	x := b.Bytes()

	fmt.Printf("data: [%s]\n[%v]\n", x, x)

	var ff c2.Config

	y := bytes.NewReader(x)
	if err := ff.Read(y); err != nil {
		panic(err)
	}

	fmt.Printf("derp1 [%s]\n[%v]\n", ff, ff)

}

func main222() {
	http.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	http.Handle("/debug/pprof/block", pprof.Handler("block"))
	http.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	http.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	go http.ListenAndServe("127.0.0.1:9091", nil)

	logx.Global.SetLevel(logx.LTrace)

	c := c2.NewServer(logx.Global)

	c2c := c2.Config{
		c2.WrapXOR([]byte("derp123")),
		c2.WrapBase64,
		c2.WrapHex,
		c2.TransformDNS("google.com", "bing.com", "lol.com", "duckduckgo.com"),
		c2.Sleep(time.Second * 10),
		c2.Jitter(30),
	}

	fmt.Printf("derp [%s]\n", c2c)

	p, err := c2c.Profile()
	if err != nil {
		panic(err)
	}

	fmt.Printf("pro [%+v]\n", p)

	//w := wc2.New(com.DefaultTimeout)
	//w.Rule(wc2.DefaultGenerator.Rule())
	//w.ServeDirectory("/", "/tmp")

	l, err := c.Listen("test1", "127.0.0.1:8080", com.TCP, p)
	if err != nil {
		panic(err)
	}

	fmt.Printf("LISTENER: %s\n", l.String())

	s, err := c.Connect("127.0.0.1:8080", com.TCP, p)
	if err != nil {
		panic(err)
	}

	//s1, err := c.ConnectQuick("127.0.0.1:8081", com.TCP)
	//if err != nil {
	//	panic(err)
	//}

	fmt.Printf("SESSION: %s\n", s.String())

	//s.SetChannel(true)
	//fmt.Printf("SESSION: %s\n", s1.String())

	time.Sleep(time.Second * 8)

	s.WritePacket(&com.Packet{ID: 5500})

	l.Wait()

}

func main9() {
	p := &cmd.Process{
		Args:    []string{"pwsh", "-c", "while($x=$host.UI.Readline()){write-host $x -fore green}"},
		Timeout: time.Second * 60,
	}
	p.SetParentRandom(nil)

	w, err := p.StdinPipe()
	if err != nil {
		panic(err)
	}

	w.Write([]byte("lol\nrofl\n"))

	fmt.Printf("done!")
	b, err := p.CombinedOutput()

	fmt.Printf("Error: [%s]\nOutput: [%s]\n", err, b)

}
