// +build windows

package devtools

import (
	"sync/atomic"
	"time"

	"golang.org/x/sys/windows/svc"
)

// Service is a struct that assists in running a Windows service. This struct can be created and given functions
// to run (Exec - the function to run for each Timeout when greater than zero, Start - function to run on service start,
// End - function to run on service shutdown.) Trigger the service to start by using the 'Service.Run' function.
// The 'Run' function always returns 'ErrNoWindows' on non-Windows devices.
type Service struct {
	Name             string
	Timeout          time.Duration
	Exec, Start, End func()

	done uint32
}

// Run will trigger the service to start and will block until the service completes. Will always returns
// 'ErrNoWindows' on non-Windows devices.
func (s *Service) Run() error {
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
	var (
		i chan bool
		t *time.Ticker
		p <-chan time.Time
	)
	if s.Timeout > 0 {
		t = time.NewTicker(s.Timeout)
		p = t.C
	} else {
		i = make(chan bool, 1)
		i <- true
	}
	x <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	for atomic.LoadUint32(&s.done) == 0 {
		select {
		case r := <-c:
			switch r.Cmd {
			case svc.Interrogate:
				x <- r.CurrentStatus
				time.Sleep(100 * time.Millisecond)
				x <- r.CurrentStatus
			case svc.Stop, svc.Shutdown:
				atomic.StoreUint32(&s.done, 1)
			default:
			}
		case <-i:
			s.Exec()
		case <-p:
			s.Exec()
		}
	}
	if s.Timeout > 0 {
		t.Stop()
	} else {
		close(i)
	}
	if s.End != nil {
		s.End()
	}
	x <- svc.Status{State: svc.StopPending}
	return false, 0
}
