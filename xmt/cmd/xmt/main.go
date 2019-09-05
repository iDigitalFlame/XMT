package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/iDigitalFlame/logx/logx"
	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/com/c2"
	"github.com/iDigitalFlame/xmt/xmt/com/c2/tcp"
	"github.com/iDigitalFlame/xmt/xmt/com/c2/udp"
	"github.com/iDigitalFlame/xmt/xmt/crypto"
	"github.com/iDigitalFlame/xmt/xmt/crypto/cbk"
	"github.com/iDigitalFlame/xmt/xmt/device"
	"github.com/iDigitalFlame/xmt/xmt/util"
)

type cw bool
type uu bool

func (q cw) Read(b []byte) ([]byte, error) {
	//fmt.Printf("rd[%s]\n", b)
	i := make([]byte, base64.StdEncoding.DecodedLen(len(b)))
	if _, err := base64.StdEncoding.Decode(i, b); err != nil {
		return nil, err
	}
	return i, nil
}
func (q cw) Write(b []byte) ([]byte, error) {
	o := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(o, b)
	//fmt.Printf("wr[%s]\n", o)
	return o, nil
}

func (u uu) Wrap(w io.WriteCloser) io.WriteCloser {
	x, _ := cbk.NewCipherEx(190, 32, crypto.NewSource("password123"))
	x.A = 120
	x.B = 90
	x.C = 10
	return crypto.NewWriter(x, w)
}
func (u uu) Unwrap(r io.ReadCloser) io.ReadCloser {
	x, _ := cbk.NewCipherEx(190, 32, crypto.NewSource("password123"))
	x.A = 120
	x.B = 90
	x.C = 10
	return crypto.NewReader(x, r)
}

func main2() {
	u, err := url.Parse(os.Args[1])
	fmt.Printf("%s | %s |%s |%s(%s)\n", u.Scheme, u.Host, u.Path, u.Port(), err)
}

func main() {
	s := "/api/%10s/"

	if len(os.Args) >= 2 {
		s = os.Args[1]
	}

	c2.Controller.Log = logx.NewConsole(logx.LTrace)
	c := &c2.CustomProfile{
		CSleep:     time.Duration(5) * time.Second,
		CJitter:    0,
		CWrapper:   uu(true),
		CTransform: cw(true),
	}

	p := util.Matcher(s)
	a := util.Matcher("Firefox-%10fs-%d.%100d")

	m, _ := p.Match()
	o, _ := a.Match()

	w := tcp.NewWeb(time.Duration(10)*time.Second, nil)
	w.Handle("/", http.FileServer(http.Dir("/tmp/")))
	w.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello! %v", r)
	})

	w.Rule(&tcp.WebRule{
		URL:   m,
		Agent: o,
	})
	w.Generator = &tcp.WebGenerator{
		URL:   p,
		Agent: a,
	}

	u, err := c2.Controller.Listen("http", "127.0.0.1:8080", w, c)
	if err != nil {
		panic(err)
	}

	u.Receive = func(s *c2.Session, p *com.Packet) {
		fmt.Printf("Received packet[%s]\n", p)
		if p.ID == 980 {
			i, err := p.UTFString()
			fmt.Printf("Payload: %s: %s\n", i, err)

			e := &com.Packet{ID: 450}
			e.WriteString("Hello from server!")
			e.Close()
			s.WritePacket(e)
		}
	}

	time.Sleep(time.Second * 1)

	y := &tcp.WebClient{
		Generator: &tcp.WebGenerator{
			URL:   p,
			Agent: a,
		},
	}

	v, err := c2.Controller.Connect("http://127.0.0.1:8080", y, c)
	if err != nil {
		panic(err)
	}
	v.Receive = func(s *c2.Session, p *com.Packet) {
		fmt.Printf("Received packet[%s]\n", p)
		if p.ID == 450 {
			i, err := p.UTFString()
			fmt.Printf("Payload: %s: %s\n", i, err)
		}
	}
	_, err = v.Proxy("127.0.0.1:9090", udp.Raw, c)
	if err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)

	t := &com.Packet{ID: 980}
	t.WriteString("Hello from client!")
	t.Close()
	v.WritePacket(t)

	c2.Controller.Wait()
}

