#!/usr/bin/python3
# Copyright (C) 2020 - 2022 iDigitalFlame
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.
#

from zlib import crc32
from shlex import split
from io import StringIO
from hashlib import sha512
from datetime import datetime
from json import dumps, loads
from secrets import token_bytes
from traceback import format_exc
from argparse import ArgumentParser
from base64 import b64decode, b64encode
from sys import argv, exit, stderr, stdin, stdout

HELP_TEXT = """XMT cfg.Config Builder v1 Release

Usage: {binary} [add|append] <options>

GROUP MODE:
  To enable adding multiple Profiles under a config (Group Mode),
  add the "add" or "append" word before the command options. This
  will append the options as a separate group in the Profile.

  To overrite the group, omit the "add" or "append" value (the
  default option).

  In Group Mode, the -f/--file/-o/--out arguments may behave
  differently. If no output file is specified and "input" points to
  a valid file, the "input" WILL ALSO be treated as the "output". The
  same behavior occurs if no input is specified but "output" points
  to a valid file, then the output file will be read in as input
  before being appended to.

  To disable this behavior, just specify the output and input files
  when using Group Mode.

BASIC ARGUMENTS:
  -h                            Show this help message and exit.
  --help

INPUT/OUTPUT ARGUMENTS:
  -f                <file>
  --in              <file>      Input file path. Use '-' for stdin.
                                  See the 'Group Mode' section for the
                                  modifiers to this argument.
  -o                <file>
  --out             <file>      Output file path. Stdout is used if
                                  empty. See the 'Group Mode' section
                                  for the modifiers to this argument.
  -j
  --json                        Output in JSON format. Omit for raw
                                  binary. (Or base64 when output to
                                  stdout.)
  -I                            Accept stdin input as commands. Each
  --stdin                         line from stdin will be treated as a
                                  'append' line to the supplied config.
                                  Input and Output are ignored and are
                                  only set via the command line.
                                  This option disables using stdin for
                                  Config data.

OPERATION ARGUMENTS:
  -p
  --print                       List values contained in the file
                                  input. Fails if no input is found or
                                  invalid. Output format can be modified
                                  using -j/-p.

BUILD ARGUMENTS:
 SYSTEM:
  --host            <hostname>  Hostname hint, used for implants only.
  --sleep           <secs|mod>  Sleep time period. Defaults to seconds
                                  for integers, but can take modifiers
                                  such as 's', 'h', 'm'. (ex: 2m or 3s).
  --jitter          <jitter %>  Jitter as a percentage [0-100]. Values
                                  greater than 100 fail. The '%' symbol
                                  may be included or omitted.
  --weight          <weight>    Group weight value [0-100]. Takes effect
                                  only if multiple Profiles exist. Only
                                  affects the Group it is used in.
  --selector        <selector>  Use the specified selector to indicate
                                  which order the Profile Groups should
                                  be used in. Takes effect globally and
                                  only needs to be used once in ANY
                                  group.

                                  See SELECTOR VALUES for more info.
  --killdate        <ISO-8601>  Specify a time in ISO-8601 format that
                                  be used to indicate when the implant
                                  will cease to function. The time should
                                  be specified in ISO-8601 format which is
                                  "YYYY-MM-DDTHH:MM:SS". The seconds may
                                  be omitted.
  --wh-days         <S-M-S str> Specify a Working hours value that targets
                                  specific days. This may be used in combonation
                                  with the "--wh-start" and "--wh-end" arguments.
                                  The accepted values are "SMTWRFS". Note:
                                  Sunday is the first day and must be the first
                                  'S' to be parsed correctly. All other chars
                                  may be out of order if needed.
  --wh-start        <HH:MM>     Specify a time in an HOURS:MINS format that
                                  specifies when the implant may start connecting
                                  with the C2 Server. This setting takes affect
                                  if there is a day or end value set.
  --wh-end          <HH:MM>     Specify a time in an HOURS:MINS format that
                                  specifies when the implant will stop connecting
                                  with the C2 Server. This setting will apply
                                  regardless of day or start setting and if a
                                  day or start time does NOT exist, this will
                                  instruct the implant to wait until midnight to
                                  try again.

 CONNECTION HINTS (Max 1 per Profile Group):
  --tcp                         Use the TCP Connection hint.
  --tls                         Use the TLS Connection hint.
  --udp                         Use the UDP Connection hint.
  --icmp                        Use the ICMP (Ping) Connection hint.
  --pipe                        Use the Windows Named Pipes or UNIX file
                                  pipes Connection hint.
  --tls-insecure
  -K                            Use the TLSNoVerify Connection hint.

  --ip              <protocol>  Use the IP Connection hint with the
                                  specified protocol number [0-255].
  --wc2-url         <url>         Use the WC2 Connection hint with the
                                  URL expression or static string.
                                  This can be used with other WC2
                                  arguments without an error.
  --wc2-host        <host>      Use the WC2 Connection hint with the
                                  Host expression or static string.
                                  This can be used with other WC2
                                  arguments without an error.
  --wc2-agent       <agent>     Use the WC2 Connection hint with the
                                  User-Agent expression or static string.
                                  This can be used with other WC2
                                  arguments without an error.
  --wc2-header      <key>=<val>
  -H                <key>=<val> Use the WC2 Connection hint with the
                                  HTTP header expression or static string in
                                  a key=value formnat. This value will be
                                  parsed and will fail if 'key' is empty or
                                  no '=' is present in the string. This may be
                                  specified multiple times. This can be used
                                  with other WC2 arguments without an error.

  --mtls                        Use Mutual TLS Authentication (mTLS) with a TLS
                                  Connection hint. This just enables the flag for
                                  client auth and will fail if '--tls-pem' and
                                  '--tls-key' are empty or not specified.
  --tls-ver         <version>   Use the TLS version specified when using a TLS
                                  Connection hint. This will set the version
                                  required and can be used by itself. A value of
                                  zero means TLSv1. Can be used with other TLS
                                  options.
  --tls-ca          <file|pem>  Use the provided certificate to verify the server
                                  (for clients) or verify clients (for the server).
                                  Can be used on it's own and with '--mtls'.
                                  This argument can take a file path to a PEM
                                  formatted certificate or raw base64 encoded PEM
                                  data.
  --tls-pem         <file|pem>  Use the provided certificate for the generated
                                  TLS socket. This can be used for client or
                                  server listeners. Requires '--tls-key'.
                                  This argument can take a file path to a PEM
                                  formatted certificate or raw base64 encoded PEM
                                  data.
  --tls-key         <file|pem>  Use the provided certificate key for the generated
                                  TLS socket. This can be used for client or
                                  server listeners. Requires '--tls-pem'.
                                  This argument can take a file path to a PEM
                                  formatted certificate private key or a raw
                                  base64 encoded PEM private key.

 WRAPPERS (Multiple different types may be used):
  --hex                         Use the HEX Wrapper.
  --zlib                        Use the Zlib compression Wrapper.
  --gzip                        Use the Gzip compression Wrapper.
  --b64                         Use the Base64 encoding Wrapper.
  --xor             [key]       Encrypt with the XOR Wrapper using the provided
                                  key string. If omitted the key will be a
                                  randomly generated 64 byte array.
  --cbk             [key]       Encrypt with the XOR Wrapper using the provided
                                  key string. If omitted the key will be a
                                  randomly seeded ABCD from a 64 byte array.
  --aes             [key]       Encrypt with the AES Wrapper using the provided
                                  key string. If omitted the key will be a
                                  randomly generated 32 byte array. The AES IV
                                  may be supplied using the '--aes-iv' argument.
                                  If not specified a 16 byte IV will be generated.
  --aes-iv          [iv]        Encrypt with the AES Wrapper using the provided
                                  IV string. If omitted the IV will be a
                                  randomly generated 16 byte array. The AES key
                                  may be supplied using the '--aes-key' argument.
                                  If not specified a 32 byte key will be generated.

 TRANSFORMS (Max 1 per Profile Group):
  --b64t            [shift]     Transform the data using a Base64 Transform. An
                                  option shift value [0-255] may be specified, but
                                  if omitted will not shift.
  --dns             [domain,*]  Use the DNS Packet Transform. optional DNS
                                  domain names may be specified (seperated by space)
                                  that will be used in the packets. This option may
                                  be used more than once to specify more domains.
  -D                [domain,*]

SELECTOR VALUES
 last                           Switch Profile Group ONLY if the last
                                  attempt failed, the default setting.
 random                         Switch Profile Group in a random order
                                  on EVERY connection attempt.
 round-robin                    Switch Profile Group in a weighted order
                                  on EVERY connection attempt. Affected
                                  by Group Weights (lower is better).
 semi-random                    Switch Profile Group in a random order
                                  dependent on a 25% chance. If the
                                  chance fails (75%), do not switch.
 semi-round-robin               Switch Profile Group in a weighted order
                                  dependent on a 25% chance. If the chance
                                  fails (75%), do not switch. Affected by
                                  Group Weights (lower is better).
"""


