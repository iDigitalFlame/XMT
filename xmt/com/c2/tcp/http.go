package tcp

import "net/http"

var (
	// Transform is the transform for the web profile.
	// This transform will convert web forms to and from the
	// required byte arrays.
	Transform = &transform{}

	// BasicProfile is the default settings for a C2 Profile
	// that follow the useage of the C2 web server.
	BasicProfile = &Profile{}
)

// HTTP is a C2 profile that mimics a standard web server and
// client setup. This struct inherits the http.Server struct and can
// be used to serve real files and pages. Use the Mapper struct to
// provide a URL mapping that can be used by clients to access the C2
// functions.
type HTTP struct {
	*http.Server
}

// Mapper is a struct that can be used to help determine which URL paths
// or options are pages versus which are C2 access routes.
type Mapper struct {
}

// Profile is the web C2 profile struct. This struct can be used or
// overrides to add more advanced functions to the C2 profile.
type Profile struct{}
type transform struct {
}
