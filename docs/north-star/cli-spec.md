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
├── skill <name>                    M1
├── transcript run youtube|zoom     M1
├── transcript summary run <run-id> M1
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

`errors` describes the command's failure. The command computes its overall verdict first, then puts an error with the matching exit code first. When the failures of the command itself map to different exit codes, the most cautious one leads: `130` or `143`, then `1`, `78`, `2`, and `75` only when every failure is temporary. When several failures share that exit code, the first one in this order leads. For exit 1, `RESUME_UNSAFE`, `DOWNLOAD_FAILED`, then `RUNTIME`. For exit 78, `CONFIG_INVALID`, `TERMINAL_REQUIRED`, `PREREQ_MISSING`, `CAPABILITY_UNSUPPORTED`, `KEYRING_UNAVAILABLE`, `AUTH_MISSING`, then `AUTH_REJECTED`, so the first fix that a human needs leads. For exit 2, `USAGE_INVALID`, `CONFIRMATION_REQUIRED`, `NOT_FOUND`, then `DEPTH_LIMIT`. Input order or the order in which failures arrive never determines retry safety. Item errors remain inside `results`. `warnings` holds objects of the same shape, may appear in any answer, and never changes the exit code.

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
| `2` | Invalid invocation: unknown command, bad flag, missing value, or missing confirmation | Fix the call. Inspect any returned `run_id` before continuing recorded work |
| `75` | Temporary failure that meets every condition in Retry safety | Yes, with the same command |
| `78` | A human must act: a key, a prerequisite, a config fix, or a step at a terminal | Relay the hint. Inspect any returned `run_id` before continuing recorded work |
| `130` | Interrupted by Ctrl-C | |
| `143` | Terminated by SIGTERM | |

Codes 124 to 127, and every code from 128 up other than 130 and 143, stay reserved for the shell.

Only the overall process exit 75 permits an automatic repeat of the same invocation. An item error describes that item's cause, not the safety of repeating the whole command. A caller retains the complete answer and `run_id` before extracting artifact paths. After fixing a recorded failure, it inspects the run and requests an explicit resume instead of repeating the original command.

## Error codes (M0)

Each code maps to exactly one exit code. Core uses only the codes in this table, and a milestone adds its codes here. An extension may use its own codes, prefixed with its domain, such as `FLEET_HOST_DOWN`. Its exit code carries the meaning that this file defines.

| Code | Exit | When | Since |
|---|---|---|---|
| `USAGE_INVALID` | 2 | Unknown command, bad flag, or a missing value in non-interactive mode | M0 |
| `CONFIRMATION_REQUIRED` | 2 | A step needs a confirmation in non-interactive mode, and `--force` is absent | M0 |
| `CONFIG_INVALID` | 78 | A config file has a syntax error or a value of the wrong type, or cannot be read | M0 |
| `TERMINAL_REQUIRED` | 78 | A step needs a human at a terminal, such as typing or granting a key | M0 |
| `RUNTIME` | 1 | Any other failure during work | M0 |
| `NOT_FOUND` | 2 | The named skill, service, profile, component, run, prompt, or recording does not exist | M1 |
| `AUTH_MISSING` | 78 | A required key is in neither the environment nor the keychain | M1 |
| `PREREQ_MISSING` | 78 | A required tool is absent or too old | M1 |
| `KEYRING_UNAVAILABLE` | 78 | The OS keychain does not answer when a command needs it: to read a key missing from the environment, for `ncly auth`, or for the `auth` check of `ncly doctor` | M1 |
| `TEMPORARY` | 75 | A temporary failure for which the whole invocation meets Retry safety | M0 |
| `INTERRUPTED` | 130 | Ctrl-C stopped the command | M0 |
| `TERMINATED` | 143 | SIGTERM stopped the command | M0 |
| `RESUME_UNSAFE` | 1 | A run cannot safely continue because evidence is missing, changed, uncertain, or unsupported | M1 |
| `DEPTH_LIMIT` | 2 | Starting a harness would exceed the configured agent depth | M1 |
| `CAPABILITY_UNSUPPORTED` | 78 | The selected program cannot enforce a required mode or protocol | M1 |
| `AUTH_REJECTED` | 78 | A service refused a key, such as Deepgram answering 401 | M1 |
| `DOWNLOAD_FAILED` | 1 | `yt-dlp` could not fetch the audio of a video, for a reason other than the network | M1 |

Warnings use codes from this table.

| Code | When | Since |
|---|---|---|
| `CONFIG_UNKNOWN_KEY` | A config file holds a key that this version of `ncly` does not know | M0 |
| `SKILL_SHADOWED` | Two skills share a name, in two folders of `[skill] paths` or inside one | M1 |
| `SKILL_NO_SOURCE` | No skill folder is configured | M1 |
| `SKILL_NAME_MISMATCH` | A skill's subfolder name differs from the `name` in its `SKILL.md` | M1 |
| `SKILL_INVALID` | A `SKILL.md` cannot be read, its frontmatter is invalid, or it lacks `name` or `description` | M1 |
| `SKILL_SOURCE_MISSING` | A folder of `[skill] paths` is missing or cannot be read | M1 |
| `CONFIG_DEFAULTS_USED` | Discovery used defaults because a config file could not be read | M1 |
| `BROWSER_COOKIES_SKIPPED` | The configured browser's cookies could not be read, so the download used anonymous access | M1 |
| `SUMMARY_SKIPPED` | No profile applies, so a transcript run saved the transcript without a summary | M1 |

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
- A syntax error, a value of the wrong type, or a file that cannot be read exits 78 with `CONFIG_INVALID`, naming the file and, when the parser knows it, the line. `--help`, `--version`, `ncly completion`, and M1 discovery still work with defaults. Discovery reports its fallback. Doctor reports this failure on stdout as a failed `core.config` check and uses defaults only for checks independent of the invalid configuration
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

One declaration per command supplies help, completion, validation, and discovery. M0 implements the fields its commands use. M1 extends them with input constraints, result types, prerequisites, and configuration sources as each real command needs them. Domain code owns the behavior. The declaration is not a workflow language.

Effects distinguish user writes, network access, and paid requests. Modes distinguish dry run, recorded execution, and resume. A missing declaration is not permission to assume a capability. A command may support recorded execution without supporting resume.

Discovery reads declarations and static extension manifests. It starts no harness or extension and reads no key. Detailed discovery is scoped to one command, and summary listings omit full schemas and skill bodies. Resolved configuration identifies the source file or environment variable for non-secret values. Secret values never appear.

