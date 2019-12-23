package c2

import "github.com/iDigitalFlame/xmt/xmt-c2/control"

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

	// Actions
	MsgUpload    = uint16(control.Upload)
	MsgRefresh   = uint16(control.Refresh)
	MsgExecute   = uint16(control.Execute)
	MsgDownload  = uint16(control.Download)
	MsgProcesses = uint16(control.ProcessList)

	MsgProxy = 0xFE11 // registry required
	MsgSpawn = 0xFE12 // registry required

	MsgError = 0xFEEF
)
