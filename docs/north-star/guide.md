# Nicely guide

Read this file at the start of every session. It says what Nicely is, which rules every change follows, and where the details live.

## What Nicely is

Nicely is a curated, multilingual CLI toolbox that humans and AI agents operate equally well. Its command is `ncly`.

Pascal builds it for his own work first and publishes it for anyone. The repository is public, MIT licensed, and sends no telemetry.

Nicely runs on macOS and Linux. Omarchy, which is based on Arch, is the reference Linux. Windows is not a target.

## Core and extensions

Nicely has two layers.

- **Core** holds the built-in commands everyone gets. A feature belongs in core when a stranger would use it.
- **Extensions** are executables named `ncly-<domain>` that add a domain. Pascal's personal tools, such as fleet sync, are extensions shipped through his private tap.

Core says where a feature lives, not when it ships. The milestones decide timing.

## Agents in both directions

Agents operate `ncly`. Every command runs without a terminal, answers in JSON on request, and fails with an error code and a fix. Any agent with a shell can use Nicely, so `ncly` never needs to know which agent calls it.

`ncly` also runs agents. From M2, `ncly agent` sends a task to a harness through a named profile. Harness logic lives in that one place, and every other command uses it.

## Principles

1. **Domain first.** Commands read `ncly <domain> <action>`, such as `ncly video convert`, and `ncly <domain> --help` lists everything a domain does. A few commands stand alone, such as `ncly doctor`. A script that joins Nicely is reshaped into domains, whatever its interface was before.
2. **One implementation, two faces.** In a terminal, a missing value opens a short form. Without a terminal, the same command fails with exit code 2 and a hint that shows the full command. Both faces produce the same result.
3. **A stable agent contract.** Command names, flags, JSON keys, error codes, and exit codes stay in English. They change only through a new entry in [decisions.md](decisions.md). [cli-spec.md](cli-spec.md) defines them.
4. **Safe by default.** Every command that writes accepts `--dry-run`. Overwriting or deleting needs `--force` or a yes in the terminal. Deleted files go to the OS trash.
5. **Keys stay with the human.** Keys live in the OS keychain or in environment variables, never in flags, config files, the repository, or the binary. An agent that meets a missing key relays the hint, and the human runs `ncly auth login`.
6. **Check everything, install nothing silently.** `ncly doctor` checks tools and keys. Nicely installs a prerequisite only after a human says yes.
7. **Every word a human reads can be translated.** User-facing text lives in message catalogs, never in code. English ships first. Canadian French follows in M6.
8. **Core is written in Go.** A Python component may sit in core while it moves in. It talks to Go in JSON only, and Go renders and translates everything a human sees.
9. **Same paths on every machine.** Config, data, and cache follow XDG on macOS and Linux, under `~/.config/nicely`, `~/.local/share/nicely`, and `~/.cache/nicely`.
10. **Import Go, wrap the rest, credit both.** Nicely imports Go libraries and wraps other tools, such as `ffmpeg`. The help of a wrapping command names the tool, and every release ships the license notices of its dependencies.
11. **Build what the current milestone needs.** Any other idea goes to the [parking lot](../milestones/M99-parking-lot.md) as one line.

## Domains and commands

| Command | What it does | Layer | Milestone |
|---|---|---|---|
| `completion` | Prints shell completion scripts | core | M0 |
| `doctor` | Checks tools, keys, and components | core | M1 |
| `auth` | Stores and checks API keys | core | M1 |
| `skill` | Lists and shows skills for agents | core | M1 |
| `transcript` | Transcribes YouTube and Zoom audio, then summarizes it | core | M1 |
| `agent` | Runs a task through a harness and a profile | core | M2 |
| `tap` | Adds and syncs repositories of skills and extensions | core | M3 |
| `fleet` | Syncs Pascal's machines | extension | M3 |
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
- **Interactive mode**: stdin and stdout are a terminal and `--no-input` is absent. Every other case is non-interactive mode.
- **Milestone**: a numbered stage of work in `docs/milestones/`. M99 is the parking lot.
- **Profile**: a named combination of harness, model, and effort.
- **Skill**: a folder with a `SKILL.md` that teaches an agent a task.
- **Tap**: a Git repository that holds skills and extensions.

## Where the details live

- [cli-spec.md](cli-spec.md) holds the agent contract and every specified command. Read the sections you touch before changing a command.
- [decisions.md](decisions.md) explains why each rule exists. Read it before you propose to change a rule.
- The milestone files in `docs/milestones/` list the work in order. The current milestone is the lowest-numbered file whose status is not `done`.
