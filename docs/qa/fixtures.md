# QA fixtures

This file answers one question: which real data do UATs use, and how is it protected?

Fixtures stay outside the repository, since they are large and can be private. This registry gives their name, folder, and properties. The fleet skill gives the real path on each host.

## Rules

- An original is read-only, with `chmod 444`, and locked, with `chflags uchg` on macOS. Only Pascal unlocks it.
- A `SHA256SUMS` file sits beside it, locked too.
- A run works on a copy. It checks the checksum of the original before and after.
- Pascal supplies each new fixture and its purpose. It enters this registry before its first run.

## Registry

| Name | Folder | For | Properties to keep |
|---|---|---|---|
| `IMG_1642.MOV` | `nicely/video/` | video-convert | HEVC 10-bit 1920x1080, 30 fps, HDR HLG, rotation -180, 2 min 49 s, AAC audio |

## Made fixtures

Each run makes these in its work folder, never in the fixture folder:

- A fake file: text with the expected extension
- An empty file
- A truncated copy of a real fixture
- A path that does not exist
