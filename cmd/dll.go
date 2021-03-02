package cmd

type DLL struct {
	Path string

	code *Code
}

// DLL is just a simple wrapper for Code
//  Payload is DLL path and exec pointer is the pointer to LoadLibraryA

func (d *DLL) Start() error {
	return nil
}
