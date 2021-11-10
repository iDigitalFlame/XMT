package data

import "io"

const (
	// ErrTooLarge is raised if memory cannot be allocated to store data in a Chunk.
	ErrTooLarge = dataError(3)
	// ErrInvalidType is an error that occurs when the Bytes, ReadBytes, StringVal or ReadString functions could not
	// propertly determine the underlying type of array from the Reader.
	ErrInvalidType = dataError(1)
	// ErrInvalidIndex is raised if a specified Grow or index function is supplied with an negative or out of
	// bounds number or when a Seek index is not valid.
	ErrInvalidIndex = dataError(2)
)

// ErrLimit is an error that is returned when a Limit is set on a Chunk and the size limit was hit when
// attempting to write to the Chunk. This error wraps the io.EOF error, which allows this error to match
// io.EOF for sanity checking.
var ErrLimit = new(limitError)

type dataError uint8
type limitError struct{}

func (limitError) Error() string {
	return "buffer size limit reached"
}
func (limitError) Unwrap() error {
	return io.EOF
}
func (e dataError) Error() string {
	switch e {
	case ErrInvalidType:
		return "could not find the buffer type"
	case ErrInvalidIndex:
		return "index provided is invalid"
	case ErrTooLarge:
		return "buffer size is too large"
	}
	return "unknown error"
}
