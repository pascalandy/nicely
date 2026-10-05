# ncly CLI spec

This file is the agent contract of Nicely and the reference for every specified command. Each command section, and each command in the tree, names the milestone that introduced it. When a milestone starts, it moves its draft commands here before any code is written, and the milestone file links to them.

## Usage (M0)

```
ncly [global flags] <domain> [<resource>] <verb> [arguments] [flags]
ncly [global flags] <command> [arguments] [flags]
```

A domain is a noun, such as `skill`. A resource is an optional second noun inside a domain, such as `prompt` in `ncly transcript prompt list`. Nouns are singular. A standalone command, such as `doctor`, has no domain.

Each action has one verb in every domain. Reuse a verb from this table when it fits. A milestone that needs a new verb adds it here.

| Verb | Meaning |
|---|---|
| `list` | Lists items |
| `view` | Shows one item |
| `add` | Adds an item |
| `remove` | Removes an item |
| `run` | Runs a task |
| `sync` | Brings local copies up to date with their source |
| `login`, `logout` | Stores or removes a key |
| `status` | Reports the current state |

`ncly --help` lists domains and commands with one line each. `ncly <domain> --help` lists the actions of that domain, and `ncly help <domain>` prints the same text. Every help page ends with two to five examples. `ncly` and `ncly <domain>` without a verb print their help on stdout and exit 0.

## Command tree

```
ncly
├── completion zsh|bash|fish        M0
├── doctor [component]              M1
├── auth login|logout|status        M1
├── skill list|view                 M1
├── transcript run youtube|zoom     M1
└── transcript prompt list          M1
```

