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

The current build implements help, the version, shell completion, and `ncly describe`, which lists the installed commands, core's and each extension's, with their summaries, and gives one command's arguments, flags, effects, modes, and requirements, with each config value that it reads and where that value comes from, and the JSON Schema of its answers. It finds the extensions named `ncly-<domain>` in `~/.local/share/nicely/extensions/` and in `PATH`, shows their help and completion from their manifests, starts them with a call that it has checked against their declared arguments, flags, and flag groups, checks their JSON answer before passing it on, and stops them and their child processes on Ctrl-C. It bundles the `skill` extension, whose `ncly skill list` lists the skills in the folders that `[skill] paths` names in the config and the skill of each installed extension, and `ncly skill view <name>` prints one of them, which a terminal shows as styled Markdown, without its frontmatter. Shell completion suggests the names of the skills after `ncly skill view`, and an extension's argument or flag can complete from one of its list commands. `ncly doctor` checks the machine and reports on stdout. It says whether the config files parse and warns about [unknown config keys](docs/north-star/contract.md#configuration). It lists each extension that it finds, and the one that hides another. It checks the programs that each extension requires, and names the install command of a missing or too old one. Then it runs each extension's status command, such as `ncly skill status`, which checks the skill folders, names, and files. Last, its `auth` component checks that the keychain answers, that each grant still names an extension that declares its key, and that each declared key is in its variable or the keychain and granted to each extension that needs it, with the verdict of `ncly auth status`. `ncly auth status` lists the keys that the installed extensions declare, says whether each one comes from its environment variable, such as `DEEPGRAM_API_KEY`, or from the OS keychain, without printing any key, and which extensions hold its grant, and `ncly auth login <service>` stores a key in the keychain, from a masked field that never shows it or from `--stdin`. `ncly auth logout <service>` removes it, and both ask for a typed `[y/N]` before they replace or remove a key. `ncly auth grant <service> <domain>` lets the installed extension of a domain receive a key, once a human confirms it with a typed `[y/N]` in a terminal, and `ncly auth revoke <service> <domain>` takes the grant back. A grant holds the path where `ncly` found the extension, so another `ncly-<domain>` earlier in `PATH` inherits nothing, and granting to it asks again with both paths, while a Homebrew upgrade behind the same link keeps the grant. Core passes an extension only the keys that a human granted to it at its path, from their variables or the keychain, and names each declared key that it does not pass in `NCLY_MISSING_KEYS` with its code, which the extension answers when it needs that key: `KEY_NOT_GRANTED`, with the grant command to run, for a key that no human granted, and `AUTH_MISSING` or `KEYRING_UNAVAILABLE` for a key that neither its variable nor the keychain gives. An extension can mark a key optional in its manifest when it runs without it, such as a harness that falls back to its own login: `status` and doctor then pass that key when it is missing or not granted, unless another extension needs it, and `ncly describe` says which keys are optional. Completion suggests the declared services after each of them, then the domains of `grant` and `revoke`. In a terminal, a call that lacks a required value opens a form that asks only for what it lacks, then prints the equivalent command after `Next time:`, and elsewhere the call fails with the full command as its hint. On a terminal, `skill list`, `auth status`, and the doctor report are styled, and a pipe gets the same text without the styles. From a checkout, `just install` puts `ncly` in `~/.local/bin/` and the `agent` extension in `~/.local/share/nicely/extensions/`, and `ncly doctor agent` checks that Codex 0.162.1 or later is installed. `ncly agent run <task> --tools --model <id> --dry-run` checks one trusted task, the model and effort from its flags or the `[agent]` table of the config, the folder, and Codex, then shows the invocation of Codex without starting it. It refuses with `AGENT_DEPTH_LIMIT` a call that already runs inside `[agent] max_depth` agents, 2 by default, as `NCLY_AGENT_DEPTH` counts them, and gives Codex that depth plus one. Without `--dry-run`, it runs the task, under the ID of `--run-id` or one that it generates, and keeps a record of the run in `~/.local/state/nicely/runs/`: it saves the record before Codex starts, then the answer of Codex in `answer.txt`, before it prints that answer. When that run completed, the same call with the same `--run-id` prints that saved answer again and starts no Codex. When the call stopped before it saved its intent, the same call runs the task. A call whose run may have started a paid task without a saved answer starts no other one. `ncly run list` lists the saved runs, newest first, by command path, item key, or start time, and marks the ones that a process holds now, without an index. `ncly run view <run-id>` shows one record, its steps, artifacts, and errors, and whether a process holds it, without starting any work. It also establishes the JSON error contract, configuration, and English message catalog. The commands below are planned work.

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

The bundled `nicely` skill teaches only what this release can do: find a command with `ncly describe`, dry-run it when it writes or bills, check the machine with `ncly doctor` when a problem has no explanation, and act on the stable `code` of each error and on the exit code, which says who acts. When the hint names a command that runs as is, `next` holds that command as words. Each extension ships a skill for its own commands, which `ncly skill list` lists. It teaches [run IDs](docs/north-star/contract.md#run-ids) that the caller chooses, the inspection of a run with `ncly run list` and `ncly run view`, and the verified replay of a completed call. An uncertain paid result never permits an automatic second charge. [Explicit resume](docs/milestones/M99-parking-lot.md#resume) waits for a workflow that needs it. The [agent contract](docs/north-star/contract.md) defines the fields and recovery rules.

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

From a checkout, with Go and `just`, run `just install`. It builds `ncly` into `~/.local/bin/` and the first-party extensions, such as `agent`, into `~/.local/share/nicely/extensions/`, or into `$XDG_DATA_HOME/nicely/extensions/` when `XDG_DATA_HOME` holds an absolute path. The `agent` extension uses the login of an installed Codex, which `ncly doctor agent` checks.

## Privacy

Nicely sends no telemetry. It contacts a service only when a command you run needs it, such as Deepgram for `ncly transcript`.

## Contributing

Issues and pull requests are welcome. [AGENTS.md](AGENTS.md) explains how changes are made.

## License

MIT. See [LICENSE](LICENSE).
