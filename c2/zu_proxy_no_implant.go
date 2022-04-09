//go:build !implant && !noproxy

package c2

func (p *Proxy) prefix() string {
	return p.parent.ID.String() + "/P"
}
