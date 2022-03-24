//go:build windows

package svc

import (
	"context"
	"os"
	"sync"
	"syscall"
	"unsafe"

	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Standard Windows Service State values
//
// DO NOT REORDER
const (
	Stopped State = 1 + iota
	StartPending
	StopPending
	Running
	ContinuePending
	PausePending
	Paused
)

// Standard Windows Service Reason values
//
// DO NOT REORDER
const (
	ReasonDemand Reason = 1 << iota
	ReasonAuto
	ReasonTrigger
	ReasonRestartOnFailure
	ReasonDelayedAuto
)

// Standard Windows Service Command values
//
// DO NOT REORDER
const (
	Stop Command = 1 + iota
	Pause
	Continue
	Interrogate
	Shutdown
	ParamChange
	NetBindAdd
	NetBindRemove
	NetBindEnable
	NetBindDisable
	DeviceEvent
	HardwareProfileChange
	PowerEvent
	SessionChange
	PreShutdown
)

// Standard Windows Service Accepted values
//
// DO NOT REORDER
const (
	AcceptStop Accepted = 1 << iota
	AcceptPauseAndContinue
	AcceptShutdown
	AcceptParamChange
	AcceptNetBindChange
	AcceptHardwareProfileChange
	AcceptPowerEvent
	AcceptSessionChange
	AcceptPreShutdown
)

var (
	service  *svcInstance
	callBack struct {
		sync.Once
		f, m uintptr
	}
)

// State describes the current service execution state (Stopped, Running, etc.)
type State uint32

// Reason is the reason that the service was started.
type Reason uint32

// Command represents a service state change request. It is sent to a service
// by the service manager, and should be acted upon by the service.
type Command uint32

// Accepted is used to describe commands accepted by the service.
//
// Interrogate is always accepted.
type Accepted uint32

// Change is sent to the service Handler to request service status changes and
// updates to the service control manager.
type Change struct {
	Command   Command
	EventType uint32
	EventData uintptr
	Status    Status
	Context   uintptr
}

// Status combines State and Accepted commands to fully describe running
// service.
type Status struct {
	State      State
	Accepts    Accepted
	CheckPoint uint32
	WaitHint   uint32
	ProcessID  uint32
	ExitCode   uint32
}

// Service is an interface that is passed to the Handler function and can be used
// to receive and send updates to the service control manager.
//
// NOTE(dij): The function 'DynamicStartReason' is only avaliable on Windows >7
// and will return an error if it does not exist.
type Service interface {
	Update(Status)
	Handle() uintptr
	Requests() <-chan Change
	UpdateState(State, ...Accepted)
	DynamicStartReason() (Reason, error)
}
type svcInstance struct {
	f     Handler
	e, in chan Change
	out   chan Status
	n     string
	h     uintptr
}

// Handler is a function interface that must be implemented to run as a Windows
// service.
//
// This function will be called by the package code at the start of the service,
// and the service will exit once Execute completes.
//
// Inside the function, you may use the context or read service change requests
// using the 's.Requests()' channel and act accordingly.
//
// You must keep service control manager up to date about state of your service
// by using the 'Update' or 'UpdateState' functions.
//
// The supplied string list contains the service name followed by argument
// strings passed to the service.
//
// You can provide service exit code in the return parameter, with 0 being
// "no error".
type Handler func(context.Context, Service, []string) uint32

func (s *svcInstance) Handle() uintptr {
	return s.h
}
func (s *svcInstance) Update(v Status) {
	s.out <- v
}

// Run executes service name by calling the appropriate handler function.
//
// This function will block until complete.
// Any errors returned indicate that bootstrappping of the service failed.
func Run(name string, f Handler) error {
	if service.f != nil {
		return os.ErrInvalid
	}
	n, err := winapi.UTF16PtrFromString(name)
	if err != nil {
		return err
	}
	callBack.Do(func() {
		service = &svcInstance{e: make(chan Change)}
		callBack.m = syscall.NewCallback(serviceMain)
		callBack.f = syscall.NewCallback(serviceHandler)
	})
	service.n, service.f = name, f
	t := []winapi.ServiceTableEntry{
		{Name: n, Proc: callBack.m}, {Name: nil, Proc: 0},
	}
	return winapi.StartServiceCtrlDispatcher(&t[0])
}
func (s *svcInstance) Requests() <-chan Change {
	return s.in
}
func serviceHandler(c, e, d, _ uintptr) uintptr {
	//            NOTE(dij): ^ This pointer is SUPER FUCKING UNRELIABLE! Don't
	//                       fucking use it!
	service.e <- Change{Command: Command(c), EventType: uint32(e), EventData: d}
	return 0
}
func serviceMain(argc uint32, argv **uint16) uintptr {
	if service == nil || callBack.f == 0 || callBack.m == 0 {
		return 0xE0000239
	}
	var err error
	service.h, err = winapi.RegisterServiceCtrlHandlerEx(service.n, callBack.f, uintptr(unsafe.Pointer(&service)))
	//                                                                          NOTE(dij): ^ For some reason, keeping
	//                                                                                       this here prevents it from
	//                                                                                       being garbage collected.
	if err != nil {
		if e, ok := err.(syscall.Errno); ok {
			return uintptr(e)
		}
		return 0xE0000239
	}
	var (
		e []*uint16
		h = (*winapi.SliceHeader)(unsafe.Pointer(&e))
	)
	h.Data = unsafe.Pointer(argv)
	h.Len, h.Cap = int(argc), int(argc)
	a := make([]string, len(e))
	for i, v := range e {
		a[i] = winapi.UTF16PtrToString(v)
	}
	if err := service.update(&Status{State: StartPending}, false, 0); err != nil {
		if e, ok := err.(syscall.Errno); ok {
			service.update(&Status{State: Stopped}, false, uint32(e))
			return uintptr(e)
		}
		service.update(&Status{State: Stopped}, false, 0xE0000239)
		return 0xE0000239
	}
	var (
		b, y = context.WithCancel(context.Background())
		c    = Status{State: StartPending}
		x    = make(chan uint32)
		f    uint32
	)
	// NOTE(dij): Making the 'in' channel buffered so the sends to it don't
	//            block.
	service.in, service.out = make(chan Change, 1), make(chan Status)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				x <- 0x1
				close(x)
			}
		}()
		x <- service.f(b, service, a)
		close(x)
	}()
