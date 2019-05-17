package data

type Wrapper interface {
	Wrap([]byte) error
	Unwrap([]byte) error
}
