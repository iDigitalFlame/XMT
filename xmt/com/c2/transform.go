package c2

// Transform is an interface that can modify the data BEFORE
// it is written or AFTER is read from a Connection.
// Transforms may be used to mask and unmask communications
// as benign protocols such as DNS, FTP or HTTP.
type Transform interface {
	Read([]byte) ([]byte, error)
	Write([]byte) ([]byte, error)
}
type rawTransform struct{}

func (r *rawTransform) Read(b []byte) ([]byte, error) {
	return b, nil
}
func (r *rawTransform) Write(b []byte) ([]byte, error) {
	return b, nil
}
