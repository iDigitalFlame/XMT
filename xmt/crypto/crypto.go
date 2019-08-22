package crypto

// Source is an interface that supports seed assistance in Ciphers and other
// cryptographic functions.
type Source interface {
	Reset() error
	Next(uint16) uint16
}
