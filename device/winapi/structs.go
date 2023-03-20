//go:build windows
// +build windows

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

package winapi

import (
	"sync"
	"unsafe"
)

// SID matches the SID struct
//
//	https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-sid
type SID struct{}

// ACL matches the ACL struct
//
//	https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-acl
//
//	typedef struct _ACL {
//	  BYTE AclRevision;
//	  BYTE Sbz1;
//	  WORD AclSize;
//	  WORD AceCount;
//	  WORD Sbz2;
//	} ACL;
//
// DO NOT REORDER
type ACL struct {
	_, _    byte
	_, _, _ uint16
}

// LUID matches the LUID struct
//
//	https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-luid
//
//	typedef struct _LUID {
//	  DWORD LowPart;
//	  LONG  HighPart;
//	} LUID, *PLUID;
//
// DO NOT REORDER
type LUID struct {
	Low  uint32
	High int32
}
type curDir struct {
	// DO NOT REORDER
	DosPath lsaString
	Handle  uintptr
}
type modInfo struct {
	// DO NOT REORDER
	Base  uintptr
	Size  uint32
	Entry uintptr
}
type clientID struct {
	// DO NOT REORDER
	Process uintptr
	Thread  uintptr
}
type objAttrs struct {
	// DO NOT REORDER
	Length                   uint32
	RootDirectory            uintptr
	ObjectName               *lsaString
	Attributes               uint32
	SecurityDescriptor       *SecurityDescriptor
	SecurityQualityOfService *SecurityQualityOfService
}
type certBlob struct {
	// DO NOT REORDER
	_ uint32
	_ uintptr
}
type certAlgo struct {
	// DO NOT REORDER
	_ *uint16
	_ certBlob
}
type certInfo struct {
	// DO NOT REORDER
	_       uint32
	Serial  certBlob
	_       certAlgo
	Issuer  certBlob
	_, _    uint64
	Subject certBlob
	_       certAlgo
	_, _, _ certBlob
	_       uint32
	_       uintptr
	// NOTE(dij): This is here as go1.10 has a bug with this.
	//            It tries to compare certInfo structs for some reason?
	_ [0]func()
}
type lsaString struct {
	// DO NOT REORDER
	Length        uint16
	MaximumLength uint16
	Buffer        *uint16
}
type dumpParam struct {
	_ [0]func()
	sync.Mutex
	h, b uintptr
	s, w uint64
}

// TokenUser matches the TOKEN_USER struct.
//
//	https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-token_user
//
//	typedef struct _TOKEN_USER {
//	  SID_AND_ATTRIBUTES User;
//	} TOKEN_USER, *PTOKEN_USER
//
// DO NOT REORDER
type TokenUser struct {
	User SIDAndAttributes
}

// ProxyInfo matches the WINHTTP_PROXY_INFO struct
//
//	https://docs.microsoft.com/en-us/windows/win32/api/winhttp/ns-winhttp-winhttp_proxy_info
//
//	typedef struct _WINHTTP_PROXY_INFO {
//	  DWORD  dwAccessType;
//	  LPWSTR lpszProxy;
//	  LPWSTR lpszProxyBypass;
//	} WINHTTP_PROXY_INFO, *LPWINHTTP_PROXY_INFO, *PWINHTTP_PROXY_INFO;
//
// DO NOT REORDER
type ProxyInfo struct {
	AccessType  uint32
	Proxy       *uint16
	ProxyBypass *uint16
}
type certSigner struct {
	// DO NOT REORDER
	_       uint32
	Issuer  certBlob
	Serial  certBlob
	_, _    certAlgo
	_, _, _ certBlob
	// NOTE(dij): This is here as go1.10 has a bug with this.
	//            It tries to compare certSigner structs for some reason?
	_ [0]func()
}
type dumpOutput struct {
	Status int32
}
type privileges struct {
	// DO NOT REORDER
	PrivilegeCount uint32
	Privileges     [5]LUIDAndAttributes
}
type processPeb struct {
	// DO NOT REORDER
	_                      [2]byte
	BeingDebugged          byte
	_                      [1]byte
	_                      [2]uintptr
	Ldr                    uintptr
	ProcessParameters      *processParams
	_                      [3]uintptr
	AtlThunkSListPtr       uintptr
	_                      uintptr
	_                      uint32
	_                      uintptr
	_                      uint32
	AtlThunkSListPtr32     uint32
	_                      [9]uintptr
	_                      [10]byte
	NtGlobalFlag           uint32
	_                      [35]uintptr
	_                      [84]byte
	PostProcessInitRoutine uintptr
	_                      [128]byte
	_                      [1]uintptr
	SessionID              uint32
}

