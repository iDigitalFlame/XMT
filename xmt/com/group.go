package com

import (
	"fmt"
	"io"
)

type PacketGroup struct {
	ID    uint16
	Limit int
	Group uint16

	last    uint16
	packets []*Packet
}

func (p *PacketGroup) P() {
	for i := range p.packets {
		fmt.Printf("packet %d: [%s]\n", i, p.packets[i].buf)
	}
	p.last = 0
}

func (p *PacketGroup) New() *Packet {
	if p.packets == nil {
		p.packets = make([]*Packet, 0, 1)
	} else {
		p.packets[p.last].Close()
		p.last++
	}
	n := &Packet{ID: p.ID}
	n.Flags |= FlagFrag
	n.Flags.SetFragPosition(p.last)
	p.packets = append(p.packets, n)
	return n
}
func (p *PacketGroup) Read(b []byte) (int, error) {
	if p.packets == nil || len(p.packets) == 0 {
		return 0, io.EOF
	}
	var c int
	var n *Packet
	for ; c < len(b); p.last++ {
		if p.last >= uint16(len(p.packets)) {
			return c, io.EOF
		}
		n = p.packets[p.last]
		n.rpos = copy(b[c:], n.buf[n.rpos:])
		c += n.rpos
	}
	return c, nil
}
func (p *PacketGroup) Write(b []byte) (int, error) {
	var n *Packet
	if p.packets == nil {
		n = p.New()
	} else {
		n = p.packets[p.last]
	}
	if p.Limit == 0 {
		return n.Write(b)
	}
	if n.Len() >= p.Limit {
		n = p.New()
	}
	if len(b) < p.Limit {
		return n.Write(b)
	}
	var c int
	for a := len(b) - c; c < len(b); a = len(b) - c {
		if a <= 0 {
			break
		}
		if a > p.Limit {
			a = p.Limit
		}
		v, err := n.Write(b[c : c+int(a)])
		if err != nil {
			return c, err
		}
		c += v
		if c < len(b) {
			n = p.New()
		}
	}
	return c, nil
}
