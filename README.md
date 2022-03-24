# XMT: eXtensible Malware Toolkit

XMT is a framework written in Golang that allows for target/endpoint control and
data exfil. Can be used to make full C2 clients/servers with little out-of-the-box
changes.

[ThunderStorm](https://github.com/iDigitalFlame/ThunderStorm) would be an implementation
of this.

## TODO

These are some things I need to work on.

- Keyloging?
- Shellcode for Linux?
- Add in memory (Reflective) DLL injection (Sorta have it?)

## Bugs

Issues that I know are broken.
If you're using this, feel free to submit issue tickets or pull requests. (I don't bite, mostly owo)

## Thanks and Credits

- Package Monkey by @skx [github.com/skx/monkey](https://github.com/skx/monkey)
- Package Otto by @robertkrimen [github.com/robertkrimen/otto](https://github.com/robertkrimen/otto)
- Intern method by @bradfitz [tailscale.com/blog/netaddr-new-ip-type-for-go/](https://tailscale.com/blog/netaddr-new-ip-type-for-go/)
  - Also the IP struct code and info.
- mTLS insights by @kofoworola [https://kofo.dev/how-to-mtls-in-golang](https://kofo.dev/how-to-mtls-in-golang)
- DLL loader by @monoxgas [https://github.com/monoxgas/sRDI](https://github.com/monoxgas/sRDI)

# Licenses

XMT is covered by the GNU GPLv3 License

- [sRDI](https://raw.githubusercontent.com/monoxgas/sRDI/master/LICENSE) (GPLv3)
- [Monkey](https://raw.githubusercontent.com/skx/monkey/master/LICENSE) (MIT)
- [Otto](https://raw.githubusercontent.com/robertkrimen/otto/master/LICENSE) (MIT)
- [LogX](https://raw.githubusercontent.com/PurpleSec/LogX/main/LICENSE) (Apache v2)
- [Escape](https://raw.githubusercontent.com/PurpleSec/Escape/main/LICENSE) (Apache v2)
