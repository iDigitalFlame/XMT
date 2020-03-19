package c2

import (
	"io"

	"github.com/iDigitalFlame/xmt/com"
)

type proxy struct{}

func (*proxy) close() {

}
func (*proxy) process() {

}
func (*proxy) accept(n *com.Packet) bool {
	/*if s.proxies != nil && len(s.proxies) > 0 {
		if c, ok := s.proxies[p.Device.Hash()]; ok {
			c.send <- p
			c.ready = true
			return nil
		}
	}*/
	return false
}

//	Register   func(*Session)

//	OnOneshot    func(*com.Packet)

func ReadPacket(c io.Reader, w Wrapper, t Transform) (*com.Packet, error) {
	return readPacket(c, w, t)
}
func WritePacket(c io.Writer, w Wrapper, t Transform, p *com.Packet) error {
	return writePacket(c, w, t, p)
}

/*/ Move to Proxy??
if len(s.delete) > 0 {
	for x := 0; len(s.delete) > 0; x++ {
		delete(s.proxies, <-s.delete)
	}
}
if len(s.new) > 0 {
	var n *proxyClient
	for x := 0; len(s.new) > 0; x++ {
		n = <-s.new
		s.proxies[n.hash] = n
	}
}*/
