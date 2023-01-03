//go:build bugs

// Copyright (C) 2020 - 2023 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

// Package bugtrack enables the bug tracking system, which is comprised of a
// global logger that will write to Standard Error and on the filesystem in a
// temporary directory, "$TEMP" in *nix and "%TEMP%" on Windows, that is named
// "bugtrack-<PID>.log".
//
// To enable bug tracking, use the "bugs" build tag.
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
		f = filepath.Join(p, "bugtrack-"+strconv.FormatUint(uint64(os.Getpid()), 10)+".log")
		l logx.Log
	)
	if l, err = logx.File(f, logx.Append, logx.Trace); err != nil {
		panic("bugtrack: creating file log failed with error: " + err.Error())
	}
	log = logx.Multiple(l, logx.Writer(os.Stderr, logx.Trace))
	log.SetPrefix("BUGTRACK")
	log.Info(`Bugtrack log init complete, log file can be found at "%s".`, f)
}

// Recover is a "guard" function to be used to gracefully shut down a program
// when a panic is detected.
//
// Can be en enabled by using:
//
//	if bugtrack.Enabled {
//	    defer bugtrack.Recover("thread-name")
//	}
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
