# Nicely

A curated, multilingual CLI toolbox that humans and AI agents operate equally well. The command is `ncly`.

```bash
ncly --help
ncly --version
ncly completion bash
```

## Status

Nicely is in early development and has no release yet. [The guide](docs/north-star/guide.md) explains what it is and why, and [the milestones](docs/milestones/) list the work in order.

The current build implements help, the version, and shell completion. It also establishes the JSON error contract, configuration, and English message catalog. The commands below are planned work.

## Planned commands

Nicely is a small core that every domain extends, in the spirit of Pi. Core brings `describe`, `doctor`, `auth`, `run`, and `tap`. It finds extensions named `ncly-<domain>`, and it bundles one extension, `skill`, which agents use to learn Nicely. Everything else is an extension, including the first ones: `agent`, which runs a task through Claude Code, Pi, and other harnesses, and `transcript`, which transcribes YouTube and Zoom audio and summarizes it through `agent`. Taps arrive between the two, so `transcript` installs from the official tap. Everyday domains such as `markdown` and `video`, a private docs hub, and Canadian French follow. The [guide's command table](docs/north-star/guide.md#domains-and-commands) records their scope.

Interactive forms will ask for missing values. Scripts and agents set `NCLY_NO_INPUT=1` to prevent questions, including in a pseudo-terminal, and use `--json` or `NCLY_JSON=1` for machine output. Commands group actions by domain, such as `ncly transcript run youtube`.

## Planned agent instructions

Once the `skill` extension ships, one line in your agent instructions will teach an agent to use Nicely:

```text
Before using ncly, run ncly skill view nicely
```

The bundled `nicely` skill covers discovery, readiness, dry run, retry safety, inspection, and explicit resume, and each extension ships a skill for its own commands. The [agent contract](docs/north-star/cli-spec.md) defines the fields and recovery rules. Saved records and verified output files retain work across sessions.

## Install

There is no release yet. The first release installs with:

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

## Privacy

Nicely sends no telemetry. It contacts a service only when a command you run needs it, such as Deepgram for `ncly transcript`.

## Contributing

Issues and pull requests are welcome. [AGENTS.md](AGENTS.md) explains how changes are made.

## License

MIT. See [LICENSE](LICENSE).
