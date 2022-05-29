// Package xerr is a simplistic (and more efficient) re-write of the "errors"
// built-in package.
//
// This is used to create comparable and (sometimes) un-wrapable error structs.
//
// This package acts differently when the "implant" build tag is used. If enabled,
// Most error string values are stripped to prevent identification and debugging.
//
// It is recommended if errors are needed to be compared even when in an implant
// build, to use the "Sub" function, which will ignore error strings and use
// error codes instead.
//
package xerr
