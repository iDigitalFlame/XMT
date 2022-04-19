//go:build !implant && !noproxy

package c2

func (p *Proxy) prefix() string {
	if len(p.name) == 0 {
		return p.parent.ID.String() + "/P"
	}
	return p.parent.ID.String() + "/P(" + p.name + ")"
}
