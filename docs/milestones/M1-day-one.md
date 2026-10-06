# M1 Day one

Status: planned
Release: v0.1.0

## Goal

The first useful release. `doctor` and `auth` check and configure a new machine. Agents discover commands and skills, and transcript proves the shared operation contract with inspectable results and explicit resume.

## Depends on

M0.

## Scope

**Spec.** The M1 sections of [cli-spec.md](../north-star/cli-spec.md) are drafts. Settle the [to-define list](#to-define-before-code), update the spec, and only then write Go.

**`auth`.** Keys as the Keys section of cli-spec.md defines them. `deepgram` is the one core service.

**`doctor`.** Components `core`, `auth`, `skill`, and `transcript`. The `skill` component checks structure only, as D021 says.

**`skill`.** `list` and `view` over the folders in `[skill] paths`, plus the `ncly skill <name>` shortcut for humans.

**Agent profiles.** `[agent.profiles]` in the config, as cli-spec.md defines them. `transcript` reads them for its summary.

**Discovery and records.** Implement [`ncly describe`](../north-star/cli-spec.md#ncly-describe-m1) and [`ncly run`](../north-star/cli-spec.md#ncly-run-m1). Command descriptions come from the M0 declarations. `internal/operation` first implements the [Operations contract](../north-star/cli-spec.md#operations-m0-contract-m1-execution) for transcript. Keep summaries bounded and full results in artifacts. Unknown effects remain visible and cannot be replayed through `--force`.

**Internal agent service.** `internal/agent` resolves profiles and owns the adapters that transcript needs, initially Claude Code and Pi. Verify tools-off support before using outside material. A missing harness uses `PREREQ_MISSING`. The public `ncly agent` command and other adapters still wait for M2.

**`internal/run`.** Starts here for both Python and the summary harnesses, as the Programs section of cli-spec.md defines it: environment, grants, and cancellation. M2 adds adapters and the public agent command. M3 reuses the runner for extensions.

**Test keychain.** A keychain stored in a file under `$WORK`, registered only in the test binary. go-keyring's mock lives in one process, and a scenario runs `ncly` several times. `NCLY_TEST_KEYCHAIN=unavailable`, which only the test binary reads, makes it stop answering.

**Test commands.** `waitfile <path>` waits until a file exists. `gone <path>` checks that the process whose ID the file holds no longer runs.

**`transcript`.** Move it in these steps:

1. Inspect the actual Python transcript CLI and harness invocations before fixing their adapter contract. Use stubs to test failure paths, then integrate a real local, non-billing path early in M1
2. Copy the Python transcript CLI from `authoring/andy/transcript/` in `pascalandy/skills` into `python/transcript/`, with its prompts, tests, and README. From then on, fixes land in Nicely first. The handoff happens early enough to test real behavior before the release gate
3. Embed it with `go:embed`. On first run, extract it to `~/.cache/nicely/python/<version>/` through a temporary folder and a rename, then run it with `uv run`.
4. Python returns versioned JSON only and stops at the transcript. Its protocol exposes the evidence Go needs to record paid dispatch, completed artifacts, and unknown effects. Go maps each Python code to the error registry and renders the text from the catalog
5. Go passes the `deepgram` key as an environment variable of the child process. Python stops reading the keychain.
6. Go calls `internal/agent` for the summary using the resolved profile. Python no longer launches a harness. Transcript and summary remain separate recorded steps so the first can be reused when the second never started
7. The checks of the old `doctor --source` become the `transcript` component of `ncly doctor`.
8. A run shows one spinner. User-facing progress streaming stays in [M99](M99-parking-lot.md). The internal protocol still acknowledges effect boundaries needed for durable records before dispatch

**Python checks.** `just check` runs ruff, pyright, and pytest on `python/` through `uv`.

**Agent instructions.** A short text that a user pastes into an `AGENTS.md`: `ncly` exists, set `NCLY_NO_INPUT=1`, discover commands with `ncly describe`, and find skills with `ncly skill list --json`. Relay exit 78 hints to the human. Inspect a recorded failure through `ncly run view` before requesting resume. README.md shows it, and Pascal adds it to his global agent instructions.

**Completion values.** The values that the completion section of cli-spec.md lists.

## To define before code

- The generic default output folder of `transcript`. Pascal's own values move to his config: his export folders, his Zoom folders, and the `synthese-rencontre` prompt of `andy-mode`, read from a folder of private prompts.
- Linux parity: browser cookies from Chromium and Chrome on Linux, and opening the result folder through `internal/platform`.
- The protocol between Go and Python: version, arguments, environment, JSON shapes, durable acknowledgement before a paid effect, signals, timeout, and behavior on EOF or a lost acknowledgement. No paid call may outrun its persisted intent
- The registry code and the catalog entry of each Python error code, all in place before M6.
- The mapping of the other transcript options onto Nicely's flags, such as `--output-dir` to `--output` and `--debug` to `NCLY_DEBUG`, and the fate of `--no-progress`, `--open`, and `--preview`.
- The profile transcript uses when the config has none, and the verified modes and supported versions of the initial `claude` and `pi` adapters
- How a `yt-dlp` fix reaches users, since the Python program pins `yt-dlp` and YouTube breaks it every few weeks.
- The time limit of a keychain call. An unlock prompt on a desktop takes longer than a script can wait.
- How `internal/run` reaches a descendant that left the process group of its parent and outlived it, since the transcript CLI starts its children in their own process groups.
- The keys each program declares: Python gets `deepgram`; the separate harness adapter declares its own key variables
- The nested JSON fields of command descriptions and run summaries/details, the stored record encoding, and finding the latest relevant run after its stdout was lost
- Transcript artifact fingerprints, the evidence needed to reuse each step, and the cases that refuse resume. Resolve lost paid responses conservatively; absence of an artifact does not prove absence of billing
- Pascal's key migration, once per machine: `chezmoi secret keyring get --service=deepgram --user=api_key | ncly auth login deepgram --stdin`.

## Follow-up in `pascalandy/skills`

Once `ncly transcript` works, the `transcript` skill points to it, the `transcript` and `transcript-cli` recipes retire, `verify-transcript` checks `ncly`, and the Python copy there is deleted. Pascal decides when.

## Out of scope

- The `ncly agent` command: M2.
- Taps, `skill link`, and `nicely.toml`: M3.
- Porting transcript to Go: [M99](M99-parking-lot.md).

## Acceptance

Scenarios use the test keychain, a stub `uv`, and stubs for the summary harnesses. They declare fixture skills and profiles in the `.txtar` archive. No scenario starts a real paid service.

```
# A missing key stops the run and names the fix
env DEEPGRAM_API_KEY=
exits 78 ncly transcript run youtube --url https://www.youtube.com/watch?v=VIDEO_ID --json
! stdout .
stderr '"code":"AUTH_MISSING"'
stderr '"hint":"ncly auth login deepgram"'

# A dry run needs no key and makes no user or state changes
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

# Doctor warns without failing when a folder name differs from its skill name
# The archive holds skills/beta/SKILL.md with name: gamma
exec ncly doctor skill --json
stdout '"ok":true'
stdout '"code":"SKILL_NAME_MISMATCH"'

# Doctor reports on stdout and exits 78 when a key is missing
exits 78 ncly doctor auth --json
stdout '"ok":false'
stdout '"status":"fail"'

# A key in the environment works when the keychain is unavailable
env NCLY_TEST_KEYCHAIN=unavailable
env DEEPGRAM_API_KEY=test-key
exec ncly doctor auth --json
stdout '"ok":true'
stdout '"status":"warn"'
exec ncly transcript run youtube --url https://www.youtube.com/watch?v=VIDEO_ID --json
stdout '"ok":true'

# Without the variable, the run names the variable to export
env DEEPGRAM_API_KEY=
exits 78 ncly transcript run youtube --url https://www.youtube.com/watch?v=VIDEO_ID --json
stderr '"code":"KEYRING_UNAVAILABLE"'
stderr 'DEEPGRAM_API_KEY'

# Ctrl-C stops the whole tree and keeps the result folder
env NCLY_TEST_KEYCHAIN=
env DEEPGRAM_API_KEY=test-key
exec ncly transcript run youtube --url https://www.youtube.com/watch?v=VIDEO_ID --json &
waitfile $WORK/grandchild.pid
kill -INT
! wait
stderr '"code":"INTERRUPTED"'
stderr '"output_dir"'
gone $WORK/grandchild.pid
```

The stub `uv` of the last scenario starts a grandchild in its own process group, as the transcript CLI does, and writes its process ID to `grandchild.pid`.

Extend these scenarios with the first real-domain proofs from [M0 Contract coverage](M0-foundation.md#contract-coverage):

- Compare user files, config, keychain, and state before and after a dry run. Permit only the documented cache writes and assert no paid call or harness start
- Stop before paid dispatch, after dispatch but before the response, and after saving the transcript but before starting the summary. Inspect the saved run in a new process
- Resume the saved transcript with a summary that never started. A counted service stub proves the audio is not billed again. An unknown paid response refuses resume, even with `--force`
- Change an input, profile, or output artifact and verify that resume refuses unsafe reuse. Verify that a completed run returns its saved result without another paid call
- Run two contenders against one destination and one run record. Verify preserved artifacts, resource locks, and the correct global exit after any earlier effect
- Check report stdout on failed checks, malformed Python messages, timeouts, and signal precedence over partial batch errors
- Discover one command and inspect one failed run using JSON without reading full transcripts or logs. A new session given only the agent instructions completes this path

Stubs exercise failures and count calls. The real Python and harness integrations must separately prove that they honor the protocol and tools-off mode.

Before tagging v0.1.0, run the paid end-to-end check from the transcript README against its listed test video.

## Done when

- [ ] Every item of the to-define list is settled in cli-spec.md or in this file
- [ ] Every acceptance scenario passes in `just check`
- [ ] The paid end-to-end check passes on macOS and on Omarchy
- [ ] An agent in a fresh session, given only the agent instructions, finds and reads a skill using only `ncly skill`
- [ ] That fresh session discovers a command, finds a recorded run after lost stdout, and resumes a supported case without repeating completed work
- [ ] The first real-domain contract proofs above pass, including no forbidden dry-run effects and refusal of uncertain paid work
- [ ] v0.1.0 is tagged and released
