# Contract

This file is the contract that every command follows, core's and each extension's: what humans and agents type, what core and an extension exchange, and what scripts parse. Core's commands live in [core-spec.md](core-spec.md), and each extension's in its own spec, such as [transcript's](../../extensions/transcript/spec.md). The code tables name the milestone that brings each entry. A milestone becomes ready only once the sections that its cards read are written, here, in core-spec.md, or in its extension's spec.

## Primitives

Every command and every extension composes these primitives. Each belongs to one or both scopes of the contract: the agent contract, what a caller of the CLI types and parses, and the wire contract, what core and an extension exchange, as [D040](decisions/D040-wire-contract.md) decides. Each primitive is defined once, in the section that the table links.

| Primitive | Scope | Defined in |
|---|---|---|
| Domain, resource, and verb | Agent contract | [Usage](#usage) |
| Flag | Agent contract | [Global flags](#global-flags) |
| Manifest | Wire contract | [Manifest](#manifest) and [Command descriptions](#command-descriptions) |
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

`ncly --help` lists domains and commands with one line each. `ncly <domain> --help` lists the actions of that domain, and `ncly help <domain>` prints the same text. Every help page ends with two to five examples. From [M04](../milestones/M04-skill.md), the first example of `ncly --help` is `ncly skill view nicely`, so an agent that starts with the help finds the skill that teaches it the rest. `ncly` and `ncly <domain>` without a verb print their help on stdout and exit 0.

A usage line shows a required argument as `<name>`, an optional one as `[name]`, and a group of flags that needs exactly one of them as `(--latest | --path <folder>)`. The help of a command that writes, uses the network, or may bill says so in one line, and names `--dry-run` when the command supports it. Both come from the command's declaration, as `ncly describe` does, so help and discovery never disagree.

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
| `--run-id <id>` | Names the run of a recorded command, so that a repeated call continues that run, as [Run IDs](#run-ids) defines. |

Commands that call the network or run for minutes also accept `--timeout <duration>`. A duration is `30s`, `5m`, `2h`, or a number of seconds. Each command states its default in its help.

## Modes

Interactive mode applies when stdin and stdout are terminals, `--no-input` and `--json` both resolve off, as [Global flags](#global-flags) resolves a flag and its variable, and `CI` is not set. Every other case is non-interactive mode, so machine output never waits for a human, as [D046](decisions/D046-machine-output-never-asks.md) decides. Some agent tools run commands in a pseudo-terminal, so an agent sets `NCLY_NO_INPUT=1` in its environment.

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

`errors` describes the command's failure. The command computes its overall verdict first, then puts an error with the matching exit code first. When the failures of the command itself map to different exit codes, the most cautious one leads: `130` or `143`, then `1`, `78`, `2`, and `75` only when every failure is temporary. When several failures share that exit code, the first one in this order leads. For exit 1, `RESUME_UNSAFE`, then the answering extension's own codes in the order of its manifest, then `RUNTIME`. For exit 78, `CONFIG_INVALID`, `EXTENSION_INVALID`, `TERMINAL_REQUIRED`, `PREREQ_MISSING`, `CAPABILITY_UNSUPPORTED`, `KEYRING_UNAVAILABLE`, `AUTH_MISSING`, `KEY_NOT_GRANTED`, `AUTH_REJECTED`, then the extension's own codes, so the first fix that a human needs leads. For exit 2, `USAGE_INVALID`, `CONFIRMATION_REQUIRED`, `NOT_FOUND`, then the extension's own codes. Input order or the order in which failures arrive never determines retry safety. Item errors remain inside `results`. `warnings` holds objects of the same shape, may appear in any answer, and never changes the exit code.

| Key | Meaning |
|---|---|
| `code` | A stable code from the [error registry](#error-codes). Agents branch on this key. |
| `message` | Human text in the active language. Agents never parse it. |
| `hint` | The command or the action that fixes the problem. A command is never translated. |
| `next` | Optional. The command that the hint names, as an array of words, such as `["ncly","skill","list"]`, present only when that command runs as is, with no value left to fill in. Agents run it instead of copying a command out of the hint, as [D053](decisions/D053-next-command.md) decides. |

`next` never changes what the exit code allows. After exit 78, a human runs it. It never names a repeat of a call that the exit code forbids repeating. A `next` command that writes or bills still gets its dry run first.

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

In a recorded operation, the record settles conditions 2 and 3 for the steps that this invocation started: exit 75 requires every step whose intent this invocation persisted to declare the `none` or `repeatable` effect, as [Record format](#record-format) defines. Such a step counts even when it then failed, because it may have made part of its effect, such as a reserved result folder before a refused upload. A step that an earlier attempt completed, and that this invocation only reuses, does not count, because a repeat reuses it again.

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

Only the overall process exit 75 permits an automatic repeat of the same invocation. An item error describes that item's cause, not the safety of repeating the whole command. A caller retains the complete answer and `run_id` before extracting artifact paths. After fixing a recorded failure, it inspects the run and requests an explicit resume, through `ncly run resume` or through the same call with the same `--run-id`, instead of repeating the original command under a new run.

## Error codes

Each code maps to exactly one exit code. This table holds core's codes, and a milestone adds core's new codes here. Every extension may answer with these codes too. An extension's own codes start with its domain, such as `TRANSCRIPT_DOWNLOAD_FAILED` or `FLEET_HOST_DOWN`, live in its spec and its [manifest](#manifest), and map to one exit code each, with the meaning that this file defines.

The set of codes is open: later milestones add core codes, and every extension adds its own. A caller that meets an error code it does not know acts on the exit code, whose meaning never changes, and relays the hint. An unknown warning code is still a warning.

| Code | Exit | When | Since |
|---|---|---|---|
| `USAGE_INVALID` | 2 | Unknown command, bad flag, or a missing value in non-interactive mode | M00 |
| `CONFIRMATION_REQUIRED` | 2 | A step needs a confirmation in non-interactive mode, and `--force` is absent | M00 |
| `CONFIG_INVALID` | 78 | A config file has a syntax error or a value of the wrong type, or cannot be read | M00 |
| `EXTENSION_INVALID` | 78 | The extension of the named domain cannot run: its manifest or its `en` catalog is missing, unreadable, or invalid, or speaks another format or contract version | M01 |
| `TERMINAL_REQUIRED` | 78 | A step needs a human at a terminal, such as typing or granting a key | M00 |
| `RUNTIME` | 1 | Any other failure during work | M00 |
| `TEMPORARY` | 75 | A temporary failure for which the whole invocation meets Retry safety | M00 |
| `INTERRUPTED` | 130 | Ctrl-C stopped the command | M00 |
| `TERMINATED` | 143 | SIGTERM stopped the command | M00 |
| `NOT_FOUND` | 2 | The named command, skill, service, profile, component, run, prompt, or recording does not exist | M03 |
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
| `CONFIG_DEFAULTS_USED` | Discovery used defaults because a config file could not be read | M03 |

## Configuration

`config.toml` is the shared config: the setup the user wants, which can travel between machines, for example through dotfiles. `config.local.toml`, in the same folder, is the local config: what belongs to this machine only.

Core owns the top-level keys, such as `lang`. Each extension owns the table named after its domain, such as `[skill]` or `[agent]`, and its spec defines that table.

Precedence, highest first: flags, environment variables, `config.local.toml`, `config.toml`, then defaults. A table, such as `[agent.profiles.everyday]`, merges key by key across the files. A list, such as `[skill] paths`, from a higher source replaces the whole list below it. Each resolved value keeps its source: the flag, the variable, the file that set it, or the default. [ncly describe](core-spec.md#ncly-describe) reports it, and a hint about a configured value names it, such as the file that lists a missing folder.

A path in a config file is absolute or starts with `~/`, which stands for the home folder. Any other path is `CONFIG_INVALID`, because its meaning would depend on the folder where `ncly` runs.

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

The language comes from the first source that is set: `--lang`, `NCLY_LANG`, `lang` in the config, `LC_ALL`, `LC_MESSAGES`, `LANG`, and finally `en`. `ncly` picks the closest catalog, so `fr_CA.UTF-8` and `fr` both select `fr-CA` once that catalog exists. A language with no close catalog, including `C` and `POSIX`, falls back to `en` without an error. The pseudo-locale `en-XA` exists for tests: it marks every catalog string between `⟦` and `⟧`, and only an exact request selects it, so `en-GB` gets English. An extension picks among its own catalogs with the same rule, as [Text of an extension](#text-of-an-extension) defines.

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

- A **bundled** extension is a Go package compiled into `ncly`, with its manifest and catalogs embedded. Only what an agent needs to operate Nicely is bundled: `skill`, in M04
- An **external** extension is an executable named `ncly-<domain>`, with its manifest `ncly-<domain>.toml` and its catalogs in the folder `ncly-<domain>.locales/` beside it, as [D048](decisions/D048-static-manifest-and-catalogs.md) decides. Core finds it as [Discovery](#discovery) defines, and taps bring more from M12. The first-party ones, such as `agent` and `transcript`, live in `extensions/<name>/` in the Nicely repository, which is their official tap

Core commands and bundled extensions always win over an external extension with the same name. Discovery, from [ncly describe](core-spec.md#ncly-describe) in M03 and [ncly doctor](core-spec.md#ncly-doctor) in M05, reports the source it selected and every extension that another one hides.

An extension receives the environment that [Programs that ncly runs](#programs-that-ncly-runs) defines, which carries the global flags as variables. It answers with the envelope of [Output](#output) and the exit codes of [Exit codes](#exit-codes). It may use core's [error codes](#error-codes), and its manifest declares its own. It translates its own text into the language of `NCLY_LANG`. It may be written in any language: a Python extension can be one file with the shebang `#!/usr/bin/env -S uv run --script` and its dependencies inline.

An extension uses another one through the CLI, as an agent would: transcript summarizes with `"$NCLY_BIN" agent run --json`. It writes its run records in the shared format of [Record format](#record-format), through the SDK for a Go extension, so `ncly run` inspects every recorded operation alike. A Go extension imports only the public `sdk/` packages, so a bundled extension could leave the binary without a rewrite.

Each extension keeps its spec in `extensions/<name>/spec.md`, and ships a `SKILL.md` that teaches an agent to use it.

### Discovery

A domain is lowercase letters and digits that start with a letter, such as `transcript`, as a service name is. Core looks for a file named `ncly-<domain>` that the user may execute, or a link to one: first in the `extensions` folder of the data folder, such as `~/.local/share/nicely/extensions/`, then in each folder of `PATH`, in order. It skips an empty or relative `PATH` entry, a reserved name of [core-spec.md](core-spec.md#command-tree), and every name whose domain breaks the rule above, such as `ncly-hello.toml`. The first file found for a domain is its extension and hides any later one. Core never resolves the link: the manifest and the catalogs sit beside the path that it found.

To dispatch `ncly <domain>`, core reads that domain's manifest and catalogs only. `ncly --help` and completion read every one, and `ncly --help` lists the extensions in their own section, apart from core's commands. Core never starts an extension to learn its commands. An extension whose manifest or `en` catalog is missing, unreadable, or invalid stays its domain's extension: `ncly <domain>`, its help included, answers `EXTENSION_INVALID` with exit 78 without starting it, and `ncly --help` and completion leave it out. A broken extension never affects another domain.

### Manifest

The manifest is a TOML file that declares the extension's commands, in the same declarations as core's commands, as [Command descriptions](#command-descriptions) defines, and its own error codes. Its text fields hold catalog IDs, never text.

```toml
format_version = 1
contract_version = 1
domain = "hello"

[[commands]]
path = "hello"
summary = "hello.summary"

[[commands]]
path = "hello run"
summary = "hello.run.summary"
examples = ["ncly hello run Ada", "ncly hello run --greeting Hi Ada --json"]

[[commands.args]]
name = "name"
summary = "hello.run.name"

[[commands.flags]]
name = "greeting"
shorthand = "g"
value = "text"
summary = "hello.run.greeting"

[[commands]]
path = "hello check"
summary = "hello.check.summary"
result = { kind = "report" }

[[errors]]
code = "HELLO_GRUMPY"
exit = 1
```

| Key | Type | Meaning |
|---|---|---|
| `format_version` | integer | The manifest format, `1` |
| `contract_version` | integer | The contract version that the extension speaks, `1`, as `NCLY_CONTRACT_VERSION` carries it |
| `domain` | string | The domain, equal to the name of the executable after `ncly-` |
| `commands` | array | One declaration per command. Help lists them in alphabetical order, as it lists core's |
| `commands[].path` | string | The words after `ncly`, separated by one space, such as `hello run`. The first word is the domain, and the command whose path is the domain alone gives the domain's line in `ncly --help` |
| `commands[].summary` | string | The catalog ID of its one-line summary |
| `commands[].examples` | array | Full command lines, as strings that are never translated |
| `commands[].args` | array | The positional arguments in order, each with `name` and `summary`, a catalog ID |
| `commands[].flags` | array | The command's own flags, each with `name`, the long name without `--`, `summary`, a catalog ID, and the optional `shorthand`, one ASCII letter, and `value`, the name of the value, such as `text`. A flag without `value` is a switch |
| `commands[].result` | table | The command's result, the same object as `result` in [ncly describe](core-spec.md#ncly-describe), whose fields join with their first consumer. M01 reads `kind`: `object`, the default, `batch`, `report`, or `text`, which tells core how to check the answer |
| `errors` | array | The extension's own error codes, each with `code` and `exit`, in the order that decides which one leads an answer, as [Output](#output) defines |

A manifest is invalid unless every rule below holds. Core rejects it before work, as [Compatibility](#compatibility) asks of a format version that it does not support:

- `format_version` and `contract_version` are both `1`, and `domain` matches the executable's name
- Each word of a path is lowercase letters, digits, and hyphens, starting with a letter, and no path appears twice. The parent of each path is declared too, such as `agent profile` for `agent profile list`, down to the command whose path is the domain alone, so that help lists each one with its summary. A path with children and no arguments groups them: alone, it prints its help, as [Usage](#usage) says
- A command's `result.kind` is one of the four kinds of the table
- A flag's name, an argument's name, and a value's name follow the rule of a path word. A shorthand is one ASCII letter. No flag takes the name or the shorthand of a global flag, `--dry-run`, or `--run-id`, and no name or shorthand appears twice in one command. Core checks every declaration before it builds a command from it, so a broken one never reaches the parser
- Each own code starts with the domain in uppercase and an underscore, such as `HELLO_GRUMPY`, differs from every core code, appears once, and maps to `1`, `2`, or `78`. Exit 75 belongs to `TEMPORARY` alone, because [Retry safety](#retry-safety) judges the whole invocation, and 130 and 143 belong to the signals
- The `en` catalog holds every catalog ID that the manifest names

Core reads only the keys of this table and ignores any other, as [Compatibility](#compatibility) asks of readers, so a later milestone adds a field with its first consumer, such as the requirements that [M05](../milestones/M05-doctor.md) checks. A field that an older core must not ignore needs a new `format_version`, which that core rejects instead of misreading.

### Text of an extension

An extension keeps its text in catalogs, in the format of core's: one TOML file per language, named after its language tag, such as `en.toml` or `fr-CA.toml`, whose entries each have a description for the translator. An external extension's catalogs sit in `ncly-<domain>.locales/` beside its manifest, and a first-party extension keeps them in `extensions/<name>/locales/`. The `en` catalog is required. The files list the languages that the extension speaks, so the manifest declares none, and a translation that cannot be read is skipped, so its entries fall back to `en`.

Core renders an extension's help and completion descriptions from its manifest and catalogs, so help never starts it. `NCLY_LANG` carries the requested language as a language tag, such as `fr-CA` for `LANG=fr_CA.UTF-8`, and `en` when no source names a known language, as for `C` or `POSIX`. Core and each extension then pick their own closest catalog with the rule of [Configuration](#configuration), falling back to `en` independently: a French request shows core in French and an extension that speaks only English in English. An entry that a catalog lacks falls back to its `en` entry. Under the pseudo-locale `en-XA`, core marks the text of an extension's `en` catalog as it marks its own.

### Dispatch

Core validates a call to an extension from its manifest, as it validates its own commands: an unknown command or flag, a flag without its value, or an extra argument fails with `USAGE_INVALID` before the extension starts. `--help`, `--version`, `ncly help <domain>`, and a group without a verb never start it.

Core then starts the executable with the canonical arguments, as [D049](decisions/D049-canonical-arguments.md) decides: the words of the command path after the domain, then each flag that the caller set, in the order of the manifest, as one `--<name>=<value>` word, then `--`, then the positional arguments, including the operands that followed the caller's `--`. A switch passes as `--<name>=true` or `--<name>=false`, and a repeated flag passes once, with its last value. A flag that the manifest declares variadic, such as the `--url` of `ncly transcript run youtube`, instead passes one word per value, in the caller's order. The manifest key that declares it arrives with its first consumer, and a Go handler receives the values in order, as [SDK](#sdk) defines. `--dry-run` and `--run-id`, which a command accepts when its `modes` sets `dry_run` or `recorded`, pass the same way after the own flags, in that order, when the caller set them. A manifest declares `--force`, `--output`, and `--timeout` as own flags of the commands that accept them, because only the command knows whether it confirms, writes a file, or runs long. A Go handler finds each of these flags in `Call.Flags` under its long name, where the SDK's record support reads `run-id`. The other global flags never reach the extension as arguments: their resolved values reach it as variables. For example, `ncly -v hello run -g Hi Ada --json` starts `ncly-hello run --greeting=Hi -- Ada`, with `NCLY_VERBOSE=1` and `NCLY_JSON=1`.

The extension inherits stdin and the working directory of `ncly`, and receives the environment of [Programs that ncly runs](#programs-that-ncly-runs). Core reports no config error before it starts an extension: it resolves `lang` as help does, skipping a config file that it cannot read. The extension reports `CONFIG_INVALID` when it reads a broken file, including a wrong core key such as `lang`, as [Configuration](#configuration) requires of every command that reads the config.

### Answer check

Core checks what an extension answers, as [D050](decisions/D050-answer-check.md) decides. Without `--json`, the extension writes to the stdout and stderr of `ncly` directly, so a human sees its text, styles, and forms unchanged, and core checks only its exit code. A command whose result kind is `text` ignores `--json`, as [Output](#output) allows, so core treats it the same way under `--json`.

With `--json`, core holds the extension's stdout until the extension exits, so a broken answer never reaches the caller. It reads stderr line by line and holds the latest line: when a new line arrives, it forwards the held one. When the answer belongs on stdout, core forwards the held line too. Otherwise the held line is the answer, which core checks. A last line without a newline counts as a line. An extension follows [Output](#output), so it writes nothing else to stderr without `--verbose` or `NCLY_DEBUG`, but core forwards what it writes anyway, so a human still sees it.

Core accepts the answer only when every rule below holds:

1. The exit code is one of [Exit codes](#exit-codes): 0, 1, 2, 75, 78, 130, or 143
2. The answer is one line that holds one JSON object, with `ok`, a boolean that is `true` exactly when the exit code is 0, and `contract_version`, the integer `1`, in exactly one place. A success puts it on stdout. A failure puts it on the last line of stderr, with stdout empty. A report puts its findings on stdout with exit 0 or 78, as [Output](#output) says, and a report command that fails itself answers as any failure, on stderr
3. An answer on stderr has a non-empty `errors`, and an answer on stdout has none
4. `errors` and `warnings`, when present, are arrays of objects that each hold the strings `code`, `message`, and `hint`, and, when present, `next`, a non-empty array of strings. Each code is uppercase letters, digits, and underscores
5. `results`, when present, is an array of objects that each hold the boolean `ok`, which is `false` exactly when the item has a non-empty `errors`. A success has no failed item
6. An error code that starts with the extension's domain and an underscore is declared in its manifest. A warning code is checked for its form only
7. When core knows the leading error code, as its own or as a declared one, it maps to the exit code. No top-level error code that core knows maps to a more cautious exit code, in the order of [Output](#output), whatever the codes that core does not know
8. When an item failed, the exit code is 1 led by `RUNTIME`, 75 led by `TEMPORARY`, 130 led by `INTERRUPTED`, or 143 led by `TERMINATED`, as [Output](#output) defines for a batch. Exit 75 holds only when every error, at the top and in every item, is `TEMPORARY`

An extension that a signal ended counts as exiting with 128 plus the signal's number, as a shell reports it. Core passes a valid answer on unchanged. Anything else is a protocol failure, and so is an extension that core cannot start, such as a script whose interpreter is missing: core drops the output that it holds and answers `RUNTIME` with exit 1, because the extension may already have produced effects, and the message names the rule that broke. Without `--json`, an exit code outside the list is a protocol failure too, and core prints its error block after the extension's text.

When `ncly` receives Ctrl-C or SIGTERM before the extension exits, the first signal decides the verdict: core answers 130 with `INTERRUPTED` or 143 with `TERMINATED`, whatever the extension's exit code. A valid answer keeps its data keys, its `results`, and its other errors after the leading one. Without a valid answer, core answers alone, with no data key. Without `--json`, core prints its own error block only when the extension's exit code differs from the signal's. A signal that arrives after the extension exited changes nothing: core keeps its handler registered and publishes the answer that it has decided.

### SDK

The SDK is the Go implementation of this contract, in the public packages of `sdk/` that the [code map](architecture.md#code-map) lists, as [D040](decisions/D040-wire-contract.md) decides. Core uses it too, so each shared layer has one implementation. A Go extension imports `sdk/` only, never core's `internal/`, which `just lint` enforces.

A Go extension is one package that exports one `extension.Extension` value, built from `manifest.toml` and `locales/` in its folder with `go:embed`. The API of `sdk/extension` comes down to this:

```go
type Extension struct {
	Manifest []byte             // manifest.toml
	Locales  fs.FS              // holds locales/<lang>.toml
	Handlers map[string]Handler // one per command that runs, keyed by path, such as "hello run"
}

// Handler runs one validated call. ctx ends on the first Ctrl-C or SIGTERM.
type Handler func(ctx context.Context, call Call) contract.Outcome

type Call struct {
	Path    string              // such as "hello run"
	Args    []string            // the positional arguments
	Flags   map[string]string   // each flag that the caller set, by long name, except a variadic one; a switch holds "true" or "false"
	Lists   map[string][]string // each variadic flag that the caller set, by long name, with its values in the caller's order
	JSON    bool
	Mode    contract.Mode
	Printer *i18n.Printer // in the language of NCLY_LANG
	Stdin   io.Reader
	Stdout  io.Writer // human text only; the SDK prints the answer
	Stderr  io.Writer
}

func Main(e Extension) // runs the extension as this process, then exits with its code
func Check(e Extension) error
```

`Main` parses the call with the manifest, prints the help of a group, and calls the handler of the path. A variadic flag reaches the handler in `Lists` alone, with its values in the caller's order, so `Flags` keeps one string for each other flag, and no handler reads one value of a variadic flag as its only one. `Main` prints the result by its kind, from the command's `result`:

- `object` and `batch`: in human mode, the handler writes its text to `Stdout`. With `--json`, it writes nothing there, and the SDK prints the answer from the outcome
- `report`: in human mode, the handler writes its report to `Stdout`. With `--json`, the SDK prints the findings from the outcome's data on stdout, in an answer without `errors`, as [Output](#output) says. In both modes, the handler returns its findings in that data, and sets the outcome's `HumanNeeded` when a finding needs a human, which makes the exit code 78 instead of 0, and `ok` false with `--json`. Core's doctor uses the same outcome from M05
- `text`: the handler writes its text to `Stdout`, whatever `--json` says

A failure of the command itself prints as an ordinary failure for every kind: the error block, or the answer on stderr with `--json`. On Ctrl-C or SIGTERM, the SDK ends `ctx`, stops the extension's programs as [Programs that ncly runs](#programs-that-ncly-runs) defines, and sets the outcome's signal, so a handler returns what it has. A handler that has not returned once the 15 seconds of cleanup and the 2 seconds of the drain limit have passed, the time that the SIGKILL of cleanup gives its programs to end, gets no more time: the SDK answers `INTERRUPTED` or `TERMINATED` alone and exits. Under `ncly`, core's own 15 seconds govern, and either path gives the same answer, because both carry the signal's code and no data key. Run directly, `ncly-<domain>` also parses the global flags and the ordinary syntax that core accepts, because canonical arguments are a subset of it, and `--version` prints the version of the module that built it. `Check` reports a manifest that breaks the rules of [Manifest](#manifest), a command that runs without a handler, a handler for no such command, and a catalog that breaks the catalog rules below. Every Go extension has a test that calls it. `Main` checks the manifest and the handlers when it starts and answers `RUNTIME` for those faults, while a missing translation falls back to `en` at run time, as [Text of an extension](#text-of-an-extension) says.

The same value serves both transports, so they cannot drift, as [D051](decisions/D051-bundled-child-process.md) decides:

- **External:** the extension's `main` calls `extension.Main`. Installing it writes the embedded manifest and catalogs beside the executable, as `ncly-<domain>.toml` and `ncly-<domain>.locales/`
- **Bundled:** `cmd/ncly` passes the value to core in an explicit list, never through an `init` side effect. Core reads its embedded manifest and catalogs, validates its call, then starts the `ncly` binary itself as the extension's program, with `ncly-<domain>` as its first argument, `argv[0]`, through the program runner. Started under the name of a bundled extension, `ncly` runs that extension's `Main` instead of core, as busybox picks its applet from its name. A bundled extension therefore gets the canonical arguments, the environment, the process group, the cancellation, and the answer check of an external one

The SDK accepts the error codes that the manifest declares, and every warning code that starts with the domain and an underscore, beside core's codes, so the verdict and the answer follow [Output](#output). The SDK's catalog holds the text that the SDK renders: help pages, global flags, usage errors, error labels, outcome summaries, size units, and the error of an extension that does not match its own manifest. An extension's printer takes an ID that starts with its domain and a dot from the extension's catalogs, and any other ID from the SDK's, so an extension never overrides the SDK's text. Core's printer takes an ID from core's catalog when that catalog holds it, and from the SDK's otherwise. The catalog rules are these: every entry has a description, every language holds exactly the IDs of `en`, an extension's IDs start with its domain and a dot, and core's IDs differ from the SDK's. `Check` applies them to an extension, and one parity test applies them to every catalog folder of the repository, core's, the SDK's, and each extension's.

## Operations

The domain prepares an operation from resolved inputs and configuration, performs its steps, and verifies their results. Dry run uses that same preparation and reports pending checks, such as authentication. Execution repeats checks affected by changed inputs or destinations before applying effects. A dry run is not a reservation or a guarantee that the environment will stay unchanged.

Preparation checks the prerequisites that the invocation selects, such as a program or another extension, so a dry run fails with the code that execution would answer, and `pending_checks` holds only what execution alone can check, such as a key. A dry run is therefore the readiness check of one invocation, and [ncly doctor](core-spec.md#ncly-doctor) the readiness check of the machine, as [D034](decisions/D034-shared-preparation.md) decides. An agent needs no doctor call before each command.

Commands that incur a bill or can leave reusable partial work record an execution before the first such effect. The first are `ncly agent run` and transcript. Read-only discovery and dry runs create no run record. Each domain declares its recording and resume support before code is written.

The SDK's record support owns run IDs, durable records, step transitions, and result publication, for core and every Go extension. It persists each step's intent, which turns the step `unknown`, and its final status, in the order that [Records and evidence](#records-and-evidence) requires, and derives [Retry safety](#retry-safety) from the effects that the steps declare, so a domain never states retry safety a second time. The domain names its steps, declares the effect of each, verifies their results, and owns the evidence that makes them resumable. The program runner owns child processes, their environment, timeouts, and signals. It does not decide whether a business action is safe to repeat.

### Records and evidence

A run record contains its format version, `run_id`, command path, status, step outcomes, and artifact references. It also keeps the non-secret input fingerprints, resolved profile and relevant tool versions needed to check a resume. A call with the same run ID compares the record's `inputs` and the ordered keys of its items, such as the URLs of a transcription, as [Run IDs](#run-ids) defines. Together they hold every input that tells two calls apart, so the domain fingerprints each file or text input, such as a task, rather than keeping its path alone. Keep only information needed to explain or continue the operation. Full task bodies, transcripts, raw environment dumps, and keys are not copied into the record.

Every effect of a recorded operation happens inside a step. Persist an intent before an effect that cannot safely be repeated, and persist completion only after checking its result. If the process dies between those writes, the effect is unknown until evidence resolves it. A missing response is never proof that nothing happened. If the record cannot be written, stop before starting the next effect.

Records distinguish `running`, `completed`, `failed`, `interrupted`, and `unknown`. A step distinguishes `pending`, `completed`, `failed`, and `unknown`. Pending means its effect has not started; preparation may have failed. A completed step names the evidence needed to reuse it, such as an artifact path and content fingerprint. The domain may leave an in-flight effect as unknown until it has a verified result.

Keep records under the XDG state directory and give every run its own temporary directory. Write each record atomically under a lock. Lock shared destinations or configuration at the point of mutation and revalidate while holding the lock. A conflict follows Retry safety, including effects already completed elsewhere in the run.

Records survive the process. Nicely retains them until the user removes their JSON files and never automatically deletes output artifacts. Listings bound both record count and item-key previews. `ncly run resume` of a missing record or of an unsupported record version never starts the original operation again.

### Record format

Each run has one record, `runs/<run_id>.json` in the state directory, such as `~/.local/state/nicely/runs/`. A run keeps an artifact that no user folder holds, such as the answer of `ncly agent run`, in the folder `runs/<run_id>/` beside its record. The record is one JSON object. While holding `runs/<run_id>.lock`, the writer, core or an extension through the SDK, writes a temporary file in the same directory, syncs that file, renames it, then syncs the parent directory. Newly created state directories are synced with their parents too. This persistence barrier finishes before acknowledging a non-repeatable effect. A failed barrier prevents the effect. Atomic visibility alone is not durable acknowledgement.

The platform helper uses the durable flush supported by the system. On macOS, it also uses `F_FULLFSYNC` after syncing metadata, because [Apple's fsync documentation](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/fsync.2.html) distinguishes an ordinary sync from flushing the drive cache. A failed required flush stops the operation. Persistence relies on the filesystem and device honoring these operations; Nicely cannot promise survival of arbitrary hardware failure.

The execution process holds the run lock until it exits. A new operation that consumes a source run also holds that source's lock while it executes, to exclude a concurrent source resume. The operating system releases each lock when its owner dies. Lock files may remain; their existence alone does not establish ownership. Before recording a step as completed, the domain independently verifies its artifacts and syncs their files and containing directories.

`run_id` is the ID that the caller chose or one that the SDK generated, as [Run IDs](#run-ids) defines, such as `20261007-214501-3f9a2c`. Readers never parse it. A run's temporary directory sits in the system temporary directory, carries the run ID in its name, and is removed when the run ends, including after Ctrl-C or SIGTERM.

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

A supported resume reuses verified completed steps and continues only steps whose execution is known to be safe. It records each attempt under the same run ID. Changed or missing evidence returns `RESUME_UNSAFE` with an explanation and the artifacts still available. In a batch, a refusal that concerns one item, such as its unknown summary, fails that item in `results` with `RESUME_UNSAFE`, while the other items continue, and a changed input of the whole run refuses the whole resume. `--force` does not override uncertainty or make a non-repeatable effect safe.

For example, a saved transcript and a summary that was never started permit a resume of the summary alone. An unanswered paid summary request is unknown and is not sent again automatically. A domain may resolve an unknown step from complete, independently verified local artifacts, from the record of a child run that the step's intent names, or from a reliable external lookup. This establishes a reusable result; it never authorizes redispatch of an unresolved request. Otherwise the user must inspect the evidence and explicitly choose any new operation that might repeat a charge.

Nicely has no background worker or scheduler. A record makes work inspectable after a session ends. It does not keep the process alive or promise that an arbitrary harness task can resume.

### Run IDs

A recorded command accepts `--run-id <id>`, so that its caller knows the run's ID before the run starts, as [D052](decisions/D052-caller-run-ids.md) decides. Without it, the SDK generates one. An ID holds 1 to 64 lowercase letters, digits, and hyphens, and starts with a letter or a digit. Any other value is `USAGE_INVALID`.

A call whose ID no record holds starts a new run under that ID. A call whose ID a record holds starts no new run. Before any effect, it takes the first case that applies:

1. The record's format version is unsupported: `RESUME_UNSAFE`
2. The record's command path, `inputs`, or ordered item keys differ from those that the call prepares: `USAGE_INVALID`, because the ID names another call
3. Another process holds the run: exit 75 with `TEMPORARY` and a `run view` hint, as for `ncly run resume`
4. The run completed: its verified saved result, without running a step again
5. The command supports resume: it continues the run, as [Inspection and explicit resume](#inspection-and-explicit-resume) defines
6. Otherwise: `RESUME_UNSAFE`, with the artifacts still available

A repeated call with the same ID therefore never repeats an effect that the record does not prove safe, and a dry run with it describes what the call would do, as `ncly run resume --dry-run` does. After lost output, a caller that chose the ID reads the run with `ncly run view <id>`, or makes the same call again, and needs no search with `ncly run list`.

A step that starts a recorded command for its own work, such as transcript calling `ncly agent run`, chooses the child's ID and persists it with the step's intent, before the call. It passes the child only inputs that a resume reproduces, such as an absolute folder rather than its own working directory, so a repeated call matches the child's record. After a lost answer, the step first reads the child's record, such as with `ncly run view <id> --json`. Only an existing record permits the same call with the same ID, and a completed child then returns its saved result without a new request. Any other outcome leaves the step `unknown` with `RESUME_UNSAFE`, whose `next` names the explicit new operation. A refusal of the repeated call, such as `USAGE_INVALID` for a changed input, proves nothing about the first call, and neither does a missing record: the step's intent shows that the child may have started, and the user may have deleted its record after the child billed. While another process holds the child run, the step stays `unknown`, and its item fails with `TEMPORARY` instead, whose `next` is `ncly run view <id>`, because the child may still finish and bill. A later resume repeats the check.

Deleting a record forgets its run: a caller's later direct call with its ID starts a new run, while a step that persisted the ID refuses, as the paragraph above says.

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
- The variables of the global flags carry their resolved values: `NCLY_JSON`, `NCLY_NO_INPUT`, and `NCLY_VERBOSE` equal `1` when their flag resolves on and are removed otherwise, `NO_COLOR` equals `1` when color resolves off and is removed otherwise, and `NCLY_LANG` holds the requested language tag, as [Text of an extension](#text-of-an-extension) defines. `NCLY_DEBUG` passes unchanged. `NCLY_VERSION` and `NCLY_CONTRACT_VERSION` identify the release and public protocol.
- `NCLY_BIN` holds the absolute path of the running `ncly`, so that a program calls another domain through the same host, as cargo passes `CARGO` to its subcommands
- A harness receives `NCLY_AGENT_DEPTH` plus one, as the [agent extension](../../extensions/agent/spec.md#depth-limit) defines

A key may be optional, such as a harness key: the extension receives it when it is granted and present, and otherwise the harness uses its own login. An unavailable keychain for an optional key does not block that fallback.

**Process groups.** The program runner starts each program in its own process group, so a Ctrl-C at the terminal reaches `ncly` alone, and the runner signals the program. A program that may read the terminal stays in the process group of `ncly` instead: an extension, which asks through its forms, and an install command that doctor runs after a yes, which may ask for a password. A process outside the terminal's foreground group stops when it reads the terminal, as [D047](decisions/D047-terminal-process-group.md) explains. A program can therefore receive the same signal more than once, from the terminal, from `ncly`, and from the extension that started it, and it treats a repeat during its cleanup as no new event.

**Cancellation.** On Ctrl-C or SIGTERM, `ncly` sends the signal to every program it started and to every descendant that it can still reach. It sends SIGKILL to what still runs 15 seconds later, which leaves a Python program its own 10 seconds to stop its children. Before each signal, `ncly` finds descendants through their parent process, so a descendant that left the process group of its parent, as the children of transcript's Python program do, still receives the signal. On Linux, `ncly` registers as a child subreaper, so it also reaches a descendant whose parent already exited. macOS has no subreaper, so a descendant whose parent exited before the signal is the one exception there, and it keeps running. `ncly` keeps every finished output and removes only its own temporary files. It exits 130 with `INTERRUPTED` or 143 with `TERMINATED`, and under `--json` the answer keeps the data keys that apply, such as `output_dir` or the `results` of finished items. A second signal during this cleanup is ignored. A SIGKILL sent to `ncly` itself leaves no time to clean up.

**Output streams.** A descendant that inherits a program's stdout or stderr can hold it open after the program exits, as an orphan on macOS can. `ncly` stops reading the streams that it reads from a program 2 seconds after that program exits, including the exit that the SIGKILL of cleanup forces. A descendant therefore never keeps `ncly` waiting, and output written after that point is lost. This drain limit is its own clock, apart from the 15 seconds of cleanup. A stream that `ncly` passes on without reading, such as an extension's stdout without `--json`, stays with the descendant, which may still write to the terminal after `ncly` exits.

**Normal end.** An extension answers only after every program that it started has ended, except a viewer that it starts detached for the user, such as the folder that transcript's `--open` shows. The SDK's program runner waits for each program that it started. When a program ends normally while a descendant still runs, `ncly` neither waits for that descendant nor stops it, because it cannot tell a forgotten program from a window that the user keeps open. The output stream rule keeps such a descendant from blocking `ncly`.

**Timeout.** The timeout covers the command's work and starts before child execution. On expiry, use the same process-tree cleanup as cancellation. A temporary failure before any paid request exits 75 only when Retry safety holds. Otherwise it exits 1. A caller's SIGTERM remains exit 143. Record finished work and unknown effects before returning when the process can still do so.

**Protocol failures.** An extension answers core as any command answers its caller, and core checks it as [Answer check](#answer-check) defines. A Python program that an extension embeds returns versioned machine messages only, which the extension's Go code maps into this public contract. Missing, malformed, or incompatible messages never count as success. A launched program may already have produced effects, so preserve available evidence and use a runtime failure unless safety can be established. Harness adapters declare the modes they can actually enforce. An unsupported tools-off mode fails with `CAPABILITY_UNSUPPORTED` before launching a task with outside material.
