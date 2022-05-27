//go:build bugs

package bugtrack

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/PurpleSec/logx"
)

// Enabled is the stats of the bugtrack package.
//
// This is true if bug tracking is enabled.
const Enabled = true

var log logx.Log

func init() {
	var (
		p   = os.TempDir()
		err = os.MkdirAll(p, 0755)
	)
	if err != nil {
		panic("bugtrack: init failed with error: " + err.Error())
	}
	var (
		f = filepath.Join(p, "bugtrack-"+strconv.Itoa(os.Getpid())+".log")
		l logx.Log
	)
	if l, err = logx.File(f, logx.Append, logx.Trace); err != nil {
		panic("bugtrack: creating file log failed with error: " + err.Error())
	}
	log = logx.Multiple(l, logx.Writer(os.Stderr, logx.Trace))
	log.SetPrefix("BUGTRACK")
	log.Info("Bugtrack log init complete, log file can be found at %q.", f)
}

// Recover is a "guard" function to be used to gracefully shutdown a program
// when a panic is detected.
//
// Can be en enabled by using:
//    if bugtrack.Enabled {
//        defer bugtrack.Recover("thread-name")
//    }
//
// The specified name will be entered into the bugtrack log and a stack trace
// will be generated before gracefully returning execution to the program.
func Recover(v string) {
	if r := recover(); r != nil {
		log.Error("Recovered %s: [%s]", v, r)
		log.Error("Trace: %s", debug.Stack())
		time.Sleep(time.Minute)
	}
}

// Track is a simple logging function that takes the same arguments as a
// 'fmt.Sprintf' function. This can be used to track bugs or output values.
//
// Not recommended to be used in production environments.
//
// The "-tags bugs" option is required in order for this function to be used.
func Track(s string, m ...any) {
	log.Trace(s, m...)
}