class Cfg:
    class Const:
        HOST = 0xA0
        SLEEP = 0xA1
        JITTER = 0xA2
        WEIGHT = 0xA3
        KILLDATE = 0xA4
        WORKHOURS = 0xA5
        SEPARATOR = 0xFA

        LAST_VALID = 0xAA
        ROUND_ROBIN = 0xAB
        RANDOM = 0xAC
        SEMI_ROUND_ROBIN = 0xAD
        SEMI_RANDOM = 0xAE

        TCP = 0xC0
        TLS = 0xC1
        UDP = 0xC2
        ICMP = 0xC3
        PIPE = 0xC4
        TLS_INSECURE = 0xC5

        IP = 0xB0
        WC2 = 0xB1
        TLS_EX = 0xB2
        MTLS = 0xB3
        TLS_CA = 0xB4
        TLS_CERT = 0xB5

        HEX = 0xD0
        ZLIB = 0xD1
        GZIP = 0xD2
        B64 = 0xD3
        XOR = 0xD4
        CBK = 0xD5
        AES = 0xD6

        B64T = 0xE0
        DNS = 0xE1
        B64S = 0xE2

        NAMES = {
            HOST: "host",
            SLEEP: "sleep",
            JITTER: "jitter",
            WEIGHT: "weight",
            KILLDATE: "killdate",
            WORKHOURS: "workhours",
            SEPARATOR: "|",
            LAST_VALID: "select-last",
            ROUND_ROBIN: "select-round-robin",
            RANDOM: "select-random",
            SEMI_ROUND_ROBIN: "select-semi-round-robin",
            SEMI_RANDOM: "select-semi-random",
            TCP: "tcp",
            TLS: "tls",
            UDP: "udp",
            ICMP: "icmp",
            PIPE: "pipe",
            TLS_INSECURE: "tls-insecure",
            IP: "ip",
            WC2: "wc2",
            TLS_EX: "tls-ex",
            MTLS: "mtls",
            TLS_CA: "tls-ca",
            TLS_CERT: "tls-cert",
            HEX: "hex",
            ZLIB: "zlib",
            GZIP: "gzip",
            B64: "base64",
            XOR: "xor",
            CBK: "cbk",
            AES: "aes",
            B64T: "b64t",
            DNS: "dns",
            B64S: "b64s",
        }
        NAMES_TO_ID = {
            "host": HOST,
            "sleep": SLEEP,
            "jitter": JITTER,
            "weight": WEIGHT,
            "killdate": KILLDATE,
            "workhours": WORKHOURS,
            "|": SEPARATOR,
            "select-last": LAST_VALID,
            "select-round-robin": ROUND_ROBIN,
            "select-random": RANDOM,
            "select-semi-round-robin": SEMI_ROUND_ROBIN,
            "select-semi-random": SEMI_RANDOM,
            "tcp": TCP,
            "tls": TLS,
            "udp": UDP,
            "icmp": ICMP,
            "pipe": PIPE,
            "tls-insecure": TLS_INSECURE,
            "ip": IP,
            "wc2": WC2,
            "tls-ex": TLS_EX,
            "mtls": MTLS,
            "tls-ca": TLS_CA,
            "tls-cert": TLS_CERT,
            "hex": HEX,
            "zlib": ZLIB,
            "gzip": GZIP,
            "base64": B64,
            "xor": XOR,
            "cbk": CBK,
            "aes": AES,
            "b64t": B64T,
            "dns": DNS,
            "b64s": B64S,
        }
        WRAPPERS = ["hex", "zlib", "gzip", "b64", "xor", "aes", "cbk"]

        @staticmethod
        def as_single(v):
            if not isinstance(v, int):
                raise ValueError("single: invalid bit")
            s = Setting(1)
            s[0] = v
            return s

    @staticmethod
    def host(v):
        if not Utils.nes(v):
            raise ValueError("host: invalid name object")
        f = v.encode("UTF-8")
        n = len(f)
        if n > 0xFFFF:
            n = 0xFFFF
        s = Setting(3 + n)
        s[0] = Cfg.Const.HOST
        s[1] = (n >> 8) & 0xFF
        s[2] = n & 0xFF
        for x in range(0, n):
            s[x + 3] = f[x]
        del n, f
        return s

    @staticmethod
    def sleep(t):
        if Utils.nes(t):
            t = Utils.str_to_dur(t)
        if not isinstance(t, int) or t <= 0:
            raise ValueError("sleep: invalid duration")
        s = Setting(9)
        s[0] = Cfg.Const.SLEEP
        s[1] = (t >> 56) & 0xFF
        s[2] = (t >> 48) & 0xFF
        s[3] = (t >> 40) & 0xFF
        s[4] = (t >> 32) & 0xFF
        s[5] = (t >> 24) & 0xFF
        s[6] = (t >> 16) & 0xFF
        s[7] = (t >> 8) & 0xFF
        s[8] = t & 0xFF
        return s

    @staticmethod
    def jitter(p):
        if Utils.nes(p):
            if "%" in p:
                p = p.replace("%", "")
            try:
                p = int(p)
            except ValueError:
                raise ValueError("jitter: invalid percentage")
        if not isinstance(p, int) or p < 0 or p > 100:
            raise ValueError("jitter: invalid percentage")
        s = Setting(2)
        s[0] = Cfg.Const.JITTER
        s[1] = p & 0xFF
        return s

    @staticmethod
    def weight(p):
        if Utils.nes(p):
            try:
                p = int(p)
            except ValueError:
                raise ValueError("weight: invalid value")
        if not isinstance(p, int) or p < 0 or p > 100:
            raise ValueError("weight: invalid value")
        s = Setting(2)
        s[0] = Cfg.Const.WEIGHT
        s[1] = p & 0xFF
        return s

    @staticmethod
    def wrap_hex():
        return Cfg.Const.as_single(Cfg.Const.HEX)

    @staticmethod
    def wrap_b64():
        return Cfg.Const.as_single(Cfg.Const.B64)

    @staticmethod
    def wrap_zlib():
        return Cfg.Const.as_single(Cfg.Const.ZLIB)

    @staticmethod
    def wrap_gzip():
        return Cfg.Const.as_single(Cfg.Const.GZIP)

    @staticmethod
    def separator():
        return Cfg.Const.as_single(Cfg.Const.SEPARATOR)

    @staticmethod
    def killdate(v):
        if v is None:
            t = 0
        elif isinstance(v, datetime):
            t = int(v.timestamp())
        elif not isinstance(v, str):
            raise ValueError("killdate: invalid date value type")
        elif len(v) == 0:
            t = 0
        else:
            t = int(datetime.fromisoformat(v).timestamp())
        s = Setting(9)
        s[0] = Cfg.Const.KILLDATE
        s[1] = (t >> 56) & 0xFF
        s[2] = (t >> 48) & 0xFF
        s[3] = (t >> 40) & 0xFF
        s[4] = (t >> 32) & 0xFF
        s[5] = (t >> 24) & 0xFF
        s[6] = (t >> 16) & 0xFF
        s[7] = (t >> 8) & 0xFF
        s[8] = t & 0xFF
        del t
        return s

    @staticmethod
    def connect_ip(p):
        if not isinstance(p, int) or p <= 0 or p > 0xFF:
            raise ValueError("ip: invalid protocol")
        s = Setting(2)
        s[0] = Cfg.Const.IP
        s[1] = p & 0xFF
        return s

    @staticmethod
    def connect_tcp():
        return Cfg.Const.as_single(Cfg.Const.TCP)

    @staticmethod
    def connect_tls():
        return Cfg.Const.as_single(Cfg.Const.TLS)

    @staticmethod
    def connect_udp():
        return Cfg.Const.as_single(Cfg.Const.UDP)

    @staticmethod
    def connect_icmp():
        return Cfg.Const.as_single(Cfg.Const.ICMP)

    @staticmethod
    def connect_pipe():
        return Cfg.Const.as_single(Cfg.Const.PIPE)

    @staticmethod
    def transform_b64():
        return Cfg.Const.as_single(Cfg.Const.B64T)

    @staticmethod
    def transform_dns(n):
        if not isinstance(n, list):
            raise ValueError("dns: invalid names list")
        s = Setting(2)
        s[0] = Cfg.Const.DNS
        s[1] = len(n) & 0xFF
        for i in n:
            if not Utils.nes(i):
                raise ValueError("dns: invalid name value")
            if len(i) > 0xFF:
                i = i[:0xFF]
            s.append(len(i) & 0xFF)
            s += i.encode("UTF-8")
        return s

    @staticmethod
    def selector_random():
        return Cfg.Const.as_single(Cfg.Const.RANDOM)

    @staticmethod
    def connect_tls_ex(v):
        if not isinstance(v, int) or v <= 0 or v > 0xFF:
            raise ValueError("tls-ex: invalid version")
        s = Setting(2)
        s[0] = Cfg.Const.TLS_EX
        s[1] = v & 0xFF
        return s

    @staticmethod
    def wrap_xor(key=None):
        if isinstance(key, list) and len(key) > 0:
            key = key[0]
        if key is None:
            key = token_bytes(64)
        elif Utils.nes(key):
            key = key.encode("UTF-8")
        elif not isinstance(key, (bytes, bytearray)):
            raise ValueError("xor: invalid KEY value")
        n = len(key)
        if n > 0xFFFF:
            n = 0xFFFF
        s = Setting(3 + n)
        s[0] = Cfg.Const.XOR
        s[1] = (n >> 8) & 0xFF
        s[2] = n & 0xFF
        for x in range(0, n):
            s[x + 3] = key[x]
        del n
        return s

    @staticmethod
    def connect_tls_ca(v, ca):
        if isinstance(ca, (bytes, bytearray)):
            f = ca
        elif Utils.nes(ca):
            f = ca.encode("UTF-8")
        else:
            raise ValueError("tls-ca: invalid CA")
        if not isinstance(v, int) or v <= 0 or v > 0xFF:
            raise ValueError("tls-ca: invalid version")
        n = len(f)
        if n > 0xFFFF:
            n = 0xFFFF
        s = Setting(4 + n)
        s[0] = Cfg.Const.TLS_CA
        s[1] = v & 0xFF
        s[2] = (n >> 8) & 0xFF
        s[3] = n & 0xFF
        for x in range(0, n):
            s[x + 4] = f[x]
        del f, n
        return s

    @staticmethod
    def selector_last_valid():
        return Cfg.Const.as_single(Cfg.Const.LAST_VALID)

    @staticmethod
    def transform_b64_shift(v):
        if not isinstance(v, int) or v <= 0 or v > 0xFF:
            raise ValueError("base64s: invalid shift")
        s = Setting(2)
        s[0] = Cfg.Const.B64S
        s[1] = v & 0xFF
        return s

    @staticmethod
    def selector_round_robin():
        return Cfg.Const.as_single(Cfg.Const.ROUND_ROBIN)

    @staticmethod
    def selector_semi_random():
        return Cfg.Const.as_single(Cfg.Const.SEMI_RANDOM)

    @staticmethod
    def connect_tls_insecure():
        return Cfg.Const.as_single(Cfg.Const.TLS_INSECURE)

    @staticmethod
    def select_semi_round_robin():
        return Cfg.Const.as_single(Cfg.Const.SEMI_ROUND_ROBIN)

    @staticmethod
    def wrap_aes(key=None, iv=None):
        if isinstance(iv, list) and len(iv) > 0:
            iv = iv[0]
        if isinstance(key, list) and len(key) > 0:
            key = key[0]
        if key is None:
            key = token_bytes(32)
        elif Utils.nes(key):
            key = key.encode("UTF-8")
        elif not isinstance(key, (bytes, bytearray)):
            raise ValueError("aes: invalid KEY value")
        if iv is None:
            iv = token_bytes(16)
        elif Utils.nes(iv):
            iv = iv.encode("UTF-8")
        elif not isinstance(iv, (bytes, bytearray)):
            raise ValueError("aes: invalid IV value")
        if len(key) > 32:
            raise ValueError("aes: invalid KEY size")
        if len(iv) != 16:
            raise ValueError("aes: invalid IV size")
        s = Setting(3 + len(key) + len(iv))
        s[0] = Cfg.Const.AES
        s[1] = len(key) & 0xFF
        s[2] = len(iv) & 0xFF
        for x in range(0, len(key)):
            s[x + 3] = key[x]
        for x in range(0, len(iv)):
            s[x + len(key) + 3] = iv[x]
        return s

    @staticmethod
    def workhours(days, start, end):
        if not Utils.nes(days) and not Utils.nes(start) and not Utils.nes(end):
            raise ValueError("workhours: empty values specified")
        d = 0
        if Utils.nes(days):
            d = Utils.parse_weekdays(days)
        h, j = 0, 0
        if Utils.nes(start):
            if ":" not in start:
                raise ValueError("workhours: invalid start format")
            x = start.split(":")
            if len(x) != 2:
                raise ValueError("workhours: invalid start format")
            try:
                h, j = int(x[0]), int(x[1])
            except ValueError:
                raise ValueError("workhours: invalid start format")
            del x
            if h > 23 or j > 59:
                raise ValueError("workhours: invalid start format")
        n, m = 0, 0
        if Utils.nes(end):
            if ":" not in end:
                raise ValueError("workhours: invalid end format")
            x = end.split(":")
            if len(x) != 2:
                raise ValueError("workhours: invalid end format")
            try:
                n, m = int(x[0]), int(x[1])
            except ValueError:
                raise ValueError("workhours: invalid end format")
            del x
            if n > 23 or m > 59:
                raise ValueError("workhours: invalid end format")
        s = Setting(6)
        s[0] = Cfg.Const.WORKHOURS
        s[1] = d
        s[2] = h
        s[3] = j
        s[4] = n
        s[5] = m
        del d, h, j, n, m
        return s

    @staticmethod
    def connect_mtls(v, ca, pem, key):
        if isinstance(ca, (bytes, bytearray)):
            f = ca
        elif Utils.nes(ca):
            f = ca.encode("UTF-8")
        else:
            raise ValueError("mtls: invalid CA")
        if isinstance(pem, (bytes, bytearray)):
            if len(pem) == 0:
                raise ValueError("mtls: invalid PEM")
            p = pem
        elif Utils.nes(pem):
            p = pem.encode("UTF-8")
        else:
            raise ValueError("mtls: invalid PEM")
        if isinstance(key, (bytes, bytearray)):
            if len(key) == 0:
                raise ValueError("mtls: invalid KEY")
            k = key
        elif Utils.nes(key):
            k = key.encode("UTF-8")
        else:
            raise ValueError("mtls: invalid KEY")
        if not isinstance(v, int) or v <= 0 or v > 0xFF:
            raise ValueError("mtls invalid version")
        if len(p) == 0 or len(k) == 0:
            raise ValueError("mtls: invalid PEM or KEY version")
        o = len(f)
        if o > 0xFFFF:
            o = 0xFFFF
        n = len(p)
        if n > 0xFFFF:
            n = 0xFFFF
        m = len(k)
        if m > 0xFFFF:
            m = 0xFFFF
        s = Setting(8 + o + n + m)
        s[0] = Cfg.Const.MTLS
        s[1] = v & 0xFF
        s[2] = (o >> 8) & 0xFF
        s[3] = o & 0xFF
        s[2] = (n >> 8) & 0xFF
        s[3] = n & 0xFF
        s[4] = (m >> 8) & 0xFF
        s[5] = m & 0xFF
        for x in range(0, n):
            s[x + 8] = f[x]
        for x in range(0, n):
            s[x + o + 8] = p[x]
        for x in range(0, m):
            s[x + o + n + 8] = k[x]
        del f, p, k, o, n, m
        return s

    @staticmethod
    def connect_tls_cert(v, pem, key):
        if isinstance(pem, (bytes, bytearray)):
            if len(pem) == 0:
                raise ValueError("tls-cert: invalid PEM")
            p = pem
        elif Utils.nes(pem):
            p = pem.encode("UTF-8")
        else:
            raise ValueError("tls-cert: invalid PEM")
        if isinstance(key, (bytes, bytearray)):
            if len(key) == 0:
                raise ValueError("tls-cert: invalid KEY")
            k = key
        elif Utils.nes(key):
            k = key.encode("UTF-8")
        else:
            raise ValueError("tls-cert: invalid KEY")
        if not isinstance(v, int) or v <= 0 or v > 0xFF:
            raise ValueError("tls-cert: invalid version")
        if len(p) == 0 or len(k) == 0:
            raise ValueError("tls-cert: invalid PEM or KEY version")
        n = len(p)
        if n > 0xFFFF:
            n = 0xFFFF
        m = len(k)
        if m > 0xFFFF:
            m = 0xFFFF
        s = Setting(6 + n + m)
        s[0] = Cfg.Const.TLS_CA
        s[1] = v & 0xFF
        s[2] = (n >> 8) & 0xFF
        s[3] = n & 0xFF
        s[4] = (m >> 8) & 0xFF
        s[5] = m & 0xFF
        for x in range(0, n):
            s[x + 6] = p[x]
        for x in range(0, m):
            s[x + n + 6] = k[x]
        del p, k, n, m
        return s

    @staticmethod
    def connect_wc2(u, h, a, head=None):
        if Utils.nes(u):
            c = u.encode("UTF-8")
        else:
            c = bytearray()
        if Utils.nes(h):
            v = h.encode("UTF-8")
        else:
            v = bytearray()
        if Utils.nes(a):
            b = a.encode("UTF-8")
        else:
            b = bytearray()
        j = len(c)
        if j > 0xFFFF:
            j = 0xFFFF
        k = len(v)
        if k > 0xFFFF:
            k = 0xFFFF
        n = len(b)
        if n > 0xFFFF:
            n = 0xFFFF
        s = Setting(8 + j + k + n)
        s[0] = Cfg.Const.WC2
        s[1] = (j >> 8) & 0xFF
        s[2] = j & 0xFF
        s[3] = (k >> 8) & 0xFF
        s[4] = k & 0xFF
        s[5] = (n >> 8) & 0xFF
        s[6] = n & 0xFF
        for x in range(0, j):
            s[x + 8] = c[x]
        for x in range(0, k):
            s[x + j + 8] = v[x]
        for x in range(0, n):
            s[x + j + k + 8] = b[x]
        del j, k, n, c, v, b
        if not isinstance(head, dict):
            s[7] = 0
            return s
        i = 0
        s[7] = len(head) & 0xFF
        for k, v in head.items():
            if i >= 0xFF:
                break
            if not Utils.nes(k):
                raise ValueError("wc2: invalid header")
            if Utils.nes(v):
                z = v.encode("UTF-8")
            else:
                z = bytearray()
            o = k.encode("UTF-8")
            f = len(o)
            if f > 0xFF:
                f = 0xFF
            g = len(z)
            if g > 0xFF:
                g = 0xFF
            s.append(f & 0xFF)
            s.append(g & 0xFF)
            s.extend(o)
            s.extend(z)
            i += 1
            del o, z, f, g
        return s

    @staticmethod
    def wrap_cbk(a=None, b=None, c=None, d=None, size=128, key=None):
        if (
            not isinstance(a, int)
            and not isinstance(b, int)
            and not isinstance(c, int)
            and not isinstance(d, int)
        ):
            if Utils.nes(key):
                v = key.encode("UTF-8")
            elif isinstance(key, (bytes, bytearray)):
                v = key
            else:
                v = token_bytes(64)
            if len(v) == 0:
                v - token_bytes(64)
            h = sha512()
            for _ in range(0, 256):
                h.update(v)
            del v
            n = crc32(h.digest()).to_bytes(4, byteorder="big", signed=False)
            del h
            a = n[0]
            b = n[1]
            c = n[2]
            d = n[3]
            del n
        if (
            not isinstance(a, int)
            or not isinstance(b, int)
            or not isinstance(c, int)
            or not isinstance(d, int)
            or a < 0
            or a > 0xFF
            or b < 0
            or b > 0xFF
            or c < 0
            or c > 0xFF
            or d < 0
            or d > 0xFF
        ):
            raise ValueError("cbk: invalid ABCD keys")
        if not isinstance(size, int) or size not in [16, 32, 64, 128]:
            raise ValueError("cbk: invalid size")
        s = Setting(6)
        s[0] = Cfg.Const.CBK
        s[1] = size & 0xFF
        s[2] = a
        s[3] = b
        s[4] = c
        s[5] = d
        return s


