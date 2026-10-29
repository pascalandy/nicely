# QA fixtures

This file answers one question: which real data do UATs use, and how is it protected?

Fixtures stay outside the repository, since they are large and can be private. This registry gives their name, folder, and properties. Each folder sits under the fixture root of a host, and the UAT issue names that root for each host. A host without the fixture gets a copy, as [Copy a fixture to a host](#copy-a-fixture-to-a-host) says.

## Rules

- An original is read-only, with `chmod 444`, and locked, with `chflags uchg` on macOS. Only Pascal unlocks it.
- A `SHA256SUMS` file sits beside it, locked too.
- A run works on a copy. It checks the checksum of the original before and after.
- Pascal supplies each new fixture and its purpose. It enters this registry before its first run.
- A verified copy of a registered fixture is an original on its new host, under the same rules.

## Registry

| Name | Folder | For | Properties |
|---|---|---|---|
| `IMG_1642.MOV` | `nicely/video/` | video-convert | HEVC 10-bit 1920x1080, 30 fps, HDR HLG, rotation -180, 2 min 49 s, AAC audio |

## Copy a fixture to a host

Copy the whole folder, with its `SHA256SUMS`, from a host that holds it. Replace `host-a` and both fixture roots with those of the issue:

```sh
mkdir -p ~/fixtures/nicely
scp -r host-a:fixtures/nicely/video ~/fixtures/nicely/
cd ~/fixtures/nicely/video
sha256sum -c SHA256SUMS        # macOS: shasum -a 256 -c SHA256SUMS
chmod 444 ./*                  # macOS: then chflags uchg ./*
```

Each file must report `OK`. Then `just preflight ~/fixtures/nicely/video` checks the checksums and the locks again. When a checksum fails, move the copy to the trash and copy it again. Never fix a `SHA256SUMS`.

## Made fixtures

Each run makes these in its work folder, never in the fixture folder:

- A fake file: text with the expected extension
- An empty file
- A truncated copy: the first 5,000,000 bytes of a real fixture
- A path that does not exist
