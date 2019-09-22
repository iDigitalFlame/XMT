package action

import (
	"github.com/iDigitalFlame/xmt/xmt/c2"
	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/data"
)

// Parameter is a struct that contains a changable
// option parameter. This option is held in string format, but
// can represent any type of value.
type Parameter struct {
	Value string

	req  bool
	name string
}

// Instruction is an interface that allows for building a packet
// that contains a instruction set or action to execute. This interface will
// be used to execute and validate the results of the function.
type Instruction interface {
	Valid() error
	String() string
	Get() []*Parameter
	Set([]*Parameter) error
	Children() []Instruction
	Instruct(*c2.Session) error
	Generate() (*com.Packet, error)
	SetParameter(string, *Parameter) error
	GetParameter(string) (*Parameter, error)
	Execute(*c2.Session, data.Reader, data.Writer) error
	Result(*c2.Session, data.Reader) (bool, string, error)
}

func (p *Parameter) Clear() {
	p.Value = ""
}
func (p *Parameter) Name() string {
	return p.name
}
func (p *Parameter) IsEmpty() bool {
	return len(p.Value) == 0
}
func (p *Parameter) Required() bool {
	return p.req
}

func NewParameter(name, def string, req bool) *Parameter {
	return &Parameter{
		req:   req,
		name:  name,
		Value: def,
	}
}