class Utils:
    UNITS = {
        "ns": 1,
        "us": 1000,
        "µs": 1000,
        "μs": 1000,
        "ms": 1000000,
        "s": 1000000000,
        "m": 60000000000,
        "h": 3600000000000,
    }

    @staticmethod
    def dur_to_str(v):
        b = bytearray(32)
        n = len(b) - 1
        b[n] = ord("s")
        n, v = Utils._fmt_frac(b, n, v)
        n = Utils._fmt_int(b, n, v % 60)
        v /= 60
        if int(v) > 0:
            n -= 1
            b[n] = ord("m")
            n = Utils._fmt_int(b, n, v % 60)
            v /= 60
            if int(v) > 0:
                n -= 1
                b[n] = ord("h")
                n = Utils._fmt_int(b, n, v)
        return b[n:].decode("UTF-8")

    @staticmethod
    def str_to_dur(s):
        if not Utils.nes(s):
            raise ValueError("str2dur: invalid duration")
        if s == "0":
            return 0
        d = 0
        while len(s) > 0:
            v = 0
            f = 0
            z = 1
            if not (s[0] == "." or (ord("0") <= ord(s[0]) and ord(s[0]) <= ord("9"))):
                raise ValueError("str2dur: invalid duration")
            p = len(s)
            v, s = Utils._leading_int(s)
            r = p != len(s)
            y = False
            if len(s) > 0 and s[0] == ".":
                s = s[1:]
                p = len(s)
                f, z, s = Utils._leading_fraction(s)
                y = p != len(s)
            if not r and not y:
                raise ValueError("str2dur: invalid duration")
            del r, y
            i = 0
            while i < len(s):
                c = ord(s[i])
                if c == ord(".") or (ord("0") <= c and c <= ord("9")):
                    break
                i += 1
            if i == 0:
                u = "s"
            else:
                u = s[:i]
                s = s[i:]
            del i
            if u not in Utils.UNITS:
                raise ValueError("str2dur: unknown unit")
            e = Utils.UNITS[u]
            del u
            if v > (((1 << 63) - 1) / e):
                raise ValueError("str2dur: invalid duration")
            v *= int(e)
            if f > 0:
                v += int(float(f) * float(float(e) / float(z)))
                if v < 0:
                    raise ValueError("str2dur: invalid duration")
            del e
            d += v
            if d < 0:
                raise ValueError("str2dur: invalid duration")
            del v, f, z
        return d

    @staticmethod
    def to_weekdays(v):
        if not isinstance(v, (int, float)) or v == 0 or v > 126:
            return "SMTWRFS"
        r = ""
        if v & 1 != 0:
            r += "S"
        if v & 2 != 0:
            r += "M"
        if v & 4 != 0:
            r += "T"
        if v & 8 != 0:
            r += "W"
        if v & 16 != 0:
            r += "R"
        if v & 32 != 0:
            r += "F"
        if v & 64 != 0:
            r += "S"
        return r

    @staticmethod
    def _leading_int(s):
        i = 0
        x = 0
        while i < len(s):
            c = ord(s[i])
            if c < ord("0") or c > ord("9"):
                break
            if x > (((1 << 63) - 1) / 10):
                raise OverflowError()
            x = int(x * 10) + int(c) - ord("0")
            if x < 0:
                raise OverflowError()
            i += 1
        return x, s[i:]

    @staticmethod
    def parse_weekdays(v):
        if not Utils.nes(v):
            return 0
        d = 0
        for x in range(0, len(v)):
            if v[x] == "s" or v[x] == "S":
                if x == 0:
                    d |= 1
                else:
                    d |= 64
            elif v[x] == "m" or v[x] == "M":
                d |= 2
            elif v[x] == "t" or v[x] == "T":
                d |= 4
            elif v[x] == "w" or v[x] == "W":
                d |= 8
            elif v[x] == "r" or v[x] == "R":
                d |= 16
            elif v[x] == "f" or v[x] == "F":
                d |= 32
            else:
                raise ValueError("bad weekday char")
        return d

    @staticmethod
    def _fmt_int(b, s, v):
        if int(v) == 0:
            s -= 1
            b[s] = ord("0")
            return s
        while int(v) > 0:
            s -= 1
            b[s] = int(v % 10) + ord("0")
            v /= 10
        return s

    @staticmethod
    def split_dns_names(v):
        n = list()
        for e in v:
            if not isinstance(e, list):
                raise ValueError("dns: invalid argument")
            if len(e) == 0:
                continue
            for s in e:
                if not Utils.nes(s):
                    raise ValueError("dns: invalid value")
                if "," not in s:
                    n.append(s)
                    continue
                for x in s.split(","):
                    h = x.strip()
                    if len(h) > 0:
                        n.append(h)
                    del h
        return n

    @staticmethod
    def _fmt_frac(b, s, v):
        p = False
        for _ in range(0, 9):
            d = v % 10
            p = p or d != 0
            if p:
                s -= 1
                b[s] = int(d) + ord("0")
            v /= 10
        if p:
            s -= 1
            b[s] = ord(".")
        del p
        return s, v

    @staticmethod
    def read_file_input(v):
        if v.strip() == "-" and not stdin.isatty():
            if hasattr(stdin, "buffer"):
                b = stdin.buffer.read()
            else:
                b = stdin.read()
            stdin.close()
        else:
            with open(v, "rb") as f:
                b = f.read()
        if len(b) == 0:
            raise ValueError("input: empty input data")
        return Config(b)

    @staticmethod
    def _leading_fraction(s):
        i = 0
        x = 0
        v = 1
        o = False
        while i < len(s):
            c = ord(s[i])
            if c < ord("0") or c > ord("9"):
                break
            if o:
                continue
            if x > (((1 << 63) - 1) / 10):
                o = True
                continue
            y = int(x * 10) + int(c) - ord("0")
            if y < 0:
                o = True
                continue
            x = y
            v *= 10
            i += 1
            del y
        del o
        return x, v, s[1:]

    @staticmethod
    def parse_wc2_headers(v):
        if not isinstance(v, list) or len(v) == 0:
            return None
        d = dict()
        for e in v:
            Utils._parse_wc2_header(d, e, False)
        if len(d) == 0:
            return None
        return d

    @staticmethod
    def nes(s, min=0, max=-1):
        if max > min:
            return isinstance(s, str) and len(s) < max and len(s) > min
        return isinstance(s, str) and len(s) > min

    @staticmethod
    def _parse_wc2_header(d, e, r):
        if isinstance(e, str):
            if len(e) == 0 or "=" not in e:
                raise ValueError("wc2: invalid header")
            p = e.find("=")
            if p == 0 or p == len(e) - 1:
                raise ValueError("wc2: empty header")
            d[e[:p].strip()] = e[p + 1 :].strip()
            return
        if isinstance(e, list) and len(e) > 0:
            if r:
                raise ValueError("wc2: too many nested lists")
            for v in e:
                Utils._parse_wc2_header(d, v, True)
            return
        raise ValueError("wc2: Invalid header")

    @staticmethod
    def parse_tls(ca, pem, key, mtls, ver):
        a = None
        p = None
        k = None
        if isinstance(ca, str) and len(ca) > 0:
            try:
                a = b64decode(ca, validate=True)
            except ValueError:
                with open(ca, "rb") as f:
                    a = f.read()
        if isinstance(pem, str) and len(pem) > 0:
            try:
                a = b64decode(pem, validate=True)
            except ValueError:
                with open(pem, "rb") as f:
                    p = f.read()
        if isinstance(key, str) and len(key) > 0:
            try:
                a = b64decode(key, validate=True)
            except ValueError:
                with open(key, "rb") as f:
                    k = f.read()
        if mtls and (p is None or k is None or a is None):
            raise ValueError("mtls: CA, PEM and KEY must be provided")
        if (p is not None and k is None) or (k is not None and p is None):
            raise ValueError("tls-cert: PEM and KEY must be provided")
        if not isinstance(ver, int):
            ver = 0
        if a is None and p is None and k is None:
            return Cfg.connect_tls_ex(ver)
        if a is not None and p is None and k is None:
            return Cfg.connect_tls_ca(ver, a)
        if a is None:
            return Cfg.connect_tls_certs(ver, p, k)
        return Cfg.connect_mtls(ver, a, p, k)

    @staticmethod
    def write_file_output(c, v, pretty, json):
        f = stdout
        if Utils.nes(v) and v != "-":
            if not pretty and not json:
                f = open(v, "wb")
            else:
                f = open(v, "w")
        try:
            if pretty or json:
                return print(
                    dumps(c.json(), sort_keys=False, indent=(4 if pretty else None)),
                    file=f,
                )
            if f == stdout and not f.isatty():
                return f.buffer.write(c)
            if f.mode == "wb":
                return f.write(c)
            f.write(b64encode(c).decode("UTF-8"))
        finally:
            if f == stdout:
                print(end="")
            else:
                f.close()
            del f


