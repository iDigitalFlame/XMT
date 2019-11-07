package c2

import "github.com/iDigitalFlame/xmt/xmt/c2/action"

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

	MsgUpload    = uint16(action.Upload)
	MsgRefresh   = uint16(action.Refresh)
	MsgExecute   = uint16(action.Execute)
	MsgDownload  = uint16(action.Download)
	MsgProcesses = uint16(action.ProcessList)

	MsgProxy = 0xFE11 // registry required
	MsgSpawn = 0xFE12 // registry required

	MsgError = 0xFEEF
)
