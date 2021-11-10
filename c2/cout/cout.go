//go:build !client
// +build !client

package cout

import "github.com/PurpleSec/logx"

// Enabled is a compile time constant that can be used to disable/enable the logx Logger and prevent any
// un-needed fmt calls as the client does not /naturally/ need to produce output. Only needed for debug
// purposes
const Enabled = true

// Log is an interface for any type of struct that supports standard Logging functions.
type Log struct {
	logx.Log
}

// New creates a Log instance from a logx Logger.
func New(l logx.Log) *Log {
	return &Log{Log: l}
}

// Set updates the internal logger. This function is a NOP if the logger is nil or logging is not
// enabled via the 'client' build tag.
func (l *Log) Set(v logx.Log) {
	if l == nil {
		return
	}
	l.Log = v
}

// Info writes a informational message to the logger.
// The function arguments are similar to fmt.Sprintf and fmt.Printf. The first argument is
// a string that can contain formatting characters. The second argument is a vardict of
// interfaces that can be omitted or used in the supplied format string.
func (l *Log) Info(s string, v ...interface{}) {
	if l == nil || l.Log == nil {
		return
	}
	l.Log.Info(s, v...)
}

// Error writes a error message to the logger.
// The function arguments are similar to fmt.Sprintf and fmt.Printf. The first argument is
// a string that can contain formatting characters. The second argument is a vardict of
// interfaces that can be omitted or used in the supplied format string.
func (l *Log) Error(s string, v ...interface{}) {
	if l == nil || l.Log == nil {
		return
	}
	l.Log.Error(s, v...)
}

// Trace writes a tracing message to the logger.
// The function arguments are similar to fmt.Sprintf and fmt.Printf. The first argument is
// a string that can contain formatting characters. The second argument is a vardict of
// interfaces that can be omitted or used in the supplied format string.
func (l *Log) Trace(s string, v ...interface{}) {
	if l == nil || l.Log == nil {
		return
	}
	l.Log.Trace(s, v...)
}

// Debug writes a debugging message to the logger.
// The function arguments are similar to fmt.Sprintf and fmt.Printf. The first argument is
// a string that can contain formatting characters. The second argument is a vardict of
// interfaces that can be omitted or used in the supplied format string.
func (l *Log) Debug(s string, v ...interface{}) {
	if l == nil || l.Log == nil {
		return
	}
	l.Log.Debug(s, v...)
}

// Warning writes a warning message to the logger.
// The function arguments are similar to fmt.Sprintf and fmt.Printf. The first argument is
// a string that can contain formatting characters. The second argument is a vardict of
// interfaces that can be omitted or used in the supplied format string.
func (l *Log) Warning(s string, v ...interface{}) {
	if l == nil || l.Log == nil {
		return
	}
	l.Log.Warning(s, v...)
}
