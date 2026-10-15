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

The current build implements help, the version, shell completion, and `ncly describe`, which lists the installed commands, core's and each extension's, with their summaries, and gives one command's arguments, flags, effects, modes, and requirements, with each config value that it reads and where that value comes from, and the JSON Schema of its answers. It finds the extensions named `ncly-<domain>` in `~/.local/share/nicely/extensions/` and in `PATH`, shows their help and completion from their manifests, starts them with a call that it has checked against their declared arguments, flags, and flag groups, checks their JSON answer before passing it on, and stops them and their child processes on Ctrl-C. It bundles the `skill` extension, whose `ncly skill list` lists the skills in the folders that `[skill] paths` names in the config and the skill of each installed extension, and `ncly skill view <name>` prints one of them. Shell completion suggests the names of the skills after `ncly skill view`, and an extension's argument or flag can complete from one of its list commands. `ncly doctor` reports on stdout whether the config files parse, warns about each key that neither ncly nor an installed extension reads, lists each extension that it finds, with the one that hides it, without starting any, and checks the programs that each extension requires, with the install command of the machine's package manager for one that is missing or too old, then runs the status command of each extension that has one, such as `ncly skill status`, which checks the skill folders, names, and files. It also establishes the JSON error contract, configuration, and English message catalog. The commands below are planned work.

## Planned commands

Nicely is a small core that every domain extends, in the spirit of Pi. Core brings `describe`, `doctor`, `auth`, `run`, and `tap`. It finds extensions named `ncly-<domain>`, and it bundles one extension, `skill`, which agents use to learn Nicely. Everything else is an extension, including the first ones: `agent`, which runs a task through Claude Code, Pi, and other harnesses, and `transcript`, which transcribes YouTube and Zoom audio and summarizes it through `agent`. Taps arrive between the two, so `transcript` installs from the official tap. Everyday domains such as `markdown` and `video`, a private docs hub, and Canadian French follow. The [domain table](docs/north-star/architecture.md#domains-and-commands) records their scope.

Interactive forms will ask for missing values. Scripts and agents set `NCLY_NO_INPUT=1` to prevent questions, including in a pseudo-terminal, and use `--json` or `NCLY_JSON=1` for machine output. Commands group actions by domain, such as `ncly transcript run youtube`.

## Agent instructions

Add one line to your agent instructions, such as `AGENTS.md` or `CLAUDE.md`, to teach an agent to use Nicely:

```text
Before using ncly, run ncly skill view nicely
```

The bundled `nicely` skill teaches only what this release can do: find a command with `ncly describe`, dry-run it when it writes or bills, check the machine with `ncly doctor` when a problem has no explanation, and act on the stable `code` of each error and on the exit code, which says who acts. When the hint names a command that runs as is, `next` holds that command as words. Each extension ships a skill for its own commands, which `ncly skill list` lists. Later releases add [run IDs](docs/north-star/contract.md#run-ids), inspection, and explicit resume to the skill, so that a lost answer costs nothing. The [agent contract](docs/north-star/contract.md) defines the fields and recovery rules.

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

## Privacy

Nicely sends no telemetry. It contacts a service only when a command you run needs it, such as Deepgram for `ncly transcript`.

## Contributing

Issues and pull requests are welcome. [AGENTS.md](AGENTS.md) explains how changes are made.

## License

MIT. See [LICENSE](LICENSE).
