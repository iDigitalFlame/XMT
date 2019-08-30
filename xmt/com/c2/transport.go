package c2

// Transport is an interface that can modify the data BEFORE
// it is written or AFTER is read from a Connection.
// Transports may be used to mask and unmask communications
// as benign protocols such as DNS, FTP or HTTP.
type Transport interface {
	Read([]byte) ([]byte, error)
	Write([]byte) ([]byte, error)
}
type rawTransport struct{}

func (r *rawTransport) Read(b []byte) ([]byte, error) {
	return b, nil
}
func (r *rawTransport) Write(b []byte) ([]byte, error) {
	return b, nil
}
