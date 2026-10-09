# Nicely guide

This file says what Nicely is, which rules every change follows, and where the details live.

## What Nicely is

Nicely is a curated, multilingual CLI toolbox that humans and AI agents operate equally well. Its command is `ncly`.

Pascal builds it for his own work first and publishes it for anyone. The repository is public, MIT licensed, and sends no telemetry.

Nicely runs on macOS and Linux. Omarchy, which is based on Arch, is the reference Linux. Windows is not a target.

## Core and extensions

Nicely is a small core, and everything it does for a user is an extension, in the spirit of Pi v1: a minimal host that is easy to extend and to compose.

- **Core** is the host: dispatch, help, completion, `describe`, `doctor`, `auth` with grants, `run`, and `tap`. It knows no domain.
- **Extensions** add domains, such as `transcript`. All of them share one manifest format and one contract, and they run with the rights of the user, because Nicely is not a sandbox. A **bundled** extension is compiled into `ncly`. Only what an agent needs to operate Nicely is bundled: `skill`, with the `nicely` skill. An **external** extension is an executable named `ncly-<domain>` beside its manifest. The first-party ones, such as `agent` and `transcript`, live in `extensions/` of this repository, and Pascal's personal tools, such as fleet sync, live in his private tap.

The parts form a stack of layers:

```
4  Extensions  bundled: skill, with the nicely skill
               external: agent, transcript, markdown, video, image, docs; fleet in a private tap
3  Core        dispatch, help, completion, describe, doctor, auth and grants, run, tap
2  Services    config, i18n, program runner, run records
1  Contract    answer envelope, code to exit, modes, declaration and manifest
0  Platform    durable writes, locks, keychain, trash, process tree
```

- A layer uses only the layers below it.
- Each part has one owner: one package and one spec section.
- Layers 0 to 2 that extensions need form the public SDK in `sdk/`, which core uses too. The keychain stays in core, and keys reach an extension only through grants.
- An extension reaches core through its environment and the CLI, and another extension only through the CLI, the same interface that agents use.
- Adding a domain never edits core.

Core says where a feature lives, not when it ships. The milestones decide timing.

## Agents in both directions

