package devtools

// ServiceMain is a quick assistant function that can be used to create and execute
// a Windows service.
//
// This function takes the service name and the function to run in the service body.
func ServiceMain(name string, f func()) error {
	return (&Service{Name: name, Exec: f}).Run()
}
