//go:build windows
// +build windows

package devtools

import (
	"context"
	"time"

	"golang.org/x/sys/windows/svc"
)

const backoff = 100 * time.Millisecond

// Service is a struct that assists in running a Windows service. This struct can be created and given functions
// to run.
//  - Exec - the function to run for each Timeout when greater than zero.
//  - Start - function to run on service start,
//  - End - function to run on service shutdown.
//
// Trigger the service to start by using the 'Service.Run' function. The 'Run' function always returns
// 'ErrNoWindows' on non-Windows devices.
type Service struct {
	ctx context.Context

	Start, End, Exec func()
	Name             string
	Interval         time.Duration
}

// Run will trigger the service to start and will block until the service completes. Will always returns
// 'ErrNoWindows' on non-Windows devices.
func (s *Service) Run() error {
	return s.RunContext(context.Background())
}
func executeCatch(f func()) (fail bool) {
	if f == nil {
		return false
	}
	defer func() {
		if recover() != nil {
			fail = true
		}
	}()
	f()
	return false
}

// RunContext will trigger the service to start and will block until the service completes. Will always returns
// 'ErrNoWindows' on non-Windows devices. This function allows to pass a Context to cancel the running service.
func (s *Service) RunContext(x context.Context) error {
	s.ctx = x
	return svc.Run(s.Name, s)
}

// Execute fulfils the 'svc.Handler' interface. This function is not to be called directly. Instead
// call the 'Service.Run' function.
func (s *Service) Execute(_ []string, c <-chan svc.ChangeRequest, x chan<- svc.Status) (bool, uint32) {
	x <- svc.Status{State: svc.StartPending}
	if executeCatch(s.Start) {
		time.Sleep(backoff)
		x <- svc.Status{State: svc.StopPending}
		return false, 1
	}
	x <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	if s.Exec == nil {
		r := executeCatch(s.End)
		time.Sleep(backoff)
		x <- svc.Status{State: svc.StopPending}
		if r {
			return false, 1
		}
		return false, 0
	}
	if s.ctx == nil {
		s.ctx = context.Background()
	}
	var (
		t    *time.Ticker
		w    <-chan time.Time
		e    bool
		z, f = context.WithCancel(s.ctx)
	)
	if s.Interval <= 0 {
		go func() {
			e = executeCatch(s.Exec)
			f()
		}()
	} else {
		t = time.NewTicker(s.Interval)
		w = t.C
	}
	x <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	for {
		select {
		case r := <-c:
			switch r.Cmd {
			case svc.Interrogate:
				x <- r.CurrentStatus
				time.Sleep(backoff)
				x <- r.CurrentStatus
			case svc.Stop, svc.Shutdown:
				goto cleanup
			default:
			}
		case <-w:
			if e = executeCatch(s.Exec); e {
				goto cleanup
			}
		case <-z.Done():
			goto cleanup
		}
	}
cleanup:
	x <- svc.Status{State: svc.StopPending}
	if f(); t != nil {
		t.Stop()
	}
	if executeCatch(s.End) || e {
		return false, 1
	}
	return false, 0
}
