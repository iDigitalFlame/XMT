package c2

// Packet Message ID Constants.  Used for
// reference.
const (
	MsgPing       = 0xFE00
	MsgSleep      = 0xFE01
	MsgHello      = 0xFE02
	MsgResult     = 0xFE13
	MsgProfile    = 0xFE15
	MsgRegister   = 0xFE05
	MsgMultiple   = 0xFE03
	MsgShutdown   = 0xFE04
	MsgRegistered = 0xFE06

	MsgUpload   = 0xFE10
	MsgRefresh  = 0xFE07
	MsgExecute  = 0xFE08
	MsgDownload = 0xFE09

	MsgProxy       = 0xFE11 // registry
	MsgSpawn       = 0xFE12 // registry
	MsgProcessList = 0xFE14

	MsgError = 0xFEEF
)
