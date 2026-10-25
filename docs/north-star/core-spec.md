# core spec

This file specifies the commands of core. They follow [contract.md](contract.md), as every extension's commands do. The tree names frozen milestones and future items.

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
├── auth grant|revoke               M08
├── run list|view                   M10
└── tap add|list                   COR-622
```

`<domain>` groups extension commands under [Extension identity](contract/extension-identity.md). `skill` is the bundled extension, as [Extensions](contract/extensions.md) defines, and [extensions/skill/spec.md](../../extensions/skill/spec.md) specifies it. External extensions, such as `agent` and `transcript`, list their commands in their own specs.

Core reserves its command and bundled-extension domains from the [domain table](architecture.md#domains-and-commands), plus `help`, `version`, and `config`.
Core reserves `core` and `extensions` because they name [doctor components](#ncly-doctor).
External extensions never use a reserved name. A name outside this version's commands and installed extensions fails with `USAGE_INVALID`.
This includes reserved names whose commands arrive in later items.

## ncly describe

```
ncly describe [command...] [--json]
```

Without a command path, returns concise descriptions of installed commands, except `help`, which `ncly --help` leaves out too. A path comes as one argument or as several words, so `ncly describe 'transcript run youtube'` and `ncly describe transcript run youtube` are the same call. A path such as `transcript run youtube` returns that command's declaration. A path that names a group, such as `transcript`, returns the concise entries of the group and of the commands under it. The JSON envelope contains `commands`, an array, in the order of their paths, compared word by word. Without `--json`, each entry prints on one line: its path, then its summary. Text is localized, while machine identifiers follow Compatibility. An unknown or empty path answers `NOT_FOUND`, whose `next` is `["ncly","describe"]`. An extension that cannot run, as [Discovery](contract/discovery.md) defines, has no entry, and a lookup at or beneath its extension root answers `EXTENSION_INVALID` with exit 78.

Each entry has `path`, such as `transcript run youtube`, `summary`, and `source`. `source` is `core`, `bundled`, or `extension`, and a reader treats an unknown `source` as not core. An entry gives no release that added its command, because discovery lists only what the installed version supports, as [Compatibility](contract/compatibility.md) says, and an extension's commands follow the extension's version, not the release of `ncly`. It gives no path to an executable either: [ncly doctor](#ncly-doctor) reports where each extension comes from and which one hides another. An external extension's entries come from its manifest. The extension root entry adds `skill`, its skill folder's absolute path, or `null` when absent or bundled.
[Extension identity](contract/extension-identity.md) defines shared groups and extension roots. The `skill` extension reads it, as [D054](decisions/D054-extension-skills-through-describe.md) decides. A detailed entry adds these keys:

| Key | Type | Meaning |
|---|---|---|
| `args` | array | Positional arguments in order, each with `name`, `summary`, `type`, `required`, `variadic`, and `values`. `type` is `string`, `boolean`, `integer`, or `number`. `values` lists accepted typed values, or is empty. `values_from`, present only when the declaration has one, holds `command`, `key`, and `field`, as [Completion values](contract/completion-values.md) defines: it only suggests values, while `values` constrains them. Only the last argument may be variadic |
| `flags` | array | The declared flags, plus `--dry-run` and `--run-id` when `modes` sets `dry_run` or `recorded`. Each has `name`, `shorthand`, `value`, `type`, `required`, `variadic`, `values`, `default`, and `summary`. Names omit `--`. Empty `shorthand` means none, empty `value` means a switch. `default` is the typed default or `null` when none applies. A flag that takes a value may add `values_from`, as an argument does. Variadic means several values |
| `flag_groups` | array | Each group has `flags`, an array of long flag names, and integer `min` and `max`, the permitted number supplied. An exactly-one group has `min: 1` and `max: 1` |
| `result` | object | `kind`: `object`, `report`, `batch`, or `text`; `keys`: top-level successful data keys; `schema`: JSON Schema for normal JSON answers, including item fields, errors, types, nullability, and closed enums. `dry_run_schema` is present when supported. A text result ignores `--json`, has empty `keys`, and omits schemas. A declaration gives the schema of its data keys, and `describe` returns the schema of the whole answer: it adds `ok`, `contract_version`, `errors`, `warnings`, and, for a batch, `results`, which keeps its declared constraints and whose items add `ok` and `errors`. A success must hold the keys that the declaration requires and no error, and so must each item that succeeded, and a successful batch holds no failed item. A failure and each failed item hold at least one error, except a report's failure, which may list its findings instead, as [Answer check](contract/answer-check.md) asks. A dry-run answer describes the work instead of doing it, as [Dry-run answer](contract/dry-run-answer.md) shows, so `dry_run_schema` adds no `results`, even for a batch. A literal value of the schema, such as a `const`, stays as declared. A declaration without a schema gets this envelope alone, with empty `keys`, or `results` alone for a batch |
| `effects` | object | Conservative possible effects: `user_writes`, `network`, and `paid`. A declaration that omits its effects reports each one `true`, because a missing declaration permits no assumption. Dry run describes the selected invocation's effects |
| `requires` | object | `bins` and `keys`, possible program and service prerequisites. Each entry of `bins` is an object with the program's `name` and the keys of its entry in the `programs` of the [manifest](contract/manifest.md), such as `version` and `conditional`. Each entry of `keys` is an object with the service's `name`, from the `requires.keys` of the manifest, and, from M08, `optional`, from the service's entry in the `keys` of the manifest, or `false` without one. Preparation selects the ones needed by this invocation, so a skipped summary requires no harness |
| `modes` | object | The booleans `dry_run`, `recorded`, and `resume` |
| `config` | array | Each config key that the command reads, with `key`, `value`, and `source`: first `lang`, which core resolves for the text of every command, then each key that the declaration names, in its order. A key that covers a table, such as `transcript.youtube`, gives one entry per value beneath it, in the order of their keys, because a table merges key by key. An empty table with no value beneath it gives one entry whose value is `{}`, from the file that set it. A key that no source sets gives `null` from `default`, because the command applies its own default. `source` is the flag, such as `--lang`, the environment variable, such as `NCLY_LANG` or `LANG`, the path of the config file that set the value, or `default`. A secret value never appears: values come only from the config files and the variables of the language, never from a key's variable or the keychain |
| `examples` | array | Full command lines |

`requires` takes its name and its keys from the manifest, so a static manifest describes an extension's commands the same way. Discovery works with default configuration when a config file is invalid, adds `CONFIG_DEFAULTS_USED` to every answer of `ncly describe` but a usage error, which comes before describe reads the configuration, and does not claim to have resolved settings that it could not read. The warning names the file and, when the parser knows it, the line, and its hint asks to fix that file.

Schemas use JSON Schema draft 2020-12, with all references resolved within the returned descriptor and no remote schema fetch. `keys` comes from that result definition, excluding `ok`, `contract_version`, `errors`, and `warnings`; a successful batch includes `results`. Flag types use the same scalar types as arguments. Schemas allow unknown optional fields as Compatibility requires. Declarations cover the current commands only; core needs no schema generator or custom constraint language for future domains. Validation uses required flags and groups from this declaration. Examples do not replace these constraints.

## ncly run

```
ncly run list [--path <command-path>] [--key <item-key>] [--since <time>] [--limit <count>] [--json]
ncly run view <run-id> [--json]
```

- `list` returns `runs`, sorted by `started_at` descending, then `run_id` ascending, with at most `--limit` entries, 20 by default. The count must be a positive integer. Filters combine with AND before limiting: `--path` exactly matches the command path, `--key` matches any item key, and `--since` includes runs whose `started_at` is at or after that RFC 3339 time. An invalid time is `USAGE_INVALID`
- Each list entry has `run_id`, `path`, `status`, `started_at`, `updated_at`, `active`, `item_count`, and `keys`. `keys` previews at most the first three item keys in input order; filters still inspect every key. The top-level `more` boolean is true when additional matching records exceed the limit. Increase the limit or narrow the filters, without a database or index service
- `list` reads each `runs/<run-id>.json` of the state folder whose name holds a valid run ID, and skips any other file, such as the temporary file of a write. A missing `runs` folder holds no run, and one that cannot be read answers `RUNTIME`. A record that cannot be read or parsed, that holds another `format_version`, or that breaks [Record format](contract/record-format.md), such as with a `started_at` that is no RFC 3339 time, never counts as absent, and neither does a run whose lock cannot be read, because its `active` would be a guess: the top-level `unreadable` lists its run ID, in name order, whatever the filters and the limit, and `ncly run view <run-id>` names its problem
- `view` returns `run`, the record of [Record format](contract/record-format.md) plus `active`, including step outcomes and artifact references. A successful inspection exits 0 even when the saved run failed. Its top-level `ok` describes the inspection, and `run.status` describes the saved operation
- `view` reads the record without taking the run lock, so it never waits for a run that executes, and reports `active` by trying the lock without waiting and without creating its file. `run` holds every key that the record file holds, including a key that this release does not know, then `active`
- A run ID outside the rule of [Run IDs](contract/run-ids.md) uses `USAGE_INVALID`. An unknown run uses `NOT_FOUND`, whose hint and `next` are `ncly run list`. An unreadable or unsupported record, such as one that breaks [Record format](contract/record-format.md), uses `RESUME_UNSAFE`, and so does a run whose lock cannot be read, because its `active` would be a guess; inspection starts no work
- `active` is true while a process holds the run lock. Inspecting a busy run does not permit a retry loop

Explicit `run resume` remains in [COR-621](../backlog/COR-621-resume-sdk.md#later). A recorded domain command can return its verified completed result under the same run ID, as [Run IDs](contract/run-ids.md) defines, without a separate core resume dispatcher.

Without `--json`, `list` prints a header row in bold, `RUN`, `STATUS`, `STARTED`, `COMMAND`, and `ITEMS`, then one row per run, with the date of `started_at`, and its previewed keys on the lines under it, indented, followed by `and <count> more` when the run has more items. The status of a row takes its color: `completed` green, `running` cyan, `failed` red, and `unknown` and `interrupted` yellow. The previewed keys and `and <count> more` are bright black. A green `●` starts the row of an active run, and a legend under the list explains it when the list shows one. The text of the legend is bright black. When `more` is true, a last line in bright black says to raise `--limit` or narrow the filters. Without a matching run, `list` prints one line that says so instead of the header. When `unreadable` holds a run ID, a line after the list names each one, and says that `ncly run view` gives its problem.

`view` prints the run ID and the command path, then the run's facts as labeled lines: its status and whether a process holds it, as `yes` or `no`, the times of `started_at` and `updated_at`, its attempts, and the version of `ncly` that wrote it. The run ID is bold, and the status is green when `completed`. Bold uppercase headings, as in the help, then group its inputs, its resolved values, the versions of its tools, and its items. Each input or resolved value is a labeled line, and a nested value takes the lines under its label, indented. The label of each fact, input, and resolved value is bright black. Each item shows its key, then `→` and its result folder once the domain selected one, then one line per step, with its status, id, and effect, and its times and duration once it started, followed by its artifacts with their sizes, and by its latest error: its code and message, then `hint:` and the hint. The status of a step is green when `completed`, red when `failed`, and bright black when `pending`. The `→`, the effect of each step, and the sizes of the artifacts are bright black. The latest error takes the styles of the error block. A step that started and has not finished shows its start alone. An array value takes one line per element under its label. Only `format_version`, the `evidence` of the steps, and the hashes of the artifacts stay in the JSON alone.

Any other text of `list` and `view` stays plain, such as a step that is `unknown` and each tool with its version. In `view`, a run status other than `completed` stays plain too, and so does the line of `unreadable` in `list`. Styles change no text, as [Output](contract/output.md) defines.

A caller that passed `--run-id` recovers a run whose stdout was lost by that ID, as [Run IDs](contract/run-ids.md) defines: `ncly run view <run-id>`, or the same call again. Without the ID, filter by the original path and an item key, such as `ncly run list --path 'transcript run youtube' --key <url> --json`. Add `--since` when the retained invocation start time is known; a fresh session may omit it. Check `more`, then confirm candidates with `run view`, including their recorded inputs and complete item list. Report ambiguous matches instead of choosing the newest one. A call that failed during preparation before its first record write has no record. Absence from one limited list proves neither absence of a run nor absence of billing.

## ncly completion

```
ncly completion zsh|bash|fish
```

Prints the completion script on stdout. The Homebrew formula installs these scripts, and each Linux archive ships them in `completions/`. The AUR package will install them too, once it ships. Users run this command only for a manual setup.

Completion suggests auth services, extension slugs for grant and revoke, and doctor components.
[Completion values](contract/completion-values.md) defines suggestions from extension list commands.

## ncly doctor

```
ncly doctor [component] [--live] [--timeout <duration>] [--json]
```

| Argument or flag | Effect |
|---|---|
| `component` | `core`, `auth`, `extensions`, or an installed extension's slug. Without it, doctor checks all of them. |
| `--live` | Adds network checks against free endpoints only. It never calls an endpoint that bills. Doctor passes it to each status command that declares it. |
| `--timeout` | Bounds the checks, including local version probes, status commands, and free network checks, 30 seconds by default. |

- The default checks stay on the machine: binaries and their versions, keys present, the keychain answering, and the config files parsing.
- Invalid or unreadable config produces the failed `core.config` finding with `CONFIG_INVALID`, even for a selected component. Doctor continues independent checks and omits checks whose required configuration could not be resolved. It never reports those omitted checks as passing
- Each unknown key, as [Configuration](contract/configuration.md) defines it, is a `warn` entry of `core.config` with `CONFIG_UNKNOWN_KEY`, whose hint names the key and the file that holds it.
- The `auth` component checks the keychain, the grants, and each key that an installed extension declares, as [Required keys](#required-keys) defines.
- The `extensions` component starts nothing. Each discovered executable has an `extensions.<slug>` entry with `domain`, `source`, and `path`, under [Extension identity](contract/extension-identity.md). The selected extension passes, a hidden one passes with `hidden_by`, and an incompatible manifest fails with `EXTENSION_INVALID`. [Discovery](contract/discovery.md) owns selection and conflicts
- Each installed extension has a component named after its slug. It holds [required-program checks](#required-programs) and its [status command's checks](contract/status-command.md), such as [skill status](../../extensions/skill/spec.md#ncly-skill-status)
- Doctor never installs anything, in any mode, as [D011](decisions/D011-local-checks.md) decides. When a tool is missing or too old and its install command for this system is known, as [Required programs](#required-programs) defines, the hint of its failed check holds that command, such as `brew install ffmpeg` or `sudo pacman -S ffmpeg`, and its `next` holds the same command as words. The human runs it.
- When `--timeout` expires, doctor kills the version probe or the status command that runs, with its descendants, at once, without the 15 seconds of cleanup, because neither writes. That check fails with `CHECK_TIMEOUT`, and so does each check that doctor had not started.
- Doctor is a report command. It exits 0 when no check fails and 78 when at least one check needs a human. Doctor's own failures use `errors`. An unknown component gets `NOT_FOUND`, exit 2. An invalid extension manifest gets `EXTENSION_INVALID`, exit 78, for its slug. Other failures get `RUNTIME`, exit 1

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
| `next` | Optional. The command that the hint names, as words, as [Output](contract/output.md) defines |
| `domain` | Only in an entry of the `extensions` component. The extension slug, under [Extension identity](contract/extension-identity.md) |
| `source` | Only in an entry of the `extensions` component. `bundled` or `extension`, as `source` in [ncly describe](#ncly-describe) |
| `path` | Only in an entry of the `extensions` component. The absolute path of the extension's executable, as discovery found it, or `null` for a bundled extension |
| `hidden_by` | Only in the entry of an extension that another one hides. An object with `source`, which is `core`, `bundled`, or `extension`, and `path`, the absolute path of the executable that wins, or `null` for a core command or a bundled extension |

Without `--json`, doctor prints one line per check, in the order of `checks`, in three aligned columns: its status, `pass`, `warn`, or `fail`, colored by status, its full `id`, and its message. A check with a problem adds a line under it, with its code in the column of the `id`, then `hint:` and the hint in the column of the message. The status is bold, green for `pass`, yellow for `warn`, and red for `fail`, its code takes the color of its status, and `hint:` takes the label style of an error.

### Required programs

Each nonconditional manifest program has a check under the extension's component, with id `<slug>.<name>`, such as `transcript.uv`:

- Doctor looks for an executable file named `name` in each folder of `PATH`, in order, and skips an empty or relative entry, as [Discovery](contract/discovery.md) does. A program that it does not find fails with `PREREQ_MISSING`
- When the program declares `version`, doctor runs it with the arguments of `probe`, with stdin closed, in a new session. It reads the version from stdout, or from stderr when stdout holds none: the first run of two or more numbers joined by dots, such as `6.1.1` in `ffmpeg version 6.1.1-3ubuntu5`. It compares that version with `version` number by number, where a missing number counts as 0, so `3.12` meets `3.12.0`. An older version fails with `PREREQ_MISSING`, and so does a probe that exits with another code than 0 or prints no version, because doctor cannot then show that the program works
- The hint of a failed check holds the install command of the first package manager in this order that `install` names and this machine has in `PATH`: `brew` on macOS, and `pacman`, `apt`, `dnf`, then `brew` on Linux. Its `next` holds the same words. Doctor ignores any other name, so a later release can add one. Without a command for this machine, the hint names the program to install, and the check has no `next`

Doctor never checks a conditional program, whether it is absent, present, or too old. Only the extension knows whether its config selects the program, and a copy that no config selects is no failure, so one rule covers every case. Preparation checks the conditional programs that an invocation selects, as [Operations](contract/operations.md) says, and the extension's status command checks the ones that its config selects, as [Status command](contract/status-command.md) defines.

### Required keys

The `auth` component holds the check of the keychain, then the check of the grants, then one check for each service that [ncly auth status](#ncly-auth) lists, in its order, so doctor and `status` give each service the same verdict:

- `auth.keychain` passes when the keychain answers. When it does not, the check is a `warn` with `KEYRING_ENV_ONLY`, whose hint says to unlock the keychain or to export the variable of each key, and which has no `next`. An unavailable keychain alone never fails the report, as [D010](decisions/D010-keychain-or-environment.md) decides
- `auth.grants` passes when every grant of [Grants](contract/grants.md) is in use. A grant is unused when its extension no longer declares the service or no longer exists at its source. Each unused grant warns with `GRANT_UNUSED`, whose hint and `next` are `ncly auth revoke <service> <slug>`. An invalid manifest counts as in use because doctor cannot read its declarations. An unreadable grants file fails with `CONFIG_INVALID`. Service checks then judge their keys alone
- `auth.<service>`, such as `auth.deepgram`, passes when the key exists and every extension that needs it holds a grant, as `status` defines. Its message names the source, never the value. A missing key gives one entry with the service's `code`, `hint`, and `next` from `status`. With a key, each required missing grant gives `KEY_NOT_GRANTED`, in slug order. Its hint and `next` are `ncly auth grant <service> <slug>`. A service that no extension needs passes, and its message says when the key is missing
- When no installed extension declares a service, the component holds `auth.keychain` and `auth.grants` alone

Doctor asks the keychain as `status` does, and runs its `auth` component last, after every other check, so its questions to the keychain never cost a check that already finished. This order changes when doctor runs the checks, not where the report lists them. Each keychain call gets the smaller of its limit in [Keys](contract/keys.md) and the time left of `--timeout`. When the keychain's limit ends first, the keychain counts as unavailable, and doctor asks it nothing more. When `--timeout` ends first, the check that waits on the keychain fails with `CHECK_TIMEOUT`, and so does each check that doctor had not started, as for a version probe. Doctor stops waiting at once, because a keychain call writes nothing. In interactive mode, an unlock prompt therefore gets what remains of the 30 seconds of the default `--timeout`, not the 60 seconds of the keychain's limit.

Under `--live`, the keychain read that starting a status command needs, as [Grants](contract/grants.md) defines, takes the same bound before doctor starts the command. When the keychain's limit ends first, the keychain counts as unavailable for the rest of doctor, and the status command receives `KEYRING_UNAVAILABLE` for that key in `NCLY_MISSING_KEYS`. When `--timeout` ends first, the `<slug>.status` check fails with `CHECK_TIMEOUT`, and doctor never starts the command.

## ncly auth

```
ncly auth login <service> [--stdin] [--dry-run] [--force]
ncly auth logout <service> [--dry-run] [--force]
ncly auth grant <service> <slug> [--dry-run]
ncly auth revoke <service> <slug> [--dry-run] [--force]
ncly auth status [--json]
```

Core declares no service of its own. Each extension declares the services that its commands need in the `requires.keys` of its manifest, such as `deepgram` for transcript, and any service name works the same way, as [Keys](contract/keys.md) defines. A name that breaks that rule, including `keychain`, exits 2 with `USAGE_INVALID` before any other step. [D057](decisions/D057-auth-edge-cases.md) records how their edge cases compare with `gh`, Docker, and clig.dev.

- `login` stores the key in the keychain. It asks the keychain before it reads a key or asks for a confirmation, so a human never types a key that cannot be stored. Replacing a stored key asks for confirmation, or needs `--force` in non-interactive mode, where the hint of `CONFIRMATION_REQUIRED` adds `--force` to the call and has no `next`, because the key must come on stdin again.
- If two `ncly auth login` calls for the same service run at once, each may read the old entry before either writes. The last write that takes effect wins, even if that call did not use `--force` or confirm a replacement. `ncly` takes no per-service lock for login.
- Without `<service>`, in interactive mode, the form for missing values of [Modes](contract/modes.md) asks for it first, in a text field that suggests the declared services, as completion does, and `login` or `logout` then goes on as below. The `Next time:` line that follows holds the service, never the key.
- In interactive mode without `--stdin`, `login` asks for the key in the masked field that [Modes](contract/modes.md) describes, whose title names the service. The key never shows, and the field leaves the terminal as it found it. With `--stdin`, it reads stdin to its end, and a terminal on stdin exits 2 with `USAGE_INVALID`, because the key would show as the human types it. Either way, `login` removes the whitespace around the key, including the final line ending, and an empty key stores nothing and exits 2 with `USAGE_INVALID`.
- In non-interactive mode without `--stdin`, `login` exits 78 with `TERMINAL_REQUIRED`, the hint ``Run `ncly auth login <service>` in a terminal``, and the `next` `ncly auth login <service>`, which only a human runs. `TERMINAL_REQUIRED` wins over `KEYRING_UNAVAILABLE`, because `login` without a terminal and without `--stdin` never reaches the keychain. An agent relays this hint and `next` to the human and never pipes a key itself.
- `logout` removes the key from the keychain. It asks for confirmation, or needs `--force` in non-interactive mode, where `CONFIRMATION_REQUIRED` has the hint and `next` `ncly auth logout <service> --force`. A service without a stored key exits 2 with `NOT_FOUND` before any confirmation, with the `next` `ncly auth status`. When the service's variable holds a key, the hint names that variable, because `logout` never changes the environment.
- When the keychain is unavailable before they write, `login` and `logout` exit 78 with `KEYRING_UNAVAILABLE`. The hint of `login` names the environment variable to use instead, and the hint of `logout` asks the human to unlock the keychain. A write that starts, then fails or gets no answer within its limit, exits 1 with `RUNTIME` and the hint and `next` that [Keys](contract/keys.md) defines, because the key may change after `ncly` answers.
- A confirmation is the `[y/N]` prompt that [Modes](contract/modes.md) describes, and its words come from the catalog. Only a yes confirms. Any other answer, an empty one, or the end of input changes nothing and exits 2 with `CONFIRMATION_REQUIRED`, as [Modes](contract/modes.md) defines. Neither the masked field nor a confirmation fills a missing value, so `ncly` prints no `Next time:` line after them. Ctrl-C at either one exits 130 with `INTERRUPTED`, changes nothing, and leaves the terminal as it found it.
- `grant` lets the installed extension of `<slug>` receive the key of `<service>`, under [Grants](contract/grants.md). It reads no key and never asks the keychain, so a human may grant before `login`. An unknown slug or undeclared service gets `NOT_FOUND`, exit 2, with `next` set to `ncly auth status`. An invalid manifest gets `EXTENSION_INVALID`, exit 78. An unreadable grants file gets `CONFIG_INVALID`, exit 78. An existing grant at this source answers without a prompt or change. These checks precede the terminal requirement in both modes
- Any other grant needs interactive mode. Without it, `grant` gets `TERMINAL_REQUIRED`, exit 78, with a hint to run `ncly auth grant <service> <slug>` in a terminal. The confirmation names the service, slug, and source path, plus the previous path when replacing a grant. A bundled extension has no path. Only yes grants. Any other answer changes nothing and gets `CONFIRMATION_REQUIRED`, exit 2, with a hint to repeat the command. `grant` accepts no `--force` because nothing replaces that answer
- `revoke` removes the grant for a service and slug, regardless of its source or whether that extension remains installed. A missing grant gets `NOT_FOUND`, exit 2, before confirmation, with `next` set to `ncly auth status`. An unreadable grants file gets `CONFIG_INVALID`, exit 78. Otherwise, revoke shows the grant's path and asks for confirmation, or requires `--force` in non-interactive mode. `CONFIRMATION_REQUIRED` names `ncly auth revoke <service> <slug> --force` in its hint and `next`
- Once the human confirms, or `revoke` has `--force`, the command takes `grants.lock` and waits at most 250 milliseconds while another process holds it, as [Record format](contract/record-format.md) waits for a run's lock. It reads the grants file again under the lock, and applies its change to what the file holds then. A lock that another process still holds exits 75 with `TEMPORARY`, because nothing changed, so the same call is safe to repeat. A file that core can no longer read exits 78 with `CONFIG_INVALID`. A grant at this source that another process added meanwhile answers `unchanged`, a grant at another source gives way to the one that the human confirmed, and a grant that another process removed meanwhile exits 2 with `NOT_FOUND`.
- A successful `grant` or `revoke` answers `service`, `domain`, `path`, and `action`, which is `grant`, `revoke`, or `unchanged` for a grant that already existed. `path` is the source that the grant holds: the path of the extension for `grant`, the path of the removed grant for `revoke`, and `null` for a bundled extension.

`status` is a report command. It asks the keychain once whether it answers, then lists, once each and in name order, the services that the commands of installed extensions declare. It reads each selected extension's manifest under [Discovery](contract/discovery.md) and skips invalid manifests. It reads the keychain entry of a listed service only when the service's variable is empty, and lists no other service, as [D057](decisions/D057-auth-edge-cases.md) decides. A key stored for a service that no installed extension declares stays in the keychain, and `status` lists it once an extension declares that service.

Each entry of `services` has `service`, `configured`, and `source`, which is `env`, `keychain`, or `none`, in the order of [Keys](contract/keys.md), the variable first and then the keychain. An entry adds `grants`, one object per selected extension that declares the service, in slug order.
Each object holds `domain`, the extension slug, and the booleans `granted` and `optional`.
`granted` means consent at that source, and `optional` comes from the manifest. An extension needs the key unless it marks it optional. A key is never printed. `ok` is `false`, with exit 78, when an extension needs the key of a listed service and the key or that extension's grant is missing, and that service's entry then adds the keys of a problem, as [Output](contract/output.md) defines:

- `AUTH_MISSING` when the keychain answers without the key, with the hint `ncly auth login <service>` and the same `next`
- `KEYRING_UNAVAILABLE` when the keychain does not answer and the variable is empty, with a hint that names the variable to export and no `next`, because the hint leaves a value to fill in
- `KEY_NOT_GRANTED` when a required grant is missing for an existing key. Its hint and `next` are `ncly auth grant <service> <slug>` for the first ungranted slug in name order. A missing key reports its own code alone because the key comes before its grants

A service that no extension needs, because each one that declares it marks its key optional, never fails `status`: when its key is missing, or the keychain does not answer, its entry keeps `configured` false and `source` `none`, without the keys of a problem.

An unavailable keychain alone leaves `ok` true, with `keychain` set to `unavailable`, because every listed key may come from the environment, as on a machine reached over SSH. A missing key fails `status` as it fails the `auth` component of [ncly doctor](#ncly-doctor), so both report the same verdict for each service.

Without `--json`, `status` prints `keychain` and its state.
When a manifest declares a service, an empty line precedes service rows.
Each row aligns the bold service, source, and comma-separated grants.
A grant shows `<slug>: granted` or `<slug>: not granted`, plus `, optional` for an optional key. A service's problem follows its line, in the column of its source: `fail[<code>]` and the message, then `hint:` and the hint, with the label styles of an error.

For `login` and `logout`, `--dry-run` takes precedence over secret input and confirmation. It describes the intended action without reading stdin for a key, opening a keychain entry, or prompting. Checks that would require a keychain read, including whether login would replace a key, are reported as pending. This simulation does not require a terminal or `--stdin`.

For `grant` and `revoke`, `--dry-run` works the same way: it needs no terminal, prompts for nothing, and changes no grant. It makes the checks that come before the confirmation, as execution does, so it fails where execution would fail before asking. `pending_checks` then holds `auth.confirmation`, the answer that only a human gives, unless `revoke` has `--force`. When `grant` would answer `unchanged`, `plan` and `pending_checks` are empty.

Their dry-run JSON has `dry_run: true`, `plan`, an array with one entry, or none when nothing would change, containing the service `key`, `action` (`login`, `logout`, `grant`, or `revoke`), `domain` for `grant` and `revoke`, and `effects` (`user_writes: true`, `network: false`, `paid: false`), plus `pending_checks`, whose entries have `id` and `message`. The check `auth.keychain` covers availability, existing keys, and any replacement confirmation. A successful login or logout answers `service` and `action`, never a key.

```json
{"ok":false,"contract_version":1,"keychain":"available","services":[{"service":"deepgram","configured":true,"source":"keychain","grants":[{"domain":"transcript","granted":true,"optional":false}]},{"service":"openai","configured":false,"source":"none","code":"AUTH_MISSING","message":"No OpenAI key found.","hint":"ncly auth login openai","next":["ncly","auth","login","openai"],"grants":[{"domain":"example","granted":false,"optional":false}]}]}
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