class Config(bytearray):
    __slots__ = ("_connector", "_transform")

    def __init__(self, b=None):
        self._connector = False
        self._transform = False
        if isinstance(b, str) and len(b) > 0:
            if b[0] == "[" and b[-1].strip() == "]":
                return self.parse(b)
            return self.extend(b64decode(b, validate=True))
        if isinstance(b, (bytes, bytearray)) and len(b) > 0:
            if b[0] == 91 and b.decode("UTF-8", "ignore").strip()[-1] == "]":
                return self.parse(b.decode("UTF-8"))
            return self.extend(b)

    def json(self):
        i = 0
        n = 0
        e = list()
        r = list()
        while n >= 0 and n < len(self):
            n = self.next(i)
            if self[i] not in Cfg.Const.NAMES:
                raise ValueError(f"json: invalid setting id {self[i]}")
            if self[i] == Cfg.Const.SEPARATOR:
                i = n
                if len(e) == 0:
                    i = n
                    continue
                if n == len(self):
                    break
                r.append(e)
                e = list()
                continue
            o = None
            if Setting.is_single(self[i]):
                pass
            elif self[i] == Cfg.Const.HOST:
                if i + 3 >= n:
                    raise ValueError("host: invalid setting")
                v = (int(self[i + 2]) | int(self[i + 1]) << 8) + i
                if v > n or v < i:
                    raise ValueError("host: invalid setting")
                o = self[i + 3 : v + 3].decode("UTF-8")
                del v
            elif self[i] == Cfg.Const.SLEEP:
                if i + 8 >= n:
                    raise ValueError("sleep: invalid setting")
                o = Utils.dur_to_str(
                    (
                        int(self[i + 8])
                        | int(self[i + 7]) << 8
                        | int(self[i + 6]) << 16
                        | int(self[i + 5]) << 24
                        | int(self[i + 4]) << 32
                        | int(self[i + 3]) << 40
                        | int(self[i + 2]) << 48
                        | int(self[i + 1]) << 56
                    )
                )
            elif self[i] == Cfg.Const.KILLDATE:
                if i + 8 >= n:
                    raise ValueError("killdate: invalid setting")
                u = (
                    int(self[i + 8])
                    | int(self[i + 7]) << 8
                    | int(self[i + 6]) << 16
                    | int(self[i + 5]) << 24
                    | int(self[i + 4]) << 32
                    | int(self[i + 3]) << 40
                    | int(self[i + 2]) << 48
                    | int(self[i + 1]) << 56
                )
                if u == 0:
                    o = ""
                else:
                    o = datetime.fromtimestamp(u).isoformat()
            elif self[i] == Cfg.Const.WORKHOURS:
                if i + 5 >= n:
                    raise ValueError("workhours: invalid setting")
                o = {
                    "start_hour": self[i + 2],
                    "start_min": self[i + 3],
                    "end_hour": self[i + 4],
                    "end_min": self[i + 5],
                    "days": Utils.to_weekdays(self[i + 1]),
                }
            elif (
                self[i] == Cfg.Const.IP
                or self[i] == Cfg.Const.B64S
                or self[i] == Cfg.Const.JITTER
                or self[i] == Cfg.Const.WEIGHT
                or self[i] == Cfg.Const.TLS_EX
            ):
                if i + 1 >= n:
                    raise ValueError("invalid setting")
                o = int(self[i + 1])
            elif self[i] == Cfg.Const.WC2:
                if i + 7 >= n:
                    raise ValueError("wc2: invalid setting")
                z = i + 8
                v = (int(self[i + 2]) | int(self[i + 1]) << 8) + i + 8
                if v > n or z > n or z < i or v < i:
                    raise ValueError("wc2: invalid setting")
                o = dict()
                if v > z:
                    o["url"] = self[z:v].decode("UTF-8")
                z = v
                v = (int(self[i + 4]) | int(self[i + 3]) << 8) + v
                if v > z:
                    if v > n or z > n or v < z or z < i or v < i:
                        raise ValueError("wc2: invalid setting")
                    o["host"] = self[z:v].decode("UTF-8")
                z = v
                v = (int(self[i + 6]) | int(self[i + 5]) << 8) + v
                if v > z:
                    if v > n or z > n or v < z or z < i or v < i:
                        raise ValueError("wc2: invalid setting")
                    o["agent"] = self[z:v].decode("UTF-8")
                if self[i + 7] > 0:
                    o["headers"] = dict()
                    j = 0
                    while v < n and z < n and j < n:
                        j = int(self[v]) + v + 2
                        z = v + 2
                        v = int(self[v + 1]) + j
                        if (
                            z == j
                            or z > n
                            or j > n
                            or v > n
                            or v < j
                            or j < z
                            or z < i
                            or j < i
                            or v < i
                        ):
                            raise ValueError("wc2: invalid header")
                        o["headers"][self[z:j].decode("UTF-8")] = self[j:v].decode(
                            "UTF-8"
                        )
                    del j
                del z, v
            elif self[i] == Cfg.Const.MTLS:
                if i + 7 >= n:
                    raise ValueError("mtls: invalid setting")
                a = (int(self[i + 3]) | int(self[i + 2]) << 8) + i + 8
                p = (int(self[i + 5]) | int(self[i + 4]) << 8) + a
                k = (int(self[i + 7]) | int(self[i + 6]) << 8) + p
                if a > n or p > n or k > n or p < a or k < p or a < i or p < i or k < i:
                    raise ValueError("mtls: invalid setting")
                o = {"version": int(self[i + 1])}
                o["ca"] = b64encode(self[i + 8 : a]).decode("UTF-8")
                o["pem"] = b64encode(self[a:p]).decode("UTF-8")
                o["key"] = b64encode(self[p:k]).decode("UTF-8")
                del a, p, k
            elif self[i] == Cfg.Const.TLS_CA:
                if i + 4 >= n:
                    raise ValueError("tls-ca: invalid setting")
                a = (int(self[i + 3]) | int(self[i + 2]) << 8) + i + 4
                if a > n or a < i:
                    raise ValueError("tls-ca: invalid setting")
                o = {"version": int(self[i + 1])}
                o["ca"] = b64encode(self[i + 4 : a]).decode("UTF-8")
                del a
            elif self[i] == Cfg.Const.TLS_CERT:
                if i + 6 >= n:
                    raise ValueError("tls-cert: invalid setting")
                p = (int(self[i + 3]) | int(self[i + 2]) << 8) + i + 6
                k = (int(self[i + 5]) | int(self[i + 4]) << 8) + p
                if p > n or k > n or p < i or k < i or k < p:
                    raise ValueError("tls-cert: invalid setting")
                o = {"version": int(self[i + 1])}
                o["pem"] = b64encode(self[i + 6 : p]).decode("UTF-8")
                o["key"] = b64encode(self[p:k]).decode("UTF-8")
                del p, k
            elif self[i] == Cfg.Const.XOR:
                if i + 3 >= n:
                    raise ValueError("xor: invalid setting")
                k = (int(self[i + 2]) | int(self[i + 1]) << 8) + i
                if k > n or k < i:
                    raise ValueError("xor: invalid setting")
                o = b64encode(self[i + 3 : k + 3]).decode("UTF-8")
                del k
            elif self[i] == Cfg.Const.CBK:
                if i + 5 >= n:
                    raise ValueError("cbk: invalid setting")
                o = {
                    "size": int(self[i + 1]),
                    "A": int(self[i + 2]),
                    "B": int(self[i + 3]),
                    "C": int(self[i + 4]),
                    "D": int(self[i + 5]),
                }
            elif self[i] == Cfg.Const.AES:
                if i + 3 >= n:
                    raise ValueError("aes: invalid setting")
                v = int(self[i + 1]) + i + 3
                z = int(self[i + 2]) + v
                if v == z or i + 3 == v or v > n or z > n or z < i or v < i or z < v:
                    raise ValueError("aes: invalid KEY/IV values")
                o = {
                    "key": b64encode(self[i + 3 : v]).decode("UTF-8"),
                    "iv": b64encode(self[v:z]).decode("UTF-8"),
                }
                del v, z
            elif self[i] == Cfg.Const.DNS:
                if i + 1 >= n:
                    raise ValueError("dns: invalid setting")
                o = list()
                v = i + 2
                z = v
                for _ in range(self[i + 1], 0, -1):
                    v += int(self[v]) + 1
                    if (
                        z + 1 > v
                        or z + 1 == v
                        or v < z
                        or v > n
                        or z > n
                        or z < i
                        or v < i
                    ):
                        raise ValueError("dns: invalid name")
                    o.append(self[z + 1 : v].decode("UTF-8"))
                    z = v
                del v, z
            y = {"type": Cfg.Const.NAMES[self[i]]}
            if o is not None:
                y["args"] = o
            del o
            e.append(y)
            i = n
        if len(e) > 0:
            r.append(e)
        return r

    def add(self, s):
        if not isinstance(s, Setting):
            raise ValueError("add: cannot add a non-Settings object")
        if not s._is_valid():
            raise ValueError("add: invalid Settings object")
        if s[0] == Cfg.Const.SEPARATOR:
            self._connector = False
            self._transform = False
        if s._is_connector():
            if self._connector:
                raise ValueError("add: attempted to add multiple Connection hints")
            self._connector = True
        if s._is_transform():
            if self._transform:
                raise ValueError("add: attempted to add multiple Transforms")
            self._transform = True
        if s.single():
            return self.append(s[0])
        self += s
        # for i in s:
        #    self.append(i)

    def __str__(self):
        i = 0
        n = 0
        b = StringIO()
        while n >= 0 and n < len(self):
            n = self.next(i)
            if i > 0:
                b.write(",")
            b.write(Setting(self[i:n]).__str__())
            i = n
        s = b.getvalue()
        b.close()
        del b
        return s

    def read(self, b):
        if not isinstance(b, (bytes, bytearray)):
            raise ValueError("read: invalid raw type")
        self.extend(b)

    def next(self, i):
        if i > len(self) or i < 0:
            return -1
        if Setting.is_single(self[i]):
            return i + 1
        if (
            self[i] == Cfg.Const.IP
            or self[i] == Cfg.Const.B64S
            or self[i] == Cfg.Const.JITTER
            or self[i] == Cfg.Const.WEIGHT
            or self[i] == Cfg.Const.TLS_EX
        ):
            return i + 2
        if self[i] == Cfg.Const.CBK or self[i] == Cfg.Const.WORKHOURS:
            return i + 6
        if self[i] == Cfg.Const.SLEEP or self[i] == Cfg.Const.KILLDATE:
            return i + 9
        if self[i] == Cfg.Const.WC2:
            if i + 7 >= len(self):
                return -1
            n = (
                i
                + 8
                + (int(self[i + 2]) | int(self[i + 1]) << 8)
                + (int(self[i + 4]) | int(self[i + 3]) << 8)
                + (int(self[i + 6]) | int(self[i + 5]) << 8)
            )
            if self[i + 7] == 0 or n >= len(self):
                return n
            for _ in range(self[i + 7], 0, -1):
                if n >= len(self) or n < 0:
                    return -1
                n += int(self[n]) + int(self[n + 1]) + 2
            return n
        if self[i] == Cfg.Const.XOR or self[i] == Cfg.Const.HOST:
            if i + 3 >= len(self):
                return -1
            return i + 3 + int(self[i + 2]) | int(self[i + 1]) << 8
        if self[i] == Cfg.Const.AES:
            if i + 2 >= len(self):
                return -1
            return i + 3 + int(self[i + 1]) + int(self[i + 2])
        if self[i] == Cfg.Const.MTLS:
            if i + 7 >= len(self):
                return -1
            return (
                i
                + 8
                + (int(self[i + 3]) | int(self[i + 2]) << 8)
                + (int(self[i + 5]) | int(self[i + 4]) << 8)
                + (int(self[i + 7]) | int(self[i + 6]) << 8)
            )
        if self[i] == Cfg.Const.TLS_CA:
            if i + 4 >= len(self):
                return -1
            return i + 4 + int(self[i + 3]) | int(self[i + 2]) << 8
        if self[i] == Cfg.Const.TLS_CERT:
            if i + 6 >= len(self):
                return -1
            return (
                i
                + 6
                + (int(self[i + 3]) | int(self[i + 2]) << 8)
                + (int(self[i + 5]) | int(self[i + 4]) << 8)
            )
        if self[i] == Cfg.Const.DNS:
            if i + 1 >= len(self):
                return -1
            n = i + 2
            for _ in range(self[i + 1], 0, -1):
                n += int(self[n]) + 1
            return n
        return -1

    def parse(self, j):
        v = loads(j)
        if not isinstance(v, list):
            raise ValueError("parse: invalid JSON value")
        if len(v) == 0:
            return
        for x in range(0, len(v)):
            if not isinstance(v[x], list):
                raise ValueError("parse: invalid JSON value")
            for e in v[x]:
                self._parse_inner(e)
            if x + 1 < len(v):
                self.add(Cfg.separator())
        del v

    @staticmethod
    def from_file(file):
        with open(file, "rb") as f:
            return Config(f.read())

    def _parse_inner(self, x):
        if not isinstance(x, dict) or len(x) == 0:
            raise ValueError("parse: invalid JSON value")
        if "type" not in x or x["type"].lower() not in Cfg.Const.NAMES_TO_ID:
            raise ValueError("parse: invalid JSON value")
        m = Cfg.Const.NAMES_TO_ID[x["type"].lower()]
        if m == Cfg.Const.SEPARATOR:
            raise ValueError("parse: unexpected separator")
        if Setting.is_single(m):
            return self.add(Cfg.Const.as_single(m))
        if "args" not in x:
            raise ValueError("parse: invalid JSON payload")
        p = x["args"]
        if m == Cfg.Const.HOST:
            if not Utils.nes(p):
                raise ValueError("host: invalid JSON value")
            return self.add(Cfg.host(p))
        if m == Cfg.Const.SLEEP:
            if not Utils.nes(p):
                raise ValueError("sleep: invalid JSON value")
            return self.add(Cfg.sleep(p))
        if m == Cfg.Const.JITTER:
            if not isinstance(p, int):
                raise ValueError("jitter: invalid JSON value")
            if p < 0 or p > 100:
                raise ValueError("jitter: invalid JSON value")
            return self.add(Cfg.jitter(p))
        if m == Cfg.Const.WEIGHT:
            if not isinstance(p, int):
                raise ValueError("weight: invalid JSON value")
            if p < 0:
                raise ValueError("weight: invalid JSON value")
            return self.add(Cfg.weight(p))
        if m == Cfg.Const.KILLDATE:
            if not Utils.nes(p):
                raise ValueError("killdate: invalid JSON value")
            return self.add(Cfg.killdate(p))
        if m == Cfg.Const.WORKHOURS:
            if not isinstance(p, dict):
                raise ValueError("workhours: invalid JSON value")
            h, j = p.get("start_hour"), p.get("start_min")
            n, m = p.get("start_hour"), p.get("start_min")
            if h is None:
                h = 0
            elif not isinstance(h, (int, float)):
                raise ValueError("workhours: invalid JSON value")
            if j is None:
                j = 0
            elif not isinstance(j, (int, float)):
                raise ValueError("workhours: invalid JSON value")
            if n is None:
                m = 0
            elif not isinstance(n, (int, float)):
                raise ValueError("workhours: invalid JSON value")
            if m is None:
                m = 0
            elif not isinstance(m, (int, float)):
                raise ValueError("workhours: invalid JSON value")
            if h > 23 or j > 59 or n > 23 or m > 59:
                raise ValueError("workhours: invalid JSON value")
            d = Utils.parse_weekdays(p.get("days", ""))
            s = Setting(6)
            s[0] = Cfg.Const.WORKHOURS
            s[1] = d
            s[2] = h
            s[3] = j
            s[4] = n
            s[5] = m
            del d, h, j, n, m
            return self.add(s)
        if m == Cfg.Const.IP:
            if not isinstance(p, int):
                raise ValueError("ip: invalid JSON value")
            if p < 0:
                raise ValueError("ip: invalid JSON value")
            return self.add(Cfg.connect_ip(p))
        if m == Cfg.Const.WC2:
            if not isinstance(p, dict):
                raise ValueError("wc2: invalid JSON value")
            u = p.get("url")
            h = p.get("host")
            a = p.get("agent")
            j = p.get("headers")
            if j is not None and not isinstance(j, dict):
                raise ValueError("wc2: invalid JSON header value")
            self.add(Cfg.connect_wc2(u, h, a, j))
            del u, h, a, j
            return
        if m == Cfg.Const.TLS_EX:
            if not isinstance(p, int) and p > 0:
                raise ValueError("tls-ex: invalid JSON value")
            if p < 0:
                raise ValueError("tls-ex: invalid JSON value")
            return self.add(Cfg.connect_tls_ex(p))
        if m == Cfg.Const.MTLS:
            if not isinstance(p, dict):
                raise ValueError("mtls: invalid JSON value")
            a = p.get("ca")
            y = p.get("pem")
            k = p.get("key")
            n = p.get("version", 0)
            if not Utils.nes(y) or not Utils.nes(k):
                raise ValueError("mtls: invalid JSON PEM/KEY values")
            if n is not None and not isinstance(n, int):
                raise ValueError("mtls: invalid JSON version value")
            self.add(
                Cfg.connect_mtls(
                    n,
                    b64decode(a, validate=True),
                    b64decode(y, validate=True),
                    b64decode(k, validate=True),
                )
            )
            del a, y, k, n
            return
        if m == Cfg.Const.TLS_CA:
            if not isinstance(p, dict):
                raise ValueError("tls-ca: invalid JSON value")
            a = p.get("ca")
            n = p.get("version", 0)
            if n is not None and not isinstance(n, int):
                raise ValueError("tls-ca: invalid JSON version value")
            self.add(Cfg.connect_tls_ca(n, b64decode(a, validate=True)))
            del a, n
            return
        if m == Cfg.Const.TLS_CERT:
            if not isinstance(p, dict):
                raise ValueError("tls-cert: invalid JSON value")
            y = p.get("pem")
            k = p.get("key")
            n = p.get("version", 0)
            if not Utils.nes(y) or not Utils.nes(k):
                raise ValueError("tls-cert: invalid JSON PEM/KEY values")
            if n is not None and not isinstance(n, int):
                raise ValueError("tls-cert: invalid JSON version value")
            self.add(
                Cfg.connect_tls_certs(
                    n, b64decode(y, validate=True), b64decode(k, validate=True)
                )
            )
            del y, k, n
            return
        if m == Cfg.Const.XOR:
            if not Utils.nes(p):
                raise ValueError("xor: invalid JSON value")
            return self.add(Cfg.wrap_xor(b64decode(p, validate=True)))
        if m == Cfg.Const.AES:
            if not isinstance(p, dict):
                raise ValueError("aes: invalid JSON value")
            y = p.get("iv")
            k = p.get("key")
            if not Utils.nes(y) or not Utils.nes(k):
                raise ValueError("aes: invalid JSON KEY/IV values")
            self.add(
                Cfg.wrap_aes(b64decode(k, validate=True), b64decode(y, validate=True))
            )
            del y, k
            return
        if m == Cfg.Const.CBK:
            if not isinstance(p, dict):
                raise ValueError("aes: invalid JSON value")
            A = p.get("A")
            B = p.get("B")
            C = p.get("C")
            D = p.get("D")
            z = p.get("size", 128)
            if not isinstance(A, int):
                raise ValueError("cbk: invalid JSON A value")
            if not isinstance(B, int):
                raise ValueError("cbk: invalid JSON B value")
            if not isinstance(C, int):
                raise ValueError("cbk: invalid JSON C value")
            if not isinstance(D, int):
                raise ValueError("cbk: invalid JSON D value")
            self.add(Cfg.wrap_cbk(a=A, b=B, c=C, d=D, size=z))
            del z, A, B, C, D
            return
        if m == Cfg.Const.DNS:
            if not isinstance(p, list):  # or len(p) == 0: Omit to allow empty DNS
                raise ValueError("dns: invalid JSON value")
            return self.add(Cfg.transform_dns(p))
        if m == Cfg.Const.B64S:
            if not isinstance(p, int):
                raise ValueError("b64s: invalid JSON value")
            if p < 0:
                raise ValueError("b64s: invalid JSON value")
            self.add(Cfg.transform_b64_shift(p))
            del p
            return
        raise ValueError(f'unhandled value type: {x["type"].lower()}')


