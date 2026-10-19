# Nicely

A curated, multilingual CLI toolbox that humans and AI agents operate equally well. The command is `ncly`.

```bash
ncly --help
ncly --version
ncly skill view nicely
ncly describe --json
ncly skill list
ncly completion bash
```

## Status

Nicely is in early development. [The vision](docs/north-star/vision.md) explains what it is and why, and [the milestones](docs/milestones/) list the work in order.

The current build brings:

- Help, the version, shell completion, and `ncly describe`, which gives the contract of each installed command
- The extensions named `ncly-<domain>`, which it finds, checks, and stops, and the bundled `skill` extension
- `ncly doctor`, which checks the machine, and `ncly auth`, which keeps the keys and their grants
- The `agent` extension, which runs one trusted task through Codex and records it
- The `transcript` extension, which transcribes one YouTube video with Deepgram and records the run
- `ncly run list` and `ncly run view`, which read the records of the runs

[The current build](docs/status.md) describes each command. The commands below are planned work.

## Planned commands

Nicely is a small core that every domain extends, in the spirit of Pi. Core brings `describe`, `doctor`, `auth`, and `run`. It bundles `skill`, which agents use to learn Nicely. Other domains use external extensions named `ncly-<domain>`.

The first `agent` command runs one trusted task through Codex. The first `transcript` command transcribes one explicitly selected YouTube video. A checkout installs the extensions with `just install`. Each capability starts with its smallest useful path; additional harnesses, transcript summaries, and taps remain in [M99](docs/milestones/M99-parking-lot.md).

The plan also keeps an everyday task chosen from real use, a local view of selected project docs, and Canadian French. Markdown is the first everyday candidate. The [domain table](docs/north-star/architecture.md#domains-and-commands) records their scope.

In a terminal, a form asks a human for each missing value. Scripts and agents set `NCLY_NO_INPUT=1` to prevent questions, including in a pseudo-terminal, and use `--json` or `NCLY_JSON=1` for machine output. Commands group actions by domain, such as `ncly transcript run youtube`.

## Agent instructions

Add one line to your agent instructions, such as `AGENTS.md` or `CLAUDE.md`, to teach an agent to use Nicely:

```text
Before using ncly, run ncly skill view nicely
```

The bundled `nicely` skill teaches only what this release can do: find a command with `ncly describe`, dry-run it when it writes or bills, check the machine with `ncly doctor` when a problem has no explanation, and act on the stable `code` of each error and on the exit code, which says who acts. When the hint names a command that runs as is, `next` holds that command as words. Each extension ships a skill for its own commands, which `ncly skill list` lists. It teaches [run IDs](docs/north-star/contract/run-ids.md) that the caller chooses, the inspection of a run with `ncly run list` and `ncly run view`, and the verified replay of a completed call. An uncertain paid result never permits an automatic second charge. [Explicit resume](docs/milestones/M99-parking-lot/resume.md) waits for a workflow that needs it. The [agent contract](docs/north-star/contract.md) defines the fields and recovery rules.

## Install

Install with:

| System | Installation |
|---|---|
| macOS | `brew install pascalandy/tap/ncly` |
| Linux | Download the archive for your architecture from the [releases page](https://github.com/pascalandy/nicely/releases) |

Homebrew installs the completions for zsh, bash, and fish. On Linux, extract the archive into an empty folder, then put `ncly` and its completions where your shell finds them:

```bash
install -Dm755 ncly ~/.local/bin/ncly
install -Dm644 completions/ncly.bash ~/.local/share/bash-completion/completions/ncly
install -Dm644 completions/_ncly ~/.local/share/zsh/site-functions/_ncly
install -Dm644 completions/ncly.fish ~/.config/fish/completions/ncly.fish
```

These commands target the default folders: bash also needs the bash-completion package, and with custom XDG folders, use the folders that your shell reads instead. `~/.local/bin` must be on your `PATH`. For zsh, add `fpath=(~/.local/share/zsh/site-functions $fpath)` to `~/.zshrc` before `compinit` runs.

From a checkout, with Go and `just`, run `just install`. It builds `ncly` into `~/.local/bin/` and the first-party extensions, such as `agent` and `transcript`, into `~/.local/share/nicely/extensions/`, or into `$XDG_DATA_HOME/nicely/extensions/` when `XDG_DATA_HOME` holds an absolute path. The `agent` extension uses the login of an installed Codex, which `ncly doctor agent` checks. The `transcript` extension needs `uv` and Python 3.12 or later, which `ncly doctor transcript` checks.

## Privacy

Nicely sends no telemetry. It contacts a service only when a command you run needs it, such as Deepgram for `ncly transcript`.

## Contributing

Issues and pull requests are welcome. [AGENTS.md](AGENTS.md) explains how changes are made.

## License

MIT. See [LICENSE](LICENSE).
