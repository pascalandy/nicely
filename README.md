# Nicely

A curated, multilingual CLI toolbox that humans and AI agents operate equally well. The command is `ncly`.

```bash
ncly --help
ncly --version
ncly completion bash
```

## Status

Nicely is in early development and has no release yet. [The guide](docs/north-star/guide.md) explains what it is and why. [The first milestone](docs/milestones/M0-foundation.md) is under way.

The current M0 build implements help, the version, and shell completion. It also establishes the JSON error contract, configuration, and English message catalog. The commands below are planned work.

## Planned commands

[M1](docs/milestones/M1-day-one.md) adds `doctor`, `auth`, `skill`, `describe`, `transcript`, and `run`. Doctor checks local prerequisites, auth stores keys in the OS keychain, and discovery describes installed capabilities. Transcript produces files and a run record. Explicit resume verifies saved work before continuing it.

Interactive forms will ask for missing values. Scripts and agents set `NCLY_NO_INPUT=1` to prevent questions, including in a pseudo-terminal, and use `--json` or `NCLY_JSON=1` for machine output. Commands group actions by domain, such as `ncly transcript run youtube`.

Later milestones add the public `agent` command in M2, taps and trusted extensions in M3, everyday domains in M4, a private docs hub in M5, and Canadian French in M6. The [guide's command table](docs/north-star/guide.md#domains-and-commands) records their scope.

## Planned agent instructions (M1)

M1 will verify this text in a fresh agent session before release. Once those commands ship, paste it into your agent instructions:

```text
Set NCLY_NO_INPUT=1 and NCLY_JSON=1 for ncly calls
Use a targeted ncly describe, the appropriate ncly doctor component, then the command's --dry-run before execution
Find skills with ncly skill list and read one with ncly skill view NAME
Parse codes and fields, never translated messages. Transcript operation answers always use results[], even for one item
Retain the invocation's RFC3339 start time, exit code, full stdout and stderr, and run_id
Give the shell a timeout longer than ncly's timeout plus 15 seconds of cleanup and a margin
Only overall exit 75 permits an automatic repeat of the same invocation. Item errors do not permit a retry. Relay human-action hints
After lost output, use ncly run list --path 'transcript run youtube' --key URL --since RFC3339 --json with the original URL and start time. If the time is lost, omit --since. Preserve ambiguous matches
Inspect candidates with ncly run view RUN_ID. active reports the execution lock. Never infer that a run is idle from its saved status
For an inactive run, inspect ncly run resume RUN_ID --dry-run before explicitly resuming supported work
Repeating the original transcript command transcribes and bills again. An explicit ncly transcript summary run RUN_ID reuses the transcript and may bill for a new summary
```

The [agent contract](docs/north-star/cli-spec.md) defines the fields and recovery rules. Saved records and verified output files retain work across sessions.

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
