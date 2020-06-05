package limits

const (
	frag   = 1024
	small  = 256
	large  = 4096
	medium = 2048
)

// Current is the basic default profile set limits for buffer and fragment values. Limit
// configuration does not affect the transmission of Channel data.
var Current = Medium

var (
	// Tiny provides the smallest values. This may provide the slowest but most undetectable transfers.
	Tiny = &Limit{Frag: 256, Small: 64, Large: 1024, Medium: 512}

	// Small provides the small values for buffers and size.
	Small = &Limit{Frag: 512, Small: 128, Large: 2048, Medium: 1024}

	// Medium provides the most efficient values for buffers and size. This is the default value.
	Medium = &Limit{Frag: 1024, Small: 256, Large: 4096, Medium: 2048}

	// Large provides the largest buffer and limit sizes. This is best for the fastest transfer rates, but
	// will increase the potential detection rate.
	Large = &Limit{Frag: 4096, Small: 512, Large: 8192, Medium: 4096}
)

// Limit is a struct that defines the default values for buffer and channel sizes. The
// built-in values can be customized for petter performance.
type Limit struct {
	_      [0]func()
	Frag   uint32
	Large  uint32
	Small  uint16
	Medium uint16
}

// FragLimit returns the Fragment size on the current Limit. This function will return the Fragment
// size of the Medium profile if the Limit is not set.
func FragLimit() int {
	if Current == nil {
		return retNoZero(frag, int(Medium.Frag))
	}
	return retNoZero(frag, int(Current.Frag))
}

// SmallLimit returns the small buffer size on the current Limit. This function will return the small buffer
// size of the Medium profile if the Limit is not set.
func SmallLimit() int {
	if Current == nil {
		return retNoZero(small, int(Medium.Small))
	}
	return retNoZero(small, int(Current.Small))
}

// LargeLimit returns the large buffer size on the current Limit. This function will return the large buffer
// size of the Medium profile if the Limit is not set.
func LargeLimit() int {
	if Current == nil {
		return retNoZero(large, int(Medium.Large))
	}
	return retNoZero(large, int(Current.Large))
}

// MediumLimit returns the medium buffer size on the current Limit. This function will return the medium buffer
// size of the Medium profile if the Limit is not set.
func MediumLimit() int {
	if Current == nil {
		return retNoZero(medium, int(Medium.Medium))
	}
	return retNoZero(medium, int(Current.Medium))
}
func retNoZero(d, v int) int {
	if v > 0 {
		return v
	}
	return d
}
