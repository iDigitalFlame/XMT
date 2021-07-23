// +build client

package c2

// Logging is a compile time constant that can be used to disable the logx Logger and prevent any
// un-needed fmt calls as the client does not /naturally/ need to produce output. Only needed for debug
// purposes
const Logging = false
