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
| `resume` | Continues a recorded run after checking which steps are safe |

`ncly --help` lists domains and commands with one line each. `ncly <domain> --help` lists the actions of that domain, and `ncly help <domain>` prints the same text. Every help page ends with two to five examples. `ncly` and `ncly <domain>` without a verb print their help on stdout and exit 0.

## Command tree

```
ncly
├── completion zsh|bash|fish        M0
├── describe [command...]           M1
├── doctor [component]              M1
├── auth login|logout|status        M1
├── run list|view|resume             M1
├── skill list|view                 M1
├── transcript run youtube|zoom     M1
└── transcript prompt list          M1
```

Core reserves every name in the domain table of the [guide](guide.md#domains-and-commands), plus `help`, `version`, and `config`. An extension never runs under a reserved name. A name that is neither a command of this version nor an installed extension fails with `USAGE_INVALID`, including a reserved name whose command arrives in a later milestone.

## Global flags (M0)

Each global flag other than `--help` and `--version` has an environment variable with the same effect. A flag wins over its variable. An `NCLY_` variable turns its flag on only when it equals `1`, and `NO_COLOR` turns color off when it is set and not empty, as [no-color.org](https://no-color.org/) defines it. [Programs that ncly runs](#programs-that-ncly-runs-m1) receive the resolved values.

| Flag | Variable | Default | Effect |
|---|---|---|---|
| `-h`, `--help` | | | Shows help on stdout and exits 0. It wins over every other argument, including unknown flags, but still honors `--lang` and `--no-color`. |
| `--version` | | | Prints `ncly vX.Y.Z` on stdout and exits 0. It works after any command and wins over every other argument except `--help`. |
| `--json` | `NCLY_JSON=1` | off | Machine output, as described in [Output](#output-m0). |
| `--no-input` | `NCLY_NO_INPUT=1` | off | Never prompts, even in a terminal. |
| `--no-color` | `NO_COLOR`, or `TERM=dumb` | off | Plain text. |
| `--lang <tag>` | `NCLY_LANG` | detected | Language of human text, such as `en` or `fr-CA`. |
| `-v`, `--verbose` | `NCLY_VERBOSE=1` | off | Adds step details on stderr. |

`NCLY_DEBUG=1` adds internals, timings, and child commands on stderr. It has no flag. Keys never appear in any output.

Every command that writes accepts `--dry-run`. The other flags below apply when the command confirms an action or produces files.

| Flag | Effect |
|---|---|
| `-n`, `--dry-run` | Uses the same preparation as execution and shows the planned effects and checks still pending. It changes no user file, config, key, grant, or persistent run record. It reads no key and sends no paid request. It may fill only Nicely's cache, including Python dependencies. |
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

stdout carries data. stderr carries progress, warnings, and errors. Without `--json`, human output uses styles, such as bold headings and colored labels, and `--no-color` turns them off. Deciding whether to style reads only the environment and the stream, and never queries the terminal or tmux. Styles appear only on a terminal, unless `CLICOLOR_FORCE=1` asks for them in a pipe. `NO_COLOR` and `TERM=dumb` turn them off even then.

With `--json`, every answer is one JSON object on one line, never a bare array. `ok` is `true` or `false` and agrees with the exit code. `contract_version` is the integer version of this public protocol, initially `1`. `--help`, `--version`, and `ncly completion` print text and ignore `--json`.

Parser failures use this envelope too. Machine mode is resolved from `NCLY_JSON` and a valid `--json` flag before reporting an invalid command, unknown flag, or missing value, wherever that flag occurs before `--`. A flag that takes a value takes the next argument even when it starts with `-`, as the parser does, so `ncly --lang --json` sets the language and leaves machine mode off. When a switch repeats, its last valid value counts, so `--json --json=bad` keeps machine mode for the parser failure. Words after `--` are operands, never flags or command names, so `ncly -- help` is `USAGE_INVALID` too. Help retains its documented precedence, including inside a group of shorthands such as `-zh`. Text after `=` is a flag value, so `-z=h` is an unknown flag, not a request for help.

- On success, stdout holds the object. Progress is not shown, and stderr stays empty unless `--verbose` or `NCLY_DEBUG` adds lines.
- On failure, stdout stays empty and the object ends stderr. With `--verbose` or `NCLY_DEBUG`, diagnostic lines come before it.
- When the output cannot be written, such as to a full disk or to a pipe whose reader closed, the command exits 1 with `RUNTIME` on stderr, even for `--help` and `--version`. A lost answer never counts as a success. When stderr refuses the answer too, the exit code stays 1, and the answer is lost.

```json
{"ok":false,"contract_version":1,"errors":[{"code":"AUTH_MISSING","message":"No Deepgram key found.","hint":"ncly auth login deepgram"}]}
```

`errors` describes the command's failure. The command computes its overall verdict first, then puts an error with the matching exit code first. When the failures of the command itself map to different exit codes, the most cautious one leads: `130` or `143`, then `1`, `78`, `2`, and `75` only when every failure is temporary. Input order or the order in which failures arrive never determines retry safety. Item errors remain inside `results`. `warnings` holds objects of the same shape, may appear in any answer, and never changes the exit code.

| Key | Meaning |
|---|---|
| `code` | A stable code from the [error registry](#error-codes-m0). Agents branch on this key. |
| `message` | Human text in the active language. Agents never parse it. |
| `hint` | The command or the action that fixes the problem. A command is never translated. |

A failure after partial work keeps the data keys that still apply, such as `output_dir`, next to `errors`. A recorded operation also returns `run_id`, including on failure. Large results are referenced by their artifact paths instead of copied into diagnostics or run listings.

A command that processes several items, such as several URLs or files, checks every input before it starts. A bad input stops the whole command with exit 2. Otherwise the answer holds `results`: one object per item, in input order, each with its own `ok` and either its data or its `errors`. Successful data remains available when another item fails. The command exits 0 when every item succeeded, 75 only when every failure is temporary and the whole invocation meets [Retry safety](#retry-safety-m0), and 1 otherwise. An incomplete batch uses the top-level `TEMPORARY` or `RUNTIME` code matching that verdict. On cancellation, 130 or 143 takes precedence and `results` contains only attempted items, in input order.

A report command, such as `ncly doctor` or `ncly auth status`, answers with a report. The report goes to stdout even when a check fails, and it lists its findings under its own key, such as `checks`, instead of `errors`. `ok` gives the verdict, and the exit code matches it. When the report command itself fails, it answers with `errors` like any other command.

## Compatibility (M0)

Within a contract version, command names, flags, error codes, exit codes, and the meaning of existing JSON remain stable. Compatibility covers key names, types, units, formats, required fields, absence versus `null`, enum values, and array ordering where specified. Object key order and translated human messages are not machine contracts.

- Readers ignore unknown optional object fields. Adding an optional field must preserve the behavior of an existing consumer
- Enums are closed unless their definition explicitly permits new values and specifies how readers handle an unknown value
- A breaking change needs a new contract version, a decision record, and a documented migration or continued support for the old contract before release. A decision record alone does not make a change compatible
- Internal Go/Python messages, extension manifests, and stored run records declare their own format versions. Readers reject unsupported versions before work instead of interpreting them as a known format
- M0 defines compatibility examples with independently stated expectations. Later milestones extend them with their own results and consumer cases

`NCLY_VERSION` remains the software release version. `NCLY_CONTRACT_VERSION` carries the public contract version to child programs. A command or mode appears in discovery only when the installed version supports it.

## Retry safety (M0)

Exit 75 is a promise that the same invocation is safe to repeat. It requires all of the following:

1. The failure is temporary
2. No paid request has been started, including a request whose response was lost
3. Every effect already produced is either harmless to repeat or covered by a demonstrated idempotent operation

Apply this rule to the whole invocation, including successful batch items and writes that happened before a lock failed. A free operation is not necessarily safe to repeat. Unknown effects, non-repeatable writes, and any started paid request exclude 75. Such runtime failures exit 1 and retain evidence for inspection or an explicit resume. Signals retain 130 and 143.

An idempotence claim includes destinations, configuration, external actions, and the result of a second invocation. Merely obtaining the same final file contents does not prove that no duplicate action occurred. Each writing command states what a repeated invocation does.

## Exit codes (M0)

| Code | Meaning | Safe to rerun |
|---|---|---|
| `0` | Success | |
| `1` | Runtime failure. Work may be partial or already billed. | No. Never rerun automatically. |
| `2` | Invalid invocation: unknown command, bad flag, missing value, or missing confirmation | After fixing the call |
| `75` | Temporary failure that meets every condition in Retry safety | Yes, with the same command |
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
| `CONFIG_INVALID` | 78 | A config file has a syntax error or a value of the wrong type, or cannot be read | M0 |
| `TERMINAL_REQUIRED` | 78 | A step needs a human at a terminal, such as typing or granting a key | M0 |
| `RUNTIME` | 1 | Any other failure during work | M0 |
| `NOT_FOUND` | 2 | The named skill, service, profile, or component does not exist | M1 |
| `AUTH_MISSING` | 78 | A required key is in neither the environment nor the keychain | M1 |
| `PREREQ_MISSING` | 78 | A required tool is absent or too old | M1 |
| `KEYRING_UNAVAILABLE` | 78 | A command needs a key that is not in the environment, and the OS keychain does not answer | M1 |
| `TEMPORARY` | 75 | A temporary failure for which the whole invocation meets Retry safety | M0 |
| `INTERRUPTED` | 130 | Ctrl-C stopped the command | M0 |
| `TERMINATED` | 143 | SIGTERM stopped the command | M0 |
| `RESUME_UNSAFE` | 1 | A run cannot safely continue because evidence is missing, changed, uncertain, or unsupported | M1 |
| `DEPTH_LIMIT` | 2 | Starting a harness would exceed the configured agent depth | M1 |
| `CAPABILITY_UNSUPPORTED` | 78 | The selected program cannot enforce a required mode or protocol | M1 |

Warnings use codes from this table.

| Code | When | Since |
|---|---|---|
| `CONFIG_UNKNOWN_KEY` | A config file holds a key that this version of `ncly` does not know | M0 |
| `SKILL_SHADOWED` | Two skill folders hold the same skill name | M1 |
| `SKILL_NO_SOURCE` | No skill folder is configured | M1 |
| `SKILL_NAME_MISMATCH` | A skill's subfolder name differs from the `name` in its `SKILL.md` | M1 |
| `CONFIG_DEFAULTS_USED` | Discovery used defaults because a config file could not be read | M1 |

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

Each path honors `XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`, or `XDG_CACHE_HOME` when it holds an absolute path, as the XDG specification asks, and ignores a relative one. macOS uses the same paths as Linux. `NCLY_CONFIG` names another shared config file, and the local config is then read from the same folder.

- A missing config file is not an error, and the defaults apply.
- A syntax error, a value of the wrong type, or a file that cannot be read exits 78 with `CONFIG_INVALID`, naming the file and, when the parser knows it, the line. `--help`, `--version`, `ncly completion`, and M1 command discovery still work with defaults. Discovery reports the fallback as specified below
- An unknown key is a `CONFIG_UNKNOWN_KEY` warning in `ncly doctor`, never an error, so an older `ncly` reads a config written for a newer one.
- `ncly` writes a config file only when the job of a command is to change the setup, such as adding a tap. It edits the file in place, keeps comments and formatting, and names the file it changed. Under `--dry-run`, it shows the change instead. Each such command states which file it writes.

The language comes from the first source that is set: `--lang`, `NCLY_LANG`, `lang` in the config, `LC_ALL`, `LC_MESSAGES`, `LANG`, and finally `en`. `ncly` picks the closest catalog, so `fr_CA.UTF-8` and `fr` both select `fr-CA` once that catalog exists. A language with no close catalog, including `C` and `POSIX`, falls back to `en` without an error. The pseudo-locale `en-XA` exists for tests: it marks every catalog string between `⟦` and `⟧`, and only an exact request selects it, so `en-GB` gets English.

Numbers follow the language, such as `1,234.5` in English and `1 234,5` in Canadian French. Sizes count 1000 bytes per kilobyte, as macOS does, and name their unit in the language, such as `1.5 MB` or `1,5 Mo`. Dates use ISO 8601, such as `2026-10-06`, in every language: English readers parse it, and it is the Canadian French standard.

```toml
lang = "en"

[skill]
paths = ["~/code/skills"]
```

## Command descriptions (M0)

One declaration per command supplies its argument and flag definitions, result contract, effects, prerequisites, examples, and supported modes. Help, completion, validation, `doctor`, and M1 discovery consume that declaration. Domain code owns the behavior. The declaration is not a workflow language.

Effects distinguish user writes, network access, and paid requests. Modes distinguish dry run, recorded execution, and resume. A missing declaration is not permission to assume a capability. A command may support recorded execution without supporting resume.

Discovery reads declarations and static extension manifests. It starts no harness or extension and reads no key. Detailed discovery is scoped to one command, and summary listings omit full schemas and skill bodies. Resolved configuration identifies the source file or environment variable for non-secret values. Secret values never appear.

## Operations (M0 contract, M1 execution)

The domain prepares an operation from resolved inputs and configuration, performs its steps, and verifies their results. Dry run uses that same preparation and reports pending checks, such as authentication. Execution repeats checks affected by changed inputs or destinations before applying effects. A dry run is not a reservation or a guarantee that the environment will stay unchanged.

Commands that incur a bill or can leave reusable partial work record an execution before the first such effect. M1 applies this to transcript. Read-only discovery and dry runs create no run record. Each domain declares its recording and resume support before code is written.

The shared operation support owns run IDs, durable records, and result publication. The domain owns its steps and the evidence that makes them resumable. `internal/run` owns child processes, their environment, timeouts, and signals. It does not decide whether a business action is safe to repeat.

### Records and evidence

A run record contains its format version, `run_id`, command path, status, step outcomes, and artifact references. It also keeps the non-secret input fingerprints, resolved profile and relevant tool versions needed to check a resume. Keep only information needed to explain or continue the operation. Full task bodies, transcripts, raw environment dumps, and keys are not copied into the record.

Persist an intent before an effect that cannot safely be repeated, and persist completion only after checking its result. If the process dies between those writes, the effect is unknown until evidence resolves it. A missing response is never proof that nothing happened. If the record cannot be written, stop before starting the next effect.

Records distinguish `running`, `completed`, `failed`, `interrupted`, and `unknown`. A step distinguishes `pending`, `completed`, `failed`, and `unknown`. Pending means the step has not started. A completed step names the evidence needed to reuse it, such as an artifact path and content fingerprint. The domain may leave an in-flight effect as unknown until it has a verified result.

Keep records under the XDG state directory and give every run its own temporary directory. Write each record atomically under a lock. Lock shared destinations or configuration at the point of mutation and revalidate while holding the lock. A conflict follows Retry safety, including effects already completed elsewhere in the run.

Records survive the process. M1 retains them until the user removes them and never automatically deletes output artifacts. Listings are bounded. A missing record or an unsupported record version never causes the original operation to be started again.

### Inspection and explicit resume

Inspection is read-only and reports saved evidence, not an assumption that a process marked running is still alive. Resume checks that no execution still owns the run, validates the record version, and verifies the saved inputs, relevant configuration, tool compatibility, and artifacts.

A supported resume reuses verified completed steps and continues only steps whose execution is known to be safe. It records each attempt under the same run ID. Changed or missing evidence returns `RESUME_UNSAFE` with an explanation and the artifacts still available. `--force` does not override uncertainty or make a non-repeatable effect safe.

For example, a saved transcript and a summary that was never started permit a resume of the summary alone. An unanswered paid summary request is unknown and is not sent again automatically. A domain may reconcile an external result when it has a reliable lookup mechanism. Otherwise the user must inspect the evidence and explicitly choose any new operation that might repeat a charge.

Nicely has no background worker or scheduler in these milestones. A record makes work inspectable after a session ends. It does not keep the process alive or promise that an arbitrary harness task can resume.

## ncly describe (M1)

```
ncly describe [command...] [--json]
```

Without a command path, returns concise descriptions of installed commands. A path such as `transcript run youtube` returns that command's declaration. The JSON envelope contains `commands`, an array. Each entry identifies its command `path`, `summary`, `since`, and `source`. A detailed entry adds arguments, flags, the result contract, effects, prerequisites, and supported modes. Text is localized, while machine identifiers follow Compatibility. Unknown paths use `NOT_FOUND`.

M1 settles the nested descriptor fields in this section before code. Discovery works with default configuration when a config file is invalid, adds `CONFIG_DEFAULTS_USED`, and does not claim to have resolved a profile that it could not read.

## ncly run (M1)

```
ncly run list [--limit <count>] [--json]
ncly run view <run-id> [--json]
ncly run resume <run-id> [--dry-run] [--force] [--timeout <duration>] [--json]
```

- `list` returns `runs`, newest first, with at most `--limit` entries, 20 by default. The count must be a positive integer
- `view` returns `run`, including step outcomes and artifact references. A successful inspection exits 0 even when the saved run failed. Its top-level `ok` describes the inspection, and `run.status` describes the saved operation
- `resume` follows Operations and returns the original domain's result with the same `run_id`. Dry run only describes the remaining steps and checks. The original command's timeout default applies unless overridden
- An unknown run uses `NOT_FOUND`. An unsafe or unsupported resume uses `RESUME_UNSAFE`. For a command that supports resume, a completed run returns its verified saved result without running its steps again

M1 settles the public run-summary and detail fields, record encoding, and lookup of a run after lost stdout before code. The invariants in Operations already apply to those choices.

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
- The variables of the global flags carry their resolved values. `NCLY_VERSION` and `NCLY_CONTRACT_VERSION` identify the release and public protocol.
- A harness receives `NCLY_AGENT_DEPTH` plus one.

**Grants.** A program of core receives the keys that core declares for it. An extension receives a key only after a human grants that key to that extension in a terminal. `--force` never grants a key. When an update of an extension declares a new key, the extension does not start until a human grants it.

**Cancellation.** On Ctrl-C or SIGTERM, `ncly` sends the signal to every program it started and to their descendants. It sends SIGKILL to what still runs 15 seconds later, which leaves a Python program its own 10 seconds to stop its children. It keeps every finished output and removes only its own temporary files. It exits 130 with `INTERRUPTED` or 143 with `TERMINATED`, and under `--json` the answer keeps the data keys that apply, such as `output_dir` or the `results` of finished items. A second signal during this cleanup is ignored. A SIGKILL sent to `ncly` itself leaves no time to clean up.

**Timeout.** The timeout covers the command's work and starts before child execution. On expiry, use the same process-tree cleanup as cancellation. A temporary failure before any paid request exits 75 only when Retry safety holds. Otherwise it exits 1. A caller's SIGTERM remains exit 143. Record finished work and unknown effects before returning when the process can still do so.

**Protocol failures.** Core Python programs return versioned machine messages only. Go maps them into this public contract. Missing, malformed, or incompatible messages never count as success. A launched program may already have produced effects, so preserve available evidence and use a runtime failure unless safety can be established. Harness adapters declare the modes they can actually enforce. An unsupported tools-off mode fails with `CAPABILITY_UNSUPPORTED` before launching a task with outside material.

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

The internal agent service checks `[agent] max_depth`, 2 by default, before every launch. `NCLY_AGENT_DEPTH` defaults to 0. A launch at or above the limit exits 2 with `DEPTH_LIMIT`. The child receives the depth plus one. This applies to transcript in M1 and the public agent command in M2.

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
- The `skill` component checks that each folder in `[skill] paths` exists and that each `SKILL.md` has a `name` and a `description`. It warns with `SKILL_NAME_MISMATCH` when a subfolder's name differs from its `name`.
- In interactive mode, when a tool is missing and its install command for this system is known, doctor shows the command, such as `brew install ffmpeg` or `sudo pacman -S ffmpeg`, and runs it only after a yes.
- In non-interactive mode, doctor never installs anything. The hint of each failed check holds the command.
- Doctor is a report command. It exits 0 when no check fails and 78 when at least one check needs a human. When doctor itself fails, it exits 1 with `errors`.

```json
{"ok":false,"contract_version":1,"checks":[{"id":"transcript.uv","component":"transcript","status":"pass","message":"uv found"},{"id":"auth.deepgram","component":"auth","status":"fail","code":"AUTH_MISSING","message":"No Deepgram key found.","hint":"ncly auth login deepgram"}]}
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

For `login` and `logout`, `--dry-run` takes precedence over secret input and confirmation. It describes the intended action without reading stdin for a key, opening a keychain entry, or prompting. Checks that would require a keychain read, including whether login would replace a key, are reported as pending. This simulation does not require a terminal or `--stdin`.

```json
{"ok":true,"contract_version":1,"keychain":"available","services":[{"service":"deepgram","configured":true,"source":"keychain"}]}
```

`keychain` is `available` or `unavailable`. `source` is `env`, `keychain`, or `none`. To move a key from another secret store, pipe it once: `<command that prints the key> | ncly auth login deepgram --stdin`.

## ncly skill (M1)

```
ncly skill list [--json]
ncly skill view <name> [--json]
ncly skill <name>
```

- Skills come from the folders in `[skill] paths`. Each subfolder with a `SKILL.md` is a skill, named by the `name` key of its frontmatter. Taps add more sources in M3.
- A skill's subfolder takes the skill's name, as the [Agent Skills specification](https://agentskills.io/specification) requires. When the names differ, the skill still works and `ncly doctor skill` warns.
- A subfolder may be a symbolic link to a folder elsewhere. Harness skill folders often hold such links, and `skill link` creates them from M3. A broken link is not a skill, and `list` skips it.
- When two folders hold the same skill name, the first folder in `paths` wins and `list` adds a `SKILL_SHADOWED` warning. Two subfolders that resolve to the same folder are one skill, listed with the path of the first, without a warning.
- When no folder is configured, `list` returns an empty list and a `SKILL_NO_SOURCE` warning whose hint names `[skill] paths`.
- `list` returns names and descriptions, sorted by name.
- `view` prints the `SKILL.md` of the skill. Its first line gives the folder, so the relative paths inside the skill resolve: `<!-- skill-dir: /Users/me/code/skills/transcript -->`. In interactive mode, Glamour renders the Markdown.
- A mode, such as `andy-mode`, is a skill like any other. `view` prints its router, which names its routes.
- `ncly skill <name>` is a shortcut for `view`, for humans. Agents use `view`. No skill may take the name of a `skill` verb.
- An unknown name exits 2 with `NOT_FOUND`.

```json
{"ok":true,"contract_version":1,"skills":[{"name":"transcript","description":"Use when the user invokes `transcript` or asks to transcribe a YouTube video or Zoom recording.","path":"/Users/me/code/skills/transcript"}]}
```

`view --json` returns `{"ok":true,"contract_version":1,"name":"…","path":"…","content":"…"}`.

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
- Exit 75 follows Retry safety, including any local writes. Never rerun an exit 1 automatically, because work may be partial or already billed
- Go owns the summary step from M1 and uses the internal agent service. Python stops at the transcript and returns its verified artifacts
- A recorded run returns `run_id`. When the summary fails, the transcript stays saved and the failure keeps `output_dir` next to `errors`. Resume reuses the transcript and runs the summary only when Operations permits it
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
# stderr: {"ok":false,"contract_version":1,"errors":[{"code":"AUTH_MISSING","message":"No Deepgram key found.","hint":"ncly auth login deepgram"}]}
# exit code: 78
```