// Overlapped matches the OVERLAPPED struct
//
//	https://docs.microsoft.com/en-us/windows/win32/api/minwinbase/ns-minwinbase-overlapped
//
//	typedef struct _OVERLAPPED {
//	  ULONG_PTR Internal;
//	  ULONG_PTR InternalHigh;
//	  DWORD Offset;
//	  DWORD OffsetHigh;
//	  HANDLE    hEvent;
//	} OVERLAPPED, *LPOVERLAPPED;
//
// DO NOT REORDER
type Overlapped struct {
	Internal     uintptr
	InternalHigh uintptr
	Offset       uint32
	OffsetHigh   uint32
	Event        uintptr
}

// StartupInfo matches the STARTUPINFOW struct
//
//	https://docs.microsoft.com/en-us/windows/win32/api/processthreadsapi/ns-processthreadsapi-startupinfow
//
//	typedef struct _STARTUPINFOW {
//	  DWORD  cb;
//	  LPWSTR lpReserved;
//	  LPWSTR lpDesktop;
//	  LPWSTR lpTitle;
//	  DWORD  dwX;
//	  DWORD  dwY;
//	  DWORD  dwXSize;
//	  DWORD  dwYSize;
//	  DWORD  dwXCountChars;
//	  DWORD  dwYCountChars;
//	  DWORD  dwFillAttribute;
//	  DWORD  dwFlags;
//	  WORD   wShowWindow;
//	  WORD   cbReserved2;
//	  LPBYTE lpReserved2;
//	  HANDLE hStdInput;
//	  HANDLE hStdOutput;
//	  HANDLE hStdError;
//	} STARTUPINFOW, *LPSTARTUPINFOW;
//
// DO NOT REORDER
type StartupInfo struct {
	Cb            uint32
	_             *uint16
	Desktop       *uint16
	Title         *uint16
	X             uint32
	Y             uint32
	XSize         uint32
	YSize         uint32
	XCountChars   uint32
	YCountChars   uint32
	FillAttribute uint32
	Flags         uint32
	ShowWindow    uint16
	_             uint16
	_             *byte
	StdInput      uintptr
	StdOutput     uintptr
	StdErr        uintptr
}
type timeZoneInfo struct {
	// DO NOT REORDER
	Bias         uint32
	_            [80]byte
	StdBias      uint32
	_            [80]byte
	DaylightBias uint32
}
type highContrast struct {
	// DO NOT REORDER
	Size  uint32
	Flags uint32
	_     *uint16
}

// StartupInfoEx matches the STARTUPINFOEXW struct
//
//	https://docs.microsoft.com/en-us/windows/win32/api/winbase/ns-winbase-startupinfoexw
//
//	typedef struct _STARTUPINFOEXW {
//	  STARTUPINFOW                 StartupInfo;
//	  LPPROC_THREAD_ATTRIBUTE_LIST lpAttributeList;
//	} STARTUPINFOEXW, *LPSTARTUPINFOEXW;
//
// DO NOT REORDER
type StartupInfoEx struct {
	StartupInfo   StartupInfo
	AttributeList *StartupAttributes
}
type lsaAttributes struct {
	// DO NOT REORDER
	Length     uint32
	_          uintptr
	_          *lsaString
	Attributes uint32
	_, _       unsafe.Pointer
}
type processParams struct {
	// DO NOT REORDER
	_                [16]byte
	Console          uintptr
	_                uint32
	StandardInput    uintptr
	StandardOutput   uintptr
	StandardError    uintptr
	CurrentDirectory curDir
	DllPath          lsaString
	ImagePathName    lsaString
	CommandLine      lsaString
	Environment      uintptr
}

// ServiceStatus matches the SERVICE_STATUS struct
//
//	https://docs.microsoft.com/en-us/windows/win32/api/winsvc/ns-winsvc-service_status
//
//	typedef struct _SERVICE_STATUS {
//	 DWORD dwServiceType;
//	 DWORD dwCurrentState;
//	 DWORD dwControlsAccepted;
//	 DWORD dwWin32ExitCode;
//	 DWORD dwServiceSpecificExitCode;
//	 DWORD dwCheckPoint;
//	 DWORD dwWaitHint;
//	} SERVICE_STATUS, *LPSERVICE_STATUS;
type ServiceStatus struct {
	ServiceType             uint32
	CurrentState            uint32
	ControlsAccepted        uint32
	Win32ExitCode           uint32
	ServiceSpecificExitCode uint32
	CheckPoint              uint32
	WaitHint                uint32
}
type diskGeometryEx struct {
	// DO NOT REORDER
	_    [24]byte
	Size uint64
	_    uintptr
}
type threadBasicInfo struct {
	// DO NOT REORDER
	ExitStatus     uint32
	TebBaseAddress uintptr
	ClientID       clientID
	_              uint64
	_              uint32
}
type ntUnicodeString struct {
	// DO NOT REORDER
	Length        uint16
	MaximumLength uint16
	_, _          uint16
	Buffer        [260]uint16
}
type systemBasicInfo struct {
	// DO NOT REORDER
	_             [8]byte
	PageSize      uint32
	PhysicalPages uint32
	LowPage       uint32
	HighPage      uint32
	_             uint32
	_             [3]uintptr
	NumProc       uint8
}

