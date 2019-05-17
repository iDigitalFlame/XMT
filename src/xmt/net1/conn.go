package net

type Server interface {
	Accept() (Connecter, error)
}

type Connecter interface {
	Write(Packet) error
	Read() (Packet, error)
}

type Transport interface {
	String() string
	Read([]byte) (int, error)
	Write([]byte) (int, error)
}

