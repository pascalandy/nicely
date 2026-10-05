# Nicely guide

This file says what Nicely is, which rules every change follows, and where the details live.

## What Nicely is

Nicely is a curated, multilingual CLI toolbox that humans and AI agents operate equally well. Its command is `ncly`.

Pascal builds it for his own work first and publishes it for anyone. The repository is public, MIT licensed, and sends no telemetry.

Nicely runs on macOS and Linux. Omarchy, which is based on Arch, is the reference Linux. Windows is not a target.

## Core and extensions

Nicely has two layers.

- **Core** holds the built-in commands everyone gets. A feature belongs in core when a stranger would use it.
- **Extensions** are executables named `ncly-<domain>` that add a domain. They receive the global flags as `NCLY_*` variables and follow the agent contract. Pascal's personal tools, such as fleet sync, are extensions shipped through his private tap.

Core says where a feature lives, not when it ships. The milestones decide timing.

## Agents in both directions

Agents operate `ncly`. Every command runs without a terminal, answers in JSON on request, and fails with an error code and a fix. [cli-spec.md](cli-spec.md#output-m0) lists the few exceptions, such as `--help`. Any agent with a shell can use Nicely, so `ncly` never needs to know which agent calls it.

`ncly` also runs agents. Profiles in the config name a harness, a model, and an effort from M1. From M2, `ncly agent` sends a task to a harness through a profile, and harness logic lives in that one place.

## Principles

1. **Domain first.** Commands read `ncly <domain> [<resource>] <verb>`, such as `ncly video convert`, with one verb per action across domains. `ncly <domain> --help` lists everything a domain does. A few commands stand alone, such as `ncly doctor`. A script that joins Nicely is reshaped into this grammar, whatever its interface was before.
2. **One implementation, two faces.** In interactive mode, a missing value opens a short form. In non-interactive mode, the same command fails with exit code 2 and a hint that shows the full command. Both faces produce the same result.
3. **A stable agent contract.** Command names, flags, JSON keys, error codes, and exit codes stay in English. JSON keys are only added. Any other change to the contract needs a new entry in [decision-records.md](decision-records.md). [cli-spec.md](cli-spec.md) defines the contract.
4. **Safe by default.** Every command that writes accepts `--dry-run`. Overwriting or deleting needs `--force` or a yes in the terminal. Deleted files go to the OS trash. Text from outside Nicely, such as a transcript, reaches an agent only with the agent's tools turned off.
5. **Keys stay with the human.** Keys live in the OS keychain or in environment variables, never in flags, config files, the repository, or the binary. An agent that meets a missing key relays the hint, and the human runs `ncly auth login`.
6. **Check everything, install nothing silently.** `ncly doctor` checks tools and keys. Nicely installs a prerequisite only after a human says yes.
7. **Every word core shows can be translated.** Core text lives in message catalogs, never in code. English ships first, and Canadian French follows in M6. A pseudo-locale exposes untranslated text from M0. Extensions receive the language in `NCLY_LANG` and translate their own text.
8. **Core is written in Go.** A Python program may sit in core while it moves to Go. It talks to Go in JSON only, and Go renders and translates everything a human sees.
9. **Same paths on every machine.** Config, data, state, and cache follow XDG on macOS and Linux, under `nicely`. The shared config describes the setup the user wants and can travel between machines. The local config holds what belongs to one machine.
10. **Import Go, wrap the rest, credit both.** Nicely imports Go libraries and wraps other tools, such as `ffmpeg`. The help of a wrapping command names the tool, and every release ships the license notices of its dependencies.
11. **Build what the current milestone needs, and decide early what is costly to change.** Write a rule before the milestone that codes it when a later milestone would otherwise force a rewrite.

## Domains and commands

| Command | What it does | Layer | Milestone |
|---|---|---|---|
| `completion` | Prints shell completion scripts | core | M0 |
| `doctor` | Checks tools, keys, and components | core | M1 |
| `auth` | Stores and checks API keys | core | M1 |
| `skill` | Lists and views skills for agents | core | M1 |
| `transcript` | Transcribes YouTube and Zoom audio, then summarizes it | core | M1 |
| `agent` | Runs a task through a harness and a profile | core | M2 |
| `tap` | Adds and syncs repositories of skills and extensions | core | M3 |
| `markdown`, `video`, `image` | Everyday tasks that need no configuration | core | M4, after the inventory |
| `docs` | Gathers project docs into one private site | core | M5 |

`transcript` belongs in core but is not the first thing a visitor should try, because it needs a Deepgram key. First impressions come from the domains of M4.

## Terms

Use one word per concept in code, docs, help, and commit messages. Add a term here before you use a new one.

- **Agent**: an AI program that works through a shell. `ncly agent` launches one.
- **Agent contract**: everything a script or an agent parses, defined in [cli-spec.md](cli-spec.md).
- **Core**: the built-in commands everyone gets.
- **Day one**: M1, the first useful release.
- **Domain**: a top-level noun that groups actions, such as `video`.
- **Extension**: an executable named `ncly-<domain>`, outside core.
- **Harness**: the program an agent runs in: Claude Code, Codex, Pi, OpenCode, or Grok.
- **Interactive mode**: the mode in which `ncly` may ask questions. [cli-spec.md](cli-spec.md#modes-m0) defines when it applies. Every other case is non-interactive mode.
- **Local config**: `config.local.toml`, which holds what belongs to one machine and overrides the shared config.
- **Milestone**: a numbered stage of work in `docs/milestones/`. Its status is `planned`, `active`, or `done`. M99 is the parking lot, and its status stays `open`.
- **Profile**: a named combination of harness, provider, model, and effort, stored in the config.
- **Pseudo-locale**: a generated test language that marks every catalog string, so text outside the catalog stands out.
- **Python program**: Python code that core embeds while it moves to Go, such as the transcript CLI.
- **Report command**: a command whose result is a report, such as `ncly doctor`. Its report goes to stdout even when a check fails.
- **Shared config**: `config.toml`, the setup the user wants, which can travel between machines.
- **Skill**: a folder with a `SKILL.md` that teaches an agent a task.
- **Tap**: a Git repository that holds skills and extensions. The Homebrew tap that ships `ncly` is a different thing, always called the Homebrew tap.
- **Task**: one pull request that ticks at least one box of a milestone.

## Where the details live

- [cli-spec.md](cli-spec.md) holds the agent contract and every specified command. Read the sections you touch before changing a command.
- [decision-records.md](decision-records.md) explains why each rule exists. Read it before you propose to change a rule.
- The milestone files in `docs/milestones/` list the work in order, and [M99](../milestones/M99-parking-lot.md) parks every other idea.
