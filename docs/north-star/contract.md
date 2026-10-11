# Contract

This file is the contract that every command follows, core's and each extension's: what humans and agents type, what core and an extension exchange, and what scripts parse. Core's commands live in [core-spec.md](core-spec.md), and each extension's in its own spec, such as [transcript's](../../extensions/transcript/spec.md). The code tables name the milestone that brings each entry. A milestone becomes ready only once the sections that its cards read are written, here, in core-spec.md, or in its extension's spec.

## Primitives

Every command and every extension composes these primitives. Each belongs to one or both scopes of the contract: the agent contract, what a caller of the CLI types and parses, and the wire contract, what core and an extension exchange, as [D040](decisions/D040-wire-contract.md) decides. Each primitive is defined once, in the section that the table links.

| Primitive | Scope | Defined in |
|---|---|---|
| Domain, resource, and verb | Agent contract | [Usage](#usage) |
| Flag | Agent contract | [Global flags](#global-flags) |
| Manifest | Wire contract | [Extensions](#extensions) and [Command descriptions](#command-descriptions) |
| Environment | Wire contract | [Programs that ncly runs](#programs-that-ncly-runs) |
| Answer | Both | [Output](#output) |
| Exit code | Both | [Exit codes](#exit-codes) |
| Run record | Wire contract | [Record format](#record-format) |

## Usage

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

## Global flags

Each global flag other than `--help` and `--version` has an environment variable with the same effect. A flag wins over its variable. An `NCLY_` variable turns its flag on only when it equals `1`, and `NO_COLOR` turns color off when it is set and not empty, as [no-color.org](https://no-color.org/) defines it. [Programs that ncly runs](#programs-that-ncly-runs) receive the resolved values.

| Flag | Variable | Default | Effect |
|---|---|---|---|
| `-h`, `--help` | | | Shows help on stdout and exits 0. It wins over every other argument, including unknown flags, but still honors `--lang` and `--no-color`. |
| `--version` | | | Prints `ncly vX.Y.Z` on stdout and exits 0. It works after any command and wins over every other argument except `--help`. |
| `--json` | `NCLY_JSON=1` | off | Machine output, as described in [Output](#output). |
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

## Modes

Interactive mode applies when stdin and stdout are terminals, `--no-input` is absent, `NCLY_NO_INPUT` is not `1`, and `CI` is not set. Every other case is non-interactive mode. Some agent tools run commands in a pseudo-terminal, so an agent sets `NCLY_NO_INPUT=1` in its environment.

| Situation | Interactive mode | Non-interactive mode |
|---|---|---|
| Every required value is given | Runs | Runs |
| A required value is missing | A form asks only for the missing values | Exit 2, `USAGE_INVALID`, and a hint with the full command |
| A step needs a confirmation and `--force` is absent | Asks for confirmation | Exit 2, `CONFIRMATION_REQUIRED`, and a hint that adds `--force` |
| A step only a human can do, such as typing or granting a key | Asks | Exit 78, `TERMINAL_REQUIRED`, and a hint that names the command to run in a terminal |

After an interactive run that used a form, `ncly` prints the equivalent command on stderr, after the line `Next time:`. The command itself is never translated.

## Output

stdout carries data. stderr carries progress, warnings, and errors. Without `--json`, human output uses styles, such as bold headings and colored labels, and `--no-color` turns them off. Deciding whether to style reads only the environment and the stream, and never queries the terminal or tmux. Styles appear only on a terminal, unless `CLICOLOR_FORCE=1` asks for them in a pipe. `NO_COLOR` and `TERM=dumb` turn them off even then.

With `--json`, every answer is one JSON object on one line, never a bare array. `ok` is `true` or `false` and agrees with the exit code. `contract_version` is the integer version of this public protocol, initially `1`. `--help`, `--version`, and `ncly completion` print text and ignore `--json`.

Parser failures use this envelope too. Machine mode is resolved from `NCLY_JSON` and a valid `--json` flag before reporting an invalid command, unknown flag, or missing value, wherever that flag occurs before `--`. A flag that takes a value takes the next argument even when it starts with `-`, as the parser does, so `ncly --lang --json` sets the language and leaves machine mode off. When a switch repeats, its last valid value counts, so `--json --json=bad` keeps machine mode for the parser failure. Words after `--` are operands, never flags or command names, so `ncly -- help` is `USAGE_INVALID` too. Help retains its documented precedence, including inside a group of shorthands such as `-zh`. Text after `=` is a flag value, so `-z=h` is an unknown flag, not a request for help.

- On success, stdout holds the object. Progress is not shown, and stderr stays empty unless `--verbose` or `NCLY_DEBUG` adds lines.
- On failure, stdout stays empty and the object ends stderr. With `--verbose` or `NCLY_DEBUG`, diagnostic lines come before it.
- When the output cannot be written, such as to a full disk or to a pipe whose reader closed, the command exits 1 with `RUNTIME` on stderr, even for `--help` and `--version`. A lost answer never counts as a success. When stderr refuses the answer too, the exit code stays 1, and the answer is lost.

```json
{"ok":false,"contract_version":1,"errors":[{"code":"AUTH_MISSING","message":"No Deepgram key found.","hint":"ncly auth login deepgram"}]}
```

`errors` describes the command's failure. The command computes its overall verdict first, then puts an error with the matching exit code first. When the failures of the command itself map to different exit codes, the most cautious one leads: `130` or `143`, then `1`, `78`, `2`, and `75` only when every failure is temporary. When several failures share that exit code, the first one in this order leads. For exit 1, `RESUME_UNSAFE`, then the answering extension's own codes in the order of its spec, then `RUNTIME`. For exit 78, `CONFIG_INVALID`, `TERMINAL_REQUIRED`, `PREREQ_MISSING`, `CAPABILITY_UNSUPPORTED`, `KEYRING_UNAVAILABLE`, `AUTH_MISSING`, `KEY_NOT_GRANTED`, `AUTH_REJECTED`, then the extension's own codes, so the first fix that a human needs leads. For exit 2, `USAGE_INVALID`, `CONFIRMATION_REQUIRED`, `NOT_FOUND`, then the extension's own codes. Input order or the order in which failures arrive never determines retry safety. Item errors remain inside `results`. `warnings` holds objects of the same shape, may appear in any answer, and never changes the exit code.

| Key | Meaning |
|---|---|
| `code` | A stable code from the [error registry](#error-codes). Agents branch on this key. |
| `message` | Human text in the active language. Agents never parse it. |
| `hint` | The command or the action that fixes the problem. A command is never translated. |

A failure after partial work keeps the data keys that still apply, such as `output_dir`, next to `errors`. A recorded operation also returns `run_id`, including on failure. Large results are referenced by their artifact paths instead of copied into diagnostics or run listings.

A command that processes several items, such as several URLs or files, checks every input before it starts. A bad input stops the whole command with exit 2. Otherwise the answer holds `results`: one object per item, in input order, each with its own `ok` and either its data or its `errors`. Successful data remains available when another item fails. The command exits 0 when every item succeeded, 75 only when every failure is temporary and the whole invocation meets [Retry safety](#retry-safety), and 1 otherwise. An incomplete batch uses the top-level `TEMPORARY` or `RUNTIME` code matching that verdict. On cancellation, 130 or 143 takes precedence and `results` contains only attempted items, in input order.

A report command, such as `ncly doctor` or `ncly auth status`, answers with a report. The report goes to stdout even when a check fails, and it lists its findings under its own key, such as `checks`, instead of `errors`. `ok` gives the verdict, and the exit code matches it. When the report command itself fails, it answers with `errors` like any other command.

## Compatibility

Within a contract version, command names, flags, error codes, exit codes, and the meaning of existing JSON remain stable. Compatibility covers key names, types, units, formats, required fields, absence versus `null`, enum values, and array ordering where specified. Object key order and translated human messages are not machine contracts.

- Readers ignore unknown optional object fields. Adding an optional field must preserve the behavior of an existing consumer
- Enums are closed unless their definition explicitly permits new values and specifies how readers handle an unknown value
- A breaking change needs a new contract version, a [decision](decisions/README.md), and a documented migration or continued support for the old contract before release. A decision alone does not make a change compatible
- Internal Go/Python messages, extension manifests, and stored run records declare their own format versions. Readers reject unsupported versions before work instead of interpreting them as a known format
- M00 defines compatibility examples with independently stated expectations. Later milestones extend them with their own results and consumer cases

`NCLY_VERSION` remains the software release version. `NCLY_CONTRACT_VERSION` carries the public contract version to child programs. A command or mode appears in discovery only when the installed version supports it.

## Retry safety

Exit 75 is a promise that the same invocation is safe to repeat. It requires all of the following:

1. The failure is temporary
2. No paid request has been started, including a request whose response was lost
3. Every effect already produced is either harmless to repeat or covered by a demonstrated idempotent operation

Apply this rule to the whole invocation, including successful batch items and writes that happened before a lock failed. A free operation is not necessarily safe to repeat. Unknown effects, non-repeatable writes, and any started paid request exclude 75. Such runtime failures exit 1 and retain evidence for inspection or an explicit resume. Signals retain 130 and 143.

In a recorded operation, the record settles conditions 2 and 3: exit 75 requires every step that has left `pending` to declare the `none` or `repeatable` effect, as [Record format](#record-format) defines. A `failed` step counts too, because it may have made part of its effect, such as a reserved result folder before a refused upload.

An idempotence claim includes destinations, configuration, external actions, and the result of a second invocation. Merely obtaining the same final file contents does not prove that no duplicate action occurred. Each writing command states what a repeated invocation does.

## Exit codes

| Code | Meaning | Safe to rerun |
|---|---|---|
| `0` | Success | |
| `1` | Runtime failure. Work may be partial or already billed. | No. Never rerun automatically. |
| `2` | Invalid invocation: unknown command, bad flag, missing value, or missing confirmation | Fix the call. Inspect any returned `run_id` before continuing recorded work |
| `75` | Temporary failure that meets every condition in Retry safety | Yes, with the same command |
| `78` | A human must act: a key, a prerequisite, a config fix, or a step at a terminal | Relay the hint. Inspect any returned `run_id` before continuing recorded work |
| `130` | Interrupted by Ctrl-C | |
| `143` | Terminated by SIGTERM | |

Codes 124 to 127, and every code from 128 up other than 130 and 143, stay reserved for the shell.

Only the overall process exit 75 permits an automatic repeat of the same invocation. An item error describes that item's cause, not the safety of repeating the whole command. A caller retains the complete answer and `run_id` before extracting artifact paths. After fixing a recorded failure, it inspects the run and requests an explicit resume instead of repeating the original command.

## Error codes

Each code maps to exactly one exit code. This table holds core's codes, and a milestone adds core's new codes here. Every extension may answer with these codes too. An extension's own codes start with its domain, such as `TRANSCRIPT_DOWNLOAD_FAILED` or `FLEET_HOST_DOWN`, live in its spec, and map to one exit code each, with the meaning that this file defines.

| Code | Exit | When | Since |
|---|---|---|---|
| `USAGE_INVALID` | 2 | Unknown command, bad flag, or a missing value in non-interactive mode | M00 |
| `CONFIRMATION_REQUIRED` | 2 | A step needs a confirmation in non-interactive mode, and `--force` is absent | M00 |
| `CONFIG_INVALID` | 78 | A config file has a syntax error or a value of the wrong type, or cannot be read | M00 |
| `TERMINAL_REQUIRED` | 78 | A step needs a human at a terminal, such as typing or granting a key | M00 |
| `RUNTIME` | 1 | Any other failure during work | M00 |
| `TEMPORARY` | 75 | A temporary failure for which the whole invocation meets Retry safety | M00 |
| `INTERRUPTED` | 130 | Ctrl-C stopped the command | M00 |
| `TERMINATED` | 143 | SIGTERM stopped the command | M00 |
| `NOT_FOUND` | 2 | The named skill, service, profile, component, run, prompt, or recording does not exist | M03 |
| `PREREQ_MISSING` | 78 | A required tool is absent or too old | M05 |
| `AUTH_MISSING` | 78 | A required key is in neither the environment nor the keychain | M06 |
| `KEYRING_UNAVAILABLE` | 78 | The OS keychain does not answer when a command needs it: to read a key missing from the environment, for `ncly auth`, or for the `auth` check of `ncly doctor` | M06 |
| `KEY_NOT_GRANTED` | 78 | An extension needs a key that no human granted to it | M08 |
| `CAPABILITY_UNSUPPORTED` | 78 | The selected program cannot enforce a required mode or protocol | M09 |
| `AUTH_REJECTED` | 78 | A service refused a key, such as Deepgram answering 401 | M13 |
| `RESUME_UNSAFE` | 1 | A run cannot safely continue because evidence is missing, changed, uncertain, or unsupported | M15 |

Warnings use codes from this table, and an extension's own warnings start with its domain, such as `SKILL_SHADOWED`.

| Code | When | Since |
|---|---|---|
| `CONFIG_UNKNOWN_KEY` | A config file holds a key that this version of `ncly` does not know | M00 |
| `CONFIG_DEFAULTS_USED` | Discovery used defaults because a config file could not be read | M04 |

## Configuration

`config.toml` is the shared config: the setup the user wants, which can travel between machines, for example through dotfiles. `config.local.toml`, in the same folder, is the local config: what belongs to this machine only.

Core owns the top-level keys, such as `lang`. Each extension owns the table named after its domain, such as `[skill]` or `[agent]`, and its spec defines that table.

Precedence, highest first: flags, environment variables, `config.local.toml`, `config.toml`, then defaults. A table, such as `[agent.profiles.everyday]`, merges key by key across the files. A list, such as `[skill] paths`, from a higher source replaces the whole list below it.

| Purpose | Path |
|---|---|
| Shared config | `~/.config/nicely/config.toml` |
| Local config | `~/.config/nicely/config.local.toml` |
| Data: taps and extensions | `~/.local/share/nicely/` |
| State: logs, locks, and run history | `~/.local/state/nicely/` |
| Cache: extracted programs, such as transcript's Python program | `~/.cache/nicely/` |

Each path honors `XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`, or `XDG_CACHE_HOME` when it holds an absolute path, as the XDG specification asks, and ignores a relative one. macOS uses the same paths as Linux. `NCLY_CONFIG` names another shared config file, and the local config is then read from the same folder.

- A missing config file is not an error, and the defaults apply.
- A syntax error, a value of the wrong type, or a file that cannot be read exits 78 with `CONFIG_INVALID`, naming the file and, when the parser knows it, the line. `--help`, `--version`, `ncly completion`, and discovery still work with defaults. Discovery reports its fallback. Doctor reports this failure on stdout as a failed `core.config` check and uses defaults only for checks independent of the invalid configuration
- An unknown key is a `CONFIG_UNKNOWN_KEY` warning in `ncly doctor`, never an error, so an older `ncly` reads a config written for a newer one.
- `ncly` writes a config file only when the job of a command is to change the setup, such as adding a tap. It edits the file in place, keeps comments and formatting, and names the file it changed. Under `--dry-run`, it shows the change instead. Each such command states which file it writes.

The language comes from the first source that is set: `--lang`, `NCLY_LANG`, `lang` in the config, `LC_ALL`, `LC_MESSAGES`, `LANG`, and finally `en`. `ncly` picks the closest catalog, so `fr_CA.UTF-8` and `fr` both select `fr-CA` once that catalog exists. A language with no close catalog, including `C` and `POSIX`, falls back to `en` without an error. The pseudo-locale `en-XA` exists for tests: it marks every catalog string between `⟦` and `⟧`, and only an exact request selects it, so `en-GB` gets English.

Numbers follow the language, such as `1,234.5` in English and `1 234,5` in Canadian French. Sizes count 1000 bytes per kilobyte, as macOS does, and name their unit in the language, such as `1.5 MB` or `1,5 Mo`. Dates use ISO 8601, such as `2026-10-06`, in every language: English readers parse it, and it is the Canadian French standard.

```toml
lang = "en"

[skill]
paths = ["~/code/skills"]
```

## Command descriptions

One declaration per command supplies help, completion, validation, and discovery. M00 implements the fields that its commands use. Each later command adds input constraints, result types, prerequisites, and configuration sources as it needs them. An extension's manifest holds the same declarations, so help, completion, validation, and discovery treat every command alike. Domain code owns the behavior. The declaration is not a workflow language.

Effects distinguish user writes, network access, and paid requests. Modes distinguish dry run, recorded execution, and resume. A missing declaration is not permission to assume a capability. A command may support recorded execution without supporting resume.

[ncly describe](core-spec.md#ncly-describe) lists the fields of a declaration. Discovery reads declarations and static extension manifests. It starts no harness or extension and reads no key. Detailed discovery is scoped to one command, and summary listings omit full schemas and skill bodies. Resolved configuration identifies the source file or environment variable for non-secret values. Secret values never appear.

## Extensions

Core is a small host. Every domain that does work for the user is an extension, as [D004](decisions/D004-minimal-core.md) decides. An extension comes in one of two transports, with one manifest format and one contract:

- A **bundled** extension is a Go package compiled into `ncly`, with its manifest embedded. Only what an agent needs to operate Nicely is bundled: `skill`, in M03
- An **external** extension is an executable named `ncly-<domain>`, with its manifest `ncly-<domain>.toml` beside it, found in `~/.local/share/nicely/extensions/` first, then in `PATH` order, and later brought by taps. The first-party ones, such as `agent` and `transcript`, live in `extensions/<name>/` in the Nicely repository, which is their official tap

Core commands and bundled extensions always win over an external extension with the same name. Discovery reports the source it selected and every extension that another one hides. Dispatch requires a manifest with a compatible protocol, and core reads it without starting the extension.

An extension receives the global flags as the variables of [Global flags](#global-flags), plus the environment that [Programs that ncly runs](#programs-that-ncly-runs) defines. It answers with the envelope of [Output](#output) and the exit codes of [Exit codes](#exit-codes). It may use core's [error codes](#error-codes), and its own codes start with its domain. It translates its own text into the language of `NCLY_LANG`. It may be written in any language: a Python extension can be one file with the shebang `#!/usr/bin/env -S uv run --script` and its dependencies inline. Under `--json`, an answer that is not valid JSON, has another `contract_version`, or disagrees with its exit code is a protocol failure, never a success.

An extension uses another one through the CLI, as an agent would: transcript summarizes with `ncly agent run --json`. It writes its run records in the shared format of [Record format](#record-format), through the SDK for a Go extension, so `ncly run` inspects every recorded operation alike. A Go extension imports only the public `sdk/` packages, so a bundled extension could leave the binary without a rewrite.

Each extension keeps its spec in `extensions/<name>/spec.md`, and ships a `SKILL.md` that teaches an agent to use it. [M01](../milestones/M01-extension-host.md) settles the manifest schema and the dispatch protocol here, and [M02](../milestones/M02-sdk.md) the SDK.

## Operations

The domain prepares an operation from resolved inputs and configuration, performs its steps, and verifies their results. Dry run uses that same preparation and reports pending checks, such as authentication. Execution repeats checks affected by changed inputs or destinations before applying effects. A dry run is not a reservation or a guarantee that the environment will stay unchanged.

Commands that incur a bill or can leave reusable partial work record an execution before the first such effect. The first are `ncly agent run` and transcript. Read-only discovery and dry runs create no run record. Each domain declares its recording and resume support before code is written.

The SDK's record support owns run IDs, durable records, step transitions, and result publication, for core and every Go extension. It persists each step's intent, which turns the step `unknown`, and its final status, in the order that [Records and evidence](#records-and-evidence) requires, and derives [Retry safety](#retry-safety) from the effects that the steps declare, so a domain never states retry safety a second time. The domain names its steps, declares the effect of each, verifies their results, and owns the evidence that makes them resumable. The program runner owns child processes, their environment, timeouts, and signals. It does not decide whether a business action is safe to repeat.

### Records and evidence

A run record contains its format version, `run_id`, command path, status, step outcomes, and artifact references. It also keeps the non-secret input fingerprints, resolved profile and relevant tool versions needed to check a resume. Keep only information needed to explain or continue the operation. Full task bodies, transcripts, raw environment dumps, and keys are not copied into the record.

Every effect of a recorded operation happens inside a step. Persist an intent before an effect that cannot safely be repeated, and persist completion only after checking its result. If the process dies between those writes, the effect is unknown until evidence resolves it. A missing response is never proof that nothing happened. If the record cannot be written, stop before starting the next effect.

Records distinguish `running`, `completed`, `failed`, `interrupted`, and `unknown`. A step distinguishes `pending`, `completed`, `failed`, and `unknown`. Pending means its effect has not started; preparation may have failed. A completed step names the evidence needed to reuse it, such as an artifact path and content fingerprint. The domain may leave an in-flight effect as unknown until it has a verified result.

Keep records under the XDG state directory and give every run its own temporary directory. Write each record atomically under a lock. Lock shared destinations or configuration at the point of mutation and revalidate while holding the lock. A conflict follows Retry safety, including effects already completed elsewhere in the run.

Records survive the process. Nicely retains them until the user removes their JSON files and never automatically deletes output artifacts. Listings bound both record count and item-key previews. A missing record or an unsupported record version never causes the original operation to be started again.

### Record format

Each run has one record, `runs/<run_id>.json` in the state directory, such as `~/.local/state/nicely/runs/`. The record is one JSON object. While holding `runs/<run_id>.lock`, the writer, core or an extension through the SDK, writes a temporary file in the same directory, syncs that file, renames it, then syncs the parent directory. Newly created state directories are synced with their parents too. This persistence barrier finishes before acknowledging a non-repeatable effect. A failed barrier prevents the effect. Atomic visibility alone is not durable acknowledgement.

The platform helper uses the durable flush supported by the system. On macOS, it also uses `F_FULLFSYNC` after syncing metadata, because [Apple's fsync documentation](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/fsync.2.html) distinguishes an ordinary sync from flushing the drive cache. A failed required flush stops the operation. Persistence relies on the filesystem and device honoring these operations; Nicely cannot promise survival of arbitrary hardware failure.

The execution process holds the run lock until it exits. A new operation that consumes a source run also holds that source's lock while it executes, to exclude a concurrent source resume. The operating system releases each lock when its owner dies. Lock files may remain; their existence alone does not establish ownership. Before recording a step as completed, the domain independently verifies its artifacts and syncs their files and containing directories.

`run_id` is an opaque string of lowercase letters, digits, and hyphens, such as `20261007-214501-3f9a2c`. Readers never parse it. A run's temporary directory sits in the system temporary directory, carries the run ID in its name, and is removed when the run ends, including after Ctrl-C or SIGTERM.

| Key | Type | Meaning |
|---|---|---|
| `format_version` | integer | The record format, initially `1`. A reader rejects any other value before work |
| `run_id` | string | The run ID |
| `path` | string | The command path, such as `transcript run youtube` |
| `status` | string | `running`, `completed`, `failed`, `interrupted`, or `unknown` |
| `started_at`, `updated_at` | string | UTC times in RFC 3339, such as `2026-10-07T21:45:01Z` |
| `ncly_version` | string | The release that last wrote the record |
| `attempts` | integer | The number of executions: 1, plus one per resume |
| `inputs` | object | The non-secret inputs and their fingerprints, under keys that the domain defines |
| `tools` | object | The version of each program that the run used, keyed by program name |
| `items` | array | One object per input item, in input order. A command with one input has one item |

Each item has `key`, a string that names the item, such as its URL, `output_dir`, the folder selected for its results, or `null` until the domain selects it, and `steps`, an array in execution order. The domain records `output_dir` before creating or writing that folder. Each step has these keys:

| Key | Type | Meaning |
|---|---|---|
| `id` | string | The step name that the domain defines, such as `transcribe` |
| `effect` | string | The strongest effect that the step can make, which the domain declares in the initial record and never changes: `none`; `repeatable`, harmless to repeat or covered by a demonstrated idempotent operation; `non_repeatable`, which a repeat would duplicate or conflict with; or `paid`, a request that may bill |
| `status` | string | `pending`, `completed`, `failed`, or `unknown` |
| `started_at`, `finished_at` | string or null | UTC times, `null` until the step starts or ends |
| `artifacts` | array | One object per verified output, with `path`, `size_bytes`, and `sha256` |
| `error` | object or null | The latest unsuccessful attempt's `code`, `message`, and `hint`, as in the public problem object. A preparation failure can leave the step `pending` with an error. An untouched pending or completed step has `null` |
| `evidence` | object, optional | Data beyond `artifacts` that the domain needs to reuse the step, with its own `format_version`. A step without such data omits the key |

The domain defines when its validated preparation creates the initial record, before its first user-side mutation or paid dispatch. It starts with pending steps and status `running`. A run's saved status follows this table:

| Status | When written |
|---|---|
| `running` | The initial record and each explicit resume attempt |
| `completed` | Every planned step has verified completion |
| `failed` | Execution ends with a known error and no unknown step |
| `interrupted` | Ctrl-C or SIGTERM stops execution, preserving each step's evidence |
| `unknown` | Execution ends without a signal and at least one step has an unresolved effect |

Abrupt death can leave `running` with no lock owner. Inspection preserves that saved status and reports `active: false`; it never invents completion or marks untouched pending steps unknown. Resume decides from the step evidence.

Each extension reuses these keys where they fit and defines its own `inputs` and evidence in its spec. Core reads records to list, show, and hand over a resume, without interpreting an extension's evidence, and without a workflow engine. Compatibility governs any later format change.

### Inspection and explicit resume

Inspection is read-only and reports saved evidence, not an assumption that a process marked running is still alive. Resume checks that no execution still owns the run, validates the record version, and verifies the saved inputs, relevant configuration, tool compatibility, and artifacts. For `ncly run resume`, the relevant configuration excludes the profile, because resume uses the profile that the run recorded.

A supported resume reuses verified completed steps and continues only steps whose execution is known to be safe. It records each attempt under the same run ID. Changed or missing evidence returns `RESUME_UNSAFE` with an explanation and the artifacts still available. `--force` does not override uncertainty or make a non-repeatable effect safe.

For example, a saved transcript and a summary that was never started permit a resume of the summary alone. An unanswered paid summary request is unknown and is not sent again automatically. A domain may resolve an unknown step from complete, independently verified local artifacts, or from a reliable external lookup. This establishes a reusable result; it never authorizes redispatch of an unresolved request. Otherwise the user must inspect the evidence and explicitly choose any new operation that might repeat a charge.

Nicely has no background worker or scheduler. A record makes work inspectable after a session ends. It does not keep the process alive or promise that an arbitrary harness task can resume.

### Dry-run answer

A recorded operation answers `--dry-run --json` with `dry_run`, set to `true`, then `plan` and `pending_checks`. `ncly run resume` and the recorded commands of extensions, such as [transcript](../../extensions/transcript/spec.md#dry-run-examples), use this shape.

| Key | Type | Meaning |
|---|---|---|
| `plan` | array | One object per item, in input order, with `key`, `output` or `output_dir`, `steps`, and `effects` |
| `plan[].steps` | array | Each step that execution would reach, with `id` and `action`, which is `run` or `reuse` |
| `plan[].effects` | object | The booleans `user_writes`, `network`, and `paid`, for the item's `run` steps |
| `pending_checks` | array | Each check that a `run` step needs and that only execution can make, with `id`, such as `auth.deepgram`, and `message`. A dry run reads no key, so a `run` step that needs a key keeps its authentication pending. The array is empty when no `run` step needs such a check |

A new item has `output`, the output folder that will hold its result folder. An item that already has a result folder, such as in `run resume`, has `output_dir`. A resume that execution would refuse fails its dry run with the same `RESUME_UNSAFE`.

## Keys

- A service has a name of lowercase letters and digits, such as `deepgram`. Its environment variable is the name in uppercase followed by `_API_KEY`, such as `DEEPGRAM_API_KEY`.
- Core reads a key from its environment variable first. It reads the key from the keychain only when that variable is empty, and only for a command whose declaration needs the key and an extension that a human granted it to, as [Grants](#grants) defines. A dry run reads no key. `ncly auth` and `ncly doctor` also check the keychain. The keychain entry uses the service `nicely` and the account named after the service.
- Each keychain call has a time limit: 60 seconds in interactive mode, where a human can answer an unlock prompt, and 10 seconds otherwise. A keychain that does not answer in time counts as unavailable. A command that needs a key exits 78 with `KEYRING_UNAVAILABLE` when the key is missing from the environment and the keychain is unavailable, and its hint names the variable to export. When the keychain answers without the key, the code is `AUTH_MISSING`.
- A key is never accepted as a flag. `ncly auth login` reads it from a masked field, or from stdin with `--stdin`.
- A key is never printed. `ncly auth status` reports where a key comes from, never its value.

## Grants

A key reaches an extension only after a human grants that key to that extension in a terminal, as [D030](decisions/D030-trusted-extensions.md) decides. `--force` never grants a key. Grants stay on one machine, in local state, and bind the key to the extension's domain and source. An extension from another source inherits no grant through its name. An update that declares a new key gets nothing until a human grants it. Core never refuses to start an extension over a key, because only the extension knows whether this invocation needs it: a summary-only resume of transcript needs no Deepgram key. The extension answers `AUTH_MISSING`, `KEYRING_UNAVAILABLE`, or `KEY_NOT_GRANTED` when it lacks a key it needs. Grants are consent and protection against accidental leaks, not isolation.

In a nested call, such as transcript calling `ncly agent run`, core passes the keys of the inner extension from the keychain, or from the environment that reached it. A key that exists only in the outer shell's environment does not cross an extension that was not granted it, so the harness key of a nested agent run comes from the keychain, or the harness uses its own login. [M08](../milestones/M08-grants.md) settles the grant command and its record here.

## Programs that ncly runs

Core runs extensions and doctor's version probes. An extension runs its own programs, such as transcript's Python program or a harness that the `agent` extension starts, through the SDK's program runner, which follows the same rules. None of them runs in a sandbox. A program runs with the rights of the user, so it could read the keychain or any file itself. Adding a tap or installing an extension means trusting its author.

**Environment.** A program receives the environment of its parent with these changes:

- Each variable named `<NAME>_API_KEY` is the key of a service. An extension receives it only when a human granted that key to it, and core then sets it from the environment or the keychain. An extension passes a key on only to the program that its spec names, such as `DEEPGRAM_API_KEY` to transcript's Python program
- The variables of the global flags carry their resolved values. `NCLY_VERSION` and `NCLY_CONTRACT_VERSION` identify the release and public protocol.
- A harness receives `NCLY_AGENT_DEPTH` plus one, as the [agent extension](../../extensions/agent/spec.md#depth-limit) defines

A key may be optional, such as a harness key: the extension receives it when it is granted and present, and otherwise the harness uses its own login. An unavailable keychain for an optional key does not block that fallback.

**Cancellation.** On Ctrl-C or SIGTERM, `ncly` sends the signal to every program it started and to every descendant that it can still reach. It sends SIGKILL to what still runs 15 seconds later, which leaves a Python program its own 10 seconds to stop its children. Before each signal, `ncly` finds descendants through their parent process, so a descendant that left the process group of its parent, as the children of transcript's Python program do, still receives the signal. On Linux, `ncly` registers as a child subreaper, so it also reaches a descendant whose parent already exited. macOS has no subreaper, so a descendant whose parent exited before the signal is the one exception there, and it keeps running. `ncly` keeps every finished output and removes only its own temporary files. It exits 130 with `INTERRUPTED` or 143 with `TERMINATED`, and under `--json` the answer keeps the data keys that apply, such as `output_dir` or the `results` of finished items. A second signal during this cleanup is ignored. A SIGKILL sent to `ncly` itself leaves no time to clean up.

**Output streams.** A descendant that inherits a program's stdout or stderr can hold it open after the program exits, as an orphan on macOS can. `ncly` stops reading a program's streams 2 seconds after that program exits, or when it sends SIGKILL during cleanup, whichever comes first. A descendant therefore never keeps `ncly` waiting, and output written after that point is lost.

**Timeout.** The timeout covers the command's work and starts before child execution. On expiry, use the same process-tree cleanup as cancellation. A temporary failure before any paid request exits 75 only when Retry safety holds. Otherwise it exits 1. A caller's SIGTERM remains exit 143. Record finished work and unknown effects before returning when the process can still do so.

**Protocol failures.** An extension, and a Python program that an extension embeds, returns versioned machine messages only, which Go maps into this public contract. Missing, malformed, or incompatible messages never count as success. A launched program may already have produced effects, so preserve available evidence and use a runtime failure unless safety can be established. Harness adapters declare the modes they can actually enforce. An unsupported tools-off mode fails with `CAPABILITY_UNSUPPORTED` before launching a task with outside material.