## Operations (M0 contract, M1 execution)

The domain prepares an operation from resolved inputs and configuration, performs its steps, and verifies their results. Dry run uses that same preparation and reports pending checks, such as authentication. Execution repeats checks affected by changed inputs or destinations before applying effects. A dry run is not a reservation or a guarantee that the environment will stay unchanged.

Commands that incur a bill or can leave reusable partial work record an execution before the first such effect. M1 applies this to transcript. Read-only discovery and dry runs create no run record. Each domain declares its recording and resume support before code is written.

The shared operation support owns run IDs, durable records, and result publication. The domain owns its steps and the evidence that makes them resumable. `internal/run` owns child processes, their environment, timeouts, and signals. It does not decide whether a business action is safe to repeat.

### Records and evidence

A run record contains its format version, `run_id`, command path, status, step outcomes, and artifact references. It also keeps the non-secret input fingerprints, resolved profile and relevant tool versions needed to check a resume. Keep only information needed to explain or continue the operation. Full task bodies, transcripts, raw environment dumps, and keys are not copied into the record.

Persist an intent before an effect that cannot safely be repeated, and persist completion only after checking its result. If the process dies between those writes, the effect is unknown until evidence resolves it. A missing response is never proof that nothing happened. If the record cannot be written, stop before starting the next effect.

Records distinguish `running`, `completed`, `failed`, `interrupted`, and `unknown`. A step distinguishes `pending`, `completed`, `failed`, and `unknown`. Pending means its effect has not started; preparation may have failed. A completed step names the evidence needed to reuse it, such as an artifact path and content fingerprint. The domain may leave an in-flight effect as unknown until it has a verified result.

Keep records under the XDG state directory and give every run its own temporary directory. Write each record atomically under a lock. Lock shared destinations or configuration at the point of mutation and revalidate while holding the lock. A conflict follows Retry safety, including effects already completed elsewhere in the run.

Records survive the process. M1 retains them until the user removes their JSON files and never automatically deletes output artifacts. Listings bound both record count and item-key previews. A missing record or an unsupported record version never causes the original operation to be started again.

### Record format (M1)

Each run has one record, `runs/<run_id>.json` in the state directory, such as `~/.local/state/nicely/runs/`. The record is one JSON object. While holding `runs/<run_id>.lock`, `ncly` writes a temporary file in the same directory, syncs that file, renames it, then syncs the parent directory. Newly created state directories are synced with their parents too. This persistence barrier finishes before acknowledging a non-repeatable effect. A failed barrier prevents the effect. Atomic visibility alone is not durable acknowledgement.

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

Later domains reuse these keys where they fit and define their own evidence at their first consumer. M1 implements transcript only, without a generic extension evidence interpreter or workflow engine. Compatibility governs any later format change.

### Inspection and explicit resume

Inspection is read-only and reports saved evidence, not an assumption that a process marked running is still alive. Resume checks that no execution still owns the run, validates the record version, and verifies the saved inputs, relevant configuration, tool compatibility, and artifacts. For `ncly run resume`, the relevant configuration excludes the profile, because resume uses the profile that the run recorded.

A supported resume reuses verified completed steps and continues only steps whose execution is known to be safe. It records each attempt under the same run ID. Changed or missing evidence returns `RESUME_UNSAFE` with an explanation and the artifacts still available. `--force` does not override uncertainty or make a non-repeatable effect safe.

For example, a saved transcript and a summary that was never started permit a resume of the summary alone. An unanswered paid summary request is unknown and is not sent again automatically. A domain may resolve an unknown step from complete, independently verified local artifacts, or from a reliable external lookup. This establishes a reusable result; it never authorizes redispatch of an unresolved request. Otherwise the user must inspect the evidence and explicitly choose any new operation that might repeat a charge.

Nicely has no background worker or scheduler in these milestones. A record makes work inspectable after a session ends. It does not keep the process alive or promise that an arbitrary harness task can resume.

### Dry-run answer (M1)

`ncly transcript run`, `ncly transcript summary run`, and `ncly run resume` answer `--dry-run --json` with `dry_run`, set to `true`, then `plan` and `pending_checks`.

| Key | Type | Meaning |
|---|---|---|
| `plan` | array | One object per item, in input order, with `key`, `output` or `output_dir`, `steps`, and `effects` |
| `plan[].steps` | array | Each step that execution would reach, with `id` and `action`, which is `run` or `reuse` |
| `plan[].effects` | object | The booleans `user_writes`, `network`, and `paid`, for the item's `run` steps |
| `pending_checks` | array | Each check that a `run` step needs and that only execution can make, with `id`, such as `auth.deepgram`, and `message`. A dry run reads no key, so a `run` step that needs a key keeps its authentication pending. The array is empty when no `run` step needs such a check |

A new item has `output`, the output folder that will hold its result folder. An item that already has a result folder, in `run resume` and `summary run`, has `output_dir`. A resume that execution would refuse fails its dry run with the same `RESUME_UNSAFE`.

A new transcription reports its output parent without reserving a result folder:

```json
{"ok":true,"contract_version":1,"dry_run":true,"plan":[{"key":"https://www.youtube.com/watch?v=VIDEO_ID","output":"/Users/me/transcripts","steps":[{"id":"transcribe","action":"run"}],"effects":{"user_writes":true,"network":true,"paid":true}}],"pending_checks":[{"id":"auth.deepgram","message":"Authentication is checked only during execution."}]}
```

A summary alone, from `ncly transcript summary run <run-id> --dry-run --json`, reuses the transcript and waits for the harness to start:

```json
{"ok":true,"contract_version":1,"dry_run":true,"plan":[{"key":"https://www.youtube.com/watch?v=VIDEO_ID","output_dir":"/Users/me/transcripts/Video_title","steps":[{"id":"transcribe","action":"reuse"},{"id":"summarize","action":"run"}],"effects":{"user_writes":true,"network":true,"paid":true}}],"pending_checks":[{"id":"agent.claude","message":"The harness login is checked when the summary starts."}]}
```

A resume that reuses every step, from `ncly run resume <run-id> --dry-run --json`, has nothing pending:

```json
{"ok":true,"contract_version":1,"dry_run":true,"plan":[{"key":"https://www.youtube.com/watch?v=VIDEO_ID","output_dir":"/Users/me/transcripts/Video_title","steps":[{"id":"transcribe","action":"reuse"},{"id":"summarize","action":"reuse"}],"effects":{"user_writes":false,"network":false,"paid":false}}],"pending_checks":[]}
```

## ncly describe (M1)

