# Shared domains

Two installed extensions contribute independent video jobs to one domain.
The conversion fixture runs at its root, while the subtitle fixture supplies completion values.

## Sub-features

- `shared-run` passes the file argument to a named root through the host
- `shared-direct` runs the same root through its executable
- `shared-help` lists both roots without running either extension
- `shared-describe` exposes the core group and each extension's skill folder
- `shared-completion` obtains file suggestions from the sibling's list command
- `shared-skills` finds both slug-named skills through describe

## How to get to it (user POV)

- `ncly video convert clip.mov --json`
- `ncly-video-convert clip.mov --json`
- `ncly video --help`
- `ncly describe video --json`
- `ncly __complete video convert cl`
- `ncly skill view video-convert --json`
- `ncly skill status --json`

## Driving it with verify.sh

Preconditions:

- A new run from `just verify`, whose Doctor step passed before installing these fixtures
- Build both fixtures and export their embedded sidecars from the checkout

```sh
extensions="$run/scratch/home/.local/share/nicely/extensions"
go build -o "$extensions/ncly-video-convert" ./testdata/extensions/video-convert/cmd/ncly-video-convert
go build -o "$extensions/ncly-video-subtitle" ./testdata/extensions/video-subtitle/cmd/ncly-video-subtitle
go run ./testdata/extensions/shared/export "$extensions"
```

Run each command below through `just verify step "$run" shared-domains`:

1. Add `video convert clip.mov --json`. Expect exit 0 with `file: "clip.mov"` and `path: "video convert"`.
2. Add `video --help`. Expect exit 0, with `convert` before `subtitle` in the command list.
3. Add `describe video --json`. Expect a `video` entry with `source: "core"` and no `skill` key.
4. Check both extension roots in that answer. Each has `source: "extension"` and its slug-named skill folder.
5. Add `describe video convert --json`. Expect the required `file` argument and `values_from.command: "video subtitle list"`.
6. Add `__complete video convert cl`. Expect `clip.mov` and no `other.mov` suggestion.
7. Add `skill view video-convert --json`. Expect the installed skill's content and its slug-named folder.
8. Add `skill view video-subtitle --json`. Expect the sibling's installed skill.
9. Add `skill status --json`. Expect exit 0 without `SKILL_NAME_MISMATCH`.
10. Add `video convert fail.mov --json`. Expect exit 1 and `VIDEO_CONVERT_FAILED` on stderr.

Run `"$extensions/ncly-video-convert" clip.mov --json` directly.
Expect the same `file` and `path` as step 1, and record this separate proof beside the transcript.
Clean up with `just verify cleanup "$run"`, then quote `shared-domains`.

## Gotchas

- These fixtures convert no real media and need no external program, key, config, or status command
- Doctor components, config ownership, and grants for named extensions belong to M23-T2
- `__complete` is Cobra's completion protocol and also prints a directive line
- Direct execution is the one step outside the host wrapper, so name its executable path in the evidence
