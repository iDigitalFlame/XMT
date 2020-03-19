package main

import (
	"fmt"

	"github.com/iDigitalFlame/xmt/c2/wrapper"

	"github.com/iDigitalFlame/xmt/c2/transform"

	"github.com/iDigitalFlame/xmt/c2"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
)

func main6() {

	p := &com.Packet{ID: 1234}

	p.WriteString("cool story bro")
	p.WriteString("cool story bro")
	p.WriteString("cool story bro")
	p.WriteString("cool story bro")
	p.WriteString("cool story bro")
	p.WriteString("cool story bro")
	p.WriteString("cool story bro")
	p.WriteString("cool story bro")
	p.WriteString("cool story bro")
	p.WriteString("cool story bro")
	p.WriteString("cool story bro")
	p.WriteString("cool story bro")
	p.WriteString("cool story bro")
	p.WriteString("cool story bro")
	p.WriteString("cool story bro")
	p.WriteString("cool story bro")
	p.WriteString("cool story bro")

	g := &data.Chunk{}

	if err := c2.WritePacket(g, wrapper.Zlib, transform.Base64, p); err != nil {
		panic(err)
	}

	fmt.Printf("payload [%s]\n", g.Payload())

	r, err := c2.ReadPacket(g, wrapper.Zlib, transform.Base64)
	fmt.Println("out", r, err)

	var t int
	var f string
	for err == nil {
		if f, err = r.StringVal(); err == nil {
			fmt.Printf("read: [%s]\n", f)
			t++
		}
	}

	fmt.Printf("amount %d\n", t)
}
