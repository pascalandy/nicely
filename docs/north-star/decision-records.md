# Decision records

Each entry records what was decided, why, and what was rejected. An entry states its decision briefly, and the [guide](guide.md), [cli-spec.md](cli-spec.md), and the extension specs hold the full rules.

Until v0.0.1 ships, fix an entry in place, because git keeps the old text. From v0.0.1, add a new entry that names the entry it replaces, and mark the old entry `Replaced by Dxxx`.

## D001 Build from scratch in Go with Cobra and Charm

Decided 2026-10-05. Nicely is written in Go on Cobra, with Lip Gloss, Huh, Bubble Tea, Bubbles, and Glamour for the interface. Until a form needs Huh or Bubble Tea, `internal/tui` styles text with `x/ansi` and `colorprofile`, both from Charm, without Lip Gloss: Lip Gloss v2.0.6 detects the terminal as its package loads, and inside tmux that runs `tmux info` with no time limit, as [lipgloss#749](https://github.com/charmbracelet/lipgloss/issues/749) reports. Even `ncly --version` could hang. The first form must prove that its packages no longer probe at load, with the scenario `terminal.txtar`. The first command that needs Huh, a Bubbles spinner, or Glamour uses a Lip Gloss version that defers the call: an upstream release, or, after Pascal approves it, a fork wired through a `replace` line in `go.mod` and removed once upstream ships the fix. Until then, that command ships its non-interactive output first.

**Why.** Cobra already carries the practices of `gh`, `kubectl`, and the Stripe CLI. The Charm libraries cover forms, spinners, styles, and Markdown in one family.

**Rejected.** Forking the Stripe CLI, because most of its code is Stripe-specific, such as login, webhooks, API resources, and fixtures. Fang, because its source hard-codes the help headings `usage`, `flags`, `examples`, and `ERROR` in English and takes `-v` for `--version`, which Nicely uses for `--verbose`. Checked on 2026-10-05.

## D002 Name the project Nicely and the command ncly

Decided 2026-10-05. The project is Nicely, the command is `ncly`, the repository is `pascalandy/nicely`, and the Go module is `github.com/pascalandy/nicely`.

**Why.** `nicely` is taken on PyPI and npm, but neither package installs a `nicely` command. Nicely ships through Homebrew and GitHub release archives, later through the AUR too, and the name is free in Homebrew and the AUR. `ncly` is free on PyPI, npm, crates.io, and Homebrew, and it is short to type.

## D003 Put the domain first and keep one verb per action

Decided 2026-10-05. Commands follow the grammar in [cli-spec.md](cli-spec.md#usage): `ncly <domain> [<resource>] <verb>`, with singular nouns and one verb per action across domains.

**Why.** An agent explores one domain at a time, so `ncly video --help` costs fewer tokens than a flat list of fifty verbs. Options stay specific to their domain. Names become predictable: after `ncly image convert`, an agent guesses `ncly video convert`. The verbs are those of `gh`, which agents already know.

**Rejected.** The verb-first form, such as `ncly convert`. Keeping the interface of a script that joins Nicely, such as `transcript list prompts`, because one mixed shape breaks the guess.

## D004 Keep core minimal and make every domain an extension

Decided 2026-10-05, revised 2026-10-09. Core is a small host: dispatch, help, completion, `describe`, `doctor`, `auth` with grants, `run`, and `tap`. Everything that Nicely does for a user is an extension, with one manifest format and one contract. Only what an agent needs to operate Nicely is bundled into `ncly`: the `skill` extension and its `nicely` skill. Every other domain is external, including the first-party `agent`, `transcript`, and the everyday domains, and Pascal's personal tools stay in his private tap. The guide's domain table says where each one lives, and the milestones decide when it ships.

**Why.** Nicely aims for the composability of Pi v1: a minimal core where adding a capability never edits core. Building domains inside core first would force a rewrite when extensions arrive, and would leave the extension contract unproven until late. Shipping first-party domains as extensions proves the contract on our own work. Nicely needs no agent to operate, so even `agent` is external.

**Rejected.** Core holding everything a stranger would use, the first version of this entry, because domains in core grow a second, privileged path. Bundling the everyday domains for a first impression, because one rule without exceptions keeps the archive minimal. Shipping the first-party domains as separate executables inside the archive, as `docker compose` does, because the records, keys, and agent services would have had to cross processes before any domain needed them. Keeping `agent` in core as the single owner of tools-off mode, because the `agent` extension owns it just as well.

## D005 Add extensions the Git way and carry them in taps

Decided 2026-10-05, revised 2026-10-09. An executable named `ncly-<domain>` becomes `ncly <domain>` when its manifest declares a compatible protocol and its capabilities. A bundled extension is a Go package compiled into `ncly` with the same manifest, as [Extensions](cli-spec.md#extensions) defines. A tap is a Git repository that carries skills and extensions, like a Pi package. Core commands and bundled extensions always win over an external extension with the same name. Source precedence is explicit, and discovery reports the selected source. An extension receives the global flags as the variables in [cli-spec.md](cli-spec.md#global-flags), and it prefixes its own error codes with its domain.

**Why.** A script can join Nicely in any language. Its manifest lets an agent inspect what it supports before running it. One private tap moves Pascal's tools to every machine. Git, `kubectl`, and cargo find plugins by executable name, while Nicely also needs to verify their shared output contract.

**Rejected.** A tap for skills only. Translating extension text from core, because core cannot know an extension's strings. Guessing capabilities from an executable's name or version. Accepting malformed protocol output as a successful result.

## D006 Serve agents through the CLI itself

Decided 2026-10-05. Agents operate `ncly` through the shell. `ncly` launches agents through their own non-interactive modes. Nicely uses no agent SDK.

**Why.** A shell command is the one interface every harness shares, so Nicely stays independent of any single agent.

**Rejected.** Limiting Nicely to Claude.

## D007 Keep one stable agent contract

Decided 2026-10-05. Command names, flags, JSON keys, error codes, and exit codes stay in English. [Compatibility](cli-spec.md#compatibility) covers types, units, formats, meaning, required fields, nullability, enum values, and defined array order. Protocols and stored run records have versions from their first use. Each core error code maps to one exit code. Exit `78` means that a human must act. Exit `75` follows the [retry safety rule](cli-spec.md#retry-safety) for the whole invocation, including finished items in a batch.

**Why.** Agents branch on codes and values, so keeping a key while changing its unit can break a consumer. A temporary failure before a paid request can still follow a file write. Exit `75` therefore requires that all effects so far are safe to repeat, through absence of incompatible effects or proven idempotence. `78` is `EX_CONFIG` in `sysexits.h`, and it lets an agent tell "a human must act" from "the call was wrong".

**Rejected.** Stability of key names alone. Treating every new enum value as compatible without an explicit rule for unknown values. Recording a breaking change in this log without a version and migration policy. Returning `75` for a lock conflict after effects that make the invocation unsafe to repeat. A dedicated exit code 3 for a missing key, or exit 2 for a step that only a human can do.

## D008 Use one flag for every confirmation

Decided 2026-10-05. In non-interactive mode, `--force` answers yes to every confirmation, such as an overwrite or a deletion. There is no `--yes`. Installing a prerequisite and granting a key are the exceptions: both wait for a human in a terminal, as principles 5 and 6 require.

**Why.** clig.dev asks for `-f` or `--force` when a confirmation cannot be asked. One concept gets one flag, so an agent never has to guess which of two flags a step needs.

**Rejected.** `--yes` beside `--force`, as in Pascal's script conventions, which serve scripts that have both a prompt and a separate safety check. `--force` granting the keys a tap requests, because an update could then gain a key in silence.

## D009 Ask in a terminal, fail elsewhere

Decided 2026-10-05. A missing value opens a form in interactive mode and exits 2 in non-interactive mode. After a form, `ncly` prints the equivalent command after `Next time:`. `CI` and `NCLY_NO_INPUT` also select non-interactive mode.

**Why.** Agents never meet a prompt. Beginners get a guided path, and the last line teaches them the flags. Some agent tools run commands in a pseudo-terminal, so a terminal alone does not prove that a human is there.

## D010 Keep keys in the keychain or the environment

Decided 2026-10-05. Keys live in environment variables or in the OS keychain through go-keyring, under the service `nicely` and an account named after the service. The environment wins. Any service name works, and its variable name derives from it, as [cli-spec.md](cli-spec.md#keys) says. `ncly` reads a key from the keychain only when the environment lacks it. `ncly auth` and `ncly doctor` also check the keychain. Every call has a time limit, so an unavailable keychain blocks only a command that needs it. Agents relay the hint and never handle a key.

**Why.** Each program owns its own keychain entries, as `gh` does with `gh:github.com`, so `logout` touches only Nicely's keys. The environment lets any other secret store feed `ncly`. Open service names let an extension declare a key that core does not know. A Linux machine reached over SSH often has no unlocked keychain, and go-keyring's unlock call has no time limit of its own. `gh` wraps the same library with a 60-second limit on every call, which leaves a human time to answer an unlock prompt. Nicely keeps 60 seconds in interactive mode and waits 10 seconds otherwise, because an agent that waits a minute for a prompt nobody sees is stuck.

**Rejected.** A fallback file. Sharing entries written by another tool, such as chezmoi's `service=deepgram, user=api_key`, because two programs would then own one entry. Remote commands that need keys, parked in M99.

## D011 Check locally by default, test live on request

Decided 2026-10-05. `ncly doctor` stays on the machine by default. `--live` calls free endpoints only. Invalid config is a failed report finding on stdout, and independent checks still run. An install runs only after a yes in a terminal.

**Why.** Doctor must be safe to run at any time. A live check proves a key works without a bill.

## D012 Make every word of core translatable from day one

Decided 2026-10-05. Core text lives in go-i18n catalogs with a description per entry. A test compares every catalog with `en`, and a pseudo-locale scenario exposes text outside the catalog from M00. Plurals, numbers, sizes, and dates go through the i18n package. English ships first, and Canadian French follows in M22. Each extension keeps its own catalog.

**Why.** Adding translation later means touching every string and every number format. AI makes catalogs cheap to keep complete, and a CLI in the user's language reaches people most CLIs ignore. The pseudo-locale catches a missing string in the milestone that adds it instead of in M22.

**Rejected.** A static check for string literals, because the pseudo-locale also catches truncated layouts.

## D013 Write core and the first-party extensions in Go

Decided 2026-10-05, revised 2026-10-09. Core and the first-party extensions aim to be fully Go. An extension may embed a Python program during its move, as transcript does, but that program returns JSON only, and the extension's Go code owns display, translation, and exit codes. An extension from a tap may use any language, as D005 allows.

**Why.** One binary and one display layer keep the experience identical across commands.

**Rejected.** Letting Python keep its own display.

## D014 Publish from the first commit

Decided 2026-10-05. The repository is public from day one, under the MIT license, with no telemetry. Issues and pull requests are welcome.

**Why.** Public work keeps good habits from the start. Everything personal lives elsewhere: private taps, the config folder, and the docs hub.

## D015 Target macOS and Linux with the same paths

Decided 2026-10-05. Nicely runs on macOS and on Linux, with Omarchy as the reference. Both systems use XDG paths under `nicely`. On macOS, a formula in the Homebrew tap `pascalandy/homebrew-tap` builds `ncly` from source. On Linux, v0.0.1 ships prebuilt archives on GitHub. The AUR package `ncly-bin` follows when AUR registration reopens, as [M99](../milestones/M99-parking-lot.md) records.

**Why.** Pascal moves toward Linux machines over the years, and the same paths everywhere keep his future dotfiles simple. A binary built on the user's machine never meets Gatekeeper, so Nicely needs no Apple Developer account, and Homebrew generates the completions itself. Homebrew builds `gh` from source the same way. v0.0.1 installs with Homebrew or the Linux archive. The AUR, the native channel on Arch, follows when AUR registration reopens.

**Rejected.** Signing and notarizing, which costs US$99 per year. A cask that strips the quarantine flag with `xattr`, which GoReleaser discourages. GoReleaser's `brews` section, deprecated in v2.10.

## D016 Run CI locally and release by hand

Decided 2026-10-05. Checks run through `just`, Lefthook, and `gh signoff`. GitHub Actions runs only when an agent starts the release workflow. External pull requests go through the maintainer's agent and `just signoff`.

**Why.** Pascal often hits GitHub Actions limits, and local checks are faster.

## D017 Write acceptance criteria as testscript scenarios

Decided 2026-10-05. Each card names the testscript scenarios that prove it. End-to-end scenarios come before unit tests. M00 verifies the real parser, streams, modes, and shared outcome logic, including partial results. A later domain proves its effects, simulation, and recovery on its first real implementation.

**Why.** Pascal writes the specifications and agents write the Go. A scenario reads like a terminal session, so Pascal checks behavior without reading Go. Parser errors and misplaced output can make a command unusable by agents even when its success path works.

**Rejected.** Treating a contract fixture as proof that a future command avoids writes or duplicate paid requests. Stubbed services keep tests repeatable, but they do not prove that the real Python program and harness support the required protocol and restrictions.

## D018 Run agents through the `agent` extension with one profile registry

Decided 2026-10-05, revised 2026-10-09. `agent` is a first-party external extension. `ncly agent run` runs a task through a harness, and the extension owns the profiles, the harness adapters, tools-off mode, and the depth limit, as [its spec](../../extensions/agent/spec.md) defines. Profiles and dry run come in M09, execution in M10, and the other harnesses in M11. Transcript summarizes through `ncly agent run --json`, as any other extension would. A depth limit stops agents from launching agents without end. `ncly agent` absorbs the `headless` skill.

**Why.** Model IDs change every few months, so they belong in the config, not in code. One registry and one owner of harness behavior keep profiles, tool restrictions, cancellation, and summary recovery consistent. Composing through the CLI tests the same interface that agents use. A simple one-step command is the easiest first consumer of run records. "Harness" stays the internal term.

**Rejected.** "Infer", "chat", and "prompt", which mean little to most people or describe too little. Each project declaring its own inference settings. Python owning the summary. An internal agent service inside core that transcript calls in process, the first version of this entry, because core then grows a harness owner that it never needs to operate.

## D019 Ship transcript as a first-party extension

Decided 2026-10-05, revised 2026-10-09. `transcript` is a first-party external extension: `ncly transcript run youtube` and `ncly transcript run zoom`, as [its spec](../../extensions/transcript/spec.md) defines.

**Why.** It needs a Deepgram key and an agent for its summary, so it is no part of operating Nicely, and a first-time visitor should start elsewhere. It handles Zoom audio as well as YouTube.

**Rejected.** `ncly video transcribe`, the wrong domain for Zoom audio.

## D020 Build the extension host before any domain

Decided 2026-10-05, revised 2026-10-09. The extension host comes first, in M01, then the SDK, then `skill` as the first bundled extension. `describe`, `doctor`, `auth`, and grants follow. `agent` comes before transcript, and brings run records. Transcript then arrives in four milestones: transcription, summary, resume, and the rest. The real Python program joins the first complete transcript path.

**Why.** A domain built before the host would be rewritten as an extension. Each milestone adds one capability on top of proven layers. `agent run` is a one-step paid command, so it proves records before transcript's two steps and its Python protocol. Early integration still exposes ownership and protocol mismatches while they are cheap to fix.

**Rejected.** The first version of this entry, a day-one milestone with doctor, auth, skill, transcript, profiles, records, and resume in eleven tasks plus separate acceptance lists. Agents picked single tasks and left the rest, and nobody could follow what remained.

## D021 Move skill requirements to a file

Decided 2026-10-05. Skills declare their requirements in frontmatter today. Extensions use one versioned manifest format from M01, and skills join it when taps carry them in M12. A skill keeps `nicely.toml` beside `SKILL.md`; an extension keeps `<executable>.toml` beside its executable so several extensions can share a directory. Until then, `ncly doctor skill` checks only the structure of a skill.

**Why.** One file format serves skills and extensions alike. A frontmatter reader for requirements would be thrown away once manifests exist.

## D022 Keep docs private by default

Decided 2026-10-05. Each project in the docs hub is `public`, `private`, or `exclude`, and `private` is the default. The hub stays plain Markdown, so any engine can render it.

**Why.** Personal details slip into docs without anyone noticing. Publishing a project must be a choice made for that project alone.

## D023 Configure with a shared TOML file and a local one

Decided 2026-10-05. The shared config describes the setup the user wants, and the local config overrides it on one machine, as [cli-spec.md](cli-spec.md#configuration) defines. `ncly` edits a config file only through a command whose job is to change the setup, and it keeps comments and formatting. An unknown key is a warning. Nicely's own environment variables start with `NCLY_`.

**Why.** TOML is easy to read and edit, and Pascal's tools already use it. One shared file lets a new machine reach the same setup, for example with `ncly tap sync`. The local file keeps machine paths out of the shared one. A warning on unknown keys keeps older versions of `ncly` working.

**Rejected.** YAML. A config that only humans write, with Nicely's own records kept per machine, because each new machine would then need every tap added again.

## D024 Ship completions with dynamic values

Decided 2026-10-05. Cobra generates zsh, bash, and fish completions. For v0.0.1, Homebrew installs them and the Linux archive ships them. The AUR package will install them when AUR registration reopens. Completion suggests values such as skill names and profiles.

**Why.** Completion is how a terminal user discovers flags and values. Cobra builds it from the command tree, so it never drifts from the commands.

**Rejected.** Hand-written completion scripts.

## D025 Keep rules in north-star and the extension specs, and work in milestones

Decided 2026-10-05, revised 2026-10-09. `docs/north-star/` holds the guide, the CLI spec of core, Pascal's developer preferences, and this log. Each extension keeps its spec in `extensions/<name>/spec.md`. `docs/milestones/` holds one file per milestone, and M99 is the parking lot. A milestone becomes ready only once the spec sections that its cards read are written, and its cards link to them.

**Why.** Each rule has one source, so milestones cannot drift from the spec. An extension's spec lives with the rest of what the extension owns. Agents read the guide every session and the rest only when a card links to it.

**Rejected.** A spec inside each milestone file.

## D026 Write the repository in English

Decided 2026-10-05. Code, comments, docs, and commit messages are in English. User-facing text goes through the catalogs.

**Why.** The repository is public, and English reaches the most contributors.

## D027 Answer in one JSON line with `ok` and `errors`

Decided 2026-10-05. With `--json`, every answer other than help, the version, and a completion script is one JSON object on one line with `ok`. A failure lists `errors`, each with `code`, `message`, and `hint`. The command determines its overall outcome before choosing the leading error and rendering the answer. That error agrees with the exit code, and item errors preserve their own causes. A report command keeps its report, with its own findings key, on stdout. [cli-spec.md](cli-spec.md#output) holds the details.

**Why.** Stripe, npm, JSON:API, and GraphQL all answer with error objects that carry a stable code. The key `errors` matches Pascal's script-output convention. `ok` stays readable when an agent merges stdout and stderr. Transcript operations always answer with `results`, including one item, so adding a URL does not change the shape and an item cause cannot authorize a whole-command retry. Only the overall exit 75 does that. ESLint, ShellCheck, `terraform validate -json`, and `npm audit --json` keep their reports on stdout when a check fails. One line keeps `tail -n1 | jq` working. Within one exit code, the leading error's hint identifies the first fix. For 78, config and terminal problems precede tools and keys. For 1, a specific refusal such as `RESUME_UNSAFE` precedes `RUNTIME`. For 2, a valid call precedes confirmation, a name, or an extension's own code.

**Rejected.** A flat error object without `ok`. A single `error` object. JSON by default without `--json`, because humans use the same commands. A report on stderr when a check fails. Letting the first observed failure authorize a retry of the whole batch.

## D028 Keep outside text away from agent tools

Decided 2026-10-05. Text from outside Nicely, such as a transcript or a web page, reaches an agent only with the agent's tools turned off. The adapter must support and enforce this mode or refuse the call. The planned `--check-unchanged` flag detects changes after a run and makes no promise to prevent writes.

**Why.** A video can carry instructions. The transcript CLI already summarizes with a tool-free `claude --print` for that reason.

**Rejected.** Calling a post-run check `--read-only`. Relying on that check to prevent writes, or silently allowing tools when a harness cannot turn them off.

## D029 Decide early what is costly to change

Decided 2026-10-05. A rule that a later milestone would otherwise rewrite goes into the north-star files or an extension's spec before the code that it constrains. Its code waits for the milestone that uses it.

**Why.** Extensions, Canadian French, the shared config, and agent profiles all put constraints on the contract, the config, and the catalogs. A rule written early costs a paragraph. The same rule found later costs a rewrite of code already written.

**Rejected.** Deciding each rule only when its milestone starts.

## D030 Treat extensions as trusted code and grant keys one by one

Decided 2026-10-05. A program that `ncly` runs gets the environment of `ncly` minus every key that is not granted to it, as [cli-spec.md](cli-spec.md#programs-that-ncly-runs) defines. No program runs in a sandbox. A key reaches an extension only after a human grants it in a terminal. Grants stay on one machine and bind the extension's source and domain. An extension from a different source cannot inherit a grant through its name, and an update that declares a new key gets nothing until a human grants it.

**Why.** Git, `kubectl`, `gh`, and cargo pass the user's environment to their plugins, and real tools need it, such as the SSH agent and proxies. An extension runs with the user's rights: on macOS, go-keyring stores keys through `/usr/bin/security`, which any process can call, and on Linux any program of the session can read an unlocked keychain. Grants are consent and protection against accidental leaks, not isolation, and the docs say so.

**Rejected.** A minimal list of variables plus the ones a manifest declares, because it breaks tools that rely on the environment and still isolates nothing. Granting a key without a human, such as every key a tap declares when an agent adds the tap. Copying grants between machines with shared config, or treating the extension's command name as its identity.

## D031 Stop the whole process tree on cancel and keep finished work

Decided 2026-10-05. On Ctrl-C or SIGTERM, `ncly` stops every program it started and every descendant that it can still reach, orphans included on Linux, keeps every finished output, and exits 130 or 143 with the data that still applies, as [cli-spec.md](cli-spec.md#programs-that-ncly-runs) defines. Run inspection distinguishes verified finished work from uncertain effects. Neither interruption code authorizes an automatic retry.

**Why.** The transcript CLI starts its children in their own process groups, so a signal to its parent alone leaves `yt-dlp`, `ffmpeg`, or a harness running. A harness that hits its time limit sends SIGTERM. A finished transcript may already be billed, so cleanup must never delete it.

`ncly` finds descendants through their parent process, because a signal to a process group misses the children that the transcript CLI starts in their own sessions. On Linux, `ncly` registers as a child subreaper, so an orphaned descendant stays reachable. macOS has no equivalent, so a descendant whose parent exited before the signal is the one exception there.

**Rejected.** Signaling only the direct child. Relying on the Python program alone to stop its children, because the SIGKILL after 15 seconds leaves them running. Deleting partial results on interrupt. Assuming a stopped process proves that its last external request had no effect.

## D032 Keep small run records and resume explicitly

Decided 2026-10-05. The [operation contract](cli-spec.md#operations) covers preparation, execution, inspection, and explicit resume. Domains own their steps and recovery rules. The program runner owns subprocess lifecycles. Records preserve the evidence needed to reuse finished work, with references to outputs rather than copies of their contents. M00 fixed the contract. `ncly agent run` in M10 and transcript from M13 implement it, and core exposes the [run commands](cli-spec.md#ncly-run).

**Why.** A new agent session must be able to tell whether Deepgram finished before a summary failed. Writing the intended non-repeatable effect before starting it leaves evidence even if the process dies before saving the result. Resume validates inputs, relevant configuration, and saved outputs, and keeps the profile that the run recorded instead of reading it again. An uncertain paid request needs reconciliation or a human decision, because a missing response does not prove that nothing happened.

**Rejected.** Restarting every invocation from scratch. Logs as the only record of completed work. Recording keys or duplicating all transcript content in history. A daemon, scheduler, or general workflow engine before a command needs one.

## D033 Describe a command once for humans and agents

Decided 2026-10-05. One [command declaration](cli-spec.md#command-descriptions) supplies help, completion, targeted JSON discovery, and doctor prerequisites. It describes inputs, results, effects, and supported capabilities. Adapters and extensions declare only guarantees they can enforce.

**Why.** An agent needs the relevant command's contract without loading every domain or probing by trial and error. Required flags, flag groups, and typed result schemas make that contract sufficient to build and parse a call. Shared declarations keep help, preflight checks, and execution from making different promises. Discovery describes possible effects and requirements; shared preparation selects those of the invocation. M00's declaration grows with its first consumers, rather than pretending every field already exists. An extension's manifest holds the same declarations.

**Rejected.** Separate manually maintained command catalogs for agents. Dumping the whole command tree for every lookup. Treating an unknown capability as supported.

## D034 Use the same preparation for simulation and execution

Decided 2026-10-05. `--dry-run` and execution use the same preparation of inputs, profiles, destinations, and effects. A simulation distinguishes verified conditions from checks deferred until execution. The [dry-run contract](cli-spec.md#global-flags) allows only documented, bounded cache preparation; it reads no key and creates no durable run record. Execution revalidates conditions that may have changed.

**Why.** A separately written simulation can approve a destination that execution resolves differently. A dry run also cannot prove that a key works or reserve files against another process. Reporting these limits gives agents enough evidence to proceed without overstating what was checked.

**Rejected.** Parallel implementations of planning and execution. Calling a simulation successful while silently skipping a required check. Treating a prior dry run as permission to overwrite a changed destination.

## D035 Track ownership when synchronizing files

Decided 2026-10-05. Tap and docs synchronization track the files they manage and detect user changes before replacing or removing them. A sync stages and validates each managed destination, then publishes that destination atomically under its lock. Unrelated destinations are not one transaction. [M12](../milestones/M12-taps.md) and [M21](../milestones/M21-docs-hub.md) define their publication boundaries and recovery.

**Why.** A repeated sync must remove stale generated files without deleting unrelated files or edits. Readers need a complete result, and an interrupted update must leave enough evidence to recover. Explicit ownership makes those decisions possible.

**Rejected.** Replacing an entire destination without knowing who owns its files. Preserving every stale output forever. Publishing files one by one while readers can see an incomplete update.

## D036 Record each run in one JSON file that its executor locks

Decided 2026-10-07. Each run has one JSON record in the state directory, written atomically and flushed before an effect is acknowledged. Its lock reports current ownership separately from saved status. A source consumer holds the source lock while executing too. `ncly run view` returns saved step evidence and recovery hints. `ncly run list` filters candidates before limiting them and bounds its item previews. [cli-spec.md](cli-spec.md#record-format) holds the format.

**Why.** JSON matches the answer format, so inspection needs no second storage model. Atomic rename protects readers from incomplete JSON; syncing files and directory metadata protects the acknowledged intent across a crash within the platform's persistence guarantees. The operating system releases a lock when its owner dies, without a heartbeat. A pending step may have a failed preparation, so its saved problem retains the cause and hint. Filtering and short previews let a new session find relevant work without dumping unrelated batches. Each extension proves its own evidence instead of core promising its complete format.

**Rejected.** SQLite, because one file per run needs no migration and an agent reads it with `jq`. A process ID in the record, because the operating system reuses process IDs. A heartbeat refreshed by a background worker, because Nicely has none.

## D037 Summarize a saved transcript as a new operation

Decided 2026-10-07. `ncly transcript summary run <run-id>` summarizes the verified transcripts of a saved run as a new recorded operation. `ncly run resume` never repeats a `summarize` step whose outcome is `unknown`, and the silent summary retries of the transcript CLI are gone. [cli-spec.md](../../extensions/transcript/spec.md#ncly-transcript-summary-run) holds the rules.

**Why.** A summary can bill, so a lost answer may already be charged. Resume never repeats that unknown request. A new operation copies verified transcript references, keeps its own prompt and profile, and writes a file named with its run ID. It leaves the source record valid and supports safe resume of its own never-started summary. Holding the source lock prevents a new summary racing its source executor. The caller explicitly chooses any new bill.

**Rejected.** A resume flag that overrides an unknown step, because `--force` never makes an unknown effect safe. Treating every summary as free, because some profiles bill per call. Rerunning the whole transcript, which bills Deepgram again.

## D038 Exchange acknowledged JSON Lines with the Python program

Decided 2026-10-07. Go and Python exchange versioned JSON Lines. An initial plan fixes the item set for both dry run and execution. Python announces an intent before user-side output or Deepgram upload. Go records the selected destination before reserving it, completes the persistence barrier, then acknowledges the final folder. [cli-spec.md](../../extensions/transcript/spec.md#protocol-with-the-python-program) holds the messages.

**Why.** One preparation avoids selecting a different Zoom meeting during execution. Recording a destination after creating it loses ownership on interruption, so Go records it first. No paid call starts before the durable intent; a lost acknowledgement stops Python before upload. Persisted verified artifacts can establish completion after a lost final message, without another request. This is the temporary Python boundary, not a protocol imposed on future extensions.

**Rejected.** Arguments plus one final JSON object, because Go could not record the intent before the upload. A separate file or socket, because stdin and stdout already connect the two processes.

## D039 Let yt-dlp update without a release

Decided 2026-10-07. The Python program names a minimum `yt-dlp` version, and `ncly` runs it with `--upgrade-package yt-dlp`, so each run uses the newest release. The other dependencies stay pinned. The Arc cookie adapter checks the `yt-dlp` function that it changes instead of an exact version, and a missing function falls back to anonymous access with a warning.

**Why.** YouTube breaks `yt-dlp` every few weeks, and a fix must reach users without an `ncly` release. Pascal keeps Arc for YouTube, so the adapter stays.

**Rejected.** An exact pin, which needs an `ncly` release for every YouTube change. The `yt-dlp` on `PATH`, because the Arc adapter changes `yt-dlp` inside the same process.

## D040 Serve extensions through a wire contract and a public Go SDK

Decided 2026-10-09. The interface between core and an extension is a contract on the wire: the manifest, the environment, the JSON answer, the exit code, and the record files. Core injects the keys that a human granted. An extension writes its run records in the shared format, and `ncly run` reads them. An extension uses another one through the CLI, as transcript calls `ncly agent run --json`. The public `sdk/` packages are the Go implementation of this contract, and core uses them too, so each shared layer has one implementation.

**Why.** Git, `gh`, `kubectl`, and cargo plugins work this way, and agents already use the same interface. A Python or shell extension can follow the contract from the spec alone. No protocol between processes needs a version before an extension needs it.

**Rejected.** Services that the host serves over an RPC protocol on stdio, as Terraform plugins do, because the protocol would need a design and a version before the first extension. Extensions importing core's `internal/` packages, which ties them to core's code.

## D041 Keep first-party extensions in the Nicely repository

Decided 2026-10-09. First-party extensions live in `extensions/<name>/`, in the same Go module as core. Each folder holds what its extension owns: the manifest, `spec.md`, `SKILL.md`, catalogs, scenarios, and code. A lint rule keeps `sdk/` and `extensions/` away from `internal/`. The repository is also the official tap of these extensions.

**Why.** An agent changes the SDK and the extension that uses it in one pull request, with one checkout and one `just next`, while the contract is still young. The lint rule keeps the boundary that separate repositories would give.

**Rejected.** One repository per extension, as `gh` extensions do, because every SDK change would then take several pull requests.

## D042 Plan work as milestones of small cards

Decided 2026-10-09. A milestone is one capability that a sentence can demonstrate, with at most five agent cards. A card is one pull request, with its dependencies, its owner, the spec sections to read, and the scenarios that prove it. A decision of Pascal's that later cards wait on is a card with the owner Pascal; a step of his that nothing waits on, such as a check on his Mac, sits under **After this milestone** and blocks nothing. No checkbox exists outside a card, so a milestone is done when all its cards are. A planned milestone with open questions becomes ready in one pull request that settles them in the spec and finishes its cards; without open questions, its first card starts it. Milestones go in order. The order puts first what later milestones build on, keeps the milestones of one extension together, and puts Pascal's own tools before the everyday domains that make a first impression. Only a milestone without a done card changes its number, in a pull request that only changes the plan. `just next` prints the one card to do, and `just status` shows the progress. `just test` checks this format.

**Why.** Large milestones with tasks, acceptance boxes, a done-when list, and a to-define list let agents pick one task and leave the rest, and Pascal could no longer follow the project. One computed pointer leaves nothing to interpret, and every piece of work has an owner and a proof.

**Rejected.** The first milestone format, described above. Cards by convention without a check. GitHub issues and milestones as the queue, because agents read the repository first and the plan should travel with the code.

## D043 Number milestones by the release that ships them, and release separately

Decided 2026-10-09. Milestone `M<n>` ships in `v0.<n>.0`, and M00 in v0.0.1. A fix between two milestones takes the next patch number, such as v0.4.1. Releases happen when Pascal asks and are never a card, so a release never blocks a milestone.

**Why.** A milestone that waited for its release, as M00 did for v0.0.1, blocked the next one until Pascal acted, and agents worked around it with an exception. Minor numbers per milestone leave patch numbers free for fixes.

**Rejected.** One release per milestone as a gate. Numbering milestones `0.1.<n>`, because a fix between milestones would take the next milestone's number.
