package main

import (
	"crypto/aes"
	"fmt"
	"net/http"
	"time"

	"github.com/iDigitalFlame/logx/logx"
	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/com/c2"
	"github.com/iDigitalFlame/xmt/xmt/com/c2/tcp"
	"github.com/iDigitalFlame/xmt/xmt/com/c2/transform"
	"github.com/iDigitalFlame/xmt/xmt/com/c2/wrapper"
	"github.com/iDigitalFlame/xmt/xmt/crypto/cbk"
	"github.com/iDigitalFlame/xmt/xmt/crypto/xor"
	"github.com/iDigitalFlame/xmt/xmt/device/local"
	"github.com/iDigitalFlame/xmt/xmt/util"
)

func main() {

	testWebC2()
}

func testWebC2() {
	c2.Controller.Log = logx.NewConsole(logx.LDebug)
	c2.Controller.Mux = c2.MuxFunc(func(s *c2.Session, p *com.Packet) {
		fmt.Printf("[%s] Packet: %s\n", s.ID.String(), p.String())
		if p.ID == 910 {
			n1, _ := p.UTFString()
			n2, _ := p.UTFString()
			fmt.Printf("Payload: %s, %s\n", n1, n2)
		}
	})

	//xo := xor.Cipher([]byte("derp string"))

	//y, _ := wrapper.NewCrypto(xo, xo)

	z1, err := aes.NewCipher([]byte("123456789012345678901234567890AB"))
	if err != nil {
		panic(err)
	}
	y, _ := wrapper.NewCryptoBlock(z1, []byte("ABCDEF1234567890"))

	x := cbk.NewCipher(100)
	x.A = 20
	x.B = 30
	x.C = 40

	xr := cbk.NewCipher(100)
	xr.A = 20
	xr.B = 30
	xr.C = 40

	x1, _ := wrapper.NewCrypto(x, xr)

	xx := xor.Cipher([]byte("this is my xor key"))
	x2, _ := wrapper.NewCrypto(xx, xx)

	p := &c2.Profile{
		Sleep:     time.Duration(5) * time.Second,
		Transform: transform.DNS,
		Wrapper:   wrapper.NewMulti(wrapper.Zlib, y, x1, x2),
	}

	g := &tcp.WebGenerator{
		URL:  util.Matcher("/post/%31fd/%12fd/%31fd/edit/%4fl"),
		Host: util.Matcher("%6fl.myblogsite.com"),
	}

	s := tcp.NewWeb(p.Sleep)
	s.Handle("/", http.FileServer(http.Dir("/tmp/")))

	s.Rule(&tcp.WebRule{
		URL:  g.URL.(util.Matcher).Match(),
		Host: g.Host.(util.Matcher).Match(),
	})

	//var err error
	h, err := c2.Controller.Listen("http", "127.0.0.1:8080", s, p)
	if err != nil {
		panic(err)
	}

	time.Sleep(3 * time.Second)

	c, err := c2.Controller.Connect("http://127.0.0.1:8080", &tcp.WebClient{Generator: g}, p)
	if err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)

	o := &com.Packet{ID: 910}
	o.Flags |= com.FlagFrag
	o.Flags.SetFragTotal(2)
	o.Flags.SetFragGroup(999)
	o.Flags.SetFragPosition(0)
	o.WriteString("derp!")
	fmt.Printf("p1: %s, %s\n", o, o.Flags)
	c.WritePacket(o)

	time.Sleep(5 * time.Second)
	q := &com.Packet{ID: 910}
	q.Flags |= com.FlagFrag
	q.Flags.SetFragTotal(2)
	q.Flags.SetFragGroup(999)
	q.Flags.SetFragPosition(1)
	q.WriteString("derp!")
	fmt.Printf("p2: %s, %s\n", q, q.Flags)
	c.WritePacket(q)

	fmt.Printf("%s\n", h.Session(local.ID()))

	time.Sleep(10 * time.Second)

	c.Shutdown()

	time.Sleep(15 * time.Second)

	fmt.Printf("%s\n", h.Session(local.ID()))

	c2.Controller.Wait()
}
