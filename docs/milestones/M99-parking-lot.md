# M99 Parking lot

Status: open

Ideas that are kept but belong to no delivery milestone. To start one, select a small useful slice, move it into a milestone, and remove only that slice here. These are options, not MVP acceptance criteria. [D061](../north-star/decisions/D061-smallest-useful-capability.md) records the scope audit of PR #156.

| Idea | What it would do | Revisit when |
|---|---|---|
| Routines | `ncly routine run <name>` runs a named sequence of commands from the config | A repeated sequence in an installed extension needs a named shortcut |
| Remote commands | `--host` runs a command on another machine over SSH, including `ncly doctor --host` | A remote task needs Nicely, or a remote command needs a key |
| Dotfiles for Linux machines | A dotfiles repository that carries the shared config to each Omarchy machine | Pascal adds a second Linux machine. This lives outside Nicely |
| Launcher | A bare `ncly` in interactive mode opens a fuzzy command palette, and prints help everywhere else | Several everyday commands ship and finding one becomes a recurring problem |
| Help by intent | The root help groups commands by purpose, such as tools, inspect and recover, and set up, and `ncly help --search <text>` finds a command by its words. Agents already search `ncly describe --json` | The root help lists more than ten commands, after a mockup that Pascal picks |
| `--explain` | Prints the underlying command and a link to the wrapped tool | The everyday domains wrap several tools |
| `ncly tools` | A showcase of every curated tool, why it was chosen, and its credits | The everyday domains ship |
| Aliases | Git-style aliases in the config, such as `md = "markdown view"` | A command feels too long in daily use |
| Completion after `--` | `ncly -- <Tab>` offers nothing, because ncly rejects a command name after `--`. Cobra offers the commands, so the fix answers Cobra's completion protocol in `Main` | A user hits it, or a command group starts taking arguments |
| `ncly ignite` | Bootstraps a project with Pascal's setup: Git, just, Lefthook, and signoff, through Copier templates | Pascal starts his next new project |
| `ncly skill pack` | Zips a skill for upload to the Claude apps | A skill must work in an app that cannot read GitHub |
| `ncly skill index` | Generates the remote skills index that `just remote-skills` builds today | The skills repository moves its tooling into Nicely |
| `ncly skill run` | Runs a script of a skill with its prerequisites checked | Agents struggle to run skill scripts by path |
| MCP skills server | Serves skills over the MCP Skills extension, SEP-2640, from the `skill` domain package rather than from CLI output. Per-file manifests join `view --json` as optional fields, never `list`, which agents read whole. Research in [#10](https://github.com/pascalandy/nicely/issues/10) | A host Pascal uses supports SEP-2640 and cannot read skills from disk. `ncly skill pack` answers the same need |
| Record check in core | Core reads the run record of an answer that exits 75 and turns it into a protocol failure when the record breaks [Retry safety](../north-star/contract/retry-safety.md), as [D045](../north-star/decisions/D045-step-effects.md) defers | A tap carries an extension that does not use the Go SDK |
| Answer limits | The answer check bounds the size of a held answer and of a held stderr line, its JSON depth, and rejects invalid UTF-8 and duplicate keys, then stops the extension and answers `RUNTIME`. An extension built on the SDK cannot produce these answers | A tap carries an extension that does not use the Go SDK |
| Hooks | Lets an extension change how core or another extension behaves, which [D044](../north-star/decisions/D044-additive-extensions.md) excludes | A need appears that neither core, the SDK, nor the extension that owns the step can meet |
| Progress events | Long commands expose live JSON Lines events and a step list. Run records already keep effect boundaries and support inspection after interruption | Live progress adds value beyond the saved run record |
| Transcript in Go | Ports the Python program of the transcript extension to Go | Python blocks a release or a translation |
| Structured summaries | A bundled prompt writes, beside its Markdown, a versioned JSON file of decisions, actions, and open questions, each citing a segment of the transcript that a local check verifies | Pascal wants to search decisions or actions across summaries |
| Transcript on Omarchy | The paid end-to-end check of `ncly transcript` passes on Omarchy, with YouTube cookies from a Linux browser. [#23](https://github.com/pascalandy/nicely/issues/23) tracks it | Pascal needs transcripts on Omarchy |
| One duration parser | `ncly doctor` reads its `--timeout` through `sdk/command.Duration`, as `ncly agent run` does, instead of its own `parseTimeout` in `internal/cli/doctor.go`, so both commands share one rule | A change touches `ncly doctor --timeout`, or a third command takes `--timeout` |
| Nested deadlines and cleanup | The program runner passes `NCLY_DEADLINE`, an absolute UTC time, to every program, and each `ncly` ends its work before that time minus its 15 seconds of cleanup. A nested `ncly` also gets a shorter cleanup than its parent. It then times out, or writes its record after the signal, before the parent's SIGKILL | An approved summary slice actually calls a nested recorded command |
| More languages | Spanish, Japanese, and others, through the M22 workflow | A user asks, or `fr-CA` proves the workflow |
| Terminal demos | VHS recordings of key commands for the README and the docs | The everyday domains ship |
| First-run experience | What a new visitor sees first, and the marketing around it | The everyday domains ship |
| Project config | A `.nicely.toml` in a repository, read between the environment and the local config | A project needs its own command settings or paths |
| `ncly config` | `ncly config get`, `set`, and `list` edit the config with comments kept | Agents edit the config by hand often enough to break it |
| Key checks for skills | `ncly doctor` checks the keys that skills declare in `nicely.toml`, replacing `just api-keys-validation` | A selected tap slice carries skills with prerequisite manifests |
| Man pages | Generated from Cobra and shipped in the packages | A user asks, or Nicely applies to homebrew-core |
| Build provenance | GitHub artifact attestations for the release archives | Strangers install from the release archives |
| AUR package | Publishes `ncly-bin` to the AUR, so that `yay -S ncly-bin` installs `ncly` and `ncly <Tab>` completes in bash on Omarchy. It needs an AUR account with the public key `nicely-release-aur`. GoReleaser builds the package that `just release-check` verifies, and `skip_upload: true` in the `aurs` section of `.goreleaser.yaml` keeps it off the AUR. To ship it, delete that line, and say in step 3 of the `release-nicely` skill that the workflow publishes the AUR package | AUR registration reopens |
| homebrew-core | A formula in Homebrew's main repository | Nicely meets Homebrew's notability rules |
| Windows | Support for Windows | Never planned. Revisit only on real demand |

## Agent extras

Parked from M09 and M11: Claude, Pi, Grok, and OpenCode adapters; a Codex API key; named
profiles; provider selection; model catalogs; profile status screens; configurable system prompts; untrusted
`--input`; built-in code review; `--check-unchanged` snapshots and guards; and migration of the existing `headless` skill.
M09 and M10 keep one trusted task through the installed Codex CLI, native login, model and
effort settings, dry run, durable execution, inspection, and cancellation.

Revisit each extra for a demonstrated workflow that the current path cannot serve. A second
adapter must solve that need on its own. Model names in a brainstorm do not create an adapter
matrix. A tools-off consumer must first prove that its chosen runtime disables tools, as
[D028](../north-star/decisions/D028-outside-text.md) requires; a filesystem check is no substitute.
The old multi-adapter M11 is removed from the delivery queue.

A Codex API key needs no new Nicely machinery: M08 already ships optional keys. PR #156
recorded that `codex exec` reads `CODEX_API_KEY`, so the adapter would map the granted `openai`
key to it. Reverify that flag before a card uses it.

## Taps

Parked from M12, whole: `ncly tap add` and `ncly tap list`, release packaging of the first-party
extensions, staged installation and source receipts, arbitrary Git remotes, automatic updates,
`tap sync`, `tap remove`, ownership and conflict recovery for updates, tap-provided skill
discovery, `skill link`, harness-specific skill destinations, standalone skill prerequisite
manifests and their doctor checks, and a private fleet extension. Today `just install` installs
`ncly` and the first-party extensions from a checkout, and Homebrew ships `ncly` itself.

Revisit when a machine or a user without a source checkout needs an extension. Start with one
first-party extension from the official tap, installed as a complete artifact that records its
source, grants no key, and overwrites no user-owned destination.
[M12 at e0a0122](https://github.com/pascalandy/nicely/blob/e0a0122c18d284b2e5b148ede34e972b388bb858/docs/milestones/M12-taps.md)
holds that slice, its open questions, and the `tap list` look that Pascal picked in M07-T1.
Revisit updates after a second real version needs installation, and skill distribution or fleet
work for a concrete consumer. That work must retain
[D035](../north-star/decisions/D035-sync-ownership.md) where it manages persistent copies.

## Transcript summary

Parked from M14 and the implicit summary in M13: summarizing transcripts, prompt selection,
agent invocation, and the evidence needed to associate each summary with its source.
M13 produces a transcript from one YouTube URL without invoking an agent.

The first candidate is one explicit summary of one verified saved transcript, with one prompt.
It requires a proven tools-off path for outside text under D028. The selected Codex path does
not supply that capability. Do not add a second harness merely to keep the old roadmap moving.
M14 is removed from the delivery queue; a concrete need and that capability must precede a
replacement slice. Structured summaries remain a separate option above.

## Resume

Parked from M15 and the broader M10 fixtures: `ncly run resume`, partial multi-step continuation,
nested child-run recovery, and compatibility tests for changing a step's effect policy.
This is a deferred capability, not a prerequisite for recording one paid step safely.

M10 and M13 retain locks, durable intent before a paid request, verified completion, run
inspection, same-ID completed-result replay, and refusal of an unknown paid effect.
Retry before any persisted non-repeatable intent remains governed by the shared contract.
A transcript can reconcile complete, verified local artifacts without another upload.

Revisit when a selected command has useful partial work that the user actually needs to
continue. Design and prove that one continuation before generalizing it to other domains.
M15 is removed from the delivery queue.

## Transcript extras

Parked from M13 and M16: multiple URLs and batch orchestration, Zoom folder discovery and
`--latest`, browser matrices, configurable prompts, output opening, spinners, and automatic
migration or retirement of the old transcript CLI and skills.
M13 keeps one YouTube URL and its three verified artifacts. M16 checks its prerequisites,
teaches the command, and proves that an independent fresh session can use it.

Revisit another input source for a real transcript need. Add browser support only when the
selected machine needs it. Retire the old CLI only after Pascal's required uses have parity;
a partial replacement does not justify deleting a working workflow. Pascal picked the spinner's
look in M07-T1; [the transcript spec at 29b311f](https://github.com/pascalandy/nicely/blob/29b311fed17791eb8d91d043a70751c3c2ce3d3e/extensions/transcript/spec.md)
records it, so a spinner needs no new mockup.

## Everyday domains

Parked from M17: exhaustive inventories of two machines, histories and packages, and the rule
that every writing command supports files, folders, globs, batches, progress, records, and
resume. A safe one-file transformation does not inherit all of those capabilities.

M17 selects one frequent job and its smallest useful input and output. M18 keeps Markdown
view as a bounded candidate, subject to that selection. Writing commands still need dry run,
destination validation, overwrite consent, and truthful retry behavior for their actual effects.

The unselected video and image suites of M19 and M20 are removed from the delivery queue.
Keep their ideas here: video encode, cut, and extract audio; image encode, resize, and compress.
Revisit one operation when a concrete daily job warrants it, and write only its contract and
one fixture. Neither domain implies that the whole command suite ships together.

## Docs hub

Parked from M21: project scanning, a visibility registry, `docs list`, `sync`, and `check`,
editable generated mirrors, synchronization recovery, public builds, publishing, and custom
search. Local documentation aggregation does not require a publishing platform.

M21 keeps explicit project roots, one existing renderer, a local `docs serve` view, read-only
sources, and a disposable cache. Resolve the actual renderer and multi-project input layout
before the milestone becomes ready. Revisit a parked feature only when the basic local view
fails a demonstrated need; public publication needs its own explicit selection of sources.
