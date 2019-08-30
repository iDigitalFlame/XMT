package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/iDigitalFlame/logx/logx"
	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/com/c2"
	"github.com/iDigitalFlame/xmt/xmt/com/c2/tcp"
	"github.com/iDigitalFlame/xmt/xmt/crypto"
	"github.com/iDigitalFlame/xmt/xmt/crypto/cbk"
	"github.com/iDigitalFlame/xmt/xmt/data"
	"github.com/iDigitalFlame/xmt/xmt/device"
	"github.com/iDigitalFlame/xmt/xmt/device/local"
)

func main1() {
	b := &bytes.Buffer{}

	x, err := cbk.NewCipherEx(
		25,
		32,
		crypto.NewMultiSource(
			crypto.NewSource("password"),
			crypto.NewSource("password123"),
		),
		//rand.NewSource(987651),
	)
	if err != nil {
		panic(err)
	}
	x.A = 10
	x.B = 20
	x.C = 30

	w := crypto.NewWriter(x, b)
	w.WriteBool(true)
	w.WriteInt(123466)
	w.WriteString("This is a loooooooooooooong string!")
	w.WriteUTF16String("This is a nice UTF16 string!!!!")
	w.Close()

	y, err := cbk.NewCipherEx(
		25,
		32,
		crypto.NewMultiSource(
			crypto.NewSource("password"),
			crypto.NewSource("password123"),
		),
		//rand.NewSource(987651),
	)
	if err != nil {
		panic(err)
	}
	y.A = 10
	y.B = 20
	y.C = 30

	fmt.Printf("write\n")

	r := crypto.NewReader(y, bytes.NewReader(b.Bytes()))
	t, err := r.Bool()
	if err != nil {
		panic(err)
	}
	i, err := r.Int()
	if err != nil {
		panic(err)
	}
	s, err := r.UTFString()
	if err != nil {
		panic(err)
	}
	s2, err := r.UTFString()
	if err != nil {
		panic(err)
	}

	fmt.Printf("b: %t, i: %d, s: %s, s2: %s\n", t, i, s, s2)

	op := new(device.Network)
	if err := op.Refresh(); err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", op)
}

func main2() {
	fmt.Print("%+v\n", local.Host())

	p := &com.Packet{}
	if err := data.Write(p, local.Host()); err != nil {
		panic(err)
	}
	p.Close()

	fmt.Printf("(%s) %d bytes\n", p.Device, p.Len())

	b, _ := json.Marshal(p)
	fmt.Printf("%s\n", b)

	var c *com.Packet
	if err := json.Unmarshal(b, &c); err != nil {
		panic(err)
	}

	fmt.Printf("(%s) %d bytes\n%s\n", c.Device, c.Len(), c.Payload())

}

type bbb struct{}
type zzz struct{}

func (r *bbb) Jitter() int8 {
	return 50
}
func (r *bbb) Sleep() time.Duration {
	return -1 //time.Second * 5
}
func (r *bbb) Close() error {
	return nil
}
func (r *bbb) Size() int {
	return -1
}
func (r *bbb) Wrapper() c2.Wrapper {
	return nil
}
func (r *bbb) Transport() c2.Transport {
	return &zzz{}
}
func (r *bbb) Listen(s string) (c2.Listener, error) {
	return tcp.Raw.Listen(s)
}
func (r *bbb) Connect(s string) (c2.Connection, error) {
	return tcp.Raw.Connect(s)
}

func (r *zzz) Read(b []byte) ([]byte, error) {
	o := make([]byte, base64.StdEncoding.DecodedLen(len(b)))
	n, err := base64.StdEncoding.Decode(o, b)
	return o[:n], err
}
func (r *zzz) Write(b []byte) ([]byte, error) {
	o := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(o, b)
	return o, nil
}

func main() {

	c2.Controller.Log = logx.NewConsole(logx.LTrace) //.SetLevel(logx.LTraceL

	c, err := c2.Controller.Listen("basic", "127.0.0.1:8080", &bbb{})
	if err != nil {
		panic(err)
	}

	c.Receive = func(s *c2.Session, p *com.Packet) {
		if p.ID == 123 {
			x := &com.Packet{ID: 456}
			s.WritePacket(x)
		}
	}

	v, err := c2.Controller.Connect("127.0.0.1:8080", &bbb{})
	if err != nil {
		panic(err)
	}
	v.Receive = func(p *com.Packet) {
		fmt.Printf("client received packet: %+v\n", p)
	}

	z := &com.Packet{ID: 123}
	z.WriteString("This is a secret message")
	v.WritePacket(z)
	v.Wake()

	c2.Controller.Wait()
	v.Close()
	c.Close()
}
