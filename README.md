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

Nicely is in early development. [The vision](docs/north-star/vision.md) explains what it is and why, and [the backlog queue](docs/backlog/README.md#queue) lists the work in order.

The current build brings:

- Help, the version, shell completion, and `ncly describe`, which gives the contract of each installed command
- Extensions named `ncly-<slug>`, including independent jobs such as `ncly-video-convert`, and the bundled `skill` extension
- `ncly doctor`, which checks the machine, and `ncly auth`, which keeps the keys and their grants
- The `agent` extension, which runs one trusted task through Codex and records it
- The `transcript` extension, which transcribes one YouTube video with Deepgram, records the run, replays it, and stops cleanly on Ctrl-C
- `ncly run list` and `ncly run view`, which read the records of the runs
- The extension template, which creates the repository of an extension outside Nicely, with its tests and its recipes. Its `AGENTS.md` explains how to add a first command, remove the example, and release a version. Its `README.md` tells a user how to install a release
- SDK checks that explain broken manifest rules and missing catalog IDs in English, with their repository file names

[The current build](docs/status.md) describes each command. The commands below are planned work.

## Planned commands

Nicely is a small core that extensions add commands to, in the spirit of Pi.
Core brings `describe`, `doctor`, `auth`, and `run`. It bundles `skill`, which agents use to learn Nicely.
External extensions supply other commands under the [extension contract](docs/north-star/contract/extensions.md).

The first `agent` command runs one trusted task through Codex. The first `transcript` command transcribes one explicitly selected YouTube video. A checkout installs the extensions with `just install`. Each capability starts with its smallest useful path; additional harnesses, transcript summaries, and taps remain in the [backlog](docs/backlog/README.md).

Named extensions share domain help, completion, and describe, such as `video convert` and `video subtitle`.
Read their skills by slug, such as `ncly skill view video-convert`.
Configuration, doctor components, and key grants use each extension's slug independently.
For example, `ncly doctor video-subtitle` checks that extension, and `ncly auth grant deepgram video-subtitle` grants only its key access.
Existing whole-domain grants keep working without migration.
The template creates named repositories with `--data domain=video --data name=convert`, such as `ncly-video-convert`.
Omit `name` for a whole-domain repository.
Both use the SDK test kit and the same update recipe.
Further work lives in [docs hub](docs/backlog/XT-104-docs-hub.md), [Canadian French](docs/backlog/COR-102-canadian-french.md),
[Markdown view](docs/backlog/XT-103-markdown-view.md), and [video and image jobs](docs/backlog/XT-602-video-image-jobs.md).

In a terminal, a form asks a human for each missing value. Scripts and agents set `NCLY_NO_INPUT=1` to prevent questions, including in a pseudo-terminal, and use `--json` or `NCLY_JSON=1` for machine output. Commands group actions by domain, such as `ncly transcript run youtube`.

## Let your agent find it

Add one line to your agent instructions, such as `AGENTS.md` or `CLAUDE.md`, to teach an agent to use Nicely:

```text
When asked to use nicely, run ncly skill view nicely
```

The bundled `nicely` skill teaches only what this release can do: find a command with `ncly describe`, dry-run it when it writes or bills, check the machine with `ncly doctor` when a problem has no explanation, and act on the stable `code` of each error and on the exit code, which says who acts. When the hint names a command that runs as is, `next` holds that command as words. Each extension ships a skill for its own commands, which `ncly skill list` lists. It teaches [run IDs](docs/north-star/contract/run-ids.md) that the caller chooses, the inspection of a run with `ncly run list` and `ncly run view`, and the verified replay of a completed call. An uncertain paid result never permits an automatic second charge. [Explicit resume](docs/backlog/COR-621-resume-sdk.md#later) waits for a workflow that needs it. The [agent contract](docs/north-star/contract.md) defines the fields and recovery rules.

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

For a manual setup, `ncly completion <shell> --help` shows the commands. In bash, load the script with `eval "$(ncly completion bash)"`, because the bash 3.2 of macOS reads nothing from `source <(...)`.

From a checkout, with Go and `just`, run `just install`. It builds `ncly` into `~/.local/bin/` and the first-party extensions, such as `agent` and `transcript`, into `~/.local/share/nicely/extensions/`, or into `$XDG_DATA_HOME/nicely/extensions/` when `XDG_DATA_HOME` holds an absolute path. The `agent` extension uses the login of an installed Codex, which `ncly doctor agent` checks. The `transcript` extension needs `uv` and Python 3.12 or later, which `ncly doctor transcript` checks, with the cookies of Arc when the config selects it.

An external extension, such as `ncly-video-convert`, installs from its repository at a release tag, with Go and `just`. [Release and install](docs/north-star/contract/extensions.md#release-and-install) gives the commands, and `ncly skill view nicely` gives them to an agent.

## Privacy

Nicely sends no telemetry. It contacts a service only when a command you run needs it, such as Deepgram for `ncly transcript`.

## Contributing

Issues and pull requests are welcome. [AGENTS.md](AGENTS.md) explains how changes are made.

## License

MIT. See [LICENSE](LICENSE).
