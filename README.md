# XMT: eXtensible Malware Toolkit

XMT is a framework written in Golang that allows for target/endpoint control and data exfil.
This framework also includes a C2 server. See [ThunderStorm](https://github.com/iDigitalFlame/ThunderStorm)

XMT will support multiple Operating Systems and will have low level (kernel and DLL) compatibility with the major targeted Operating Systems.

This is a current work in progress.

## TODO

These are some things I need to work on.

- RPC Server (WIP!)
- Shellcode for Linux?
- Add execute assembly (See Sliver C2)
- Add in memory (Reflective) DLL injection (See Sliver C2)

## Bugs

Issues that I know are broken.
If you're using this, feel free to submit issue tickets or pull requests. (I don't bite, mostly owo)

## Thanks and Credits

- Package Monkey by @skx [github.com/skx/monkey](https://github.com/skx/monkey)
- Package Otto by @robertkrimen [github.com/robertkrimen/otto](https://github.com/robertkrimen/otto)
- Intern method by @bradfitz [tailscale.com/blog/netaddr-new-ip-type-for-go/](https://tailscale.com/blog/netaddr-new-ip-type-for-go/)
  - Also the IP struct code and info.
- Package machineid by @denisbrodbeck [github.com/denisbrodbeck/machineid](https://github.com/denisbrodbeck/machineid)
- mTLS insights by @kofoworola [https://kofo.dev/how-to-mtls-in-golang](https://kofo.dev/how-to-mtls-in-golang)