class Setting(bytearray):
    def __str__(self):
        if len(self) == 0 or self[0] == 0:
            return "<invalid>"
        if self[0] not in Cfg.Const.NAMES:
            return "<invalid>"
        return Cfg.Const.NAMES[self[0]]

    @staticmethod
    def is_single(v):
        if v == 0:
            return False
        if v == Cfg.Const.B64T or v == Cfg.Const.SEPARATOR:
            return True
        if v >= Cfg.Const.LAST_VALID and v <= Cfg.Const.SEMI_RANDOM:
            return True
        if v >= Cfg.Const.TCP and v <= Cfg.Const.TLS_INSECURE:
            return True
        if v >= Cfg.Const.HEX and v <= Cfg.Const.B64:
            return True
        return False

    def single(self):
        if len(self) == 0 or self[0] == 0:
            return False
        return Setting.is_single(self[0])

    def _is_valid(self):
        return len(self) > 0 and self[0] > 0

    def _is_connector(self):
        if len(self) == 0 or self[0] == 0:
            return False
        if self[0] >= Cfg.Const.IP and self[0] <= Cfg.Const.TLS_CERT:
            return True
        return self[0] >= Cfg.Const.TCP and self[0] <= Cfg.Const.TLS_INSECURE

    def _is_transform(self):
        if len(self) == 0 or self[0] == 0:
            return False
        return self[0] >= Cfg.Const.B64T and self[0] <= Cfg.Const.B64S


