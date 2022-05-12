#!/usr/bin/python3

from sys import argv


class ErrorValue(object):
    def __init__(self, name, ident=None):
        self.name = name
        self.ident = ident

    def __str__(self):
        if self.ident is None:
            return self.name
        return f"{self.name} {self.ident}"


ERRORS = [None] * 0xFF

# Generic-ish
ERRORS[0x00] = ErrorValue("invalid error")
ERRORS[0x01] = ErrorValue("unknown error")
ERRORS[0x0A] = ErrorValue("empty host field")
ERRORS[0x0B] = ErrorValue("invalid port specified")
ERRORS[0x0C] = ErrorValue("whence is invalid")
ERRORS[0x0D] = ErrorValue("invalid type")
# Package c2, c2/cfg, c2/wrapper, c2/task
ERRORS[0x0F] = ErrorValue("invalid Packet")
ERRORS[0x10] = ErrorValue("empty or nil host", "c2.ErrNoHost")
ERRORS[0x11] = ErrorValue("other side did not come up", "c2.ErrNoConn")
ERRORS[0x12] = ErrorValue("empty or nil Profile", "c2.ErrInvalidProfile")
ERRORS[0x13] = ErrorValue("first Packet is invalid")
ERRORS[0x14] = ErrorValue("empty or invalid loader name")
ERRORS[0x15] = ErrorValue("no Profile parser loaded")
ERRORS[0x16] = ErrorValue("unexpected OK value")
ERRORS[0x17] = ErrorValue("empty or nil Packet", "c2.ErrMalformedPacket")
ERRORS[0x18] = ErrorValue("unable to listen")
ERRORS[0x19] = ErrorValue("not a Listener", "c2.ErrNotAListener")
ERRORS[0x1A] = ErrorValue("not a Connector", "c2.ErrNotAConnector")
ERRORS[0x1B] = ErrorValue("empty Listener name")
ERRORS[0x1C] = ErrorValue("listener already exists")
ERRORS[0x1D] = ErrorValue("no Job created for client Session", "c2.ErrNoTask")
ERRORS[0x1E] = ErrorValue("send buffer is full", "c2.ErrFullBuffer")
ERRORS[0x1F] = ErrorValue(
    "frag/multi total is zero on a frag/multi packet", "c2.ErrInvalidPacketCount"
)
ERRORS[0x20] = ErrorValue("empty or nil Job")
ERRORS[0x21] = ErrorValue("cannot be a client session")
ERRORS[0x22] = ErrorValue("migration in progress")
ERRORS[0x23] = ErrorValue("cannot assign a Job ID")
ERRORS[0x24] = ErrorValue("job ID is in use")
ERRORS[0x25] = ErrorValue("cannot marshal Profile")
ERRORS[0x26] = ErrorValue("empty or nil Tasklet")
ERRORS[0x27] = ErrorValue("cannot marshal Proxy data")
ERRORS[0x28] = ErrorValue("packet ID does not match the supplied ID")
ERRORS[0x29] = ErrorValue("proxy support disabled")
ERRORS[0x2A] = ErrorValue("cannot marshal Proxy Profile")
ERRORS[0x2B] = ErrorValue("only a single Proxy per session can be active")
ERRORS[0x2C] = ErrorValue(
    "frag/multi count is larger than 0xFFFF", "c2.ErrTooManyPackets"
)
ERRORS[0x2D] = ErrorValue("received Packet that does not match our own device ID")
ERRORS[0x2E] = ErrorValue("setting is invalid", "cfg.ErrInvalidSetting")
ERRORS[0x2F] = ErrorValue("cannot add multiple transforms", "cfg.ErrMultipleTransforms")
ERRORS[0x30] = ErrorValue(
    "cannot add multiple connections", "cfg.ErrMultipleConnections"
)
ERRORS[0x31] = ErrorValue("binary source not available")
ERRORS[0x32] = ErrorValue('missing "type" string')
ERRORS[0x33] = ErrorValue("key not found")
ERRORS[0x34] = ErrorValue("invalid io operation")
ERRORS[0x35] = ErrorValue("mapping ID is invalid")
ERRORS[0x36] = ErrorValue("mapping ID is already exists")
ERRORS[0x37] = ErrorValue("empty key name")
ERRORS[0x38] = ErrorValue("empty value name")
ERRORS[0x39] = ErrorValue("arguments cannot be nil or empty")
# Package com
ERRORS[0x3A] = ErrorValue("malformed Tag", "com.ErrMalformedTag")
ERRORS[0x3B] = ErrorValue("tags list is too large", "com.ErrTagsTooLarge")
ERRORS[0x3C] = ErrorValue("missing TLS certificates", "com.ErrInvalidTLSConfig")
ERRORS[0x3D] = ErrorValue("invalid permissions")
ERRORS[0x3E] = ErrorValue("invalid permission size")
ERRORS[0x3F] = ErrorValue("invalid HTTP response")
ERRORS[0x40] = ErrorValue("body is not writable")
ERRORS[0x41] = ErrorValue("could not get underlying net.Conn")
ERRORS[0x42] = ErrorValue("not a file")
ERRORS[0x43] = ErrorValue("not a directory")
# Package man
ERRORS[0x50] = ErrorValue("empty or invalid Guardian name")
ERRORS[0x51] = ErrorValue("no paths found", "man.ErrNoEndpoints")
ERRORS[0x52] = ErrorValue("invalid path name")
ERRORS[0x53] = ErrorValue("invalid link type")
# Package cmd, cmd/script, cmd/evade
ERRORS[0x60] = ErrorValue("stdout already set")
ERRORS[0x61] = ErrorValue("stderr already set")
ERRORS[0x62] = ErrorValue("stdin already set")
ERRORS[0x63] = ErrorValue("process has not been started", "cmd.ErrNotStarted")
ERRORS[0x64] = ErrorValue("process arguments are empty", "cmd.ErrEmptyCommand")
ERRORS[0x65] = ErrorValue("process is still running", "cmd.ErrStillRunning")
ERRORS[0x66] = ErrorValue("process has already been started", "cmd.ErrAlreadyStarted")
ERRORS[0x67] = ErrorValue("cannot find '.text' section", 0x67)
ERRORS[0x69] = ErrorValue(
    "could not find a suitable process", "filter.ErrNoProcessFound"
)
# Package data/crypto
ERRORS[0x80] = ErrorValue("block size must be between 16 and 128 and a power of two")
ERRORS[0x81] = ErrorValue("block size must equal IV size")
# Package device, device/winapi, device/winapi/registry
ERRORS[0x90] = ErrorValue("invalid address value")
ERRORS[0x91] = ErrorValue("cannot dump self")
ERRORS[0x92] = ErrorValue("invalid env value")
ERRORS[0x93] = ErrorValue("empty DLL name")
ERRORS[0x94] = ErrorValue("unexpected key size")
ERRORS[0x95] = ErrorValue("unexpected key type")
ERRORS[0x96] = ErrorValue("update with no service status handle")
ERRORS[0xFA] = ErrorValue("only supported on Windows devices", "device.ErrNoWindows")
ERRORS[0xFB] = ErrorValue("only supported on *nix devices", "device.ErrNoNix")


def find_error(v):
    if isinstance(v, int):
        if v < 0 or v > 0xFF:
            return None
        return ERRORS[v]
    try:
        n = int(v, base=0)
        if n < 0 or n > 0xFF:
            return None
        return ERRORS[n]
    except ValueError:
        pass
    return None


if __name__ == "__main__":
    if len(argv) == 2:
        print(find_error(argv[1]))
    else:
        print(f"{argv[0]} <error_code>")