Agents operate `ncly`. Every command runs without a terminal, answers in JSON on request, and fails with an error code and a fix. [contract.md](contract.md#output) lists the few exceptions, such as `--help`. Any agent with a shell can use Nicely, so `ncly` never needs to know which agent calls it.

`ncly` also runs agents, through the `agent` extension. Profiles in the config name a harness, a model, and an effort. Other extensions use agents through `ncly agent run`, as transcript does for its summary.

Agents set `NCLY_NO_INPUT=1` and `NCLY_JSON=1`, read the bundled `nicely` skill with `ncly skill view nicely`, discover one command, check readiness, prepare its invocation with dry run, then execute it. Each extension ships a skill that teaches agents its own commands.

| Source of truth | What it establishes |
|---|---|
| [Command declaration](core-spec.md#ncly-describe) | Installed capabilities, input constraints, result types, and supported modes |
| [Doctor](core-spec.md#ncly-doctor) | Local readiness, with free network checks only on request |
| [Dry run](contract.md#operations) | The invocation's plan and checks still pending before execution |
| [Run inspection](core-spec.md#ncly-run) | Saved step evidence and `active`, which reports whether a process holds the run lock |
| Artifacts | The full output files, referenced by path and verified fingerprints |
| [Explicit resume](contract.md#inspection-and-explicit-resume) | The saved outputs that current verification permits the extension to reuse |

Commands with paid work or reusable partial results keep records. An agent inspects a recorded failure before choosing a recovery action. Overall exit 75 alone permits an automatic repeat of the same invocation. A saved status describes the operation's last recorded outcome; the execution lock establishes whether a process owns the run now.

Records and verified outputs retain useful work across sessions. Explicit resume reuses that work and runs only safe remaining steps. The SDK's record support owns records, locks, and result publication; the extension decides what to reuse or execute. Nicely needs no memory daemon, workflow interpreter, or generic engine for future domains. A record does not keep a process alive, and Nicely has no background worker or scheduler.

## Principles

1. **Domain first.** Commands read `ncly <domain> [<resource>] <verb>`, such as `ncly video convert`, with one verb per action across domains. `ncly <domain> --help` lists everything a domain does. A few commands stand alone, such as `ncly doctor`. A script that joins Nicely is reshaped into this grammar, whatever its interface was before.
2. **One implementation, two faces.** In interactive mode, a missing value opens a short form. In non-interactive mode, the same command fails with exit code 2 and a hint that shows the full command. Both faces produce the same result.
3. **A stable agent contract.** Machine identifiers stay in English. [Compatibility](contract.md#compatibility) protects JSON types, units, values, and meanings as well as key names, and defines how versions evolve
4. **Safe by default.** Every writing command shares preparation with `--dry-run`. Overwriting or deleting needs `--force` or a yes in the terminal. Deleted files go to the OS trash. [Retry safety](contract.md#retry-safety) covers all effects, including free writes. Outside text reaches an agent only with its tools turned off, and a check after execution is never described as write prevention
5. **Keys stay with the human.** Keys live in the OS keychain or in environment variables, never in flags, config files, the repository, or the binary. An agent that meets a missing key relays the hint, and the human runs `ncly auth login`. A key reaches an extension only after a human grants it.
6. **Check everything, install nothing silently.** `ncly doctor` checks tools and keys. Nicely installs a prerequisite only after a human says yes.
7. **Every word Nicely shows can be translated.** Core and each extension keep their text in message catalogs, never in code. English ships first, and Canadian French follows in M22. A pseudo-locale exposes untranslated text from M00. Extensions receive the language in `NCLY_LANG` and translate their own text.
8. **Core and the first-party extensions are written in Go.** An extension may embed a Python program while it moves to Go. That program talks to Go in JSON only, and Go renders and translates everything a human sees.
9. **Same paths on every machine.** Config, data, state, and cache follow XDG on macOS and Linux, under `nicely`. The shared config describes the setup the user wants and can travel between machines. The local config holds what belongs to one machine.
10. **Import Go, wrap the rest, credit both.** Nicely imports Go libraries and wraps other tools, such as `ffmpeg`. The help of a wrapping command names the tool, and every release ships the license notices of its dependencies.
11. **Build what your card needs, and decide early what is costly to change.** Write a rule before the milestone that codes it when a later milestone would otherwise force a rewrite.
12. **Compose through the CLI.** An extension uses another extension through `ncly <domain> ... --json`, the interface that agents use, never through its code. Every new extension therefore adds a command that agents and other extensions can use at once.

## Domains and commands

| Command | What it does | Kind | Milestone |
|---|---|---|---|
| `completion` | Prints shell completion scripts | core | M00 |
| `describe` | Describes installed commands and their capabilities | core | M04 |
| `doctor` | Checks tools, keys, extensions, and components | core | M05 |
| `auth` | Stores keys and grants them to extensions | core | M06, grants in M08 |
| `run` | Lists, inspects, and explicitly resumes recorded operations | core | M10, resume in M15 |
| `tap` | Adds and syncs repositories of skills and extensions | core | M12 |
| `skill` | Lists and views skills for agents, including the `nicely` skill | bundled | M03 |
| `agent` | Runs a task through a harness and a profile | external | M09 to M11 |
| `transcript` | Transcribes YouTube and Zoom audio, then summarizes it | external | M13 to M16 |
| `markdown`, `video`, `image` | Everyday tasks that need no configuration | external | M18 to M20, after the inventory of M17 |
| `docs` | Gathers project docs into one private site | external | M21 |

Core reserves the names of the core and bundled rows. `transcript` is not the first thing a visitor should try, because it needs a Deepgram key. First impressions come from the everyday domains.

## Terms

Use one word per concept in code, docs, help, and commit messages. Add a term here before you use a new one.

- **Agent**: an AI program that works through a shell. The `agent` extension launches one.
- **Agent contract**: everything a script or an agent parses, defined in [contract.md](contract.md) and the extension specs.
- **Artifact**: a result produced by an operation, referenced by its path and the evidence needed to verify it
- **Bundled extension**: an extension compiled into `ncly`, with the same manifest and contract as an external one. Only `skill` is bundled.
- **Card**: one pull request of a milestone, with its owner, its dependencies, what to read, and what proves it.
- **Core**: the host that every install gets: dispatch, help, completion, `describe`, `doctor`, `auth`, `run`, and `tap`.
- **Domain**: a top-level noun that groups actions, such as `video`.
- **Extension**: a domain added to core through a manifest, bundled or external.
- **External extension**: an executable named `ncly-<domain>` with a compatible manifest beside it.
- **First-party extension**: an extension in `extensions/` of this repository, such as `agent` or `transcript`.
- **Grant**: the permission, given by a human in a terminal, for one extension to receive one key.
- **Harness**: the program an agent runs in: Claude Code, Codex, Pi, OpenCode, or Grok.
- **Interactive mode**: the mode in which `ncly` may ask questions. [contract.md](contract.md#modes) defines when it applies. Every other case is non-interactive mode.
- **Local config**: `config.local.toml`, which holds what belongs to one machine and overrides the shared config.
- **Manifest**: the file that declares an extension's commands, effects, modes, and requirements, read without starting the extension.
- **Milestone**: a numbered stage of work in `docs/milestones/`, one capability that a sentence can demonstrate. Its status is `planned`, `ready`, `active`, or `done`. M99 is the parking lot, and its status stays `open`.
- **Operation**: the work prepared and performed by a command. The domain owns its steps
- **Profile**: a named combination of harness, provider, model, and effort, stored in the config.
- **Pseudo-locale**: a generated test language that marks every catalog string, so text outside the catalog stands out.
- **Python program**: Python code that an extension embeds while it moves to Go, such as transcript's.
- **Report command**: a command whose result is a report, such as `ncly doctor`. Its report goes to stdout even when a check fails.
- **Resume**: an explicit request to continue a recorded run after validating its inputs, artifacts, and remaining steps
- **Run record**: the durable evidence of one operation, identified by `run_id`, including its step outcomes and artifact references
- **SDK**: the public Go packages in `sdk/` that implement the contract and the shared services, for core and for every Go extension.
- **Shared config**: `config.toml`, the setup the user wants, which can travel between machines.
- **Skill**: a folder with a `SKILL.md` that teaches an agent a task.
- **Tap**: a Git repository that holds skills and extensions, like a Pi package. The Homebrew tap that ships `ncly` is a different thing, always called the Homebrew tap.

## Where the details live

- [contract.md](contract.md) holds the agent contract and the commands of core. Each extension's spec lives in `extensions/<name>/spec.md`, such as [transcript's](../../extensions/transcript/spec.md). Read the sections that your card links before changing a command.
- [decision-records.md](decision-records.md) explains why each rule exists. Read it before you propose to change a rule.
- [dev-preferences.md](dev-preferences.md) records how Pascal wants changes made, such as deciding the look with a published mockup.
- The milestone files in `docs/milestones/` list the work as cards, in order. `just next` prints the card to do, `just status` shows the progress, and [M99](../milestones/M99-parking-lot.md) parks every other idea.

## Boundaries to keep through the milestones

| Owner | Responsibility |
|---|---|
| Command declaration and manifest | Arguments, flags, result contract, effects, prerequisites, and capabilities used by discovery, help, completion, and validation |
| Extension host | Discovery, manifests, dispatch, answer checks, and grants |
| Extension | Its domain's preparation, business steps, result verification, the evidence that permits resume, its catalog, and its skill |
| Record support | Run IDs, durable records, execution locks, and publication of the extension's result |
| Program runner | Child environment, timeouts, signals, and process cleanup |
| Platform support | Atomic file replacement, locks, keychain, and trash |

M00 fixed these responsibilities and the contract. Each implementation arrives with its first consumer: the host and the program runner in M01, the SDK in M02, and records with `ncly agent run` in M10.
