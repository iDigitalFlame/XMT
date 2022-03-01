// Package bugtrack enables the bug tracking system, which is comprised of a
// global logger that will write to Standard Error and on the filesystem in a
// temporary directory, "$TEMP" in *nix and "%TEMP%" on Windows, that is named
// "bugtrack-<PID>.log".
//
// To enable bug tracking, use the "bugs" build tag.
//
package bugtrack