Core reserves every name in the domain table of the [guide](guide.md#domains-and-commands), plus `help`, `version`, and `config`. An extension never runs under a reserved name.

## Global flags (M0)

Each global flag other than `--help` and `--version` has an environment variable with the same effect. A flag wins over its variable. [Programs that ncly runs](#programs-that-ncly-runs-m1) receive the resolved values.

| Flag | Variable | Default | Effect |
|---|---|---|---|
| `-h`, `--help` | | | Shows help on stdout and exits 0. It wins over every other argument, including unknown flags, but still honors `--lang` and `--no-color`. |
| `--version` | | | Prints `ncly vX.Y.Z` on stdout and exits 0. |
| `--json` | `NCLY_JSON=1` | off | Machine output, as described in [Output](#output-m0). |
| `--no-input` | `NCLY_NO_INPUT=1` | off | Never prompts, even in a terminal. |
| `--no-color` | `NO_COLOR`, or `TERM=dumb` | off | Plain text. |
| `--lang <tag>` | `NCLY_LANG` | detected | Language of human text, such as `en` or `fr-CA`. |
| `-v`, `--verbose` | `NCLY_VERBOSE=1` | off | Adds step details on stderr. |

`NCLY_DEBUG=1` adds internals, timings, and child commands on stderr. It has no flag. Keys never appear in any output.

Commands that write also accept these flags.

| Flag | Effect |
|---|---|
| `-n`, `--dry-run` | Shows what the command would change. It changes nothing the user owns, reads no key, and sends no paid request. It may fill Nicely's cache, including the first download of a Python program's dependencies. |
| `-f`, `--force` | Answers yes to every confirmation, such as an overwrite or a deletion. It never installs a prerequisite and never grants a key, because both need a human in a terminal. |
| `-o`, `--output <path>` | Writes the result to this file, or to this folder when the command writes several files. |

Commands that call the network or run for minutes also accept `--timeout <duration>`. A duration is `30s`, `5m`, `2h`, or a number of seconds. Each command states its default in its help.

## Modes (M0)

Interactive mode applies when stdin and stdout are terminals, `--no-input` is absent, `NCLY_NO_INPUT` is not `1`, and `CI` is not set. Every other case is non-interactive mode. Some agent tools run commands in a pseudo-terminal, so an agent sets `NCLY_NO_INPUT=1` in its environment.

| Situation | Interactive mode | Non-interactive mode |
|---|---|---|
| Every required value is given | Runs | Runs |
| A required value is missing | A form asks only for the missing values | Exit 2, `USAGE_INVALID`, and a hint with the full command |
| A step needs a confirmation and `--force` is absent | Asks for confirmation | Exit 2, `CONFIRMATION_REQUIRED`, and a hint that adds `--force` |
| A step only a human can do, such as typing or granting a key | Asks | Exit 78, `TERMINAL_REQUIRED`, and a hint that names the command to run in a terminal |

After an interactive run that used a form, `ncly` prints the equivalent command on stderr, after the line `Next time:`. The command itself is never translated.

## Output (M0)

stdout carries data. stderr carries progress, warnings, and errors. Without `--json`, human output uses Lip Gloss styles, and `--no-color` turns them off.

With `--json`, every answer is one JSON object on one line, never a bare array. `ok` is `true` or `false` and agrees with the exit code. `--help`, `--version`, and `ncly completion` print text and ignore `--json`.

- On success, stdout holds the object. Progress is not shown, and stderr stays empty unless `--verbose` or `NCLY_DEBUG` adds lines.
- On failure, stdout stays empty and the object ends stderr. With `--verbose` or `NCLY_DEBUG`, diagnostic lines come before it.

```json
{"ok":false,"errors":[{"code":"AUTH_MISSING","message":"No Deepgram key found.","hint":"ncly auth login deepgram"}]}
```

`errors` holds one object per problem, and the first object decides the exit code. `warnings` holds objects of the same shape, may appear in any answer, and never changes the exit code.

| Key | Meaning |
|---|---|
| `code` | A stable code from the [error registry](#error-codes-m0). Agents branch on this key. |
| `message` | Human text in the active language. Agents never parse it. |
| `hint` | The command or the action that fixes the problem. A command is never translated. |

A failure after partial work keeps the data keys that still apply, such as `output_dir`, next to `errors`.

A command that processes several items, such as several URLs or files, checks every input before it starts. A bad input stops the whole command with exit 2. Otherwise the answer holds `results`: one object per item, in input order, each with its own `ok` and either its data or its `errors`. The command exits 0 when every item succeeded, 75 when every failure was temporary and came before any paid request, and 1 otherwise.

A report command, such as `ncly doctor` or `ncly auth status`, answers with a report. The report goes to stdout even when a check fails, and it lists its findings under its own key, such as `checks`, instead of `errors`. `ok` gives the verdict, and the exit code matches it. When the report command itself fails, it answers with `errors` like any other command.

JSON keys are only added. Renaming or removing a key needs an entry in [decision-records.md](decision-records.md).

## Exit codes (M0)

| Code | Meaning | Safe to rerun |
|---|---|---|
| `0` | Success | |
| `1` | Runtime failure. Work may be partial or already billed. | No. Never rerun automatically. |
| `2` | Invalid invocation: unknown command, bad flag, missing value, or missing confirmation | After fixing the call |
| `75` | Temporary failure before any paid request, such as a network error or a lock that another `ncly` holds | Yes, with the same command |
| `78` | A human must act: a key, a prerequisite, a config fix, or a step at a terminal | After the human runs the hint |
| `130` | Interrupted by Ctrl-C | |
| `143` | Terminated by SIGTERM | |

Codes 124 to 127, and every code from 128 up other than 130 and 143, stay reserved for the shell.

## Error codes (M0)

Each code maps to exactly one exit code. Core uses only the codes in this table, and a milestone adds its codes here. An extension may use its own codes, prefixed with its domain, such as `FLEET_HOST_DOWN`. Its exit code carries the meaning that this file defines.

| Code | Exit | When | Since |
|---|---|---|---|
| `USAGE_INVALID` | 2 | Unknown command, bad flag, or a missing value in non-interactive mode | M0 |
| `CONFIRMATION_REQUIRED` | 2 | A step needs a confirmation in non-interactive mode, and `--force` is absent | M0 |
| `CONFIG_INVALID` | 78 | A config file has a syntax error or a value of the wrong type | M0 |
| `TERMINAL_REQUIRED` | 78 | A step needs a human at a terminal, such as typing or granting a key | M0 |
| `RUNTIME` | 1 | Any other failure during work | M0 |
| `NOT_FOUND` | 2 | The named skill, service, profile, or component does not exist | M1 |
| `AUTH_MISSING` | 78 | A required key is in neither the environment nor the keychain | M1 |
| `PREREQ_MISSING` | 78 | A required tool is absent or too old | M1 |
| `KEYRING_UNAVAILABLE` | 78 | A command needs a key that is not in the environment, and the OS keychain does not answer | M1 |
| `TEMPORARY` | 75 | The network, a service, or a lock failed before any paid request | M0 |
| `INTERRUPTED` | 130 | Ctrl-C stopped the command | M0 |
| `TERMINATED` | 143 | SIGTERM stopped the command | M0 |

Warnings use codes from this table.

| Code | When | Since |
|---|---|---|
| `CONFIG_UNKNOWN_KEY` | A config file holds a key that this version of `ncly` does not know | M0 |
| `SKILL_SHADOWED` | Two skill folders hold the same skill name | M1 |
| `SKILL_NO_SOURCE` | No skill folder is configured | M1 |

## Configuration (M0)

`config.toml` is the shared config: the setup the user wants, which can travel between machines, for example through dotfiles. `config.local.toml`, in the same folder, is the local config: what belongs to this machine only.

Precedence, highest first: flags, environment variables, `config.local.toml`, `config.toml`, then defaults. A table, such as `[agent.profiles.everyday]`, merges key by key across the files. A list, such as `[skill] paths`, from a higher source replaces the whole list below it.

| Purpose | Path |
|---|---|
| Shared config | `~/.config/nicely/config.toml` |
| Local config | `~/.config/nicely/config.local.toml` |
| Data: taps and extensions | `~/.local/share/nicely/` |
| State: logs, locks, and run history | `~/.local/state/nicely/` |
| Cache: extracted Python programs | `~/.cache/nicely/` |

Each path honors `XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`, or `XDG_CACHE_HOME` when set. macOS uses the same paths as Linux. `NCLY_CONFIG` names another shared config file, and the local config is then read from the same folder.

- A missing config file is not an error, and the defaults apply.
- A syntax error or a value of the wrong type exits 78 with `CONFIG_INVALID`. `--help` and `--version` still work, with the defaults.
- An unknown key is a `CONFIG_UNKNOWN_KEY` warning in `ncly doctor`, never an error, so an older `ncly` reads a config written for a newer one.
- `ncly` writes a config file only when the job of a command is to change the setup, such as adding a tap. It edits the file in place, keeps comments and formatting, and names the file it changed. Under `--dry-run`, it shows the change instead. Each such command states which file it writes.

The language comes from the first source that is set: `--lang`, `NCLY_LANG`, `lang` in the config, `LC_ALL`, `LC_MESSAGES`, `LANG`, and finally `en`. `ncly` picks the closest catalog, so `fr_CA.UTF-8` and `fr` both select `fr-CA` once that catalog exists. A language with no close catalog, including `C` and `POSIX`, falls back to `en` without an error.

```toml
lang = "en"

[skill]
paths = ["~/code/skills"]
```

## Keys (M1)

- A service has a name of lowercase letters and digits, such as `deepgram`. Its environment variable is the name in uppercase followed by `_API_KEY`, such as `DEEPGRAM_API_KEY`.
- `ncly` reads a key from its environment variable first. It asks the keychain only when that variable is empty, and for `ncly auth`. The keychain entry uses the service `nicely` and the account named after the service.
- Each keychain call has a time limit, and a keychain that does not answer in time counts as unavailable. A command that needs a key exits 78 with `KEYRING_UNAVAILABLE` when the key is missing from the environment and the keychain is unavailable, and its hint names the variable to export. When the keychain answers without the key, the code is `AUTH_MISSING`.
- A key is never accepted as a flag. `ncly auth login` reads it from a masked field, or from stdin with `--stdin`.
- A key is never printed. `ncly auth status` reports where a key comes from, never its value.

## Programs that ncly runs (M1)

`ncly` runs three kinds of programs: Python programs of core, harnesses, and, from M3, extensions. None of them runs in a sandbox. A program runs with the rights of the user, so it could read the keychain or any file itself. Adding a tap means trusting its author.

**Environment.** A program receives the environment of `ncly` with these changes:

- Each variable named `<NAME>_API_KEY` is the key of a service. The program receives it only when that key is granted to it, and `ncly` then sets it from the environment or the keychain.
- The variables of the global flags carry their resolved values, and `NCLY_VERSION` holds the version of `ncly`.
- A harness receives `NCLY_AGENT_DEPTH` plus one.

**Grants.** A program of core receives the keys that core declares for it. An extension receives a key only after a human grants that key to that extension in a terminal. `--force` never grants a key. When an update of an extension declares a new key, the extension does not start until a human grants it.

**Cancellation.** On Ctrl-C or SIGTERM, `ncly` sends the signal to every program it started and to their descendants. It sends SIGKILL to what still runs 15 seconds later, which leaves a Python program its own 10 seconds to stop its children. It keeps every finished output and removes only its own temporary files. It exits 130 with `INTERRUPTED` or 143 with `TERMINATED`, and under `--json` the answer keeps the data keys that apply, such as `output_dir` or the `results` of finished items. A second signal during this cleanup is ignored. A SIGKILL sent to `ncly` itself leaves no time to clean up.

## Agent profiles (M1)

A profile names a harness, a model, and an effort, plus a provider when the harness serves several. Profiles live in the config, and every command that runs an agent reads them there.

```toml
[agent]
default_profile = "everyday"

[agent.profiles.everyday]
harness = "claude"
model = "<model id>"
effort = "medium"

[agent.profiles.second-opinion]
harness = "pi"
provider = "openai-codex"
model = "<model id>"
effort = "high"
```

| Key | Meaning |
|---|---|
| `harness` | `claude`, `codex`, `grok`, `pi`, or `opencode` |
| `provider` | The model provider, for a harness that serves several, such as `pi` or `opencode` |
| `model` | A model ID that the harness accepts |
| `effort` | A reasoning effort that the harness accepts |

`--profile <name>` picks a profile, and `default_profile` applies without it. An unknown profile exits 2 with `NOT_FOUND`.

## ncly completion (M0)

```
ncly completion zsh|bash|fish
```

Prints the completion script on stdout. The Homebrew and AUR packages install these scripts, so users run this command only for a manual setup.

Completion also suggests values: skill names, auth services, profiles, transcript prompts, and doctor components.

## ncly doctor (M1)

```
ncly doctor [component] [--live] [--json]
```

| Argument or flag | Effect |
|---|---|
| `component` | One of `core`, `auth`, `skill`, or `transcript`. Without it, doctor checks all of them. |
| `--live` | Adds network checks against free endpoints only, such as listing Deepgram projects. It never calls an endpoint that bills. |

- The default checks stay on the machine: binaries and their versions, keys present, the keychain answering, and the config files parsing.
- When the keychain is unavailable, its check is `warn` if every key that a component needs comes from the environment, and `fail` otherwise.
- The `skill` component checks that each folder in `[skill] paths` exists and that each `SKILL.md` has a `name` and a `description`.
- In interactive mode, when a tool is missing and its install command for this system is known, doctor shows the command, such as `brew install ffmpeg` or `sudo pacman -S ffmpeg`, and runs it only after a yes.
- In non-interactive mode, doctor never installs anything. The hint of each failed check holds the command.
- Doctor is a report command. It exits 0 when no check fails and 78 when at least one check needs a human. When doctor itself fails, it exits 1 with `errors`.

```json
{"ok":false,"checks":[{"id":"transcript.uv","component":"transcript","status":"pass","message":"uv found"},{"id":"auth.deepgram","component":"auth","status":"fail","code":"AUTH_MISSING","message":"No Deepgram key found.","hint":"ncly auth login deepgram"}]}
```

`status` is `pass`, `warn`, or `fail`. A warning never changes the exit code.

## ncly auth (M1)

```
ncly auth login <service> [--stdin] [--dry-run] [--force]
ncly auth logout <service> [--dry-run] [--force]
ncly auth status [--json]
```

Core uses one service in M1, `deepgram`, for `transcript`. Any other service name works the same way, so an extension's key is stored like a core key.

- `login` stores the key in the keychain. Replacing a stored key asks for confirmation, or needs `--force` in non-interactive mode. In non-interactive mode without `--stdin`, `login` exits 78 with `TERMINAL_REQUIRED` and the hint ``Run `ncly auth login <service>` in a terminal``. An agent relays this hint to the human and never pipes a key itself.
- `logout` removes the key from the keychain. It asks for confirmation, or needs `--force` in non-interactive mode.
- When the keychain is unavailable, `login` and `logout` exit 78 with `KEYRING_UNAVAILABLE`, and the hint names the environment variable to use instead.
- `status` is a report command. It reports whether the keychain answers, then lists every service that core uses, plus, from M3, every service that an extension declares. `ok` is `false`, with exit 78, only when the keychain does not answer and a listed service has no key in the environment.

```json
{"ok":true,"keychain":"available","services":[{"service":"deepgram","configured":true,"source":"keychain"}]}
```

`keychain` is `available` or `unavailable`. `source` is `env`, `keychain`, or `none`. To move a key from another secret store, pipe it once: `<command that prints the key> | ncly auth login deepgram --stdin`.

## ncly skill (M1)

```
ncly skill list [--json]
ncly skill view <name> [--json]
ncly skill <name>
```

- Skills come from the folders in `[skill] paths`. Each subfolder with a `SKILL.md` is a skill, named by the `name` key of its frontmatter. Taps add more sources in M3.
- When two folders hold the same skill name, the first folder in `paths` wins and `list` adds a `SKILL_SHADOWED` warning.
- When no folder is configured, `list` returns an empty list and a `SKILL_NO_SOURCE` warning whose hint names `[skill] paths`.
- `list` returns names and descriptions, sorted by name.
- `view` prints the `SKILL.md` of the skill. Its first line gives the folder, so the relative paths inside the skill resolve: `<!-- skill-dir: /Users/me/code/skills/transcript -->`. In interactive mode, Glamour renders the Markdown.
- A mode, such as `andy-mode`, is a skill like any other. `view` prints its router, which names its routes.
- `ncly skill <name>` is a shortcut for `view`, for humans. Agents use `view`. No skill may take the name of a `skill` verb.
- An unknown name exits 2 with `NOT_FOUND`.

```json
{"ok":true,"skills":[{"name":"transcript","description":"Use when the user invokes `transcript` or asks to transcribe a YouTube video or Zoom recording.","path":"/Users/me/code/skills/transcript"}]}
```

`view --json` returns `{"ok":true,"name":"…","path":"…","content":"…"}`.

## ncly transcript (M1)

Transcribes YouTube videos and Zoom recordings with Deepgram, then summarizes the result through an agent profile.

```
ncly transcript run youtube --url <url> [<url>...] [--prompt <name>] [--profile <name>] [--output <folder>] [--timeout <duration>] [--dry-run] [--json]
ncly transcript run zoom (--latest | --path <folder>) [--profile <name>] [--output <folder>] [--timeout <duration>] [--dry-run] [--json]
ncly transcript prompt list [--json]
```

- A run needs the `deepgram` key and exits 78 with `AUTH_MISSING` without it.
- `--profile` names an [agent profile](#agent-profiles-m1) for the summary.
- `--dry-run` follows the [global definition](#global-flags-m0): it validates the plan without a key, a paid request, or a change to the user's files.
- Exit 75 is safe to rerun. Never rerun an exit 1 automatically, because Deepgram may already have billed the audio.
- When the summary fails, the transcript is still saved, and the failure keeps `output_dir` next to `errors`.
- Several URLs answer with `results`, as [Output](#output-m0) describes. Each item carries its `url`.
- M1 settles the default output folder and the remaining options of the transcript CLI, as its [to-define list](../milestones/M1-day-one.md#to-define-before-code) says.

## Examples

```bash
# Check a new machine, then fix what doctor reports
ncly doctor
ncly auth login deepgram

# An agent discovers skills, then reads one
ncly skill list --json | jq -r '.skills[].name'
ncly skill view transcript

# Plan a transcription without a key, a paid request, or a write
ncly transcript run youtube --url "https://www.youtube.com/watch?v=VIDEO_ID" --dry-run --json

# Transcribe, then keep only the result folder
ncly transcript run youtube --url "$URL" --json | jq -r .output_dir

# A missing key, seen by a script
ncly transcript run youtube --url "$URL" --json
# stderr: {"ok":false,"errors":[{"code":"AUTH_MISSING","message":"No Deepgram key found.","hint":"ncly auth login deepgram"}]}
# exit code: 78
```
