// +build windows

package devtools

import (
	"context"
	"time"

	"golang.org/x/sys/windows/svc"
)

// Service is a struct that assists in running a Windows service. This struct can be created and given functions
// to run.
//  - Exec - the function to run for each Timeout when greater than zero.
//  - Start - function to run on service start,
//  - End - function to run on service shutdown.
//
// Trigger the service to start by using the 'Service.Run' function. The 'Run' function always returns
// 'ErrNoWindows' on non-Windows devices.
type Service struct {
	Name             string
	Timeout          time.Duration
	Exec, Start, End func()

	ctx context.Context
}

// Run will trigger the service to start and will block until the service completes. Will always returns
// 'ErrNoWindows' on non-Windows devices.
func (s *Service) Run() error {
	return s.RunContext(context.Background())
}

// RunContext will trigger the service to start and will block until the service completes. Will always returns
// 'ErrNoWindows' on non-Windows devices. This function allows to pass a Context to cancel the running service.
func (s *Service) RunContext(x context.Context) error {
	if s != nil {
		s.ctx = x
	}
	return svc.Run(s.Name, s)
}

// Execute fulfils the 'svc.Handler' interface. This function is not to be called directly. Instead
// call the 'Service.Run' function.
func (s *Service) Execute(_ []string, c <-chan svc.ChangeRequest, x chan<- svc.Status) (bool, uint32) {
	x <- svc.Status{State: svc.StartPending}
	defer func(z chan<- svc.Status) {
		recover()
		z <- svc.Status{State: svc.StopPending}
	}(x)
	if s.Start != nil {
		s.Start()
	}
	if s.Exec == nil {
		x <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
		time.Sleep(100 * time.Millisecond)
		x <- svc.Status{State: svc.StopPending}
		return false, 0
	}
	if s.ctx == nil {
		s.ctx = context.Background()
	}
	var (
		t    *time.Ticker
		w    <-chan time.Time
		z, f = context.WithCancel(s.ctx)
	)
	if s.Timeout <= 0 {
		go func(o *Service, q context.CancelFunc) {
			o.Exec()
			q()
		}(s, f)
	} else {
		t = time.NewTicker(s.Timeout)
		w = t.C
	}
	x <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	for {
		select {
		case r := <-c:
			switch r.Cmd {
			case svc.Interrogate:
				x <- r.CurrentStatus
				time.Sleep(100 * time.Millisecond)
				x <- r.CurrentStatus
			case svc.Stop, svc.Shutdown:
				goto cleanup
			default:
			}
		case <-w:
			s.Exec()
		case <-z.Done():
			goto cleanup
		}
	}
cleanup:
	x <- svc.Status{State: svc.StopPending}
	if f(); t != nil {
		t.Stop()
	}
	if s.End != nil {
		s.End()
	}
	return false, 0
}
