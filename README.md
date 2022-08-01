# XMT: eXtensible Malware Toolkit

XMT is a full featured C2 framework written in Golang that allows for control,
data exfiltration and some other cool functions. Can be used to make full C2
clients/servers with little out-of-the-box changes.

[ThunderStorm](https://github.com/iDigitalFlame/ThunderStorm) would be an implementation
of this.

The pkg.go.dev site has the framework documentation [here](https://pkg.go.dev/github.com/iDigitalFlame/xmt).

## TODO

These are some things I need to work on.

- Keyloging
- MultiProxy Support
- Shellcode for Linux without CGO (potentially)
- Add in memory (Reflective) DLL injection (Outside of SRDi)
- ScreenShot support for MacOS/Linux without CGO?
- More Windows thread creation techniques besides NtCreateThreadEx

## Bugs

Issues that I know are broken.

- _device.GoExit() / winapi.KillRuntime()_: We can't determine which threads are ours (but we DO know they are Golang threads)
   so we kill threads in the same process as us that might not be ours.

If you're using this, feel free to submit issue tickets or pull requests. (I don't bite, mostly owo)

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
