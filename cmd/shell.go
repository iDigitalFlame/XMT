package cmd

const (
	// VerbEdit launches an editor and opens the document for editing. If the target is not a document file,
	// the function will fail.
	VerbEdit = Verb("edit")
	// VerbFind initiates a search beginning in the directory specified by the working directory.
	VerbFind = Verb("find")
	// VerbOpen opens the item specified by the target parameter. The item can be a file or folder.
	VerbOpen = Verb("open")
	// VerbPrint prints the file specified by the target. If the target is not a document file, the function fails.
	VerbPrint = Verb("print")
	// VerbRunAs launches an application as Administrator. User Account Control (UAC) will prompt the user for consent to run
	// the application elevated or enter the credentials of an administrator account used to run the application.
	VerbRunAs = Verb("runas")
	//VerbExplore explores a folder specified by the target.
	VerbExplore = Verb("explore")
)

// Verb is the equivalent to the Windows ShellExecute verb type string. This is used in the ShellExecute function.
type Verb string
