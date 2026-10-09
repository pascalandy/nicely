# core spec

This file specifies the commands of core. They follow [contract.md](contract.md), as every extension's commands do. The command tree names the milestone that brings each command.

## Command tree

```
ncly
├── completion zsh|bash|fish        M00
├── <domain> [arguments]            M01
├── skill list|view|<name>          M03
├── describe [command...]           M04
├── doctor [component]              M05
├── auth login|logout|status        M06
├── run list|view                   M10
├── run resume <run-id>             M15
└── tap add|list|sync|remove        M12
```

`<domain>` is an extension. `skill` is the bundled extension, as [Extensions](contract.md#extensions) defines, and [extensions/skill/spec.md](../../extensions/skill/spec.md) specifies it. External extensions, such as `agent` and `transcript`, list their commands in their own specs.

Core reserves the names of its own commands and of the bundled extensions, the rows marked core or bundled in the [domain table](architecture.md#domains-and-commands), plus `help`, `version`, and `config`. An external extension never runs under a reserved name. A name that is neither a command of this version nor an installed extension fails with `USAGE_INVALID`, including a reserved name whose command arrives in a later milestone.

## ncly describe

```
ncly describe [command...] [--json]
```

Without a command path, returns concise descriptions of installed commands. A path such as `transcript run youtube` returns that command's declaration. A path that names a group, such as `transcript`, returns the concise entries of the commands under it. The JSON envelope contains `commands`, an array. Text is localized, while machine identifiers follow Compatibility. Unknown paths use `NOT_FOUND`.

Each entry has `path`, such as `transcript run youtube`, `summary`, `since`, the release that added the command, such as `v0.1.0`, and `source`. `source` is `core`, `bundled`, or `extension`, and a reader treats an unknown `source` as not core. An external extension's entries come from its manifest. A detailed entry adds these keys:

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

`requires` takes its name and its keys from the manifest, so a static manifest describes an extension's commands the same way. Discovery works with default configuration when a config file is invalid, adds `CONFIG_DEFAULTS_USED`, and does not claim to have resolved a profile that it could not read.

Schemas use JSON Schema draft 2020-12, with all references resolved within the returned descriptor and no remote schema fetch. `keys` comes from that result definition, excluding `ok`, `contract_version`, `errors`, and `warnings`; a successful batch includes `results`. Flag types use the same scalar types as arguments. Schemas allow unknown optional fields as Compatibility requires. Declarations cover the current commands only; core needs no schema generator or custom constraint language for future domains. Validation uses required flags and groups from this declaration, such as transcript's, where exactly one of `latest` and `path` is required for Zoom. Examples do not replace these constraints.

## ncly run

```
ncly run list [--path <command-path>] [--key <item-key>] [--since <time>] [--limit <count>] [--json]
ncly run view <run-id> [--json]
ncly run resume <run-id> [--dry-run] [--force] [--timeout <duration>] [--json]
```

- `list` returns `runs`, sorted by `started_at` descending, then `run_id` ascending, with at most `--limit` entries, 20 by default. The count must be a positive integer. Filters combine with AND before limiting: `--path` exactly matches the command path, `--key` matches any item key, and `--since` includes runs whose `started_at` is at or after that RFC 3339 time. An invalid time is `USAGE_INVALID`
- Each list entry has `run_id`, `path`, `status`, `started_at`, `updated_at`, `active`, `item_count`, and `keys`. `keys` previews at most the first three item keys in input order; filters still inspect every key. The top-level `more` boolean is true when additional matching records exceed the limit. Increase the limit or narrow the filters, without a database or index service
- `view` returns `run`, the record of [Record format](contract.md#record-format) plus `active`, including step outcomes and artifact references. A successful inspection exits 0 even when the saved run failed. Its top-level `ok` describes the inspection, and `run.status` describes the saved operation
- `resume` hands the run to the extension that owns its command path, follows Operations, and returns that extension's result with the same `run_id`. Dry run only describes the remaining steps and checks, as [Dry-run answer](contract.md#dry-run-answer) defines. The original command's timeout default applies unless overridden
- An unknown run uses `NOT_FOUND`. An unsafe or unsupported resume uses `RESUME_UNSAFE`. For a command that supports resume, a completed run returns its verified saved result without running its steps again
- When another process holds the run, `resume` exits 75 with `TEMPORARY` before any effect, with a hint naming `ncly run view <run-id>`. `active` is true while a process holds the run lock; inspecting a busy run does not permit a retry loop

To find a run whose stdout was lost, filter by the original path and an item key, such as `ncly run list --path 'transcript run youtube' --key <url> --json`. Add `--since` when the retained invocation start time is known; a fresh session may omit it. Check `more`, then confirm candidates with `run view`, including their recorded inputs and complete item list. Report ambiguous matches instead of choosing the newest one. A call that failed during preparation before its first record write has no record. Absence from one limited list proves neither absence of a run nor absence of billing.

## ncly completion

```
ncly completion zsh|bash|fish
```

Prints the completion script on stdout. The Homebrew formula installs these scripts, and each Linux archive ships them in `completions/`. The AUR package will install them too, once it ships. Users run this command only for a manual setup.

Completion also suggests the values that declarations list, such as skill names, auth services, and doctor components, and an extension's values, such as profiles and transcript prompts.

## ncly doctor

```
ncly doctor [component] [--live] [--timeout <duration>] [--json]
```

| Argument or flag | Effect |
|---|---|
| `component` | `core`, `auth`, `extensions`, or the domain of an extension that declares checks, such as `skill` or `transcript`. Without it, doctor checks all of them. |
| `--live` | Adds network checks against free endpoints only, such as listing Deepgram projects. It never calls an endpoint that bills. |
| `--timeout` | Bounds the checks, including local version probes and free network checks, 30 seconds by default. Installation after human confirmation is separate from those checks |

- The default checks stay on the machine: binaries and their versions, keys present, the keychain answering, and the config files parsing.
- Invalid or unreadable config produces the failed `core.config` finding with `CONFIG_INVALID`, even for a selected component. Doctor continues independent checks and omits checks whose required configuration could not be resolved. It never reports those omitted checks as passing
- When the keychain is unavailable, its check is `warn` if every key that a component needs comes from the environment, and `fail` with `KEYRING_UNAVAILABLE` otherwise.
- The `extensions` component lists every extension that core finds and its source, every extension that another one hides, every incompatible manifest, the programs that each manifest requires, and every grant that no extension declares anymore. It starts no extension.
- An extension's own checks, such as those of the [skill component](../../extensions/skill/spec.md#doctor-checks), appear under its domain.
- In interactive mode, when a tool is missing and its install command for this system is known, doctor shows the command, such as `brew install ffmpeg` or `sudo pacman -S ffmpeg`, and runs it only after a yes.
- In non-interactive mode, doctor never installs anything. The hint of each failed check holds the command.
- Doctor is a report command. It exits 0 when no check fails and 78 when at least one check needs a human. When doctor itself fails, it exits 1 with `errors`.

```json
{"ok":false,"contract_version":1,"checks":[{"id":"transcript.uv","component":"transcript","status":"pass","message":"uv found"},{"id":"auth.deepgram","component":"auth","status":"fail","code":"AUTH_MISSING","message":"No Deepgram key found.","hint":"ncly auth login deepgram"}]}
```

`status` is `pass`, `warn`, or `fail`. A warning never changes the exit code.

## ncly auth

```
ncly auth login <service> [--stdin] [--dry-run] [--force]
ncly auth logout <service> [--dry-run] [--force]
ncly auth status [--json]
```

Core declares no service of its own. Each extension declares the services it needs in its manifest, such as `deepgram` for transcript, and any service name works the same way.

- `login` stores the key in the keychain. Replacing a stored key asks for confirmation, or needs `--force` in non-interactive mode. In non-interactive mode without `--stdin`, `login` exits 78 with `TERMINAL_REQUIRED` and the hint ``Run `ncly auth login <service>` in a terminal``. `TERMINAL_REQUIRED` wins over `KEYRING_UNAVAILABLE`, because `login` without a terminal and without `--stdin` never reaches the keychain. An agent relays this hint to the human and never pipes a key itself.
- `logout` removes the key from the keychain. It asks for confirmation, or needs `--force` in non-interactive mode.
- When the keychain is unavailable, `login` and `logout` exit 78 with `KEYRING_UNAVAILABLE`, and the hint names the environment variable to use instead.
- `status` is a report command. It reports whether the keychain answers, then lists every service that an installed extension declares, with its grants. `ok` is `false`, with exit 78, only when the keychain does not answer and a listed service has no key in the environment. That service's entry then adds `code`, set to `KEYRING_UNAVAILABLE`, the code of the failed `auth` check in `ncly doctor`.

For `login` and `logout`, `--dry-run` takes precedence over secret input and confirmation. It describes the intended action without reading stdin for a key, opening a keychain entry, or prompting. Checks that would require a keychain read, including whether login would replace a key, are reported as pending. This simulation does not require a terminal or `--stdin`.

Their dry-run JSON has `dry_run: true`, `plan`, an array with one entry containing the service `key`, `action` (`login` or `logout`), and `effects` (`user_writes: true`, `network: false`, `paid: false`), plus `pending_checks`, whose entries have `id` and `message`. The check `auth.keychain` covers availability, existing keys, and any replacement confirmation. A successful login or logout answers `service` and `action`, never a key.

```json
{"ok":true,"contract_version":1,"keychain":"available","services":[{"service":"deepgram","configured":true,"source":"keychain"}]}
```

`keychain` is `available` or `unavailable`. `source` is `env`, `keychain`, or `none`. To move a key from another secret store, pipe it once: `<command that prints the key> | ncly auth login deepgram --stdin`.

## Examples

```bash
# Check a new machine, then fix what doctor reports
ncly doctor
ncly auth login deepgram

# An agent learns to use Nicely, then discovers one command
ncly skill view nicely
ncly describe 'transcript run youtube' --json

# Find a run whose answer was lost, then inspect it
ncly run list --path 'transcript run youtube' --key "$URL" --json
ncly run view "$RUN_ID" --json
```

Each extension's spec holds the examples of its own commands, such as [transcript's](../../extensions/transcript/spec.md#examples).
