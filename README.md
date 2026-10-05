# Nicely

A curated, multilingual CLI toolbox that humans and AI agents operate equally well. The command is `ncly`.

```bash
ncly doctor
ncly skill list --json
ncly transcript run youtube --url "https://www.youtube.com/watch?v=VIDEO_ID"
```

## Status

Nicely is in planning and has no release yet. [The guide](docs/north-star/guide.md) explains what it is and why. [The first milestone](docs/milestones/M0-foundation.md) starts the plan.

## How it works

- `ncly` groups tools by domain. `ncly video --help` shows everything Nicely does with a video.
- Run a command without its flags in a terminal, and a short form asks for what is missing. Pass every flag, or call it from a script or an AI agent, and it runs without a question. Add `--json` for output a program can read.
- `ncly doctor` lists what is missing on your machine and the command that fixes each item.
- `ncly auth login` stores API keys in your OS keychain. Keys never sit in a config file.
- Every message can be translated. English comes first, then Canadian French.
- Personal tools plug in as extensions, the way Git subcommands do: an executable named `ncly-backup` becomes `ncly backup`. An extension runs with your rights, so add only taps you trust.
- One TOML file describes your setup and can travel between your machines. `ncly tap sync` brings a new machine to the same taps.

## Install

There is no release yet. The first release installs with:

| System | Command |
|---|---|
| macOS | `brew install pascalandy/tap/ncly` |
| Arch Linux, Omarchy | `yay -S ncly-bin`, or any AUR helper |

Both packages include completions for zsh, bash, and fish.

## Privacy

Nicely sends no telemetry. It contacts a service only when a command you run needs it, such as Deepgram for `ncly transcript`.

## Contributing

Issues and pull requests are welcome. [AGENTS.md](AGENTS.md) explains how changes are made.

## License

MIT. See [LICENSE](LICENSE).
