#!/usr/bin/python

from sys import argv, exit, stderr


def xor(key: bytes, data: bytes) -> bytearray:
    if len(data) == 0 or len(key) == 0:
        return bytearray()
    r = bytearray(len(data))
    for i in range(0, len(r)):
        r[i] = data[i] ^ key[i % len(key)]
    return r


if __name__ == "__main__":
    if len(argv) != 3:
        print("XMT String Encoder\nEncodes a string payload\n", file=stderr)
        print("%s <key> <string>" % argv[0], file=stderr)
        exit(2)

    try:
        key = argv[1].encode("UTF-8")
    except UnicodeDecodeError as err:
        print('Error encoding key "%s": %s!' % (argv[1], str(err)), file=stderr)
        exit(1)

    try:
        data = argv[2].encode("UTF-8")
    except UnicodeDecodeError as err:
        print('Error encoding string "%s": %s!' % (argv[2], str(err)), file=stderr)
        exit(1)

    out = xor(key, data)
    del key
    del data

    print("Paste this data into your Golang code, then call 'util.Decode(key, value)'", file=stderr)
    print("var value = []byte{", end="")
    for i in range(0, len(out)):
        if i == 0 or i % 20 == 0:
            print("\n\t", end="")
        else:
            print(" ", end="")
        print("0x%s," % ("%X" % out[i]).zfill(2), end="")
    print("\n}")
    del out