```
ncly describe [command...] [--json]
```

Without a command path, returns concise descriptions of installed commands. A path such as `transcript run youtube` returns that command's declaration. A path that names a group, such as `transcript`, returns the concise entries of the commands under it. The JSON envelope contains `commands`, an array. Text is localized, while machine identifiers follow Compatibility. Unknown paths use `NOT_FOUND`.

Each entry has `path`, such as `transcript run youtube`, `summary`, `since`, the release that added the command, such as `v0.1.0`, and `source`. `source` is `core` in M1. M3 adds sources for extensions, and a reader treats an unknown `source` as not core. A detailed entry adds these keys:

| Key | Type | Meaning |
|---|---|---|
| `args` | array | Positional arguments in order, each with `name`, `summary`, `type`, `required`, `variadic`, and `values`. `type` is `string`, `boolean`, `integer`, or `number`. `values` lists accepted typed values, or is empty. Only the last argument may be variadic |
| `flags` | array | Own flags plus applicable write, output, and timeout flags. Each has `name`, `shorthand`, `value`, `type`, `required`, `variadic`, `values`, `default`, and `summary`. Names omit `--`. Empty `shorthand` means none, empty `value` means a switch. `default` is the typed default or `null` when none applies. Variadic means several values |
| `flag_groups` | array | Each group has `flags`, an array of long flag names, and integer `min` and `max`, the permitted number supplied. Zoom's `latest` and `path` have `min: 1` and `max: 1` |
| `result` | object | `kind`: `object`, `report`, `batch`, or `text`; `keys`: top-level successful data keys; `schema`: JSON Schema for normal JSON answers, including item fields, errors, types, nullability, and closed enums. `dry_run_schema` is present when supported. A text result ignores `--json`, has empty `keys`, and omits schemas |
| `effects` | object | Conservative possible effects: `user_writes`, `network`, and `paid`. Dry run describes the selected invocation's effects |
| `requires` | object | `bins` and `keys`, possible program and service prerequisites. Preparation selects the ones needed by this invocation, so a skipped summary requires no harness |
| `modes` | object | The booleans `dry_run`, `recorded`, and `resume` |
| `config` | array | Each config key that the command reads, with `key`, `value`, and `source`, which is a file path, an environment variable, or `default`. A secret value never appears |
| `examples` | array | Full command lines |

`requires` takes its name and its keys from the M3 manifest, so a static manifest can describe an extension's commands the same way. Discovery works with default configuration when a config file is invalid, adds `CONFIG_DEFAULTS_USED`, and does not claim to have resolved a profile that it could not read.

Schemas use JSON Schema draft 2020-12, with all references resolved within the returned descriptor and no remote schema fetch. `keys` comes from that result definition, excluding `ok`, `contract_version`, `errors`, and `warnings`; a successful batch includes `results`. Flag types use the same scalar types as arguments. Schemas allow unknown optional fields as Compatibility requires. Declarations cover the current commands only; M1 needs no schema generator or custom constraint language for future domains. Validation uses required flags and groups from this declaration. For YouTube, `url` is required and variadic. For Zoom, exactly one of `latest` and `path` is required. Examples do not replace these constraints.

## ncly run (M1)

```
ncly run list [--path <command-path>] [--key <item-key>] [--since <time>] [--limit <count>] [--json]
ncly run view <run-id> [--json]
ncly run resume <run-id> [--dry-run] [--force] [--timeout <duration>] [--json]
```

