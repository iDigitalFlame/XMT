package net

type Packet interface {
	ID() uint8
	Flags() Flag
	Sender() Host
	String() string
	Payload() []byte
	UnmarshalBinary([]byte) error
	MarshalBinary() ([]byte, error)
}

const (
	PacketNil uint8 = iota
	PacketRaw
	PacketMap
)
