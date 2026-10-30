"""inspect.py FILE prints the size and the line count of a file as JSON.

The output is one line, such as {"size_bytes": 12, "lines": 2, "text": true}.
The worker reads the file as bytes, so it accepts any file, and a last line
without a newline counts too. text is false when the file holds a NUL byte, as
an image does. Go decides what to refuse. The worker uses the standard library
only, and Go renders and translates every word that a human sees.
"""

import json
import sys


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: inspect.py FILE", file=sys.stderr)
        return 2
    size, lines, last, text = 0, 0, b"\n", True
    with open(sys.argv[1], "rb") as file:
        for chunk in iter(lambda: file.read(65536), b""):
            size += len(chunk)
            lines += chunk.count(b"\n")
            last = chunk[-1:]
            text = text and b"\0" not in chunk
    if last != b"\n":
        lines += 1
    print(json.dumps({"size_bytes": size, "lines": lines, "text": text}))
    return 0


if __name__ == "__main__":
    sys.exit(main())