- `list` returns `runs`, sorted by `started_at` descending, then `run_id` ascending, with at most `--limit` entries, 20 by default. The count must be a positive integer. Filters combine with AND before limiting: `--path` exactly matches the command path, `--key` matches any item key, and `--since` includes runs whose `started_at` is at or after that RFC 3339 time. An invalid time is `USAGE_INVALID`
- Each list entry has `run_id`, `path`, `status`, `started_at`, `updated_at`, `active`, `item_count`, and `keys`. `keys` previews at most the first three item keys in input order; filters still inspect every key. The top-level `more` boolean is true when additional matching records exceed the limit. Increase the limit or narrow the filters, without a database or index service
- `view` returns `run`, the record of [Record format](#record-format-m1) plus `active`, including step outcomes and artifact references. A successful inspection exits 0 even when the saved run failed. Its top-level `ok` describes the inspection, and `run.status` describes the saved operation
- `resume` follows Operations and returns the original domain's result with the same `run_id`. Dry run only describes the remaining steps and checks, as [Dry-run answer](#dry-run-answer-m1) defines. The original command's timeout default applies unless overridden
- An unknown run uses `NOT_FOUND`. An unsafe or unsupported resume uses `RESUME_UNSAFE`. For a command that supports resume, a completed run returns its verified saved result without running its steps again
- When another process holds the run, `resume` exits 75 with `TEMPORARY` before any effect, with a hint naming `ncly run view <run-id>`. `active` is true while a process holds the run lock; inspecting a busy run does not permit a retry loop

To find a run whose stdout was lost, filter by the original path and an item key, such as `ncly run list --path 'transcript run youtube' --key <url> --json`. Add `--since` when the retained invocation start time is known; a fresh session may omit it. Check `more`, then confirm candidates with `run view`, including their recorded inputs and complete item list. Report ambiguous matches instead of choosing the newest one. A call that failed during preparation before its first record write has no record. Absence from one limited list proves neither absence of a run nor absence of billing.

## Keys (M1)

- A service has a name of lowercase letters and digits, such as `deepgram`. Its environment variable is the name in uppercase followed by `_API_KEY`, such as `DEEPGRAM_API_KEY`.
- `ncly` reads a key from its environment variable first. It reads the key from the keychain only when that variable is empty. `ncly auth` and `ncly doctor` also check the keychain. The keychain entry uses the service `nicely` and the account named after the service.
- Each keychain call has a time limit: 60 seconds in interactive mode, where a human can answer an unlock prompt, and 10 seconds otherwise. A keychain that does not answer in time counts as unavailable. A command that needs a key exits 78 with `KEYRING_UNAVAILABLE` when the key is missing from the environment and the keychain is unavailable, and its hint names the variable to export. When the keychain answers without the key, the code is `AUTH_MISSING`.
- A key is never accepted as a flag. `ncly auth login` reads it from a masked field, or from stdin with `--stdin`.
- A key is never printed. `ncly auth status` reports where a key comes from, never its value.

## Programs that ncly runs (M1)

`ncly` runs three kinds of programs: Python programs of core, harnesses, and, from M3, extensions. None of them runs in a sandbox. A program runs with the rights of the user, so it could read the keychain or any file itself. Adding a tap means trusting its author.

**Environment.** A program receives the environment of `ncly` with these changes:

- Each variable named `<NAME>_API_KEY` is the key of a service. The program receives it only when that key is granted to it, and `ncly` then sets it from the environment or the keychain.
- The variables of the global flags carry their resolved values. `NCLY_VERSION` and `NCLY_CONTRACT_VERSION` identify the release and public protocol.
- A harness receives `NCLY_AGENT_DEPTH` plus one.

**Grants.** A program of core receives the keys that core declares for it. The transcript Python program receives `DEEPGRAM_API_KEY`. A harness adapter declares the variables of its harness: `claude` receives `ANTHROPIC_API_KEY`, and `pi` uses its adapter's verified provider-to-variable mapping, such as `OPENROUTER_API_KEY`. It never derives a key variable by parsing translated help at runtime. A harness key is optional. `ncly` passes it when the environment or the keychain holds it, and otherwise the harness uses its own login. An unavailable keychain for an optional harness key does not block that fallback. An extension receives a key only after a human grants that key to that extension in a terminal. `--force` never grants a key. When an update of an extension declares a new key, the extension does not start until a human grants it.

**Cancellation.** On Ctrl-C or SIGTERM, `ncly` sends the signal to every program it started and to every descendant that it can still reach. It sends SIGKILL to what still runs 15 seconds later, which leaves a Python program its own 10 seconds to stop its children. Before each signal, `ncly` finds descendants through their parent process, so a descendant that left the process group of its parent, as the children of the transcript CLI do, still receives the signal. On Linux, `ncly` registers as a child subreaper, so it also reaches a descendant whose parent already exited. macOS has no subreaper, so a descendant whose parent exited before the signal is the one exception there, and it keeps running. `ncly` keeps every finished output and removes only its own temporary files. It exits 130 with `INTERRUPTED` or 143 with `TERMINATED`, and under `--json` the answer keeps the data keys that apply, such as `output_dir` or the `results` of finished items. A second signal during this cleanup is ignored. A SIGKILL sent to `ncly` itself leaves no time to clean up.

**Timeout.** The timeout covers the command's work and starts before child execution. On expiry, use the same process-tree cleanup as cancellation. A temporary failure before any paid request exits 75 only when Retry safety holds. Otherwise it exits 1. A caller's SIGTERM remains exit 143. Record finished work and unknown effects before returning when the process can still do so.

**Protocol failures.** Core Python programs return versioned machine messages only. Go maps them into this public contract. Missing, malformed, or incompatible messages never count as success. A launched program may already have produced effects, so preserve available evidence and use a runtime failure unless safety can be established. Harness adapters declare the modes they can actually enforce. An unsupported tools-off mode fails with `CAPABILITY_UNSUPPORTED` before launching a task with outside material.

## Agent profiles (M1)

A profile names a harness, a model, and an effort, plus a provider when the harness serves several. Profiles live in the config, and every command that runs an agent reads them there, except `ncly run resume`, which uses the profile that the run recorded without reading the config again.

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

`--profile <name>` picks a profile, and `default_profile` applies without it. An unknown profile exits 2 with `NOT_FOUND`. Without `--profile` and without `default_profile`, no profile applies, and each command that runs an agent states what it does then.

M1 ships the adapters for `claude` and `pi`. A profile that names another harness fails with `CAPABILITY_UNSUPPORTED` until M2 adds its adapter. Each adapter turns the tools off, sends the task on stdin, and checks the answer:

- `claude`, version 2.1.291 or later, runs with `--print`, `--model`, `--effort`, `--system-prompt`, and `--output-format json`. It turns tools off with `--tools ""`, `--strict-mcp-config --mcp-config '{"mcpServers":{}}'`, `--setting-sources ""`, `--disable-slash-commands`, `--no-session-persistence`, and `--permission-mode dontAsk`. `--settings` turns hooks off and sets `CLAUDE_CODE_EFFORT_LEVEL`, which would otherwise override `--effort`. The answer is `result`, and `is_error` must be `false`. The effort is `low`, `medium`, `high`, `xhigh`, or `max`, and a profile for `claude` has no `provider`
- `pi`, version 1.0.4 or later, runs with `--print`, `--model <provider>/<model>`, `--thinking`, and `--system-prompt`. It turns tools off with `--no-tools`, `--no-session`, `--no-skills`, `--no-prompt-templates`, `--no-context-files`, `--no-extensions`, and `--no-approve`. The answer is stdout, which must not be empty. The effort is `off`, `minimal`, `low`, `medium`, `high`, `xhigh`, or `max`

A profile with an effort that its harness rejects, or with a `provider` for `claude`, is `CONFIG_INVALID` when that profile is selected. Syntax errors and values of the wrong declared type still fail as Configuration defines. An unused profile's semantic constraint does not block skill inspection. A harness older than its listed version is `PREREQ_MISSING`. These invocations come from the transcript CLI, and their flags were checked against `--help` on 2026-10-07. M1 rechecks the real adapters and observes that material cannot invoke tools, hooks, MCP servers, or loaded extensions. A stub accepting flags or a real harness returning text is not that proof.

The internal agent service checks `[agent] max_depth`, 2 by default, before every launch. `NCLY_AGENT_DEPTH` defaults to 0. A launch at or above the limit exits 2 with `DEPTH_LIMIT`. The child receives the depth plus one. This applies to transcript in M1 and the public agent command in M2.

## ncly completion (M0)

```
ncly completion zsh|bash|fish
```

Prints the completion script on stdout. The Homebrew formula installs these scripts, and each Linux archive ships them in `completions/`. The AUR package will install them too, once it ships. Users run this command only for a manual setup.

Completion also suggests values: skill names, auth services, profiles, transcript prompts, and doctor components.

## ncly doctor (M1)

```
ncly doctor [component] [--live] [--timeout <duration>] [--json]
```

| Argument or flag | Effect |
|---|---|
| `component` | One of `core`, `auth`, `skill`, or `transcript`. Without it, doctor checks all of them. |
| `--live` | Adds network checks against free endpoints only, such as listing Deepgram projects. It never calls an endpoint that bills. |
| `--timeout` | Bounds the checks, including local version probes and free network checks, 30 seconds by default. Installation after human confirmation is separate from those checks |

- The default checks stay on the machine: binaries and their versions, keys present, the keychain answering, and the config files parsing.
- Invalid or unreadable config produces the failed `core.config` finding with `CONFIG_INVALID`, even for a selected component. Doctor continues independent checks and omits checks whose required configuration could not be resolved. It never reports those omitted checks as passing
- When the keychain is unavailable, its check is `warn` if every key that a component needs comes from the environment, and `fail` with `KEYRING_UNAVAILABLE` otherwise.
- The `skill` component warns with `SKILL_SOURCE_MISSING` for a folder of `[skill] paths` that is missing or cannot be read. It warns with `SKILL_NAME_MISMATCH` when a subfolder's name differs from its `name`, and with `SKILL_INVALID` for a `SKILL.md` that cannot be read, whose frontmatter is invalid, or that lacks `name` or `description`.
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

- `login` stores the key in the keychain. Replacing a stored key asks for confirmation, or needs `--force` in non-interactive mode. In non-interactive mode without `--stdin`, `login` exits 78 with `TERMINAL_REQUIRED` and the hint ``Run `ncly auth login <service>` in a terminal``. `TERMINAL_REQUIRED` wins over `KEYRING_UNAVAILABLE`, because `login` without a terminal and without `--stdin` never reaches the keychain. An agent relays this hint to the human and never pipes a key itself.
- `logout` removes the key from the keychain. It asks for confirmation, or needs `--force` in non-interactive mode.
- When the keychain is unavailable, `login` and `logout` exit 78 with `KEYRING_UNAVAILABLE`, and the hint names the environment variable to use instead.
- `status` is a report command. It reports whether the keychain answers, then lists every service that core uses, plus, from M3, every service that an extension declares. `ok` is `false`, with exit 78, only when the keychain does not answer and a listed service has no key in the environment. That service's entry then adds `code`, set to `KEYRING_UNAVAILABLE`, the code of the failed `auth` check in `ncly doctor`.

For `login` and `logout`, `--dry-run` takes precedence over secret input and confirmation. It describes the intended action without reading stdin for a key, opening a keychain entry, or prompting. Checks that would require a keychain read, including whether login would replace a key, are reported as pending. This simulation does not require a terminal or `--stdin`.

Their dry-run JSON has `dry_run: true`, `plan`, an array with one entry containing the service `key`, `action` (`login` or `logout`), and `effects` (`user_writes: true`, `network: false`, `paid: false`), plus `pending_checks`, whose entries have `id` and `message`. The check `auth.keychain` covers availability, existing keys, and any replacement confirmation. A successful login or logout answers `service` and `action`, never a key.

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
- Inside one folder of `paths`, when two subfolders declare the same `name`, the subfolder whose name sorts first byte by byte wins, and `list` adds `SKILL_SHADOWED`.
- A `SKILL.md` that cannot be read, whose frontmatter is invalid, or that lacks `name` or `description` is not a skill. `list` skips it and adds a `SKILL_INVALID` warning whose hint names the file to fix, and `view` answers `NOT_FOUND` for it.
- A folder of `paths` that is missing or cannot be read is skipped. `list` continues with the other folders and adds a `SKILL_SOURCE_MISSING` warning whose hint names the folder, and `view` answers `NOT_FOUND` for a skill that no other folder holds.
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
ncly transcript run youtube --url <url> [<url>...] [--prompt <name>] [--profile <name>] [--no-summary] [--output <folder>] [--open] [--timeout <duration>] [--dry-run] [--json]
ncly transcript run zoom (--latest | --path <folder>) [--prompt <name>] [--profile <name>] [--no-summary] [--output <folder>] [--open] [--timeout <duration>] [--dry-run] [--json]
ncly transcript summary run <run-id> [--prompt <name>] [--profile <name>] [--timeout <duration>] [--dry-run] [--json]
ncly transcript prompt list [--json]
```

- `ncly transcript run` needs the `deepgram` key. Without it, the command exits 78 with `AUTH_MISSING` or `KEYRING_UNAVAILABLE`, as [Keys](#keys-m1) defines.
- `--profile` names an [agent profile](#agent-profiles-m1) for the summary. Without a profile, the run saves the transcript, skips the summary, and adds `SUMMARY_SKIPPED`, whose hint shows a profile to add. `--no-summary` skips the summary without a warning. It cannot combine with `--profile` or `--prompt`; these are two optional exclusive flag groups
- `--dry-run` follows the [global definition](#global-flags-m0): it validates the plan without a key, a paid request, or a change to the user's files, and answers as [Dry-run answer](#dry-run-answer-m1) defines.
- `--timeout` covers preparation and execution, 570 seconds by default, with at most 15 additional seconds for process cleanup. An agent gives its shell a longer limit than that total, with a margin, or chooses a shorter `ncly` timeout. No universal shell limit or default is assumed. A batch may need a longer timeout
- `--open` opens each result folder as soon as it appears, through `open` on macOS and `xdg-open` on Linux. Dry run never opens it
- Exit 75 follows Retry safety, including any local writes. Never rerun an exit 1 automatically, because work may be partial or already billed
- Go owns the summary step from M1 and uses the internal agent service. Python stops at the transcript and returns its verified artifacts
- A recorded run returns `run_id`. Its normal execution answer always contains `results`, even for one item, as [Output](#output-m0) describes. `summary run` and resume use that same item shape. This is a declared batch result; the shared M0 object answers remain unchanged
- Each item carries `url` or the Zoom `recording` folder, `output_dir`, `transcript`, and `summary`. The last three are paths or `null` until available. An unsuccessful item adds its `errors`; a saved transcript remains referenced when its summary fails. A URL given twice is `USAGE_INVALID` before any effect, because each item key names one item

Preparation validates all arguments and local inputs before effects. When a summary is selected, it resolves the prompt and profile and checks depth, harness version, and declared tools-off support before transcription dispatch. It starts no harness to check its login; that remains pending. `--no-summary` resolves no prompt or profile and needs no harness. A preparation refusal creates no record and bills nothing.

An attempted item's error stays in `results`. The overall verdict follows Output across all attempted items. A human-action cause after transcription therefore does not authorize repeating the original paid command. Each recorded failure's recovery hint names inspection or explicit resume; an unknown summary names the explicit new summary operation.

### Output, prompts, and cookies

Each item gets its own result folder inside the output folder. The folder holds `raw_transcript.txt`, `raw_sentences.txt`, `raw_transcript.json`, and any summary. `--output` names the output folder. Without it, the config decides, and the current folder is the default, as with `yt-dlp` and `gh run download`. In every JSON key, `output` names an output folder, and `output_dir` names a result folder.

```toml
[transcript]
browser = "arc"
prompt_paths = ["~/code/prompts"]

[transcript.youtube]
output = "~/Documents/transcripts"
prompt = "follow_along_note"

[transcript.zoom]
recordings = "~/Documents/Zoom"
output = "~/Documents/meetings"
prompt = "short_summary"
```

| Key | Default | Meaning |
|---|---|---|
| `browser` | none | The browser whose YouTube cookies the download uses: `arc`, on macOS only, or a browser that `yt-dlp` reads, such as `chrome`, `chromium`, `brave`, `firefox`, or `safari`. Without it, the download uses anonymous access, which YouTube often refuses |
| `prompt_paths` | none | Folders searched before the bundled prompts. A prompt is `<name>.md`, or `<name>/prompt.md` |
| `youtube.output`, `zoom.output` | the current folder | The output folder of each source |
| `youtube.prompt` | `follow_along_note` | The default prompt for YouTube |
| `zoom.recordings` | `~/Documents/Zoom` | The folder where Zoom saves recordings. `--latest` picks the newest meeting folder, and `--path` takes a folder name or a full path |
| `zoom.prompt` | `short_summary` | The default prompt for Zoom |

Go reserves each new result folder, named after the video title or meeting, or `<name>-2`, `<name>-3`, and so on when taken. Under the output resource lock, it records the candidate path before creating it and rechecks any collision. Creating the directory reserves its name; two new transcriptions never share a folder or write into an existing one. A failure after reservation exits 1 for the new transcription, even if no audio was sent, because repeating the command creates another folder. Explicit resume reuses the recorded folder instead. Only effects of the current invocation decide retry safety, not paid work reused from an earlier attempt.

Repeating `transcript run` transcribes and bills Deepgram again. Resume and `summary run` reuse the recorded folder. Every summary uses `<prompt-name>-<run-id>.md`, including the initial summary, so a new operation preserves every earlier record's artifact fingerprint.

`ncly transcript prompt list --json` returns `prompts`, sorted by name, each with `name`, `path`, and `source`, which is `config` for a folder of `prompt_paths` and `bundled` otherwise. A name appears once. The first folder of `prompt_paths` that holds it wins, then the bundled prompts. Inside one folder, `<name>.md` wins over `<name>/prompt.md`, because the file comes before the folder.

Machine paths belong in the local config. A Zoom meeting folder starts with its date and time, such as `2026-05-03 14.46.55`, and the rest of its name, which Zoom writes in the user's language, becomes the title.

When the configured browser's cookies cannot be read, the download falls back to anonymous access and the answer adds `BROWSER_COOKIES_SKIPPED`. In a batch, the warning appears once in the top-level `warnings`, because the browser setting covers the whole run. `ncly doctor transcript` reports the same problem as `warn`. The Arc adapter checks the `yt-dlp` function that it changes, instead of an exact `yt-dlp` version. When that function is missing, the adapter reports the problem and the download falls back the same way.

`ncly` runs the program with `uv run --script --upgrade-package yt-dlp --no-python-downloads --cache-dir <folder>`, where the folder is `uv` inside Nicely's cache, such as `~/.cache/nicely/uv`. The flag wins over an inherited `UV_CACHE_DIR`, so uv writes only inside Nicely's cache, as a dry run requires. uv never downloads a Python. A missing Python 3.12 or later is `PREREQ_MISSING`, and `ncly doctor transcript` offers its install command. Each Python invocation uses the newest `yt-dlp` release at or above the minimum that the program names, so a fix reaches users without an `ncly` release. The program pins its other dependencies. A dry run may contact the package index to prepare that cache; it is not an offline guarantee. Failed dependency preparation returns `RUNTIME` with a setup hint; the command's own timeout follows Programs. Doctor reads installed versions without upgrading packages.

These options of the transcript CLI change:

| Transcript CLI | ncly |
|---|---|
| `--output-dir` | `--output` |
| `--debug`, `TRANSCRIPT_DEBUG` | `NCLY_DEBUG=1` |
| `--provider`, `--model`, `--effort` | A profile in the config |
| `--no-progress` | Removed, because `--json` and non-interactive mode already hide the spinner |
| `--preview` | Removed, because `ncly markdown view` renders a summary from M4 |
| `list prompts` | `ncly transcript prompt list` |
| `list profiles` | Removed until M2 adds `ncly agent profile list` |
| `list models` | Removed. M2 decides whether `ncly agent model list` replaces it |
| `doctor --source` | `ncly doctor transcript` |

### Steps and resume

Each transcription item has `transcribe`, then `summarize` unless skipped. Python prepares the complete item set before emitting its initial plan. Go validates that set and creates the initial `running` record with pending steps before allowing any user-side mutation. Temporary audio preparation may overlap that validation after Python sends the plan. `started_at` is the invocation's start time and stays unchanged on resume. A step reaches `completed` only after Go independently verifies its required files, syncs them, and durably records their fingerprints.

- `transcribe` turns `unknown` before result-folder reservation and upload, as the protocol below requires. A download failure before intent leaves it `pending` with a saved problem. After acknowledgement, it turns `failed` only when transport failed before any audio was sent or Deepgram returned a 4xx refusal. Any 5xx status and any transport timeout or lost connection after sending audio leave it `unknown`. The transport disables automatic upload retries, including SDK retries; one acknowledgement permits at most one request
- `summarize` turns `unknown` before the harness starts. It turns `failed` only when the harness never started, such as when it is missing. Once the harness starts, any end without a saved summary leaves the step `unknown`, because the harness may already have billed
- The record keeps the `inputs` that the table below lists. It keeps as tools the program and `yt-dlp` versions that the `intent` reports, and the harness version
- Resume takes the URL, selected Zoom folder, and absolute output folder from the record. It never selects `--latest` again. It rereads the prompt and refuses a changed fingerprint. For Zoom, it verifies the recorded audio and compares the program's `audio` in `plan` and `intent` before acknowledgement. Resume uses the recorded profile without resolving it again; `summary run --profile` changes it. Only programs needed by `run` steps require a current version check, so a summary-only resume needs no `uv` or Deepgram key. A version above its minimum does not block resume; below it returns `PREREQ_MISSING`. Unsupported record or protocol versions return `RESUME_UNSAFE`
- Resume verifies the recorded inputs and artifacts, reuses completed steps, and runs pending or failed ones. For an unknown transcription only, the complete recorded artifact set of `raw_transcript.txt`, `raw_sentences.txt`, and `raw_transcript.json` may establish local completion after a lost final message. Go independently verifies every path, size, hash, and usable transcript format. Dry run plans reuse without rewriting the record; explicit resume persists the reconciled completion without another upload
- An unanswered transcription without that complete evidence, any unknown summary, a changed input or artifact, or an unexpected file at a path that resume would write returns `RESUME_UNSAFE`, retaining the available artifacts. `--force` overrides none of these refusals. The existence of an unrecorded file alone never proves paid completion
- An `unknown` `summarize` step names `ncly transcript summary run <run-id>` in its hint. An `unknown` `transcribe` step names `ncly run view <run-id>`, because a new transcription may bill Deepgram again

The transcript domain defines these `inputs` keys:

| Key | Type | Meaning |
|---|---|---|
| `source` | string | `youtube` or `zoom` |
| `output` | string | The absolute output folder |
| `audio` | object | `transcript run zoom` only. The `path`, `size_bytes`, and `sha256` of the first `*.m4a` of the meeting folder in name order, selected once by Python. A folder without one is `NOT_FOUND` |
| `prompt` | object or null | The prompt `name` and the `sha256` of its text, or `null` when the summary is skipped |
| `profile` | object or null | The resolved profile: `name`, or `null` when it has none, `harness`, `provider`, a key omitted when there is none, `model`, and `effort`. `null` when the summary is skipped |
| `source_run` | string | `summary run` only. The run ID of the source run |

A Zoom run records this:

```json
{
  "format_version": 1,
  "run_id": "20261007-214501-3f9a2c",
  "path": "transcript run zoom",
  "status": "completed",
  "started_at": "2026-10-07T21:45:01Z",
  "updated_at": "2026-10-07T21:49:12Z",
  "ncly_version": "v0.1.0",
  "attempts": 1,
  "inputs": {
    "source": "zoom",
    "output": "/Users/me/meetings",
    "audio": {"path": "/Users/me/Documents/Zoom/2026-05-03 14.46.55 Weekly sync/audio1.m4a", "size_bytes": 48213504, "sha256": "<sha256>"},
    "prompt": {"name": "short_summary", "sha256": "<sha256>"},
    "profile": {"name": "everyday", "harness": "claude", "model": "<model id>", "effort": "medium"}
  },
  "tools": {"transcript": "4.1.0", "yt-dlp": "2026.7.4", "claude": "2.1.291"},
  "items": [
    {
      "key": "/Users/me/Documents/Zoom/2026-05-03 14.46.55 Weekly sync",
      "output_dir": "/Users/me/meetings/Weekly_sync",
      "steps": [
        {"id": "transcribe", "status": "completed", "started_at": "2026-10-07T21:45:20Z", "finished_at": "2026-10-07T21:47:58Z", "artifacts": [
          {"path": "/Users/me/meetings/Weekly_sync/raw_transcript.txt", "size_bytes": 51234, "sha256": "<sha256>"},
          {"path": "/Users/me/meetings/Weekly_sync/raw_sentences.txt", "size_bytes": 52011, "sha256": "<sha256>"},
          {"path": "/Users/me/meetings/Weekly_sync/raw_transcript.json", "size_bytes": 403877, "sha256": "<sha256>"}
        ], "error": null},
        {"id": "summarize", "status": "completed", "started_at": "2026-10-07T21:47:59Z", "finished_at": "2026-10-07T21:49:12Z", "artifacts": [{"path": "/Users/me/meetings/Weekly_sync/short_summary-20261007-214501-3f9a2c.md", "size_bytes": 4096, "sha256": "<sha256>"}], "error": null}
      ]
    }
  ]
}
```

### ncly transcript summary run

`ncly transcript summary run <run-id>` summarizes saved transcripts as a new recorded operation. It resolves its profile first: `--profile`, then the source run's recorded profile, then the current `default_profile`. If none applies, a form asks in interactive mode; otherwise it exits 2 with `USAGE_INVALID` and a hint showing `--profile`. Until a profile is known it reads no item, writes no record, and bills nothing.

Before reading items or starting a summary, execution takes the source run lock and holds it until the new operation ends. A busy source returns 75 with `TEMPORARY` and a `run view` hint, before effects. Dry run only checks ownership and never holds or reserves the source for execution.

Its prompt name comes from `--prompt`, the recorded source prompt, then the source's current default, such as `youtube.prompt`. It reads the current prompt text: a missing prompt is `NOT_FOUND` before effects; a changed hash is allowed and stored in the new record. This is a new operation, while resume preserves the recorded prompt fingerprint.

An item is eligible when its transcription is completed or locally reconciles from the complete recorded artifact set defined in Steps and resume. Any other item fails in `results` with `RESUME_UNSAFE`. If none is eligible, the failed answer includes `source_run` and `results` but no `run_id`; no record or harness starts.

A mixed operation records every source item in its original order. Eligible items copy verified transcript references and original timestamps into a completed `transcribe` step, then add a pending `summarize` step. An ineligible item records only a failed `summarize` step with its `RESUME_UNSAFE` problem and its source `output_dir`, possibly `null`. That item never becomes executable in this operation. Resume reproduces its saved error while continuing eligible pending or failed summaries; including the rejected item later requires a new operation. This is transcript-domain handling, without a new shared status.

The new record keeps `source_run` for provenance and its own resolved prompt and profile. It copies no transcript content or source audio dependency.

When a new record is created, normal execution returns `run_id`, `source_run`, and `results` with the transcript item shape. Each summary goes to the recorded result folder as `<prompt-name>-<new-run-id>.md`, preserving source artifacts. `run resume` supports this record: it verifies copied transcript evidence and continues eligible pending or failed summaries without transcribing again. Unknown summaries still refuse resume and hint `ncly transcript summary run <source_run>`. This new operation may bill and starts only on an explicit call.

### Protocol with the Python program

`ncly` extracts the program to `~/.cache/nicely/python/<version>/` and runs it through `uv`, with the flags that [Output, prompts, and cookies](#output-prompts-and-cookies) lists. The program reads JSON Lines on stdin and writes JSON Lines on stdout. Its stderr carries diagnostics, which `ncly` shows only under `--verbose` or `NCLY_DEBUG`. Each message is one JSON object with `protocol_version`, `1` in M1, and `type`. A message with another version or an unknown type is a protocol failure.

| Message | Sender | Keys besides `protocol_version` and `type` |
|---|---|---|
| `request` | `ncly` | `source`: `youtube` or `zoom`; `urls`: ordered URLs, empty for Zoom; `zoom`: `latest`, `path`, and `recordings`, or `null`; `output`: absolute output parent; `temp_dir`: Go-owned temporary folder; `browser` or `null`; `dry_run`: boolean; `deadline`: UTC RFC 3339 execution deadline |
| `plan` | program | `items`: the complete resolved item set, in input order. Each has `item`, `output`, and, for Zoom, selected `audio` with `path`, `size_bytes`, and `sha256`. On preparation failure, `items` is empty and `code` and `detail` identify the program failure |
| `intent` | program | `item`, `step`, `name`: proposed result-folder basename, `tools`: actual program and `yt-dlp` versions, and, for Zoom, the unchanged `audio` from `plan`. No user-side output has been created |
| `ack` | `ncly` | `item` and `step` from the intent, plus `output_dir`: the final reserved or recorded result folder |
| `artifact` | program | `item`, `step`, `path`, `size_bytes`, and `sha256` |
| `step` | program | `item`, `step`, `status`, `code`, which is `null` for `completed`, `warnings`, a list of program warning codes, and `detail`, a diagnostic that `ncly` shows only under `--verbose` |

`item` is the URL as given or the absolute Zoom meeting folder selected once in `plan`. `step` is `transcribe` in M1. The key travels in `DEEPGRAM_API_KEY`, never in a message. Python puts all temporary audio and child scratch files under `temp_dir`, so Go can remove them after killing a child. It derives phase timeouts from the remaining `deadline` and starts no upload once that deadline has expired. Go owns the overall timeout verdict and process cleanup.

1. Go sends one `request`. Python returns exactly one initial `plan` in both modes, before downloads, output reservation, or paid calls. A preparation failure reports its code and exits 0 without effects or final item steps
2. Go validates the complete expected item set, order, uniqueness, selected Zoom audio, and output parent. A new Zoom selection is fixed for this execution. Resume passes the recorded folder with `latest: false`, and verifies the stored audio. Dry run creates no record or user directory and starts no download or service call. Python exits after its plan; dependency cache preparation follows the documented exception
3. For execution, Go durably creates the initial record from the validated plan, with pending steps. After emitting its plan, Python may prepare audio only in `temp_dir` while Go validates and records that plan. This free temporary work needs no second handshake. Before any user-side write or upload, Python sends `intent` and waits for `ack`
4. Go validates the intent against the fixed plan. For a new result folder, it selects and durably records the candidate `output_dir` and unknown step before reserving the directory under the resource lock. For resume, it revalidates and uses the recorded folder. It completes the persistence barrier before sending `ack` with that final folder. Failure prevents acknowledgement. Python stops before writing outputs or uploading when stdin closes or no matching acknowledgement arrives within 10 seconds
5. After saving files only under the acknowledged folder, Python sends one `artifact` per required file, then `step` with status `completed`. Artifact paths are absolute within that folder, sizes are nonnegative byte integers, and hashes are 64 lowercase hexadecimal characters. Go independently verifies those files and the complete required set, syncs them, and durably records completion. Each artifact's verified evidence is persisted when received, so a lost final step can reconcile only a complete recorded set
6. A failed item sends `step` with its status and program code: `pending` before intent, otherwise `failed` or `unknown` as the table defines. Go keeps its public code, message, and recovery hint on the record, including pending preparation failures. Exactly one terminal step is required for each planned item. Python exits 0 whatever their outcomes; Go derives the verdict. On Ctrl-C or SIGTERM, Python stops its children within 10 seconds and reports open steps when possible

A nonzero exit, missing initial plan, unexpected or duplicate item, illegal message order, missing final step, or malformed message is a protocol failure. Go keeps verified evidence, leaves acknowledged unresolved effects unknown, and exits 1 with `RUNTIME`. Untouched pending steps stay pending with their failure cause. Cancellation and timeout retain the verdict from [Programs that ncly runs](#programs-that-ncly-runs-m1). A dry-run plan or zero exit alone never proves that execution completed an item.

| Program code | Step status | Error code | Catalog ID |
|---|---|---|---|
| `recording_not_found` | `pending` | `NOT_FOUND` | `transcript.recording_not_found` |
| `download_temporary` | `pending` | `TEMPORARY` | `transcript.download_temporary` |
| `download_failed` | `pending` | `DOWNLOAD_FAILED` | `transcript.download_failed` |
| `deepgram_unreachable` | `failed` | `TEMPORARY` | `transcript.deepgram_unreachable` |
| `deepgram_key_rejected` | `failed` | `AUTH_REJECTED` | `transcript.deepgram_key_rejected` |
| `deepgram_failed` | `failed` | `RUNTIME` | `transcript.deepgram_failed` |
| `deepgram_unknown` | `unknown` | `RUNTIME` | `transcript.deepgram_unknown` |
| `transcript_unusable` | `unknown` | `RUNTIME` | `transcript.transcript_unusable` |
| `write_failed` | `pending` before the intent, `unknown` after it | `RUNTIME` | `transcript.write_failed` |

The only program warning code in M1 is `browser_cookies_skipped`, which `ncly` maps to `BROWSER_COOKIES_SKIPPED` with the catalog ID `transcript.browser_cookies_skipped`.

`deepgram_unreachable` means transport failed before audio was sent, such as a refused connection, DNS or TLS failure, or a received HTTP refusal with status 408 or 429. A received HTTP 408 response is distinct from a transport timeout without a response. `deepgram_key_rejected` means 401 or 403; `deepgram_failed` means any other 4xx refusal. Any 5xx status, including 503, or a timeout or lost connection after sending audio is `deepgram_unknown`.

The item's `TEMPORARY` describes its cause. Only the overall exit 75 permits repeating the whole invocation. After a reserved result folder, started paid request, or unknown effect, the normal answer retains that item cause in `results` and carries a top-level `RUNTIME` with exit 1. Inspect the saved run before choosing resume or a new operation.

## Examples

```bash
# Check a new machine, then fix what doctor reports
ncly doctor
ncly auth login deepgram

# An agent discovers skills, then reads one
ncly skill list --json | jq -r '.skills[].name'
ncly skill view transcript

# Plan a transcription without a key, a paid request, or a user-file write
ncly transcript run youtube --url "https://www.youtube.com/watch?v=VIDEO_ID" --dry-run --json

# Keep the full successful answer and run ID before extracting paths
ncly transcript run youtube --url "$URL" --json > answer.json 2> diagnostics.txt
jq -r '.run_id, .results[].output_dir' answer.json

# A missing key, seen by a script
ncly transcript run youtube --url "$URL" --json
# stderr: {"ok":false,"contract_version":1,"errors":[{"code":"AUTH_MISSING","message":"No Deepgram key found.","hint":"ncly auth login deepgram"}]}
# exit code: 78
```
