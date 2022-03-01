#!/usr/bin/python

from os.path import exists
from json import loads, dumps
from sys import argv, exit, stderr


class StringTable(dict):
    def __init__(self, file=None):
        dict.__init__(self)
        if isinstance(file, str) and exists(file):
            with open(file, "r") as f:
                m = loads(f.read())
                if not isinstance(m, dict):
                    raise ValueError(f'File "{file}" is not a dict!')
                if len(m) == 0:
                    return
                e = [None] * len(m)
                for k, v in m.items():
                    if not isinstance(v, dict) or "value" not in v:
                        continue
                    e[int(k)] = v
                for x in range(0, len(e)):
                    if not isinstance(e[x].get("tags"), list) or len(e[x]["tags"]) == 0:
                        print(f'Read {x}: {e[x]["value"]}')
                    else:
                        print(
                            f'Read {x}: {e[x]["value"]} (tags: {", ".join(e[x]["tags"])})'
                        )
                    self[x] = e[x]
            print(f"{len(self)} entries loaded.\n")

    def read(self):
        try:
            while True:
                i = input("String? ")
                if not isinstance(i, str) or len(i) == 0:
                    continue
                v = i.strip()
                e = self._exists(v)
                del i
                if e is not None:
                    print(f'EXIST => {e} == "{v}"', end="")
                    if not isinstance(self[e].get("tags"), list):
                        print()
                        continue
                    if len(self[e]["tags"]) == 0:
                        print()
                        continue
                    print(f' tags: {", ".join(self[e]["tags"])}')
                    continue
                k = self._next_key()
                self[k] = {"value": v}
                print(f'ADD   => {k} == "{v}"')
        except (Exception, KeyboardInterrupt):
            print("[+] Exiting.", file=stderr)

    def _next_key(self):
        for x in range(0, len(self)):
            if x not in self:
                return str(x)
        return len(self)

    def save(self, path):
        with open(path, "w") as f:
            f.write(dumps(self, sort_keys=False, indent=4))

    def _exists(self, value):
        for k, v in self.items():
            if v["value"] == value:
                return k
        return None


if __name__ == "__main__":
    if len(argv) != 2:
        print(f"{argv[0]} <db>", file=stderr)
        exit(2)

    t = StringTable(argv[1])
    t.read()
    t.save(argv[1])