loop:
	for {
		select {
		case f = <-x:
			break loop
		case v := <-service.e:
			// NOTE(dij): Instead of dropping all the excess new entries on the
			//            floor, we should clear them instead as we want the
			//            service handler to see the latest entry.
			for len(service.in) > 0 {
				<-service.in
			}
			// NOTE(dij): Cancel the context and signal that we're working on
			//            closing up.
			switch v.Status = c; v.Command {
			case Stop, Shutdown:
				y()
				service.update(&Status{State: StopPending}, false, 0)
			}
			service.in <- v
		case v := <-service.out:
			if err := service.update(&v, false, v.ExitCode); err != nil {
				if e, ok := err.(syscall.Errno); ok {
					f = uint32(e)
				} else {
					f = 0xE0000239
				}
				break loop
			}
			c = v
		}
	}
	service.update(&Status{State: StopPending}, f > 0, f)
	y()
	close(service.in)
	close(service.out)
	service.update(&Status{State: Stopped}, f > 0, f)
	close(service.e)
	service.h = 0
	return 0
}
func (s *svcInstance) UpdateState(v State, a ...Accepted) {
	if len(a) == 0 {
		s.out <- Status{State: v}
		return
	}
	if len(a) == 1 {
		s.out <- Status{State: v, Accepts: a[0]}
		return
	}
	u := Status{State: v, Accepts: a[0]}
	for i := 1; i < len(a); i++ {
		u.Accepts |= a[i]
	}
	s.out <- u
}
func (s *svcInstance) DynamicStartReason() (Reason, error) {
	r, err := winapi.QueryServiceDynamicInformation(s.h, 1)
	if err != nil {
		return 0, err
	}
	return Reason(r), nil
}
func (s *svcInstance) update(u *Status, r bool, e uint32) error {
	if s.h == 0 {
		return xerr.Sub("update with no service status handle", 0x5)
	}
	v := winapi.ServiceStatus{ServiceType: 16, CurrentState: uint32(u.State)}
	if u.Accepts&AcceptStop != 0 {
		v.ControlsAccepted |= 1
	}
	if u.Accepts&AcceptPauseAndContinue != 0 {
		v.ControlsAccepted |= 2
	}
	if u.Accepts&AcceptShutdown != 0 {
		v.ControlsAccepted |= 4
	}
	if u.Accepts&AcceptParamChange != 0 {
		v.ControlsAccepted |= 8
	}
	if u.Accepts&AcceptNetBindChange != 0 {
		v.ControlsAccepted |= 16
	}
	if u.Accepts&AcceptHardwareProfileChange != 0 {
		v.ControlsAccepted |= 32
	}
	if u.Accepts&AcceptPowerEvent != 0 {
		v.ControlsAccepted |= 64
	}
	if u.Accepts&AcceptSessionChange != 0 {
		v.ControlsAccepted |= 128
	}
	if u.Accepts&AcceptPreShutdown != 0 {
		v.ControlsAccepted |= 256
	}
	if e == 0 {
		v.Win32ExitCode, v.ServiceSpecificExitCode = 0, 0
	} else if r {
		v.Win32ExitCode, v.ServiceSpecificExitCode = 1064, e
	} else {
		v.Win32ExitCode, v.ServiceSpecificExitCode = e, 0
	}
	v.CheckPoint = u.CheckPoint
	v.WaitHint = u.WaitHint
	return winapi.SetServiceStatus(s.h, &v)
}
