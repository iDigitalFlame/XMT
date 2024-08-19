# XMT: eXtensible Malware Toolkit

[![Go Report Card](https://goreportcard.com/badge/github.com/iDigitalFlame/xmt)](https://goreportcard.com/report/github.com/iDigitalFlame/xmt)
[![Go Reference](https://pkg.go.dev/badge/github.com/iDigitalFlame/xmt.svg)](https://pkg.go.dev/github.com/iDigitalFlame/xmt)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Code Analysis](https://github.com/iDigitalFlame/XMT/actions/workflows/checks.yaml/badge.svg)](https://github.com/iDigitalFlame/XMT/actions/workflows/checks.yaml)
[![codecov](https://codecov.io/github/iDigitalFlame/XMT/branch/main/graph/badge.svg?token=REQESSIT7C)](https://codecov.io/github/iDigitalFlame/XMT)
[![Latest](https://img.shields.io/github/v/tag/iDigitalFlame/XMT)](https://github.com/iDigitalFlame/XMT/releases)

XMT is a full-featured C2 framework written in Golang that allows for control,
data exfiltration and some other cool functions. Can be used to make full C2
clients/servers with little out-of-the-box changes.

[ThunderStorm](https://github.com/iDigitalFlame/ThunderStorm) would be an implementation
of this.

This framework also contains many utility functions, including:

- Advanced Process Control (Windows)
- Device Identification
- User Identification
- Windows "Window" utils
- Efficient Data Marshaling interfaces
- Easy Network communication resources
- Super low file size! ~5mb completely using [JetStream](https://github.com/iDigitalFlame/ThunderStorm)
- Backwards compatibility with systems as old as Windows Xp!

The pkg.go.dev site has some of the framework documentation and definitions
[here](https://pkg.go.dev/github.com/iDigitalFlame/xmt).

__DISCLAIMER: Please use for legal reasons only. I'm not responsible if you get__
__in trouble for using this improperly or if someone owns your environment and is__
__using XMT (or a derivative of it).__

## Roadmap

_Updated 02/17/23_

- Reflective DLL Injection (Windows)
- Updates to handeling x86 PEB (Windows)
- Linux mem_fd loader
- Thread Injection improvements
- "Device Check" package
  - Detect VM
  - Anti-VM checks

These are some things that would be nice to have, but are not as important as the
above list:

- Keylogging
- MultiProxy support
- X/Wayland/OSX Screenshot support
- EDR Detection
- Linux shellcode support
- More thread injection options (Windows)

## Compatibility

This project is compatable with __ALL__ Golang versions starting from __go1.10__!
You can download the older versions of Golang from [the Golang website](https://go.dev/dl/).

Unless convined otherwise, I plan to keep the compatibility down to Go1.10.
__Since I don't control the Script engines, Scripts are bound to >= go1.18__

__The following depreciated build types will NOT be supported__

- nacl/386
- nacl/amd64p32
- nacl/arm

__The following depreciated build types WORK but are specific__

- darwin/386 (<= go1.14)
- darwin/arm (<= go1.14, needs CGO)

### Older OS Support Issues

So far the only issues I've seen are:

- Xp
  - Lacks the "CreateProcessWithTokenW" so any processes created while impersonating
    a user will fail. _(This does NOT affect Server 2003 WTF)_
- Xp < SP3
  - Lacks the "WinHttpGetDefaultProxyConfiguration" function, which disables
    automatic HTTP Proxy detection.
- Xp and Server 2003
  - Lacks the "RegDeleteTree" function so deleting non-empty Keys may fail.
  - The concept of Token "Integrity" does not exist and users that are in the
    "Administrators" group are considered elevated.
  - Per the previous entry, the "Untrust" helper will NOT set the Token Integrity
    _(since it doesn't exist!)_, but it will STILL remove Token permissions.
  - Setting the parent process does __NOT__ work.
- Vista, Server 2008 and older
  - Cannot evade ETW logs as the function calls do not exist.
- Windows 8.1, Server 2012 and older
  - Cannot evade ASMI as it is only present in Windows 10 and newer.

### Compiling for Go1.10 (pre-modules)

Golang version 1.11 introduced the concept of Golang Modules and made dependency
management simple. Unfortunately, Go1.10 (the last to support Xp, 2003, 2008
and Vista) does __not__.

To work around this, we can just _vendor_ the packages, since the only dependencies,
are the following PurpleSec modules:

- [LogX: github.com/PurpleSec/logx](https://github.com/PurpleSec/logx)
- [Escape: github.com/PurpleSec/escape](https://github.com/PurpleSec/escape)

Which we already make backwards compatible :D

These dependencies can be downloaded and used with the following commands:

```bash
go mod vendor
mkdir "deps"
mv "vendor" "deps/src"
mkdir "deps/src/github.com/iDigitalFlame"
ln -s "$(pwd)" "deps/src/github.com/iDigitalFlame/xmt"
export GOPATH="$(pwd)/deps"
export GOROOT="<path to downloaded Go1.10 folder>"
```

_(Yes, I know you CAN use "-o" to specific the vendor directory, but that isn't_
_supported until go1.18!)_

This should allow you to compile using the fullpath of the Go1.10 Golang binary.
_(As long as you set your `GOROOT` and `GOPATH` correctly)_

## TODO

These are some things I need to work on.

- Documentation
- Build tags list

## References / Hightlights / Presentations

BSides Las Vegas 2022: So you Wanta Build a C2?

[Video](https://www.youtube.com/watch?v=uAfGtGlHLxs) /
[Slides](https://public.idigitalflame.com/docs/so_you_wanta_build_a_c2.pdf)

## Bugs

_Updated 02/17/23_

- Potential KeyPair sync issue over long periods of time. __Still needs more testing__

Feel free to submit issue tickets or pull requests if something is broken or
doesn't act right. (I don't bite, mostly owo)

## Thanks and Credits

- [Geoff Chappell](https://www.geoffchappell.com) for his insights into various Windows API stuff
- Package Monkey by @skx [github.com/skx/monkey](https://github.com/skx/monkey)
- Package Otto by @robertkrimen [github.com/robertkrimen/otto](https://github.com/robertkrimen/otto)
- Intern method by @bradfitz [tailscale.com/blog/netaddr-new-ip-type-for-go/](https://tailscale.com/blog/netaddr-new-ip-type-for-go/)
  - Also the IP struct code and info.
- mTLS insights by @kofoworola [kofo.dev/how-to-mtls-in-golang](https://kofo.dev/how-to-mtls-in-golang)
- DLL loader by @monoxgas [github.com/monoxgas/sRDI](https://github.com/monoxgas/sRDI)
- Initial idea for MiniDump/DLL Reload by the Sliver C2 framework [github.com/BishopFox/sliver/](https://github.com/BishopFox/sliver/)
- Untrust idea by @zha0gongz1 [golangexample.com/...](https://golangexample.com/without-closing-windows-defender-to-make-defender-useless-by-removing-its-token-privileges-and-lowering-the-token-integrity/)

# Licenses

XMT is covered by the GNU GPLv3 License

Third-party Licenses:

- [sRDI](https://raw.githubusercontent.com/monoxgas/sRDI/master/LICENSE) (GPLv3)
- [Monkey](https://raw.githubusercontent.com/skx/monkey/master/LICENSE) (MIT)
  - Only if [Monkey](https://github.com/skx/monkey) support is compiled in and enabled.
- [Otto](https://raw.githubusercontent.com/robertkrimen/otto/master/LICENSE) (MIT)
  - Only if [Otto](https://github.com/robertkrimen/otto) support is compiled in and enabled.
- [LogX](https://raw.githubusercontent.com/PurpleSec/LogX/main/LICENSE) (Apache v2)
- [Escape](https://raw.githubusercontent.com/PurpleSec/Escape/main/LICENSE) (Apache v2)

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/Z8Z4121TDS)
