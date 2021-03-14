package cmd

type DLL struct {
	code *Code
	Path string
}

// DLL is just a simple wrapper for Code
//  Payload is DLL path and exec pointer is the pointer to LoadLibraryA

func (d *DLL) Start() error {
	return nil
}
