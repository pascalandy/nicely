# ncly CLI spec

This file is the agent contract of Nicely and the reference for every specified command. A milestone adds its commands here before any code is written. Each section names the milestone that introduced it.

## Usage

```
ncly [global flags] <domain> <action> [arguments] [flags]
ncly [global flags] <command> [arguments] [flags]
```

`ncly --help` lists domains and commands with one line each. `ncly <domain> --help` lists the actions of that domain. Every help page ends with two to five examples.

## Command tree

```
ncly
├── completion zsh|bash|fish        M0
├── doctor [component]              M1
├── auth login|logout|status        M1
├── skill list|show                 M1
└── transcript run|list             M1
```

These names are reserved for later milestones, and nothing else may use them: `agent`, `tap`, `markdown`, `video`, `image`, `docs`, and `fleet`.

## Global flags (M0)

| Flag | Default | Effect |
|---|---|---|
| `-h`, `--help` | | Shows help and exits 0. Other arguments are ignored. |
| `--version` | | Prints `ncly vX.Y.Z` to stdout and exits 0. |
| `--json` | off | Machine output, as described in [Output](#output-m0). |
| `--no-input` | off | Never prompts, even in a terminal. |
| `--no-color` | off | Plain text. `NO_COLOR` and `TERM=dumb` have the same effect. |
| `--lang <tag>` | detected | Language of human text, such as `en` or `fr-CA`. |
| `-v`, `--verbose` | off | Adds diagnostics on stderr. |

Commands that write also accept these flags.

| Flag | Effect |
|---|---|
| `--dry-run` | Prints what would happen. Writes nothing, reads no key, and sends no paid request. |
| `--force` | Overwrites or deletes without asking. |
| `--output <path>` | Writes results to this file or folder instead of the default. |

## Modes (M0)

Interactive mode applies when stdin and stdout are a terminal and `--no-input` is absent. Every other case is non-interactive mode.

| Situation | Interactive mode | Non-interactive mode |
|---|---|---|
| Every required value is given | Runs | Runs |
| A required value is missing | A form asks only for the missing values | Exit 2, `USAGE_INVALID`, and a hint with the full command |
| A step would overwrite or delete without `--force` | Asks for confirmation | Exit 2, `CONFIRMATION_REQUIRED`, and a hint that adds `--force` |

After an interactive run that used a form, `ncly` prints the equivalent command on stderr, after the line `Next time:`. The command itself is never translated.

## Output (M0)

stdout carries data. stderr carries progress, warnings, and errors.

Without `--json`, human output uses Lip Gloss styles, and `--no-color` turns them off.

With `--json`:

- On success, stdout holds exactly one JSON object and stderr stays empty. Warnings go in a `warnings` array inside that object. Progress is not shown.
- On failure, stdout stays empty and stderr holds exactly one JSON object.

```json
{"code":"AUTH_MISSING","message":"No Deepgram key found.","hint":"ncly auth login deepgram"}
```

| Key | Meaning |
|---|---|
| `code` | A stable code from the [error registry](#error-codes-m0). Agents branch on this key. |
| `message` | Human text in the active language. Agents never parse it. |
| `hint` | The command or the action that fixes the error. A command is never translated. |

An error object may carry more keys for context, such as `output_dir` after a partial success.

A report command, such as `ncly doctor` or `ncly auth status`, prints its report on stdout even when a check fails. Its exit code summarizes the result.

## Exit codes (M0)

| Code | Meaning | Safe to rerun |
|---|---|---|
| `0` | Success | |
| `1` | Runtime failure. Work may be partial or already billed. | No. Never rerun automatically. |
| `2` | Invalid invocation: unknown command, bad flag, missing value, or missing confirmation | After fixing the call |
| `75` | Temporary failure before any write or paid request | Yes, with the same command |
| `78` | Setup needed. A key or a prerequisite is missing, and a human must act. | After the human runs the hint |
| `130` | Interrupted by Ctrl-C | |
| `143` | Terminated by SIGTERM | |

## Error codes (M0)

Each code maps to exactly one exit code. A milestone adds its codes here.

| Code | Exit | When | Since |
|---|---|---|---|
| `USAGE_INVALID` | 2 | Unknown command, bad flag, or a missing value in non-interactive mode | M0 |
| `CONFIRMATION_REQUIRED` | 2 | A step would overwrite or delete in non-interactive mode without `--force` | M0 |
| `RUNTIME` | 1 | Any other failure during work | M0 |
| `NOT_FOUND` | 2 | The named skill, service, or component does not exist | M1 |
| `AUTH_MISSING` | 78 | A required key is in neither the environment nor the keychain | M1 |
| `PREREQ_MISSING` | 78 | A required tool is absent or too old | M1 |
| `KEYRING_UNAVAILABLE` | 78 | The OS keychain does not answer | M1 |
| `TEMPORARY` | 75 | The network or a service failed before any paid request | M1 |

## Configuration (M0)

Precedence, highest first: flags, environment variables, the config file, then defaults.

| Purpose | Path |
|---|---|
| Config file | `~/.config/nicely/config.toml` |
| Data: taps and extensions | `~/.local/share/nicely/` |
| Cache: extracted components | `~/.cache/nicely/` |

Each path honors `XDG_CONFIG_HOME`, `XDG_DATA_HOME`, or `XDG_CACHE_HOME` when set. macOS uses the same paths as Linux. A missing config file is not an error, and the defaults apply. A config file that does not parse exits 2 with `USAGE_INVALID`.

| Variable | Effect |
|---|---|
| `NCLY_CONFIG` | Reads this config file instead of the default |
| `NCLY_LANG` | Same as `--lang` |
| `NCLY_NO_INPUT` | `1` has the same effect as `--no-input` |
| `NO_COLOR` | Turns off colors |

The language comes from the first source that is set: `--lang`, `NCLY_LANG`, `lang` in the config file, `LC_ALL`, `LC_MESSAGES`, `LANG`, and finally `en`. A language without a catalog falls back to `en` without an error.

```toml
lang = "en"

[skill]
paths = ["~/code/skills/skills"]
```

## Keys (M1)

- `ncly` reads a key from its environment variable first, then from the keychain. The keychain entry uses the service `nicely` and the account named after the service, such as `deepgram`.
- A key is never accepted as a flag. `ncly auth login` reads it from a masked field, or from stdin with `--stdin`.
- A key is never printed. `ncly auth status` reports where a key comes from, never its value.
- A Python component receives keys as environment variables of its own process only.

## ncly completion (M0)

```
ncly completion zsh|bash|fish
```

Prints the completion script to stdout. The Homebrew and AUR packages install these scripts, so users run this command only for a manual setup.

Completion also suggests values: skill names, auth services, transcript prompts and profiles, and doctor components.

## ncly doctor (M1)

```
ncly doctor [component] [--live] [--json]
```

| Argument or flag | Effect |
|---|---|
| `component` | One of `core`, `auth`, `skill`, or `transcript`. Without it, doctor checks all of them. |
| `--live` | Adds network checks against free endpoints only, such as listing Deepgram projects. It never calls an endpoint that bills. |

- The default checks stay on the machine: binaries and their versions, keys present, the keychain answering, and the config file parsing.
- In interactive mode, when a tool is missing and its install command for this system is known, doctor shows the command, such as `brew install ffmpeg` or `sudo pacman -S ffmpeg`, and runs it only after a yes.
- In non-interactive mode, doctor never installs anything. The hint of each failed check holds the command.
- Exit 0 when no check fails, 78 when at least one check needs a human, and 1 when doctor itself fails.

```json
{"ok":false,"checks":[
  {"id":"transcript.uv","component":"transcript","status":"pass","message":"uv found"},
  {"id":"auth.deepgram","component":"auth","status":"fail","code":"AUTH_MISSING","message":"No Deepgram key found.","hint":"ncly auth login deepgram"}
]}
```

`status` is `pass`, `warn`, or `fail`. A warning never changes the exit code.

## ncly auth (M1)

```
ncly auth login <service> [--stdin]
ncly auth logout <service> [--dry-run] [--force]
ncly auth status [--json]
```

| Service | Environment variable | Used by |
|---|---|---|
| `deepgram` | `DEEPGRAM_API_KEY` | `transcript` |
| `openrouter` | `OPENROUTER_API_KEY` | `transcript` profiles that call OpenRouter |

- `login` stores the key in the keychain. In non-interactive mode without `--stdin`, it exits 2 with `USAGE_INVALID` and the hint ``Run `ncly auth login <service>` in a terminal``. An agent relays this hint to the human and never pipes a key itself.
- `logout` removes the key from the keychain. It asks for confirmation, or needs `--force` in non-interactive mode.
- `status` lists every known service.

```json
{"services":[
  {"service":"deepgram","configured":true,"source":"keychain"},
  {"service":"openrouter","configured":false,"source":"none"}
]}
```

`source` is `env`, `keychain`, or `none`.

## ncly skill (M1)

```
ncly skill list [--json]
ncly skill show <name> [--json]
ncly skill <name>
```

- Skills come from the folders in `[skill] paths`. Each subfolder with a `SKILL.md` is a skill, named by the `name` key of its frontmatter. Taps add more sources in M3.
- When two folders hold the same skill name, the first folder in `paths` wins and `list` adds a warning.
- `list` returns names and descriptions, sorted by name.
- `show` prints the `SKILL.md` of the skill. Its first line gives the folder, so the relative paths inside the skill resolve: `<!-- skill-dir: /Users/me/code/skills/skills/transcript -->`. In interactive mode, Glamour renders the Markdown.
- A mode, such as `andy-mode`, is a skill like any other. `show` prints its router, which names its routes.
- `ncly skill <name>` is a shortcut for `show`, for humans. Agents use `show`. No skill may be named `list` or `show`.
- An unknown name exits 2 with `NOT_FOUND`.

```json
{"skills":[
  {"name":"transcript","description":"Use when the user invokes `transcript` or asks to transcribe a YouTube video or Zoom recording.","path":"/Users/me/code/skills/skills/transcript"}
]}
```

`show --json` returns `{"name":"…","path":"…","content":"…"}`.

## ncly transcript (M1)

Transcribes YouTube videos and Zoom recordings with Deepgram, then summarizes the result through a profile when asked.

```
ncly transcript run youtube --url <url> [<url>...] [--prompt <name>] [--profile <name>] [--dry-run] [--json]
ncly transcript run zoom (--latest | --path <folder>) [--profile <name>] [--dry-run] [--json]
ncly transcript list prompts [--json]
ncly transcript list profiles [--json]
ncly transcript list models --provider <provider> [--json]
```

- Other run options keep the names and defaults of the transcript CLI that moves into Nicely. `ncly transcript run youtube --help` lists them all.
- A run needs the `deepgram` key and exits 78 with `AUTH_MISSING` without it. A profile that calls OpenRouter also needs `openrouter`.
- `--dry-run` resolves the plan without keys, network calls, or writes.
- The output folder stays the default of the transcript component.
- Exit 75 is safe to rerun. Never rerun an exit 1 automatically, because Deepgram may already have billed the audio.
- When the summary fails, the transcript is still published, and the error object carries `output_dir`. A run with several URLs reports each one in a `results` array, with its own `url` and either `output_dir` or an error.
- In M1, transcript keeps its own profiles. M2 moves profiles to `ncly agent` and removes `list profiles` from transcript.

## Examples

```bash
# Check a new machine, then fix what doctor reports
ncly doctor
ncly auth login deepgram

# An agent discovers skills, then reads one
ncly skill list --json | jq -r '.skills[].name'
ncly skill show transcript

# Plan a transcription without keys, network calls, or writes
ncly transcript run youtube --url "https://www.youtube.com/watch?v=VIDEO_ID" --dry-run --json

# Transcribe, then keep only the result folder
ncly transcript run youtube --url "$URL" --json | jq -r .output_dir

# A missing key, seen by a script
ncly transcript run youtube --url "$URL" --json
# stderr: {"code":"AUTH_MISSING","message":"No Deepgram key found.","hint":"ncly auth login deepgram"}
# exit code: 78
```
