# M1 Day one

Status: planned
Release: v0.1.0

## Goal

The first useful release. `doctor` and `auth` check and configure a new machine, agents discover skills with `skill`, and `transcript` moves into Nicely.

## Depends on

M0.

## Scope

**Spec.** The M1 sections of [cli-spec.md](../north-star/cli-spec.md) are drafted. Confirm them against the transcript code you find, and update them before writing Go.

**`auth`.** Keys in the OS keychain through go-keyring, read after environment variables. Services `deepgram` and `openrouter`.

**`doctor`.** Components `core`, `auth`, `skill`, and `transcript`. Local checks by default, free endpoints only with `--live`. In interactive mode, doctor offers the `brew` or `pacman` command for a missing tool and runs it only after a yes.

**`skill`.** `list` and `show` over the folders in `[skill] paths`, plus the `ncly skill <name>` shortcut for humans.

**`transcript`.** Move it in these steps:

1. Move the Python transcript CLI from `pascalandy/skills` into `python/transcript/`, with its prompts, tests, and README.
2. Embed it with `go:embed`. On first run, extract it to `~/.cache/nicely/python/<version>/` and run it with `uv run`.
3. Python returns JSON only. Go maps the result to exit codes and renders human text from the error code. When a code has no catalog entry yet, Go shows the Python `message`.
4. Go passes keys to Python as environment variables of the child process.
5. The checks of the old `doctor --source` become the `transcript` component of `ncly doctor`.
6. Transcript keeps its own profiles until M2.

**Completion values.** Skill names, auth services, transcript prompts and profiles, and doctor components.

**Follow-up in `pascalandy/skills`.** Once `ncly transcript` works, the `transcript` skill points to it, the `transcript` and `transcript-cli` recipes retire, and `verify-transcript` checks `ncly`. Pascal decides when.

## Out of scope

- `agent` and shared profiles: M2.
- Taps, `skill link`, and `nicely.toml`: M3.
- Porting transcript to Go: [M99](M99-parking-lot.md).

## Acceptance

Scenarios replace the keychain with go-keyring's mock and declare fixture skills in the `.txtar` archive.

```
# A missing key stops the run and names the fix
env DEEPGRAM_API_KEY=
exits 78 ncly transcript run youtube --url https://www.youtube.com/watch?v=VIDEO_ID --json
! stdout .
stderr '"code":"AUTH_MISSING"'
stderr '"hint":"ncly auth login deepgram"'

# A dry run needs no key and writes nothing
env DEEPGRAM_API_KEY=
exec ncly transcript run youtube --url https://www.youtube.com/watch?v=VIDEO_ID --dry-run --json
! stderr .

# An agent cannot log in without a terminal
exits 2 ncly auth login deepgram --json
stderr '"code":"USAGE_INVALID"'

# Skills are listed as JSON and shown with their folder first
exec ncly skill list --json
stdout '"name":"alpha"'
exec ncly skill show alpha
stdout '^<!-- skill-dir: .*/alpha -->'
exits 2 ncly skill show nope --json
stderr '"code":"NOT_FOUND"'

# Doctor reports on stdout and exits 78 when a key is missing
exits 78 ncly doctor auth --json
stdout '"status":"fail"'
```

Before tagging v0.1.0, run the paid end-to-end check from the transcript README against its listed test video.

## Done when

- [ ] Every acceptance scenario passes in `just check`
- [ ] The paid end-to-end check passes on macOS and on Omarchy
- [ ] An agent in a fresh session finds and reads a skill using only `ncly skill`
- [ ] v0.1.0 is tagged and released

## Open questions

1. Does the transcript CLI call OpenRouter with its own key? If it does not, drop the `openrouter` service from M1.
2. Should transcript show each step, such as download, transcription, and summary, while it runs? Recommendation: one spinner per run in M1, and progress events later if the wait feels long.
