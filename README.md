# XMT: eXtensible Malware Toolkit

XMT is a framework written in Golang that allows for target/endpoint control and data exfil.

This framework also includes a C2 server.

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
If you're using this, feel free to submit issue tickets or pull requests. (I don't bite)

- UDP/ICMP connectors currently have issues transferring data blobs larger than the chunk size.
  - WC2 Also has this issue.
- Channel is still broken. :(
- CBK does not play well with WC2 for some reason.
