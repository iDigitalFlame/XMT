# XMT: eXtensible Malware Toolkit

[![Go Report Card](https://goreportcard.com/badge/github.com/iDigitalFlame/xmt)](https://goreportcard.com/report/github.com/iDigitalFlame/xmt)
[![Go Reference](https://pkg.go.dev/badge/github.com/iDigitalFlame/xmt.svg)](https://pkg.go.dev/github.com/iDigitalFlame/xmt)

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

The pkg.go.dev site has some of the framework documentation and definitions
[here](https://pkg.go.dev/github.com/iDigitalFlame/xmt).

## Roadmap

_Updated 10/18/22_

- Reflective DLL Injection (Windows)
- Linux mem_fd loader
- "Device Check" package
  - Detect VM
  - Anti-VM checks
- Linux shellcode support

These are some things that would be nice to have, but are not as important as the
above list:

- Keylogging
- MultiProxy support
- X/Wayland/OSX Screenshot support
- EDR Detection
- More thread injection options (Windows)

## TODOs

These are some things I need to work on.

- Documentation
- Build tags list

## References / Hightlights / Presentations

BSides Las Vegas 2022: So you Wanta Build a C2?

[Video](https://www.youtube.com/watch?v=uAfGtGlHLxs) /
[Slides](https://public.idigitalflame.com/docs/so_you_wanta_build_a_c2.pdf)

## Bugs

_Updated 10/18/22_

So far, we're looking good. Feel free to submit issue tickets or pull requests if
something is broken or doesn't act right. (I don't bite, mostly owo)

## Thanks and Credits

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

- [sRDI](https://raw.githubusercontent.com/monoxgas/sRDI/master/LICENSE) (GPLv3)
- [Monkey](https://raw.githubusercontent.com/skx/monkey/master/LICENSE) (MIT)
- [Otto](https://raw.githubusercontent.com/robertkrimen/otto/master/LICENSE) (MIT)
- [LogX](https://raw.githubusercontent.com/PurpleSec/LogX/main/LICENSE) (Apache v2)
- [Escape](https://raw.githubusercontent.com/PurpleSec/Escape/main/LICENSE) (Apache v2)
