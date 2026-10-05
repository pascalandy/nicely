# M1 Day one

Status: planned
Release: v0.1.0

## Goal

The first useful release. `doctor` and `auth` check and configure a new machine, agents discover skills with `skill`, `transcript` moves into Nicely, and agent profiles live in the config.

## Depends on

M0.

## Scope

**Spec.** The M1 sections of [cli-spec.md](../north-star/cli-spec.md) are drafts. Settle the [to-define list](#to-define-before-code), update the spec, and only then write Go.

**`auth`.** Keys as the Keys section of cli-spec.md defines them. `deepgram` is the one core service.

**`doctor`.** Components `core`, `auth`, `skill`, and `transcript`. The `skill` component checks structure only, as D021 says.

**`skill`.** `list` and `view` over the folders in `[skill] paths`, plus the `ncly skill <name>` shortcut for humans.

**Agent profiles.** `[agent.profiles]` in the config, as cli-spec.md defines them. `transcript` reads them for its summary.

**`internal/run`.** Starts here with the Python program. M2 reuses it for harnesses and M3 for extensions.

**Test keychain.** A keychain stored in a file under `$WORK`, registered only in the test binary. go-keyring's mock lives in one process, and a scenario runs `ncly` several times.

**`transcript`.** Move it in these steps:

1. Build the Go side first, tested with a stub `uv` in `PATH`.
2. As the last step before the release, copy the Python transcript CLI from `authoring/andy/transcript/` in `pascalandy/skills` into `python/transcript/`, with its prompts, tests, and README. From then on, fixes land in Nicely first.
3. Embed it with `go:embed`. On first run, extract it to `~/.cache/nicely/python/<version>/` through a temporary folder and a rename, then run it with `uv run`.
4. Python returns JSON only. Go maps each Python code to the error registry and renders the text from the catalog.
5. Go passes the `deepgram` key as an environment variable of the child process. Python stops reading the keychain.
6. Go resolves `--profile` from the config and passes the harness, model, and effort through the Python flags `--provider`, `--model`, and `--effort`.
7. The checks of the old `doctor --source` become the `transcript` component of `ncly doctor`.
8. A run shows one spinner. Step-by-step progress stays in [M99](M99-parking-lot.md).

**Python checks.** `just check` runs ruff, pyright, and pytest on `python/` through `uv`.

**Agent instructions.** A short text that a user pastes into an `AGENTS.md`: `ncly` exists, set `NCLY_NO_INPUT=1`, start with `ncly skill list --json`, and relay any hint of an exit 78 to the human. README.md shows it, and Pascal adds it to his global agent instructions.

**Completion values.** The values that the completion section of cli-spec.md lists.

## To define before code

- The generic default output folder of `transcript`. Pascal's own values move to his config: his export folders, his Zoom folders, and the `synthese-rencontre` prompt of `andy-mode`, read from a folder of private prompts.
- Linux parity: browser cookies from Chromium and Chrome on Linux, and opening the result folder through `internal/platform`.
- The protocol between Go and the Python program: arguments, environment, JSON shape, signals, and timeout.
- The registry code and the catalog entry of each Python error code, all in place before M6.
- The mapping of the other transcript options onto Nicely's flags, such as `--output-dir` to `--output` and `--debug` to `NCLY_DEBUG`, and the fate of `--no-progress`, `--open`, and `--preview`.
- The harnesses that `transcript` accepts in a profile until M2, `claude` and `pi`, and the profile it uses when the config has none.
- How a `yt-dlp` fix reaches users, since the Python program pins `yt-dlp` and YouTube breaks it every few weeks.
- Pascal's key migration, once per machine: `chezmoi secret keyring get --service=deepgram --user=api_key | ncly auth login deepgram --stdin`.

## Follow-up in `pascalandy/skills`

Once `ncly transcript` works, the `transcript` skill points to it, the `transcript` and `transcript-cli` recipes retire, `verify-transcript` checks `ncly`, and the Python copy there is deleted. Pascal decides when.

## Out of scope

- The `ncly agent` command: M2.
- Taps, `skill link`, and `nicely.toml`: M3.
- Porting transcript to Go: [M99](M99-parking-lot.md).

## Acceptance

Scenarios use the test keychain and a stub `uv`, and they declare fixture skills in the `.txtar` archive.

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
stdout '"ok":true'
! stderr .

# An agent cannot log in without a terminal
exits 78 ncly auth login deepgram --json
stderr '"code":"TERMINAL_REQUIRED"'

# Skills are listed as JSON and viewed with their folder first
exec ncly skill list --json
stdout '"name":"alpha"'
exec ncly skill view alpha
stdout '^<!-- skill-dir: .*/alpha -->'
exits 2 ncly skill view nope --json
stderr '"code":"NOT_FOUND"'

# Doctor reports on stdout and exits 78 when a key is missing
exits 78 ncly doctor auth --json
stdout '"ok":false'
stdout '"status":"fail"'
```

Before tagging v0.1.0, run the paid end-to-end check from the transcript README against its listed test video.

## Done when

- [ ] Every item of the to-define list is settled in cli-spec.md or in this file
- [ ] Every acceptance scenario passes in `just check`
- [ ] The paid end-to-end check passes on macOS and on Omarchy
- [ ] An agent in a fresh session, given only the agent instructions, finds and reads a skill using only `ncly skill`
- [ ] v0.1.0 is tagged and released
