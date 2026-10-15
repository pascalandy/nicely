# core spec

This file specifies the commands of core. They follow [contract.md](contract.md), as every extension's commands do. The command tree names the milestone that brings each command.

## Command tree

```
ncly
├── completion zsh|bash|fish        M00
├── <domain> [arguments]            M01
├── describe [command...]           M03
├── skill list|view                 M04
├── skill status                    M05
├── doctor [component]              M05
├── auth login|logout|status        M06
├── run list|view                   M10
├── run resume <run-id>             M15
└── tap add|list|sync|remove        M12
```

`<domain>` is an extension. `skill` is the bundled extension, as [Extensions](contract.md#extensions) defines, and [extensions/skill/spec.md](../../extensions/skill/spec.md) specifies it. External extensions, such as `agent` and `transcript`, list their commands in their own specs.

Core reserves the names of its own commands and of the bundled extensions, the rows marked core or bundled in the [domain table](architecture.md#domains-and-commands), plus `help`, `version`, and `config`, and `core` and `extensions`, which name components of [ncly doctor](#ncly-doctor) beside each installed extension's domain. An external extension never runs under a reserved name. A name that is neither a command of this version nor an installed extension fails with `USAGE_INVALID`, including a reserved name whose command arrives in a later milestone.

## ncly describe

```
ncly describe [command...] [--json]
```

Without a command path, returns concise descriptions of installed commands, except `help`, which `ncly --help` leaves out too. A path comes as one argument or as several words, so `ncly describe 'transcript run youtube'` and `ncly describe transcript run youtube` are the same call. A path such as `transcript run youtube` returns that command's declaration. A path that names a group, such as `transcript`, returns the concise entries of the group and of the commands under it. The JSON envelope contains `commands`, an array, in the order of their paths, compared word by word. Without `--json`, each entry prints on one line: its path, then its summary. Text is localized, while machine identifiers follow Compatibility. An unknown or empty path answers `NOT_FOUND`, whose `next` is `["ncly","describe"]`. An extension that cannot run, as [Discovery](contract.md#discovery) defines, has no entry, and a path inside its domain answers `EXTENSION_INVALID` with exit 78, as `ncly <domain>` does.

Each entry has `path`, such as `transcript run youtube`, `summary`, and `source`. `source` is `core`, `bundled`, or `extension`, and a reader treats an unknown `source` as not core. An entry gives no release that added its command, because discovery lists only what the installed version supports, as [Compatibility](contract.md#compatibility) says, and an extension's commands follow the extension's version, not the release of `ncly`. It gives no path to an executable either: [ncly doctor](#ncly-doctor) reports where each extension comes from and which one hides another. An external extension's entries come from its manifest. The entry whose path is an extension's domain adds `skill`: the absolute path of the folder that holds that extension's skill, or `null` when it ships none or is bundled. The `skill` extension reads it, as [D054](decisions/D054-extension-skills-through-describe.md) decides. A detailed entry adds these keys:

| Key | Type | Meaning |
|---|---|---|
| `args` | array | Positional arguments in order, each with `name`, `summary`, `type`, `required`, `variadic`, and `values`. `type` is `string`, `boolean`, `integer`, or `number`. `values` lists accepted typed values, or is empty. `values_from`, present only when the declaration has one, holds `command`, `key`, and `field`, as [Completion values](contract.md#completion-values) defines: it only suggests values, while `values` constrains them. Only the last argument may be variadic |
| `flags` | array | The declared flags, plus `--dry-run` and `--run-id` when `modes` sets `dry_run` or `recorded`. Each has `name`, `shorthand`, `value`, `type`, `required`, `variadic`, `values`, `default`, and `summary`. Names omit `--`. Empty `shorthand` means none, empty `value` means a switch. `default` is the typed default or `null` when none applies. A flag that takes a value may add `values_from`, as an argument does. Variadic means several values |
| `flag_groups` | array | Each group has `flags`, an array of long flag names, and integer `min` and `max`, the permitted number supplied. Zoom's `latest` and `path` have `min: 1` and `max: 1` |
| `result` | object | `kind`: `object`, `report`, `batch`, or `text`; `keys`: top-level successful data keys; `schema`: JSON Schema for normal JSON answers, including item fields, errors, types, nullability, and closed enums. `dry_run_schema` is present when supported. A text result ignores `--json`, has empty `keys`, and omits schemas |
| `effects` | object | Conservative possible effects: `user_writes`, `network`, and `paid`. Dry run describes the selected invocation's effects |
| `requires` | object | `bins` and `keys`, possible program and service prerequisites. Each entry of `bins` is an object with the program's `name` and the keys of its entry in the `programs` of the [manifest](contract.md#manifest), such as `version` and `conditional`. Each entry of `keys` is an object with the service's `name`, from the `requires.keys` of the manifest. Preparation selects the ones needed by this invocation, so a skipped summary requires no harness |
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

A caller that passed `--run-id` recovers a run whose stdout was lost by that ID, as [Run IDs](contract.md#run-ids) defines: `ncly run view <run-id>`, or the same call again. Without the ID, filter by the original path and an item key, such as `ncly run list --path 'transcript run youtube' --key <url> --json`. Add `--since` when the retained invocation start time is known; a fresh session may omit it. Check `more`, then confirm candidates with `run view`, including their recorded inputs and complete item list. Report ambiguous matches instead of choosing the newest one. A call that failed during preparation before its first record write has no record. Absence from one limited list proves neither absence of a run nor absence of billing.

## ncly completion

```
ncly completion zsh|bash|fish
```

Prints the completion script on stdout. The Homebrew formula installs these scripts, and each Linux archive ships them in `completions/`. The AUR package will install them too, once it ships. Users run this command only for a manual setup.

Completion also suggests the values that core knows, such as auth services and doctor components, and the values that only an extension knows, such as skill names, profiles, and transcript prompts, from the list command that the declaration names, as [Completion values](contract.md#completion-values) defines.

## ncly doctor

```
ncly doctor [component] [--live] [--timeout <duration>] [--json]
```

| Argument or flag | Effect |
|---|---|
| `component` | `core`, `auth`, `extensions`, or the domain of an installed extension, such as `skill` or `transcript`. Without it, doctor checks all of them. |
| `--live` | Adds network checks against free endpoints only, such as listing Deepgram projects. It never calls an endpoint that bills. Doctor passes it to each status command that declares it. |
| `--timeout` | Bounds the checks, including local version probes, status commands, and free network checks, 30 seconds by default. |

- The default checks stay on the machine: binaries and their versions, keys present, the keychain answering, and the config files parsing.
- Invalid or unreadable config produces the failed `core.config` finding with `CONFIG_INVALID`, even for a selected component. Doctor continues independent checks and omits checks whose required configuration could not be resolved. It never reports those omitted checks as passing
- Each unknown key, as [Configuration](contract.md#configuration) defines it, is a `warn` entry of `core.config` with `CONFIG_UNKNOWN_KEY`, whose hint names the key and the file that holds it.
- The `auth` component checks the keychain and each key that an installed extension declares, as [Required keys](#required-keys) defines.
- The `extensions` component starts no extension. It holds one `extensions.<domain>` entry for each extension that core finds, with its `domain`, `source`, and `path`: a `pass` for the extension that runs, a `pass` with `hidden_by` for each extension that a core command, a bundled extension, or an earlier extension hides, and a `fail` with `EXTENSION_INVALID` for an incompatible manifest. It also lists every grant that no extension declares anymore.
- Each installed extension is a component named after its domain. It holds the checks of the programs that its manifest requires, as [Required programs](#required-programs) defines, and the checks of its status command, as [Status command](contract.md#status-command) defines, such as those of [ncly skill status](../../extensions/skill/spec.md#ncly-skill-status).
- Doctor never installs anything, in any mode, as [D011](decisions/D011-local-checks.md) decides. When a tool is missing or too old and its install command for this system is known, as [Required programs](#required-programs) defines, the hint of its failed check holds that command, such as `brew install ffmpeg` or `sudo pacman -S ffmpeg`, and its `next` holds the same command as words. The human runs it.
- When `--timeout` expires, doctor kills the version probe or the status command that runs, with its descendants, at once, without the 15 seconds of cleanup, because neither writes. That check fails with `CHECK_TIMEOUT`, and so does each check that doctor had not started.
- Doctor is a report command. It exits 0 when no check fails and 78 when at least one check needs a human. When doctor itself fails, it answers with `errors`, such as `NOT_FOUND` with exit 2 for an unknown component, `EXTENSION_INVALID` with exit 78 for the domain of an extension whose manifest is invalid, or `RUNTIME` with exit 1.

```json
{"ok":false,"contract_version":1,"checks":[{"id":"transcript.uv","component":"transcript","status":"pass","message":"uv found"},{"id":"auth.deepgram","component":"auth","status":"fail","code":"AUTH_MISSING","message":"No Deepgram key found.","hint":"ncly auth login deepgram","next":["ncly","auth","login","deepgram"]}]}
```

Each check is an object with these keys:

| Key | Meaning |
|---|---|
| `id` | What the check examines, as `<component>.<name>`, such as `transcript.uv`. A check that finds several problems of one kind lists one entry per problem under the same `id`, each with its own hint, and one `pass` entry when it finds none |
| `component` | The component of the check |
| `status` | `pass`, `warn`, or `fail`. A warning never changes the exit code |
| `code` | Absent from a `pass`. A failed check's code maps to exit 78, and a warning's code is a warning code |
| `message` | Human text in the active language |
| `hint` | Absent from a `pass`. The command or the action that fixes the problem |
| `next` | Optional. The command that the hint names, as words, as [Output](contract.md#output) defines |
| `domain` | Only in an entry of the `extensions` component. The domain of the extension that the entry describes |
| `source` | Only in an entry of the `extensions` component. `bundled` or `extension`, as `source` in [ncly describe](#ncly-describe) |
| `path` | Only in an entry of the `extensions` component. The absolute path of the extension's executable, as discovery found it, or `null` for a bundled extension |
| `hidden_by` | Only in the entry of an extension that another one hides. An object with `source`, which is `core`, `bundled`, or `extension`, and `path`, the absolute path of the executable that wins, or `null` for a core command or a bundled extension |

### Required programs

For each program in the `programs` of an extension's manifest that is not `conditional`, as [Manifest](contract.md#manifest) declares them, the extension's component holds a check whose id is `<domain>.<name>`, such as `transcript.uv`:

- Doctor looks for an executable file named `name` in each folder of `PATH`, in order, and skips an empty or relative entry, as [Discovery](contract.md#discovery) does. A program that it does not find fails with `PREREQ_MISSING`
- When the program declares `version`, doctor runs it with the arguments of `probe`, with stdin closed, in its own process group. It reads the version from stdout, or from stderr when stdout holds none: the first run of two or more numbers joined by dots, such as `6.1.1` in `ffmpeg version 6.1.1-3ubuntu5`. It compares that version with `version` number by number, where a missing number counts as 0, so `3.12` meets `3.12.0`. An older version fails with `PREREQ_MISSING`, and so does a probe that exits with another code than 0 or prints no version, because doctor cannot then show that the program works
- The hint of a failed check holds the install command of the first package manager in this order that `install` names and this machine has in `PATH`: `brew` on macOS, and `pacman`, `apt`, `dnf`, then `brew` on Linux. Its `next` holds the same words. Doctor ignores any other name, so a later release can add one. Without a command for this machine, the hint names the program to install, and the check has no `next`

Doctor never checks a conditional program, whether it is absent, present, or too old. Only the extension knows whether its config selects the program, and a copy that no config selects is no failure, so one rule covers every case. Preparation checks the conditional programs that an invocation selects, as [Operations](contract.md#operations) says, and the extension's status command checks the ones that its config selects, as [Status command](contract.md#status-command) defines.

### Required keys

The `auth` component holds the check of the keychain, then one check for each service that [ncly auth status](#ncly-auth) lists, in its order, so doctor and `status` give each service the same verdict:

- `auth.keychain` passes when the keychain answers. When it does not, the check is a `warn` with `KEYRING_ENV_ONLY`, whose hint says to unlock the keychain or to export the variable of each key, and which has no `next`. An unavailable keychain alone never fails the report, as [D010](decisions/D010-keychain-or-environment.md) decides
- `auth.<service>`, such as `auth.deepgram`, passes when the key comes from the environment or the keychain, and its message names the source, never the value. Otherwise it fails with the `code`, `hint`, and `next` of the service's entry in `status`, such as `AUTH_MISSING` with the hint and `next` `ncly auth login <service>`
- When no installed extension declares a service, the component holds `auth.keychain` alone

Doctor asks the keychain as `status` does, and runs its `auth` component last, after every other check, so its questions to the keychain never cost a check that already finished. This order changes when doctor runs the checks, not where the report lists them. Each keychain call gets the smaller of its limit in [Keys](contract.md#keys) and the time left of `--timeout`. When the keychain's limit ends first, the keychain counts as unavailable, and doctor asks it nothing more. When `--timeout` ends first, the check that waits on the keychain fails with `CHECK_TIMEOUT`, and so does each check that doctor had not started, as for a version probe. Doctor stops waiting at once, because a keychain call writes nothing. In interactive mode, an unlock prompt therefore gets what remains of the 30 seconds of the default `--timeout`, not the 60 seconds of the keychain's limit.

## ncly auth

```
ncly auth login <service> [--stdin] [--dry-run] [--force]
ncly auth logout <service> [--dry-run] [--force]
ncly auth status [--json]
```

Core declares no service of its own. Each extension declares the services that its commands need in the `requires.keys` of its manifest, such as `deepgram` for transcript, and any service name works the same way, as [Keys](contract.md#keys) defines. A name that breaks that rule, including `keychain`, exits 2 with `USAGE_INVALID` before any other step. [D057](decisions/D057-auth-edge-cases.md) records how their edge cases compare with `gh`, Docker, and clig.dev.

- `login` stores the key in the keychain. It asks the keychain before it reads a key or asks for a confirmation, so a human never types a key that cannot be stored. Replacing a stored key asks for confirmation, or needs `--force` in non-interactive mode.
- In interactive mode without `--stdin`, `login` asks for the key in a plain prompt, until the masked field of [M07](../milestones/M07-forms.md) replaces it. It turns echo off before it shows the key prompt, and restores the terminal's state when the prompt ends, so the key never shows. With `--stdin`, it reads stdin to its end, and a terminal on stdin exits 2 with `USAGE_INVALID`, because the key would show as the human types it. Either way, `login` removes the whitespace around the key, including the final line ending, and an empty key stores nothing and exits 2 with `USAGE_INVALID`.
- In non-interactive mode without `--stdin`, `login` exits 78 with `TERMINAL_REQUIRED` and the hint ``Run `ncly auth login <service>` in a terminal``. `TERMINAL_REQUIRED` wins over `KEYRING_UNAVAILABLE`, because `login` without a terminal and without `--stdin` never reaches the keychain. An agent relays this hint to the human and never pipes a key itself.
- `logout` removes the key from the keychain. It asks for confirmation, or needs `--force` in non-interactive mode. A service without a stored key exits 2 with `NOT_FOUND` before any confirmation, with the `next` `ncly auth status`. When the service's variable holds a key, the hint names that variable, because `logout` never changes the environment.
- When the keychain is unavailable before they write, `login` and `logout` exit 78 with `KEYRING_UNAVAILABLE`. The hint of `login` names the environment variable to use instead, and the hint of `logout` asks the human to unlock the keychain. A write that starts, then fails or gets no answer within its limit, exits 1 with `RUNTIME` and the hint and `next` that [Keys](contract.md#keys) defines, because the key may change after `ncly` answers.
- Until the forms of M07, a confirmation is a plain `y/N` prompt whose words come from the catalog. Only a yes confirms. Any other answer, an empty one, or the end of input changes nothing and exits 2 with `CONFIRMATION_REQUIRED`, as [Modes](contract.md#modes) defines. A plain prompt is no form, so `ncly` prints no `Next time:` line after it. Ctrl-C at a prompt exits 130 with `INTERRUPTED` and leaves the terminal's echo as it found it.

`status` is a report command. It asks the keychain once whether it answers, then lists, once each and in name order, the services that the commands of installed extensions declare. It reads the manifest of the extension that runs for each domain, as [Discovery](contract.md#discovery) finds it, and skips an invalid one. It reads the keychain entry of a listed service only when the service's variable is empty, and lists no other service, as [D057](decisions/D057-auth-edge-cases.md) decides. A key stored for a service that no installed extension declares stays in the keychain, and `status` lists it once an extension declares that service.

Each entry of `services` has `service`, `configured`, and `source`, which is `env`, `keychain`, or `none`, in the order of [Keys](contract.md#keys), the variable first and then the keychain. From [M08](../milestones/M08-grants.md), an entry adds `grants`. A key is never printed. `ok` is `false`, with exit 78, when a listed service has no key, and that service's entry then adds the keys of a problem, as [Output](contract.md#output) defines:

- `AUTH_MISSING` when the keychain answers without the key, with the hint `ncly auth login <service>` and the same `next`
- `KEYRING_UNAVAILABLE` when the keychain does not answer and the variable is empty, with a hint that names the variable to export and no `next`, because the hint leaves a value to fill in

An unavailable keychain alone leaves `ok` true, with `keychain` set to `unavailable`, because every listed key may come from the environment, as on a machine reached over SSH. A missing key fails `status` as it fails the `auth` component of [ncly doctor](#ncly-doctor), so both report the same verdict for each service.

For `login` and `logout`, `--dry-run` takes precedence over secret input and confirmation. It describes the intended action without reading stdin for a key, opening a keychain entry, or prompting. Checks that would require a keychain read, including whether login would replace a key, are reported as pending. This simulation does not require a terminal or `--stdin`.

Their dry-run JSON has `dry_run: true`, `plan`, an array with one entry containing the service `key`, `action` (`login` or `logout`), and `effects` (`user_writes: true`, `network: false`, `paid: false`), plus `pending_checks`, whose entries have `id` and `message`. The check `auth.keychain` covers availability, existing keys, and any replacement confirmation. A successful login or logout answers `service` and `action`, never a key.

```json
{"ok":false,"contract_version":1,"keychain":"available","services":[{"service":"deepgram","configured":true,"source":"keychain"},{"service":"openai","configured":false,"source":"none","code":"AUTH_MISSING","message":"No OpenAI key found.","hint":"ncly auth login openai","next":["ncly","auth","login","openai"]}]}
```

`keychain` is `available` or `unavailable`. To move a key from another secret store, pipe it once: `<command that prints the key> | ncly auth login deepgram --stdin`.

## Examples

```bash
# Check a new machine, then fix what doctor reports
ncly doctor
ncly auth login deepgram

# An agent learns to use Nicely, then discovers one command
ncly skill view nicely
ncly describe 'transcript run youtube' --json

# Name the run before it starts, so a lost answer costs nothing
ncly transcript run youtube --url "$URL" --run-id "$RUN_ID" --json
ncly run view "$RUN_ID" --json

# Without a run ID, find a run whose answer was lost, then inspect it
ncly run list --path 'transcript run youtube' --key "$URL" --json
```

Each extension's spec holds the examples of its own commands, such as [transcript's](../../extensions/transcript/spec.md#examples).
