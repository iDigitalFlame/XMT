package xerr

type err struct {
	e error
	s string
}
type strErr string

func (e err) Error() string {
	return e.s
}
func (e err) Unwrap() error {
	return e.e
}
func (e err) String() string {
	return e.s
}
func (e strErr) Error() string {
	return string(e)
}
func (e strErr) String() string {
	return string(e)
}
