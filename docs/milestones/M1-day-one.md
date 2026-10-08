# M1 Day one

Status: planned
Release: v0.1.0

## Goal

The first useful release. `doctor` and `auth` check and configure a new machine. Agents discover commands and skills, and transcript proves the shared operation contract with inspectable results and explicit resume.

## Depends on

M0. Pascal approved writing this milestone's spec before M0 is done. The spec is written now, in docs only. Go code waits until M0 is `done`, and v0.1.0 never ships before v0.0.1.

## Scope

**Spec.** The M1 sections of [cli-spec.md](../north-star/cli-spec.md) are drafts. Settle the [to-define list](#to-define-before-code), update the spec, and only then write Go.

**`auth`.** Keys as the Keys section of cli-spec.md defines them. `deepgram` is the one core service.

**`doctor`.** Components `core`, `auth`, `skill`, and `transcript`. The `skill` component checks structure only, as D021 says.

**`skill`.** `list` and `view` over the folders in `[skill] paths`, plus the `ncly skill <name>` shortcut for humans.

**Agent profiles.** `[agent.profiles]` in the config, as cli-spec.md defines them. `transcript` reads them for its summary.

**Discovery and records.** Implement [`ncly describe`](../north-star/cli-spec.md#ncly-describe-m1) and [`ncly run`](../north-star/cli-spec.md#ncly-run-m1). Grow the M0 declarations with their first M1 consumers. `internal/operation` owns records, execution locks, and result publication. Transcript owns its steps and reuse decisions. Keep listings bounded and full results in artifacts. Unknown effects remain visible and cannot be replayed through `--force`.

**Internal agent service.** `internal/agent` resolves profiles and owns the adapters that transcript needs, initially Claude Code and Pi. Verify tools-off support before using outside material. A missing harness uses `PREREQ_MISSING`. The public `ncly agent` command and other adapters still wait for M2.

**`internal/run`.** Starts with doctor's bounded local version probes, then extends to Python and the summary harnesses. It follows [Programs that ncly runs](../north-star/cli-spec.md#programs-that-ncly-runs-m1) for environment, grants, and cancellation. Each consumer adds only what it needs. M2 adds adapters and the public agent command. M3 reuses the runner for extensions.

**Test keychain.** A keychain stored in a file under `$WORK`, registered only in the test binary. go-keyring's mock lives in one process, and a scenario runs `ncly` several times. `NCLY_TEST_KEYCHAIN=unavailable`, which only the test binary reads, makes it stop answering.

**Test commands.** `waitfile <path>` waits until a file exists. `gone <path>` checks that the process whose ID the file holds no longer runs.

**`transcript`.** Move it in these steps:

1. Inspect the actual Python transcript CLI and harness invocations before fixing their adapter contract. Use stubs to test failure paths, then integrate a real local, non-billing path early in M1
2. Copy the Python transcript CLI from `authoring/andy/transcript/` in `pascalandy/skills` into `python/transcript/`, with its prompts, tests, and README. From then on, fixes land in Nicely first. The handoff happens early enough to test real behavior before the release gate
3. Embed it with `go:embed`. On first run, extract it to `~/.cache/nicely/python/<version>/` through a temporary folder and a rename, then run it with `uv run`.
4. Python exchanges versioned JSON Lines with Go and stops at the transcript. Its initial manifest fixes the items before effects. Go reserves result directories and syncs each paid intent before sending `ack`. Go maps each Python code to the error registry and renders the text from the catalog
5. Go passes the `deepgram` key as an environment variable of the child process. Python stops reading the keychain.
6. Go calls `internal/agent` for the summary using the resolved profile. Python no longer launches a harness. Transcript and summary remain separate recorded steps so the first can be reused when the second never started
7. The checks of the old `doctor --source` become the `transcript` component of `ncly doctor`.
8. A run shows one spinner. User-facing progress streaming stays in [M99](M99-parking-lot.md). The internal protocol still acknowledges effect boundaries needed for durable records before dispatch

**Python checks.** `just check` runs ruff, pyright, and pytest on `python/` through `uv`.

**Agent instructions.** [README.md](../../README.md#planned-agent-instructions-m1) holds the planned text to paste into an `AGENTS.md`. It covers discovery, readiness, planning, execution, inspection, and explicit resume. Agents set `NCLY_NO_INPUT=1` and `NCLY_JSON=1`, retain answers and invocation start times, and parse machine fields. Overall exit 75 alone permits an automatic repeat. Saved evidence and verified artifacts retain work across sessions. T10 verifies the text before Pascal adds it to his global instructions.

**Completion values.** The values that the completion section of cli-spec.md lists.

## To define before code

- [ ] The generic default output folder of `transcript`. Pascal's own values move to his config: his export folders, his Zoom folders, and the `synthese-rencontre` prompt of `andy-mode`, read from a folder of private prompts.
- [ ] Linux parity: browser cookies from Chromium and Chrome on Linux, and opening the result folder through `internal/platform`. The paid check on Omarchy waits for [#23](https://github.com/pascalandy/nicely/issues/23)
- [ ] The protocol between Go and Python: version, arguments, environment, initial item manifest, Go-owned result directory reservation, JSON shapes, sync before acknowledging a paid effect, signals, timeout, and behavior on EOF or a lost acknowledgement. No paid call may outrun its persisted intent
- [ ] The registry code and the catalog entry of each Python error code, all in place before M6.
- [ ] The mapping of the other transcript options onto Nicely's flags, such as `--output-dir` to `--output` and `--debug` to `NCLY_DEBUG`, and the fate of `--no-progress`, `--open`, and `--preview`.
- [ ] The profile transcript uses when the config has none, and the verified modes and supported versions of the initial `claude` and `pi` adapters
- [ ] How a `yt-dlp` fix reaches users, since the Python program pins `yt-dlp` and YouTube breaks it every few weeks, and how the Arc cookie adapter keeps working with a newer `yt-dlp`
- [ ] The time limit of a keychain call. An unlock prompt on a desktop takes longer than a script can wait.
- [ ] How `internal/run` reaches a descendant that left the process group of its parent and outlived it, since the transcript CLI starts its children in their own process groups.
- [ ] The keys each program declares: Python gets `deepgram`; the separate harness adapter declares its own key variables
- [ ] The nested JSON fields of command descriptions and run summaries, the stored record encoding, and filtered lookup after stdout is lost. Declarations cover required values, defaults, flag groups, and JSON Schema for current result types. Run filters apply before the limit, with `more` and bounded item-key previews
- [ ] Transcript artifact fingerprints, the evidence needed to reuse each step, and the cases that refuse resume. Verified required artifacts can resolve a lost completion message. An unanswered paid request without that evidence remains unknown
- [ ] Pascal's key migration, once per machine: `chezmoi secret keyring get --service=deepgram --user=api_key | ncly auth login deepgram --stdin`.
- [ ] The route around the `tmux info` call that Lip Gloss makes at load, [lipgloss#749](https://github.com/charmbracelet/lipgloss/issues/749). Every Charm form, spinner, and Markdown renderer loads Lip Gloss
- [ ] The summary failure path: no silent retry, and a way to summarize a saved transcript again without a second Deepgram charge when the summary failed or its outcome is unknown
- [ ] What M2 to M5 need from run records, discovery, and `internal/run`: harness tasks that are recorded but not resumable, extension manifests and grants, batch items, and publication per destination

## Tasks

Do the tasks in this order, each through the checklist in [AGENTS.md](../../AGENTS.md#work-on-a-milestone-task). Every item of the to-define list is settled before T1 starts. Step 1 of each task re-checks the sources that its spec sections cite, such as the transcript CLI and the harness versions, and fixes the spec first when one changed.

Before T1 changes `internal/tui`, publish one mockup page for every look that M1 adds: the `skill list` and `skill view` output, the masked key field, the confirmation, `auth status`, the doctor report, `run list` and `run view`, and the transcript spinner. The tasks continue while Pascal picks.

- [ ] **T1 Skill.** `ncly skill list`, `ncly skill view`, and the `ncly skill <name>` shortcut over `[skill] paths`, with `SKILL_SHADOWED`, `SKILL_NO_SOURCE`, and `NOT_FOUND`. As the first real config-consuming command, map unreadable, malformed, or wrongly typed config to `CONFIG_INVALID`. Validate profile semantics only when a command uses that profile. Ship the non-interactive output first. Glamour renders `view` in interactive mode once the Lip Gloss route is in place.
- [ ] **T2 Auth.** `ncly auth login`, `logout`, and `status`, the keychain through go-keyring with its time limit, and the test keychain. As the first writing command, it proves dry run against the keychain and the files with `snapshot` and `unchanged`. Its dry-run JSON reports pending keychain checks without reading a key or stdin. Its masked field and confirmation are the first Huh form, so `terminal.txtar` proves that no package asks the terminal anything at load.
- [ ] **T3 Doctor.** `ncly doctor` with the `core`, `auth`, and `skill` components, and the `CONFIG_UNKNOWN_KEY` warning. Add the path that keeps failed reports on stdout. Ordinary `WriteJSON` sends failed answers to stderr. Introduce only the minimal `internal/run` path for bounded local `--version` probes. Doctor's `--timeout` defaults to 30 seconds for checks, including `--live`. Prove that no-input checks never install anything. Install offers remain interactive. The `transcript` component lands in T9.
- [ ] **T4 Describe.** Reuse declaration fields added by T1 to T3 and extend the incomplete M0 declarations where needed. Cover requiredness, value types, defaults, flag groups, prerequisites, and typed current outputs using JSON Schema. Derive concise result `keys` from those outputs. One declaration supplies help, validation, doctor, and targeted discovery. Add `CONFIG_DEFAULTS_USED` for invalid config and completion values for skills, auth services, and doctor components. Build no custom schema language or framework for future commands.
- [ ] **T5 Python transcript.** Copy the transcript CLI into `python/transcript/`, and make `just check` run ruff, pyright, and pytest through `uv`. Embed it, extract it to the cache, and extend `internal/run` for its environment, grants, timeout, and process tree, proven with `waitfile` and `gone`. Protocol version 1 carries the transcript step, but T5 never saves a transcript end to end. Until T6 syncs the intent, `ncly` never sends `ack`, so no Deepgram upload can start. A stub `uv` proves that Go handles the protocol. The real program and real `uv` prove a dry run from an empty cache, including no new empty result directories. Python's tests save a transcript through a fake Deepgram transport supplied in test code only. No production endpoint override or fixture CLI is added.
- [ ] **T6 Run records.** Record the initial item manifest before effects. Go owns result directory reservation and records its destination before mutation. For each paid intent, sync the temporary record, rename it, then sync its parent directory before Go sends `ack`. A failed barrier prevents dispatch. Complete one non-billing run through the actual Go command, real `uv`, actual Python, a test-only fake transport, actual artifacts, and a saved record. `ncly run list` filters before limiting and reports `more` with bounded item-key previews. `run view` separates saved status from the execution lock's `active`. Interrupt before and after dispatch, then inspect the record in a new process.
- [ ] **T7 Summary.** `internal/agent` resolves the selected profile and checks its harness minimum version, tools-off support, and depth during preparation, before transcription dispatch. Use the `claude` and `pi` adapters only. Each real adapter proves tools-off behavior through observable evidence, beyond accepting flags or returning success. Go records the summary as a separate step. `ncly transcript summary run <run-id>` explicitly starts a new operation and may bill.
- [ ] **T8 Resume.** `ncly run resume` reuses verified steps and refuses unsafe ones. Reconcile all required recorded artifacts when the final completion message was lost. Missing or mismatched artifacts refuse reuse. Preserve the code, message, and hint of failed attempts whose paid step remains pending. Guard summary runs against an active source executor. Operation answers use `results` for one item and for many. Counted calls prove no repeated transcription, and two contenders prove locks and the overall exit.
- [ ] **T9 Transcript finish.** `ncly transcript run zoom`, `ncly transcript prompt list`, `--open`, the `transcript` doctor component, the spinner, and the completion values for profiles and prompts.
- [ ] **T10 Agent instructions.** The text that a user pastes into an `AGENTS.md`, shown in README.md. An agent in a fresh session, given only that text, completes the agent paths of [Done when](#done-when).
- [ ] **T11 Release.** Pascal moves his key, then the paid end-to-end check passes on macOS from a local session on the Mac. Tag v0.1.0 only after v0.0.1.

## Follow-up in `pascalandy/skills`

Once `ncly transcript` works, the `transcript` skill points to it, the `transcript` and `transcript-cli` recipes retire, `verify-transcript` checks `ncly`, and the Python copy there is deleted. Pascal decides when.

## Out of scope

- The `ncly agent` command: M2.
- Taps, `skill link`, and `nicely.toml`: M3.
- Porting transcript to Go: [M99](M99-parking-lot.md).

## Acceptance

Scenarios use the test keychain and stubs for failure injection and counted calls. They declare fixture skills and profiles in the `.txtar` archive. T6 also exercises the actual Go, `uv`, and Python path with a fake transport supplied by test code. No scenario starts a real paid service, and production gains no endpoint override or fixture CLI.

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

# A linked skill folder is listed once, and a broken link is skipped
# [skill] paths lists skills, then shared
symlink $WORK/skills/delta -> $WORK/shared/delta
symlink $WORK/skills/ghost -> $WORK/nowhere
exec ncly skill list --json
stdout -count=1 '"name":"delta"'
stdout '"path":"[^"]*/skills/delta"'
! stdout SKILL_SHADOWED

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

- [ ] Compare user files, config, keychain, and state before and after dry run. Permit only documented cache writes, with no key read, paid call, harness start, or new user directory. Detect newly created empty result directories explicitly. `snapshot` currently ignores directories
- [ ] Prove the initial manifest and Go-owned directory reservation precede effects. Fail the persistence sync and assert that no `ack` or paid dispatch occurs
- [ ] Complete the T6 non-billing path through actual Go, real `uv`, actual Python, test-only transport, verified artifacts, and a record inspected in a new process
- [ ] Stop before paid dispatch, after dispatch but before the response, and after saving the transcript but before starting the summary. Preserve the cause of a pre-intent failure as pending-step code, message, and hint
- [ ] Return 504 after upload. Assert overall exit 1, an unknown transcription, and `RESUME_UNSAFE` with no replay, even under `--force`
- [ ] Resume a saved transcript whose summary never started. Count service calls to prove that transcription is not repeated. Reconcile a lost completion message only when every required recorded artifact verifies. Refuse missing or mismatched evidence
- [ ] Change an input or output artifact and refuse unsafe reuse. Change the config profile and run the remaining steps with the recorded profile. Return a completed run's verified result without another paid call
- [ ] Run two contenders against one destination and one record. Preserve artifacts and verify resource locks and the overall exit after earlier effects. Refuse a new summary while the source executor is active
- [ ] Check report stdout on failed checks, bounded doctor probes, non-installing no-input checks, malformed Python messages, timeouts, and signal precedence over partial batch errors
- [ ] Verify transcript `results` for one item and for many. Treat item failure codes as cause metadata. Only the overall exit can permit an automatic retry
- [ ] Lose stdout, add more than 20 unrelated runs, and recover through filters applied before the limit. Verify `more` and bounded item-key previews. Repeat without a remembered start time, preserve ambiguous matches, and distinguish saved `run.status` from lock-based `active`
- [ ] Give a fresh agent only the README instructions. Discover a command, plan it, inspect a recorded failure without full transcripts or logs, then dry-run and explicitly resume supported work without repeating a paid step

Stubs exercise failures and count calls. The real Python and harness integrations must separately prove that they honor the protocol and tools-off mode.

Before tagging v0.1.0, run the paid end-to-end check from the transcript README against its listed test video.

## Done when

- [ ] Every item of the to-define list is settled in cli-spec.md or in this file
- [ ] Every acceptance scenario passes in `just check`
- [ ] The paid end-to-end check passes on macOS, from a local session on the Mac. Omarchy waits for [#23](https://github.com/pascalandy/nicely/issues/23)
- [ ] An agent in a fresh session, given only the agent instructions, finds and reads a skill using only `ncly skill`
- [ ] That fresh session discovers a command, finds a recorded run after lost stdout, and resumes a supported case without repeating completed work
- [ ] The first real-domain contract proofs above pass, including no forbidden dry-run effects and refusal of uncertain paid work
- [ ] v0.1.0 is tagged and released
