// +build client

package device

// IsServer is a compile time constant that can be used to disable the logx Logger and prevent any
// un-needed fmt calls as the client does not /naturally/ need to produce output. Only needed for debug
// purposes
const IsServer = false
