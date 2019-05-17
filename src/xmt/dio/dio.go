package dio

// StreamMarshaller is an interface that defines functions for reading and writing
// data to/from a struct to a data streaming interface.
type StreamMarshaller interface {
	MarshalStream(Writer) error
	UnmarshalStream(Reader) error
}