class _Builder(ArgumentParser):
    def __init__(self):
        ArgumentParser.__init__(self, description="XMT c2.Config Tool")
        self.add_argument(
            nargs="?",
            dest="action",
            default="",
            metavar="action",
            choices=["", "add", "append"],
        )
        self.add_argument("-j", "--json", dest="json", action="store_true")
        self.add_argument("-p", "--print", dest="print", action="store_true")
        self.add_argument("-I", "--stdin", dest="stdin", action="store_true")

        self.add_argument("-f", "--in", type=str, dest="input")
        self.add_argument("-o", "--out", type=str, dest="output")

        self.add_argument("-T", "--host", type=str, dest="host")
        self.add_argument("-S", "--sleep", type=str, dest="sleep")
        self.add_argument("-J", "--jitter", type=int, dest="jitter")
        self.add_argument("-W", "--weight", type=int, dest="weight")
        self.add_argument(
            "-X",
            "--selector",
            type=str,
            dest="selector",
            default=None,
            choices=[
                "last",
                "random",
                "round-robin",
                "semi-random",
                "semi-round-robin",
            ],
        )

        self.add_argument("--killdate", type=str, dest="killdate")
        self.add_argument("--wh-days", type=str, dest="workhours_days")
        self.add_argument("--wh-start", type=str, dest="workhours_start")
        self.add_argument("--wh-end", type=str, dest="workhours_end")

        c = self.add_mutually_exclusive_group(required=False)
        c.add_argument("--tcp", dest="tcp", action="store_true")
        c.add_argument("--tls", dest="tls", action="store_true")
        c.add_argument("--udp", dest="udp", action="store_true")
        c.add_argument("--ip", type=int, dest="ip", default=None)
        c.add_argument("--icmp", dest="icmp", action="store_true")
        c.add_argument("--pipe", dest="pipe", action="store_true")
        c.add_argument("-K", "--tls-insecure", dest="tls_insecure", action="store_true")
        del c

        self.add_argument("--wc2-url", type=str, dest="wc2_url", default=None)
        self.add_argument("--wc2-host", type=str, dest="wc2_host", default=None)
        self.add_argument("--wc2-user", type=str, dest="wc2_agent", default=None)
        self.add_argument("--wc2-server", dest="wc2_server", action="store_true")
        self.add_argument(
            "-H",
            "--wc2_header",
            nargs="+",
            type=str,
            dest="wc2_headers",
            action="append",
            default=None,
        )
        self.add_argument("--mtls", dest="mtls", action="store_true")
        self.add_argument("--tls-ca", type=str, dest="tls_ca", default=None)
        self.add_argument("--tls-ver", type=int, dest="tls_ver", default=None)
        self.add_argument("--tls-pem", type=str, dest="tls_pem", default=None)
        self.add_argument("--tls-key", type=str, dest="tls_key", default=None)

        self.add_argument("--hex", dest="hex", action="store_true")
        self.add_argument("--b64", dest="b64", action="store_true")
        self.add_argument("--zlib", dest="zlib", action="store_true")
        self.add_argument("--gzip", dest="gzip", action="store_true")
        self.add_argument("--xor", nargs="?", type=str, dest="xor", default=None)
        self.add_argument("--cbk", nargs="?", type=str, dest="cbk", default=None)
        self.add_argument("--aes", nargs="?", type=str, dest="aes", default=None)

        self.add_argument("--aes-iv", nargs="?", type=str, dest="aes_iv", default=None)
        self.add_argument(
            "--b64t", nargs="?", type=int, dest="b64t", action="append", default=None
        )
        self.add_argument(
            "-D",
            "--dns",
            nargs="*",
            type=str,
            dest="dns",
            action="append",
            default=None,
        )

    def run(self):
        a = self.parse_args()
        e = Utils.nes(a.action) and a.action[0] == "a"
        if e and Utils.nes(a.input) and not Utils.nes(a.output) and a.input != "-":
            a.output = a.input
        if e and not Utils.nes(a.input) and Utils.nes(a.output):
            a.input = a.output
        if a.input:
            c = Utils.read_file_input(a.input)
        else:
            c = Config()
        if a.stdin and a.input != "-":
            if stdin.isatty():
                raise ValueError("stdin: no input found")
            if hasattr(stdin, "buffer"):
                b = stdin.buffer.read().decode("UTF-8")
            else:
                b = stdin.read()
            stdin.close()
            for v in b.split("\n"):
                x = split(v)
                _Builder.build(c, super(__class__, self).parse_args(x), True, x)
                del x
        elif not Utils.nes(a.output) or (not a.print and not a.json):
            _Builder.build(c, a, e, argv)
        if len(c) == 0:
            return
        Utils.write_file_output(c, a.output, a.print, a.json)
        del e, a, c

    @staticmethod
    def _organize(args):
        a = False
        w = list()
        d = dict()
        for i in range(0, len(args)):
            if len(args[i]) < 3:
                continue
            if args[i][0] != "-":
                continue
            if args[i].lower() == "--aes":
                if a:
                    continue
                a = True
            if args[i].lower() == "--aes-iv":
                if a:
                    continue
                v = "aes"
                a = True
            else:
                v = args[i].lower()[2:]
            if v not in Cfg.Const.WRAPPERS:
                continue
            if v in d:
                raise ValueError('duplicate argument "--{v}" found')
            w.append(v)
            d[v] = len(w) - 1
        e = [None] * len(w)
        del w, a
        return d, e

    def parse_args(self):
        if len(argv) <= 1:
            return self.print_help()
        return super(__class__, self).parse_args()

    def print_help(self, file=None):
        print(HELP_TEXT.format(binary=argv[0]), file=file)
        exit(2)

    @staticmethod
    def build(config, args, add, arv):
        if add and len(config) > 0:
            config.add(Cfg.separator())
        p, w = _Builder._organize(arv)
        if args.host:
            config.add(Cfg.host(args.host))
        if args.sleep:
            config.add(Cfg.sleep(args.sleep))
        if isinstance(args.jitter, int):
            config.add(Cfg.jitter(args.jitter))
        if isinstance(args.weight, int):
            config.add(Cfg.weight(args.weight))
        if args.killdate:
            config.add(Cfg.killdate(args.killdate))
        if args.workhours_days or args.workhours_start or args.workhours_end:
            config.add(
                Cfg.workhours(
                    args.workhours_days, args.workhours_start, args.workhours_end
                )
            )
        if args.selector:
            if args.selector == "last":
                config.add(Cfg.selector_last_valid())
            elif args.selector == "random":
                config.add(Cfg.selector_random())
            elif args.selector == "round-robin":
                config.add(Cfg.selector_round_robin())
            elif args.selector == "semi-random":
                config.add(Cfg.selector_semi_random())
            elif args.selector == "semi-round-robin":
                config.add(Cfg.selector_semi_round_robin())
            else:
                raise ValueError("selector: invalid value")
        if args.tcp:
            config.add(Cfg.connect_tcp())
        if args.tls:
            config.add(Cfg.connect_tls())
        if args.udp:
            config.add(Cfg.connect_udp())
        if args.icmp:
            config.add(Cfg.connect_icmp())
        if args.pipe:
            config.add(Cfg.connect_pipe())
        if args.tls_insecure:
            config.add(Cfg.connect_tls_insecure())
        if isinstance(args.ip, int):
            config.add(Cfg.connect_ip(args.ip))
        if args.wc2_url or args.wc2_host or args.wc2_agent or args.wc2_headers:
            config.add(
                Cfg.connect_wc2(
                    args.wc2_url,
                    args.wc2_host,
                    args.wc2_agent,
                    Utils.parse_wc2_headers(args.wc2_headers),
                )
            )
        if args.tls_ca or args.tls_pem or args.tls_key:
            config.add(
                Utils.parse_tls(
                    args.tls_ca, args.tls_pem, args.tls_key, args.mtls, args.tls_ver
                )
            )
        elif args.mtls:
            raise ValueError("mtls: missing CA, PEM and KEY values")
        elif isinstance(args.tls_ver, int):
            config.add(Cfg.connect_tls_ex(args.tls_ver))
        if args.b64t and len(args.b64t) == 1:
            if args.b64t[0] is None:
                config.add(Cfg.transform_b64())
            else:
                config.add(Cfg.transform_b64_shift(int(args.b64t[0])))
        if args.dns and len(args.dns) > 0:
            if len(args.dns) == 1 and len(args.dns[0]) == 0:
                config.add(Cfg.transform_dns([]))
            else:
                config.add(Cfg.transform_dns(Utils.split_dns_names(args.dns)))
        if args.hex:
            w[p["hex"]] = Cfg.wrap_hex()
        if args.zlib:
            w[p["zlib"]] = Cfg.wrap_zlib()
        if args.gzip:
            w[p["gzip"]] = Cfg.wrap_gzip()
        if args.b64:
            w[p["b64"]] = Cfg.wrap_b64()
        if args.xor:
            w[p["xor"]] = Cfg.wrap_xor(args.xor[0])
        elif "xor" in p:
            w[p["xor"]] = Cfg.wrap_xor()
        if args.cbk:
            w[p["cbk"]] = Cfg.wrap_cbk(key=args.cbk[0])
        elif "cbk" in p:
            w[p["cbk"]] = Cfg.wrap_cbk()
        if args.aes or args.aes_iv:
            w[p["aes"]] = Cfg.wrap_aes(args.aes, args.aes_iv)
        elif "aes" in p:
            w[p["aes"]] = Cfg.wrap_aes()
        for i in w:
            config.add(i)
        del w, p


if __name__ == "__main__":
    try:
        _Builder().run()
    except Exception as err:
        print(f"Error: {err}\n{format_exc(3)}", file=stderr)
        exit(1)
