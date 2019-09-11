package c2

// Packet Message ID Constants.  Used for
// reference.
const (
	MsgPing       = 0xFE00
	MsgSleep      = 0xFE01
	MsgHello      = 0xFE02
	MsgMultiple   = 0xFE03
	MsgShutdown   = 0xFE04
	MsgRegister   = 0xFE05
	MsgRegistered = 0xFE06

	MsgRefresh  = 0xFE07
	MsgExecute  = 0xFE08
	MsgDownload = 0xFE09
	MsgUpload   = 0xFE10
	MsgProxy    = 0xFE11
	MsgConnect  = 0xFE12
	MsgResults  = 0xFE13

	MsgError = 0xFEEF
)
