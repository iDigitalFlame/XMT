package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/iDigitalFlame/logx/logx"
	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/com/c2"
	"github.com/iDigitalFlame/xmt/xmt/com/c2/tcp"
	"github.com/iDigitalFlame/xmt/xmt/com/c2/udp"
	"github.com/iDigitalFlame/xmt/xmt/crypto"
	"github.com/iDigitalFlame/xmt/xmt/crypto/cbk"
)

type cw bool
type uu bool

func (q cw) Read(b []byte) ([]byte, error) {
	i := make([]byte, base64.StdEncoding.DecodedLen(len(b)))
	if _, err := base64.StdEncoding.Decode(i, b); err != nil {
		return nil, err
	}
	return i, nil
}
func (q cw) Write(b []byte) ([]byte, error) {
	o := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(o, b)
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

func main() {

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
		for {
			fmt.Printf("Gathering current sessions:\n")
			for _, v := range tl.Sessions() {
				fmt.Printf("Session: %s (Created: %s, Last: %s)\n", v.ID, v.Created, v.Last)
			}
			fmt.Println()
			time.Sleep(time.Second * 10)
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
		tpc, err := c2.Controller.Connect("127.0.0.1:9090", udp.Raw, p1)
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
	default:
		os.Exit(-1)
	}

	c2.Controller.Wait()
}
