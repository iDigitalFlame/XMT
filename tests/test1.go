package main

import (
	"fmt"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util"
)

var maxSize = 50

func main221() {
	/*
		s, d, d1 := &data.Chunk{}, &data.Chunk{Limit: 12}, &data.Chunk{Limit: 16}

		_, err := s.Write([]byte("this is a very long string value!"))
		fmt.Printf("first %s\n", err)

		n, err := s.WriteTo(d)
		fmt.Printf("read %d, %s\n", n, err)

		n, err = s.WriteTo(d1)
		fmt.Printf("read2 %d, %s\n", n, err)

		fmt.Printf("out:\n1: %s\n2: %s\n3: %s\n", s, d, d1)
		//fmt.Printf("%d %d %d %d %d %d", 32/maxSize, 64/maxSize, 48/maxSize, 128/maxSize, 256/maxSize, 100/maxSize)
	*/
	c := make(chan *com.Packet, 100)

	p := &com.Packet{ID: 16, Job: 32}

	p.WriteString("abcdefghijklmnopqrstuvwxyz0123456789")
	p.WriteString("abcdefghijklmnopqrstuvwxyz0123456789")
	p.WriteString("abcdefghijklmnopqrstuvwxyz0123456789")
	p.WriteString("abcdefghijklmnopqrstuvwxyz0123456789")
	fmt.Println(p.WriteString("12384972984789231"))
	fmt.Println(p.WriteString("12384972984789231"))

	fmt.Printf("packet: %d, %d, %s\n", p.Size(), p.Len(), p)

	add(c, p)

	p.Rewind()

	//add(c, p)

	g := &com.Packet{ID: 16}

	for len(c) > 0 {
		x := <-c
		if g == nil {
			g = x
			g.Limit = 0
		} else {
			if err := g.Add(x); err != nil {
				panic(err)
			}
		}
		fmt.Printf("packet: %s/%d [%s]\n", x, x.Len(), x.Payload())
	}

	fmt.Printf("packet: %d, %d, %s\n", g.Size(), g.Len(), g)

}

func add(z chan *com.Packet, p *com.Packet) {
	if p.Len() > maxSize {
		var (
			t = int64(0)
			x = int64(p.Len())
			m = (p.Len() / maxSize) + 1
			g = uint16(util.Rand.Uint32())
		)
		for i := 0; i < m && t < x; i++ {

			fmt.Printf("%d(%d)]: %d - %d (%d = %d)\n", i, m, i*maxSize, (i+1)*maxSize, t, x)

			c := &com.Packet{ID: p.ID, Job: p.Job, Flags: p.Flags, Chunk: data.Chunk{Limit: maxSize}}

			c.Flags.SetGroup(g)
			c.Flags.SetLen(uint16(m))
			c.Flags.SetPosition(uint16(i))

			n, err := p.WriteTo(c)
			if err != nil && err != data.ErrLimit {
				c.Flags.SetLen(0)
				c.Flags.SetPosition(0)
				c.Flags.Set(com.FlagError)
				break
			}

			t += n
			z <- c
		}
	} else {
		z <- p
	}
}

func proc(c chan *com.Packet) {

}