// SIDAndAttributes matches the SID_AND_ATTRIBUTES struct
//
//	https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-sid_and_attributes
//
//	typedef struct _SID_AND_ATTRIBUTES {
//	  PSID  Sid;
//	  DWORD Attributes;
//	} SID_AND_ATTRIBUTES, *PSID_AND_ATTRIBUTES;
//
// DO NOT REORDER
type SIDAndAttributes struct {
	Sid        *SID
	Attributes uint32
}
type processBasicInfo struct {
	// DO NOT REORDER
	ExitStatus                   uint32
	PebBaseAddress               uintptr
	_                            *uintptr
	_                            uint32
	UniqueProcessID              uintptr
	InheritedFromUniqueProcessID uintptr
}

// ServiceTableEntry matches the SERVICE_TABLE_ENTRYW struct
//
//	https://docs.microsoft.com/en-us/windows/win32/api/winsvc/ns-winsvc-service_table_entryw
//
//	typedef struct _SERVICE_TABLE_ENTRYW {
//	 LPWSTR                   lpServiceName;
//	 LPSERVICE_MAIN_FUNCTIONW lpServiceProc;
//	} SERVICE_TABLE_ENTRYW, *LPSERVICE_TABLE_ENTRYW;
type ServiceTableEntry struct {
	Name *uint16
	Proc uintptr
}

// StartupAttributes matches the LPPROC_THREAD_ATTRIBUTE_LIST opaque struct
//
//	https://docs.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-initializeprocthreadattributelist
//
// There's not much documentation for this *shrug*
type StartupAttributes struct {
	_     [4]byte
	Count uint32
	_     [64]byte
}

// LUIDAndAttributes matches the LUIDAndAttributes struct
//
//	https://docs.microsoft.com/en-us/previous-versions/windows/desktop/wmipjobobjprov/win32-luidandattributes
//
//	typedef struct LUIDAndAttributes {
//	  LUID  Luid;
//	  DWORD dwSize;
//	} PLUIDANDATTRIBUTES;
//
// DO NOT REORDER
type LUIDAndAttributes struct {
	Luid       LUID
	Attributes uint32
}

// ProcessInformation matches the PROCESS_INFORMATION struct
//
//	https://docs.microsoft.com/en-us/windows/win32/api/processthreadsapi/ns-processthreadsapi-process_information
//
//	typedef struct _PROCESS_INFORMATION {
//	  HANDLE hProcess;
//	  HANDLE hThread;
//	  DWORD  dwProcessId;
//	  DWORD  dwThreadId;
//	} PROCESS_INFORMATION, *PPROCESS_INFORMATION, *LPPROCESS_INFORMATION;
//
// DO NOT REORDER
type ProcessInformation struct {
	Process   uintptr
	Thread    uintptr
	ProcessID uint32
	ThreadID  uint32
}

// SecurityDescriptor matches the SECURITY_DESCRIPTOR struct
//
//	https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-security_descriptor
//
//	typedef struct _SECURITY_DESCRIPTOR {
//	  BYTE                        Revision;
//	  BYTE                        Sbz1;
//	  SECURITY_DESCRIPTOR_CONTROL Control;
//	  PSID                        Owner;
//	  PSID                        Group;
//	  PACL                        Sacl;
//	  PACL                        Dacl;
//	} SECURITY_DESCRIPTOR, *PISECURITY_DESCRIPTOR;
//
// DO NOT REORDER
type SecurityDescriptor struct {
	_, _ byte
	_    SecurityDescriptorControl
	_, _ SID
	_, _ *ACL
}

// SecurityAttributes matches the SECURITY_ATTRIBUTES struct
//
//	https://docs.microsoft.com/en-us/windows/win32/api/wtypesbase/ns-wtypesbase-security_attributes
//
//	typedef struct _SECURITY_ATTRIBUTES {
//	  DWORD  nLength;
//	  LPVOID lpSecurityDescriptor;
//	  BOOL   bInheritHandle;
//	} SECURITY_ATTRIBUTES, *PSECURITY_ATTRIBUTES, *LPSECURITY_ATTRIBUTES;
//
// DO NOT REORDER
type SecurityAttributes struct {
	Length             uint32
	SecurityDescriptor *SecurityDescriptor
	InheritHandle      uint32
}
type lsaAccountDomainInfo struct {
	// DO NOT REORDER
	_   lsaString
	SID *SID
}

