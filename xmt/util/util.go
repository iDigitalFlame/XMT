package util

// Must accepts dual paramater returns with an error on
// the end. This function returns the resulting object if
// the error returned in nil. This function will panic if
// the error is not nil.
func Must(v interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return v
}
