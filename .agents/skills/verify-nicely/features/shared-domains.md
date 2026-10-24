# Shared domains

Two installed extensions contribute independent video jobs to one domain. The conversion root takes a file, while subtitle supplies completion values.
Doctor checks each slug independently, and grants bind keys to one extension's slug and source.

## Sub-features

- `shared-run` passes the file argument to a named root through the host
- `shared-direct` runs the same root through its executable
- `shared-help` lists both roots without running either extension
- `shared-describe` exposes the core group and each extension's skill folder
- `shared-completion` obtains file suggestions from the sibling's list command
- `shared-skills` finds both slug-named skills through describe
- `shared-config` recognizes only the selected manifest's config keys
- `shared-doctor` reports discovery, program checks, and status under the extension slug
- `shared-grants` grants and revokes a named extension's key through a real terminal

## How to get to it (user POV)

- `ncly video convert clip.mov --json` and `ncly-video-convert clip.mov --json`
- `ncly video --help`, `ncly describe video --json`, and `ncly __complete video convert ''`
- `ncly skill view video-convert --json` and `ncly skill status --json`
- `ncly doctor core --json`, `ncly doctor extensions --json`, and `ncly doctor video-subtitle --json`
- `ncly auth grant deepgram video-subtitle` and `ncly auth revoke deepgram video-subtitle --force --json`
- `ncly auth status --json` and `ncly video subtitle fetch --json`

## Driving it with verify.sh

Preconditions:

- A new run from `just verify`, whose Doctor step passed before installing these fixtures
- Build both fixtures and export their embedded sidecars from the checkout
- For grants, follow SKILL.md's terminal feature and use a fake environment key, never a stored service key

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
6. Add `__complete video convert ''`. Expect both `clip.mov` and `other.mov` from the sibling's list.
7. Add `skill view video-convert --json`. Expect the installed skill's content and its slug-named folder.
8. Add `skill view video-subtitle --json`. Expect the sibling's installed skill.
9. Add `skill status --json`. Expect exit 0 without `SKILL_NAME_MISMATCH`.
10. Add `video convert fail.mov --json`. Expect exit 1 and `VIDEO_CONVERT_FAILED` on stderr.

Run `"$extensions/ncly-video-convert" clip.mov --json` directly. Expect the same `file` and `path` as step 1, and record this separate proof beside the transcript.

Prepare config in the scratch home:

```sh
mkdir -p "$run/scratch/home/.config/nicely"
cat > "$run/scratch/home/.config/nicely/config.toml" <<'EOF'
video-subtitle = { language = "en", unknown = true }
video-convert = { language = "fr" }
away-extra = { setting = true }
EOF
```

Continue through `just verify step "$run" shared-domains`:

11. Add `doctor core --json`. Expect warnings for `video-subtitle.unknown` and `video-convert.language`, but none for `video-subtitle.language` or `away-extra.setting`.
12. Add `doctor extensions --json`. Expect separate passing `extensions.video-convert` and `extensions.video-subtitle` entries, each with its slug in `domain`.
13. Add `doctor video-convert --json`. Expect exit 0 and an empty `checks` array.
14. Add `doctor video-subtitle --json`. Expect exit 78 for `video-subtitle.subtitle-tool`, unless that program exists.
15. Check the same answer for passing `video-subtitle.ready`. It proves doctor reached `video subtitle status` through the full root.
16. Add `auth grant deepgram video-subtitle --dry-run --json`. Expect `domain: "video-subtitle"` and a pending confirmation.
17. Add `auth grant deepgram video-convert --dry-run --json`. Expect `NOT_FOUND`, because this fixture declares no key.

Use the fake key only on each command that reads it:

```sh
DEEPGRAM_API_KEY=shared-domain-verification just verify step "$run" shared-domains video subtitle fetch --json
just verify tmux "$run" shared-domains --answer '[y/N]' y auth grant deepgram video-subtitle
DEEPGRAM_API_KEY=shared-domain-verification just verify step "$run" shared-domains video subtitle fetch --json
DEEPGRAM_API_KEY=shared-domain-verification just verify step "$run" shared-domains auth status --json
just verify step "$run" shared-domains auth revoke deepgram video-subtitle --force --json
DEEPGRAM_API_KEY=shared-domain-verification just verify step "$run" shared-domains video subtitle fetch --json
```

Expect exit 78 and `KEY_NOT_GRANTED` before granting, with `next` ending in `video-subtitle`.
After granting, fetch returns `received_key: true`, and auth status lists that slug as granted.
Revocation returns `domain: "video-subtitle"`, and the last fetch returns `KEY_NOT_GRANTED` again.
These steps write no keychain entry. Clean up with `just verify cleanup "$run"`, then quote `shared-domains`.

## Gotchas

- These fixtures convert no real media and send no network request
- The conversion root keeps no descendants, prerequisites, config, or status command
- Subtitle status declares `subtitle-tool`, and fetch declares `deepgram`; the grant steps use a fake environment key
- `__complete` is Cobra's completion protocol and also prints a directive line
- Direct execution is the one step outside the host wrapper, so name its executable path in the evidence