// SecurityQualityOfService matches the SECURITY_QUALITY_OF_SERVICE struct
//
//	https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-security_quality_of_service
//
//	typedef struct _SECURITY_QUALITY_OF_SERVICE {
//	  DWORD                          Length;
//	  SECURITY_IMPERSONATION_LEVEL   ImpersonationLevel;
//	  SECURITY_CONTEXT_TRACKING_MODE ContextTrackingMode;
//	  BOOLEAN                        EffectiveOnly;
//	} SECURITY_QUALITY_OF_SERVICE, *PSECURITY_QUALITY_OF_SERVICE;
type SecurityQualityOfService struct {
	Length              uint32
	ImpersonationLevel  uint32
	ContextTrackingMode bool
	EffectiveOnly       bool
}

// SecurityDescriptorControl matches the SECURITY_DESCRIPTOR_CONTROL bitflag.
//
//	https://docs.microsoft.com/en-us/windows/win32/secauthz/security-descriptor-control
//
// typedef WORD SECURITY_DESCRIPTOR_CONTROL, *PSECURITY_DESCRIPTOR_CONTROL;
type SecurityDescriptorControl uint16

// String returns the string representation of this SID.
func (s *SID) String() string {
	var o *uint16
	if err := convertSIDToStringSID(s, &o); err != nil {
		return ""
	}
	v := UTF16ToString((*[256]uint16)(unsafe.Pointer(o))[:])
	localFree(uintptr(unsafe.Pointer(o)))
	return v
}

// UserName attempts to return a Domain\User string from the SID.
func (s *SID) UserName() (string, error) {
	var c, x, t uint32 = 64, 64, 0
	for {
		var (
			n, d      = make([]uint16, c), make([]uint16, x)
			r, _, err = syscallN(funcLookupAccountSid.address(),
				0, uintptr(unsafe.Pointer(s)), uintptr(unsafe.Pointer(&n[0])),
				uintptr(unsafe.Pointer(&c)), uintptr(unsafe.Pointer(&d[0])),
				uintptr(unsafe.Pointer(&x)), uintptr(unsafe.Pointer(&t)),
			)
		)
		if r > 0 {
			u, q := UTF16ToString(n), UTF16ToString(d)
			if n, d = nil, nil; len(q) == 0 {
				return u, nil
			}
			return q + "\\" + u, nil
		}
		if err != ErrInsufficientBuffer || c <= uint32(len(n)) {
			return "", unboxError(err)
		}
	}
}

// IsWellKnown returns true if this SID matches the well known SID type index.
func (s *SID) IsWellKnown(t uint32) bool {
	r, _, _ := syscallN(funcIsWellKnownSID.address(), uintptr(unsafe.Pointer(s)), uintptr(t))
	return r > 0
}
func (s *SecurityDescriptor) len() uint32 {
	r, _, _ := syscallN(funcRtlLengthSecurityDescriptor.address(), uintptr(unsafe.Pointer(s)))
	return uint32(r)
}
func localFree(h uintptr) (uintptr, error) {
	r, _, err := syscallN(funcLocalFree.address(), h)
	if r != 0 {
		return r, unboxError(err)
	}
	return r, nil
}
func convertSIDToStringSID(i *SID, s **uint16) error {
	r, _, err := syscallN(funcConvertSIDToStringSID.address(), uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(s)))
	if r == 0 {
		return unboxError(err)
	}
	return nil
}
func (s *SecurityDescriptor) copyRelative() *SecurityDescriptor {
	var (
		n = int(s.len())
		m = int(unsafe.Sizeof(SecurityDescriptor{}))
	)
	if n < m {
		n = m
	}
	var (
		b []byte
		h = (*SliceHeader)(unsafe.Pointer(&b))
	)
	h.Data = unsafe.Pointer(s)
	h.Len, h.Cap = n, n
	var (
		d []byte
		x = (*SliceHeader)(unsafe.Pointer(&d))
		a = make([]uintptr, (n+int(ptrSize)-1)/int(ptrSize))
	)
	x.Data = (*SliceHeader)(unsafe.Pointer(&a)).Data
	x.Len, x.Cap = n, n
	copy(d, b)
	return (*SecurityDescriptor)(unsafe.Pointer(&d[0]))
}