func main12() {

	if len(os.Args) < 2 {
		fmt.Printf("%s <1|2|3>\n", os.Args[0])
		os.Exit(1)
	}

	c2.Controller.Log = logx.NewConsole(logx.LDebug)

	p := &c2.CustomProfile{
		CSleep:     time.Duration(5) * time.Second,
		CJitter:    0,
		CWrapper:   uu(true),
		CTransform: cw(true),
	}
	p1 := &c2.CustomProfile{
		CSleep:  time.Duration(5) * time.Second,
		CJitter: 0,
	}

	switch os.Args[1] {
	case "1":
		tl, err := c2.Controller.Listen("MyTCP", ":8080", tcp.Raw, p)
		if err != nil {
			panic(err)
		}
		tl.Oneshot = func(s *c2.Session, p *com.Packet) {
			fmt.Printf("TL: %v oneshot %s\n", s, p.String())
		}
		tl.Receive = func(s *c2.Session, p *com.Packet) {
			if s != nil {
				fmt.Printf("[%s] sent: %s\n", s.ID, p.String())
				if p.ID == 123 {
					s.WritePacket(&com.Packet{ID: 456})
				}
			}
		}

		r := bufio.NewReader(os.Stdin)
		go func(t *c2.Handle) {
			for {
				fmt.Printf("Gathering current sessions:\n")
				for _, v := range t.Sessions() {
					fmt.Printf("Session: %s (Created: %s, Last: %s)\n", v.ID, v.Created, v.Last)
				}
				fmt.Println()
				time.Sleep(time.Second * 10)
			}
		}(tl)

		for {
			fmt.Printf("\nShutdown session? ")
			v, err := r.ReadString('\n')
			if err != nil {
				continue
			}
			i, err := device.IDFromString(strings.TrimSpace(v))
			if err != nil {
				continue
			}
			fmt.Printf("\nSelected: [%s]\n", i.ID())
			s := tl.Session(i)
			if s == nil {
				continue
			}
			s.Shutdown()
		}

	case "2":
		tc, err := c2.Controller.Connect("127.0.0.1:8080", tcp.Raw, p)
		if err != nil {
			panic(err)
		}
		tc.Receive = func(s *c2.Session, p *com.Packet) {
			fmt.Printf("TC %s: got %s\n", s.ID.ID(), p.String())
		}
		tc.Times(-1, 10)
		_, err = tc.Proxy("127.0.0.1:9090", udp.Raw, p1)
		if err != nil {
			panic(err)
		}
	case "3":
		tpc, err := c2.Controller.Connect("127.0.0.1:9090", udp.Raw, p)
		if err != nil {
			panic(err)
		}
		tpc.Receive = func(s *c2.Session, p *com.Packet) {
			fmt.Printf("TCP %s: got %s\n", s.ID.ID(), p.String())
		}
		tpc.Times(-1, 10)
		time.Sleep(5 * time.Second)
		tpc.WritePacket(&com.Packet{ID: 123})
		tpc.WritePacket(&com.Packet{ID: 123})
		tpc.WritePacket(&com.Packet{ID: 123})
		tpc.WritePacket(&com.Packet{ID: 123})
		tpc.WritePacket(&com.Packet{ID: 123})
	case "4":
		if err := c2.Controller.Oneshot("127.0.0.1:9090", udp.Raw, p1, &com.Packet{ID: 990}); err != nil {
			panic(err)
		}
	case "5":
		o := tcp.NewWeb(time.Duration(5)*time.Second, nil)
		h, err := o.Listen("127.0.0.1:8080")
		if err != nil {
			panic(err)
		}
		h1, err := o.Listen("127.0.0.1:9090")
		if err != nil {
			panic(err)
		}
		fmt.Printf("http server: %v %v\n", h, h1)
	default:
		os.Exit(-1)
	}

	c2.Controller.Wait()
}
