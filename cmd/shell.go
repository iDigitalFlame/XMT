package cmd

import (
	"context"
	"time"
)

// TODO - Work on this

type Shell struct {
	Data    []string
	Verb    string
	Timeout time.Duration

	ctx    context.Context
	err    error
	opts   *options
	exit   uint32
	once   uint32
	flags  uint32
	cancel context.CancelFunc
	parent *container
}
