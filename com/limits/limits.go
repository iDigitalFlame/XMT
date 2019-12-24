package limits

var (
	// Limits is the basic profile set limits for the default
	// buffer and fragment values.
	Limits = Medium

	// Tiny provides the smallest values. This may provide the slowest but
	// most undetectable transfers.
	Tiny = &Limit{
		Frag:   256,
		Small:  64,
		Large:  1024,
		Medium: 512,
	}
	// Small provides the small values for buffers and size.
	Small = &Limit{
		Frag:   512,
		Small:  128,
		Large:  2048,
		Medium: 1024,
	}
	// Medium provides the most efficient values for buffers
	// and size. This is the default value for limits.
	Medium = &Limit{
		Frag:   1024,
		Small:  256,
		Large:  4096,
		Medium: 2048,
	}
	// Large provides the largest buffer and limit sizes. This is best for
	// the fastest transfer rates, but will increase the detection rate.
	Large = &Limit{
		Frag:   4096,
		Small:  512,
		Large:  8192,
		Medium: 4096,
	}
)

// Limit is a struct that defines the default values for
// buffer and channel sizes. The built-in values can be customized for
// petter performance.
type Limit struct {
	Frag   uint32
	Small  uint16
	Large  uint32
	Medium uint16
}

// FragLimit returns the Fragment size on the current Limit. This function will
// return the Fragment size of the Medium profile if the Limit is not set.
func FragLimit() int {
	if Limits == nil {
		return int(Medium.Frag)
	}
	return int(Limits.Frag)
}

// SmallLimit returns the small buffer size on the current Limit. This function will
// return the small buffer size of the Medium profile if the Limit is not set.
func SmallLimit() int {
	if Limits == nil {
		return int(Medium.Small)
	}
	return int(Limits.Small)
}

// LargeLimit returns the large buffer size on the current Limit. This function will
// return the large buffer size of the Medium profile if the Limit is not set.
func LargeLimit() int {
	if Limits == nil {
		return int(Medium.Large)
	}
	return int(Limits.Large)
}

// MediumLimit returns the medium buffer size on the current Limit. This function will
// return the medium buffer size of the Medium profile if the Limit is not set.
func MediumLimit() int {
	if Limits == nil {
		return int(Medium.Medium)
	}
	return int(Limits.Medium)
}
